/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package mempool

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/account/key"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

const (
	initial = iota
	loading = iota
	running = iota
)

// MemPool is main structure of mempool service
type MemPool struct {
	*component.BaseComponent

	sync.RWMutex
	cfg *cfg.Config

	//curBestBlockNo uint32
	curBestBlockNo types.BlockNo

	orphan   int
	cache    map[types.TxID]*types.Tx
	pool     map[types.AccountID]*TxList
	sdb      *state.ChainStateDB
	dumpPath string
	status   int32
	// misc configs
	testConfig bool
}

// NewMemPoolService create and return new MemPool
func NewMemPoolService(cfg *cfg.Config) *MemPool {
	actor := &MemPool{
		cfg:      cfg,
		cache:    map[types.TxID]*types.Tx{},
		pool:     map[types.AccountID]*TxList{},
		dumpPath: cfg.Mempool.DumpFilePath,
		status:   initial,
		//testConfig:    true, // FIXME test config should be removed
	}
	actor.BaseComponent = component.NewBaseComponent(message.MemPoolSvc, actor, log.NewLogger("mempool"))

	return actor
}

// Start runs mempool servivce
func (mp *MemPool) BeforeStart() {
	if mp.testConfig {
		initStubData()
		mp.curBestBlockNo = getCurrentBestBlockNoMock()
	}
	//else {
	//p.BaseComponent.Start(mp)

	/*result, err := mp.Hub().RequestFuture(message.ChainSvc, &message.GetBestBlockNo{}, time.Second).Result()
	if err != nil {
		mp.Error("get best block failed", err)
	}
	rsp := result.(message.GetBestBlockNoRsp)
	mp.curBestBlockNo = rsp.BlockNo*/
	//}
	//go mp.generateInfiniteTx()
	if mp.cfg.Mempool.ShowMetrics {
		go func() {
			for range time.Tick(1e9) {
				l, o := mp.Size()
				mp.Info().Int("len", l).Int("orphan", o).Int("acc", len(mp.pool)).Msg("mempool metrics")
			}
		}()
	}
	//mp.Info("mempool start on: current Block :", mp.curBestBlockNo)
}
func (mp *MemPool) AfterStart() {}

// Stop handles clean-up for mempool service
func (mp *MemPool) BeforeStop() {
	mp.dumpTxsToFile()
}

func (mp *MemPool) SetStateDb(sdb *state.ChainStateDB) {
	mp.sdb = sdb
}

// Size returns current maintaining number of transactions
// and number of orphan transaction
func (mp *MemPool) Size() (int, int) {
	mp.RLock()
	defer mp.RUnlock()
	return len(mp.cache), mp.orphan
}

// Receive handles requested messages from other services
func (mp *MemPool) Receive(context actor.Context) {

	switch msg := context.Message().(type) {
	case *message.MemPoolPut:
		errs := mp.puts(msg.Txs...)
		context.Respond(&message.MemPoolPutRsp{
			Err: errs,
		})
	case *message.MemPoolGet:
		txs, err := mp.get()
		context.Respond(&message.MemPoolGetRsp{
			Txs: txs,
			Err: err,
		})
	case *message.MemPoolDel:
		errs := mp.removeOnBlockArrival(msg.Block)
		context.Respond(&message.MemPoolDelRsp{
			Err: errs,
		})
	case *message.MemPoolExist:
		tx := mp.exists(msg.Hash)
		context.Respond(&message.MemPoolExistRsp{
			Tx: tx,
		})
	case *actor.Started:
		mp.loadTxs() // FIXME :work-around for actor settled

	default:
		//mp.Debug().Str("type", reflect.TypeOf(msg).String()).Msg("unhandled message")
	}
}

func (mp *MemPool) Statics() *map[string]interface{} {
	return &map[string]interface{}{
		"cache_len": len(mp.cache),
		"orphan":    mp.orphan,
	}
}

func (mp *MemPool) get() ([]*types.Tx, error) {
	mp.RLock()
	defer mp.RUnlock()
	count := 0
	txs := make([]*types.Tx, 0)
	for _, list := range mp.pool {
		for _, v := range list.Get() {
			txs = append(txs, v)
			count++
		}
	}
	mp.Debug().Int("len", len(mp.cache)).Int("orphan", mp.orphan).Int("count", count).Msg("total tx returned")
	return txs, nil
}

// check existence.
// validate
// add pool if possible, else pendings
func (mp *MemPool) put(tx *types.Tx) error {
	id := types.ToTxID(tx.Hash)
	acc := tx.GetBody().GetAccount()

	mp.Lock()
	defer mp.Unlock()
	if _, found := mp.cache[id]; found {
		return message.ErrTxAlreadyInMempool
	}

	err := mp.validate(tx)
	if err != nil {
		return err
	}

	list, err := mp.acquireMemPoolList(acc)
	if err != nil {
		return err
	}
	diff, err := list.Put(tx)
	if err != nil {
		mp.Debug().Err(err).Msg("fail to put at a mempool list")
		return err
	}
	mp.orphan -= diff
	mp.cache[id] = tx
	//mp.Debugf("tx add-ed size(%d, %d)[%s]", len(mp.cache), mp.orphan, tx.GetBody().String())
	if !mp.testConfig {
		mp.notifyNewTx(*tx)
	}
	return nil
}
func (mp *MemPool) puts(txs ...*types.Tx) []error {
	errs := make([]error, len(txs))
	for i, tx := range txs {
		errs[i] = mp.put(tx)
	}
	return errs
}

// input tx based ? or pool based?
// concurrency consideration,
func (mp *MemPool) removeOnBlockArrival(block *types.Block) error {
	mp.Lock()
	defer mp.Unlock()

	//FIXME improvement
	// it is fine because all account's state is cached for now.
	// if eviction on statedb cached is occured, it should be improved
	// to avoid disk access
	//FIXME after block has state root hash, use it
	for _, v := range mp.pool {

		acc := v.GetAccount()
		ns, err := mp.getAccountState(acc, true)
		if err != nil {
			mp.Error().Err(err).Msg("getting Account status failed")
			// TODO : ????
		}
		list, err := mp.acquireMemPoolList(acc)
		if err == nil {
			diff, delTxs := list.SetMinNonce(ns.Nonce + 1)
			mp.orphan -= diff
			if list.Empty() {
				mp.delMemPoolList(acc)
			}
			for _, tx := range delTxs {
				h := types.ToTxID(tx.Hash)
				delete(mp.cache, h) // need lock
			}
		}
	}
	return nil

}

// check tx sanity
// TODO sender's signiture
// check if sender has enough balance
// check tx account is lower than known value
func (mp *MemPool) validate(tx *types.Tx) error {
	account := tx.GetBody().GetAccount()
	if account == nil {
		return message.ErrTxFormatInvalid
	}
	if !bytes.Equal(tx.Hash, tx.CalculateTxHash()) {
		return message.ErrTxHasInvalidHash
	}

	err := key.VerifyTx(tx)
	if err != nil {
		return err
	}
	ns, err := mp.getAccountState(account, false)
	if err != nil {
		return err
	}
	if tx.GetBody().GetAmount() > ns.Balance {
		return message.ErrInsufficientBalance
	}
	if tx.GetBody().GetNonce() <= ns.Nonce {
		return message.ErrTxNonceTooLow
	}
	return nil
}

func (mp *MemPool) exists(hash []byte) *types.Tx {
	mp.RLock()
	defer mp.RUnlock()
	if v, ok := mp.cache[types.ToTxID(hash)]; ok {
		return v
	}
	return nil
}

func (mp *MemPool) acquireMemPoolList(acc []byte) (*TxList, error) {
	list := mp.getMemPoolList(acc)
	if list != nil {
		return list, nil
	}
	var nonce uint64
	ns, err := mp.getAccountState(acc, false)
	if err != nil {
		return nil, err
	}
	nonce = ns.Nonce
	id := types.ToAccountID(acc)
	mp.pool[id] = NewTxList(acc, nonce+1)
	return mp.pool[id], nil
}

func (mp *MemPool) getMemPoolList(acc []byte) *TxList {
	id := types.ToAccountID(acc)
	return mp.pool[id]
}

func (mp *MemPool) delMemPoolList(acc []byte) {
	id := types.ToAccountID(acc)
	delete(mp.pool, id)
}

func (mp *MemPool) getAccountState(acc []byte, refresh bool) (*types.State, error) {
	if mp.testConfig {
		aid := types.ToAccountID(acc)
		strAcc := aid.String()
		bal := getBalanceByAccMock(strAcc)
		nonce := getNonceByAccMock(strAcc)
		mp.Error().Str("acc:", strAcc).Int("nonce", int(nonce)).Msg("")
		return &types.State{Balance: bal, Nonce: nonce}, nil
	}

	state, err := mp.sdb.GetAccountStateClone(types.ToAccountID(acc))

	if err != nil {
		//FIXME PANIC?
		//mp.Fatal().Err(err).Msg("failed to get state")
		return nil, err
	}

	return state, nil
}

func (mp *MemPool) notifyNewTx(tx types.Tx) {
	mp.RequestTo(message.P2PSvc, &message.NotifyNewTransactions{
		Txs: []*types.Tx{&tx},
	})
}

func (mp *MemPool) loadTxs() {
	time.Sleep(time.Second) // FIXME
	if !atomic.CompareAndSwapInt32(&mp.status, initial, loading) {
		return
	}
	defer atomic.StoreInt32(&mp.status, running)
	file, err := os.Open(mp.dumpPath)
	if err != nil {
		if !os.IsNotExist(err) {
			mp.Error().Err(err).Msg("Unable to open dump file")
		}
		return
	}
	reader := csv.NewReader(bufio.NewReader(file))

	var drop, count int
	for {
		rc, err := reader.Read()
		if err != nil {
			break
		}
		count++
		dataBuf, err := base64.StdEncoding.DecodeString(rc[0])
		if err == nil {
			buf := types.Tx{}
			if proto.Unmarshal(dataBuf, &buf) == nil {
				mp.put(&buf) // nolint: errcheck
				continue
			}
		}
		drop++
	}

	mp.Info().Int("len", len(mp.cache)).Int("orphan", mp.orphan).Msg("loading mempool done")
}

func (mp *MemPool) isRunning() bool {
	if atomic.LoadInt32(&mp.status) != running {
		mp.Info().Msg("skip to dump txs because mempool is not running yet")
		return false
	}
	return true
}
func (mp *MemPool) dumpTxsToFile() {

	if !mp.isRunning() {
		return
	}
	if len, _ := mp.Size(); len == 0 {
		os.Remove(mp.dumpPath) // nolint: errcheck
		return
	}

	file, err := os.Create(mp.dumpPath)
	if err != nil {
		mp.Error().Err(err).Msg("Unable to create file")
		return
	}
	defer file.Close() // nolint: errcheck

	writer := csv.NewWriter(bufio.NewWriter(file))
	defer writer.Flush() //nolint: errcheck

	mp.Lock()
	defer mp.Unlock()
	count := 0
	for _, list := range mp.pool {
		for _, v := range list.GetAll() {
			data, err := proto.Marshal(v)
			if err != nil {
				continue
			}

			strData := enc.ToString(data)
			err = writer.Write([]string{strData})
			if err != nil {
				mp.Info().Err(err).Msg("writing encoded tx fail")
				break
			}
			count++
		}
	}
	mp.Info().Int("count", count).Str("path", mp.dumpPath).Msg("dump txs")
}
