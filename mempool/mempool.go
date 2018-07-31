/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package mempool

/*
TODO
1. nonce map -> per account list
2. account -> type def
3 inter-actor comm
*/
import (
	"bytes"
	"encoding/hex"
	"sync"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
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
	pool       map[types.AccountKey]*MemPoolList
	pending    map[types.AccountKey]*MemPoolList
	stateCache map[types.AccountKey]*types.State

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
		pool:          map[types.AccountKey]*MemPoolList{},
		pending:       map[types.AccountKey]*MemPoolList{},
		stateCache:    map[types.AccountKey]*types.State{},
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
	//mp.SetLevel("debug")
	//go mp.generateInfiniteTx()
	if mp.cfg.Mempool.ShowMetrics {
		go func() {
			for range time.Tick(1e9) {
				l, o := mp.Size()
				mp.Infof("mempool metrics len:%d orphan:%d", l, o)
			}
		}()
	}
	//mp.Info("mempool start on: current Block :", mp.curBestBlockNo)
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
	case *message.MemPoolGenerateSampleTxs:
		context.Respond(mp.generateSampleTxs(msg.MaxCount))
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
	}
}

func (mp *MemPool) get() ([]*types.Tx, error) {
	mp.RLock()
	defer mp.RUnlock()
	count := 0
	txs := make([]*types.Tx, 0)
	for _, list := range mp.pool {
		for _, v := range list.GetAll() {
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
	mp.Lock()
	defer mp.Unlock()

	//mp.Debugf("tx adding into mempool <-[%s]", tx.GetBody().String())
	if _, found := mp.cache[key]; found {
		return message.ErrTxAlreadyInMempool
	}
	err := mp.validate(tx)
	if err != nil {
		return err
	}
	acc := tx.GetBody().GetAccount()
	list, err := mp.acquireMemPoolList(acc, false)
	if err != nil {
		return err
	}
	if err := list.Put(tx); err != nil {
		if err != message.ErrTxNonceToohigh {
			return err
		}
		orphan, err := mp.acquireMemPoolList(acc, true)
		if err != nil {
			return err
		}
		if err = orphan.Put(tx); err != nil {
			return err
		}
		mp.orphan++
	} else {
		err = mp.rearrange(acc)
		if err != nil {
			panic(err)
		}

	}
	mp.cache[key] = tx
	//mp.Debugf("tx add-ed size(%d, %d)[%s]", len(mp.cache), mp.orphan, tx.GetBody().String())

	//TODO go routine for shootint tx top2p component
	// FIXME must solve pingpong problem (in here or in p2p module)
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

func (mp *MemPool) adjust(acc []byte, n uint64, bal uint64) {
	key := types.ToAccountKey(acc)
	adjust := false

	if orphan, ok := mp.pending[key]; ok {
		tmp := orphan.SetMinNonce(n)
		mp.orphan -= tmp
		if orphan.Len() > 0 {
			tmp := orphan.FilterByPrice(bal)
			mp.orphan -= tmp
			adjust = true
		} else {
			delete(mp.pending, key)
		}
	}
	if list, ok := mp.pool[key]; ok {
		list.SetMinNonce(n)
		adjust = true
	}
	if adjust {
		err := mp.rearrange(acc)
		if err == nil && mp.pool[key].Len() == 0 {
			mp.Debugf("pool for %s deleted", acc)
			delete(mp.pool, key)
		}

	}
}

// input tx based ? or pool based?
// concurrency consideration,
func (mp *MemPool) removeOnBlockArrival(blockNo types.BlockNo, txs ...*types.Tx) error {
	accSet := map[types.AccountKey]bool{}
	mp.Lock()
	defer mp.Unlock()

	for _, v := range txs {
		acc := v.GetBody().GetAccount()
		key := types.ToAccountKey(acc)
		h := types.ToTransactionKey(v.Hash)

		if !accSet[key] {
			ns, err := mp.getAccountState(v.GetBody().GetAccount(), true)
			if err != nil {
				mp.Error("getting Account status failed:", err)
				// TODO : ????
			}
			mp.adjust(acc, ns.Nonce, ns.Balance)
			accSet[key] = true
		}
		delete(mp.cache, h) // need lock
	}
	/*
		mp.Debugf(" #######middle#########", blockNo)
		for k, v := range addrBalMap {
			mp.Debugf("k:%s bal:%d, nonce:%d", k, v, addrNonceMap[k])
		}
		mp.Debugf(" #######end -middle#########", blockNo)
	*/
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

//already has lock
func (mp *MemPool) rearrange(acc []byte) error {
	key := types.ToAccountKey(acc)
	orphan := mp.getMemPoolList(acc, true)
	if orphan == nil {
		return nil
	}
	ready, err := mp.acquireMemPoolList(acc, false)
	if err != nil {
		return err
	}
	merged, err := ready.Merge(orphan)
	if err != nil {
		return err
	}
	mp.orphan -= merged
	if orphan.Len() == 0 {
		delete(mp.pending, key)
		mp.Debugf("pending for %s deleted", acc)
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

func (mp *MemPool) acquireMemPoolList(acc []byte, orphan bool) (*MemPoolList, error) {
	list := mp.getMemPoolList(acc, orphan)
	if list != nil {
		return list, nil
	}
	m := mp.pool
	if orphan {
		m = mp.pending
	}

	var nonce uint64
	if !orphan {
		ns, err := mp.getAccountState(acc, false)
		if err != nil {
			return nil, err
		}
		nonce = ns.Nonce
	}

	key := types.ToAccountKey(acc)
	m[key] = NewMemPoolList(nonce+1, !orphan)
	return m[key], nil
}
func (mp *MemPool) getMemPoolList(acc []byte, orphan bool) *MemPoolList {
	key := types.ToAccountKey(acc)
	m := mp.pool
	if orphan {
		m = mp.pending
	}
	return m[key]
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
