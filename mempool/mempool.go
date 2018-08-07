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
	"encoding/hex"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/pkg/log"
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

	orphan     int
	cache      map[types.TransactionKey]*types.Tx
	pool       map[types.AccountKey]*TxList
	stateCache map[types.AccountKey]*types.State

	dumpPath string
	status   int32
	// misc configs
	testConfig bool
}

var _ component.IComponent = (*MemPool)(nil)

// NewMemPoolService create and return new MemPool
func NewMemPoolService(cfg *cfg.Config) *MemPool {
	return &MemPool{
		BaseComponent: component.NewBaseComponent(message.MemPoolSvc, log.NewLogger(log.MemPoolSvc), cfg.EnableDebugMsg),
		cfg:           cfg,
		cache:         map[types.TransactionKey]*types.Tx{},
		pool:          map[types.AccountKey]*TxList{},
		stateCache:    map[types.AccountKey]*types.State{},
		dumpPath:      cfg.Mempool.DumpFilePath,
		status:        initial,
		//testConfig:    true, // FIXME test config should be removed
	}
}

// Start runs mempool servivce
func (mp *MemPool) Start() {
	if mp.testConfig {
		initStubData()
		mp.curBestBlockNo = getCurrentBestBlockNoMock()
	} else {
		mp.BaseComponent.Start(mp)

		/*result, err := mp.Hub().RequestFuture(message.ChainSvc, &message.GetBestBlockNo{}, time.Second).Result()
		if err != nil {
			mp.Error("get best block failed", err)
		}
		rsp := result.(message.GetBestBlockNoRsp)
		mp.curBestBlockNo = rsp.BlockNo*/
	}
	//go mp.generateInfiniteTx()
	if mp.cfg.Mempool.ShowMetrics {
		go func() {
			for range time.Tick(1e9) {
				l, o := mp.Size()
				mp.Infof("mempool metrics len:%d orphan:%d pool(%d)", l, o, len(mp.pool))
			}
		}()
	}
	//mp.Info("mempool start on: current Block :", mp.curBestBlockNo)
}

// Stop handles clean-up for mempool service
func (mp *MemPool) Stop() {
	mp.dumpTxsToFile()
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
	mp.BaseComponent.Receive(context)

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
		errs := mp.removeOnBlockArrival(msg.BlockNo, msg.Txs...)
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
		mp.Debug("unhandled message:", reflect.TypeOf(msg).String())
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
	mp.Debugf("len :%d orphan:%d, total tx returned:%d", len(mp.cache), mp.orphan, count)
	return txs, nil
}

// check existence.
// validate
// add pool if possible, else pendings
func (mp *MemPool) put(tx *types.Tx) error {
	key := types.ToTransactionKey(tx.Hash)
	acc := tx.GetBody().GetAccount()

	mp.Lock()
	defer mp.Unlock()
	if _, found := mp.cache[key]; found {
		return message.ErrTxAlreadyInMempool
	}
	list, err := mp.acquireMemPoolList(acc)
	if err != nil {
		return err
	}
	err = mp.validate(tx)
	if err != nil {
		return err
	}
	diff, err := list.Put(tx)
	if err != nil {
		mp.Debug(err)
		return err
	}
	mp.orphan -= diff
	mp.cache[key] = tx
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
func (mp *MemPool) removeOnBlockArrival(blockNo types.BlockNo, txs ...*types.Tx) error {
	accSet := map[types.AccountKey]bool{}
	mp.Lock()
	defer mp.Unlock()

	// better to have account slice
	for _, v := range txs {
		acc := v.GetBody().GetAccount()
		key := types.ToAccountKey(acc)

		if !accSet[key] {
			ns, err := mp.getAccountState(acc, true)
			if err != nil {
				mp.Error("getting Account status failed:", err)
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
					h := types.ToTransactionKey(tx.Hash)
					delete(mp.cache, h) // need lock
				}
			}
			accSet[key] = true
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

	ns, err := mp.getAccountState(account, false)
	if err != nil {
		return err
	}
	/*
		if tx.GetBody().GetAmount() > ns.Balance {
			return ErrInsufficientBalance
		}
	*/
	if tx.GetBody().GetNonce() <= ns.Nonce {
		return message.ErrTxNonceTooLow
	}
	return nil
}

func (mp *MemPool) exists(hash []byte) *types.Tx {
	mp.RLock()
	defer mp.RUnlock()
	if v, ok := mp.cache[types.ToTransactionKey(hash)]; ok {
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
	key := types.ToAccountKey(acc)
	mp.pool[key] = NewTxList(nonce + 1)
	return mp.pool[key], nil
}

func (mp *MemPool) getMemPoolList(acc []byte) *TxList {
	key := types.ToAccountKey(acc)
	return mp.pool[key]
}

func (mp *MemPool) delMemPoolList(acc []byte) {
	key := types.ToAccountKey(acc)
	delete(mp.pool, key)
}

func (mp *MemPool) setAccountState(acc []byte) (*types.State, error) {
	result, err := mp.Hub().RequestFuture(message.ChainSvc,
		&message.GetState{Account: acc}, time.Second).Result()
	if err != nil {
		return nil, err
	}
	rsp := result.(message.GetStateRsp)
	if rsp.Err != nil {
		return nil, rsp.Err
	}
	mp.stateCache[types.ToAccountKey(acc)] = rsp.State
	return rsp.State, nil
}

func (mp *MemPool) getAccountState(acc []byte, refresh bool) (*types.State, error) {
	if mp.testConfig {
		strAcc := hex.EncodeToString(acc)
		bal := getBalanceByAccMock(strAcc)
		nonce := getNonceByAccMock(strAcc)
		return &types.State{Balance: bal, Nonce: nonce}, nil
	}

	if v, ok := mp.stateCache[types.ToAccountKey(acc)]; ok && !refresh {
		return v, nil
	}
	rsp, err := mp.setAccountState(acc)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (mp *MemPool) notifyNewTx(tx types.Tx) {
	mp.Hub().Request(message.P2PSvc, &message.NotifyNewTransactions{
		Txs: []*types.Tx{&tx},
	}, mp)
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
			mp.Errorf("Unable to open dump file: %v", err)
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

	mp.Infof("loading mempool done: %d txs(%d drops)", len(mp.cache), mp.orphan)
}

func (mp *MemPool) isRunning() bool {
	if atomic.LoadInt32(&mp.status) != running {
		mp.Info("skip to dump txs because mempool is not running yet")
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
		mp.Errorf("Unable to create file: %v", err)
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

			strData := base64.StdEncoding.EncodeToString(data)
			err = writer.Write([]string{strData})
			if err != nil {
				mp.Info(err)
				break
			}
			count++
		}
	}
	mp.Infof("Dump %d txs into %s\n", count, mp.dumpPath)
}
