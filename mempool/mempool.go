/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package mempool

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"math/big"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-actor/router"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/account/key"
	"github.com/aergoio/aergo/v2/chain"
	cfg "github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/contract/enterprise"
	"github.com/aergoio/aergo/v2/contract/name"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/fee"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/protobuf/proto"
)

const (
	initial = iota
	loading = iota
	running = iota
)

var (
	evictPeriod      = time.Hour * types.DefaultEvictPeriod
	evictInterval    = evictPeriod >> 5
	evictWorkTimeout = time.Millisecond << 2
	metricInterval   = time.Second
)

// MemPool is main structure of mempool service
type MemPool struct {
	*component.BaseComponent

	sync.RWMutex
	cfg *cfg.Config

	sdb           *state.ChainStateDB
	bestBlockID   types.BlockID
	bestBlockInfo *types.BlockHeaderInfo
	stateDB       *state.StateDB
	verifier      *actor.PID
	orphan        int
	//cache       map[types.TxID]types.Transaction
	cache             sync.Map
	length            int
	pool              map[types.AccountID]*txList
	dumpPath          string
	status            int32
	coinbasefee       *big.Int
	bestChainIdHash   []byte
	acceptChainIdHash []byte
	isPublic          bool
	whitelist         *whitelistConf
	// followings are for test
	testConfig bool
	deadtx     int

	quit chan bool
	wg   sync.WaitGroup // wait for internal loop
}

// NewMemPoolService create and return new MemPool
func NewMemPoolService(cfg *cfg.Config, cs *chain.ChainService) *MemPool {

	var sdb *state.ChainStateDB
	if cs != nil {
		sdb = cs.SDB()
	} else { // Test
		fee.EnableZeroFee()
	}

	actor := &MemPool{
		cfg: cfg,
		sdb: sdb,
		//cache:    map[types.TxID]types.Transaction{},
		cache:    sync.Map{},
		pool:     map[types.AccountID]*txList{},
		dumpPath: cfg.Mempool.DumpFilePath,
		status:   initial,
		verifier: nil,
		quit:     make(chan bool),
	}
	actor.BaseComponent = component.NewBaseComponent(message.MemPoolSvc, actor, log.NewLogger("mempool"))
	if cfg.Mempool.EnableFadeout == false {
		evictPeriod = 0
	} else if cfg.Mempool.FadeoutPeriod > 0 {
		evictPeriod = time.Duration(cfg.Mempool.FadeoutPeriod) * time.Hour
	}
	return actor
}

// BeforeStart runs mempool servivce
func (mp *MemPool) BeforeStart() {
	if mp.testConfig {
		initStubData()
		mp.bestBlockID = getCurrentBestBlockNoMock()
		mp.bestBlockInfo = getCurrentBestBlockInfoMock()
	}
	//mp.Info("mempool start on: current Block :", mp.curBestBlockNo)
}

func (mp *MemPool) AfterStart() {

	mp.Info().Bool("showmetric", mp.cfg.Mempool.ShowMetrics).
		Bool("fadeout", mp.cfg.Mempool.EnableFadeout).
		Str("evict period", evictPeriod.String()).
		Int("number of verifier", mp.cfg.Mempool.VerifierNumber).
		Msg("mempool init")

	mp.verifier = actor.Spawn(router.NewRoundRobinPool(mp.cfg.Mempool.VerifierNumber).
		WithInstance(NewTxVerifier(mp)))

	rsp, err := mp.RequestToFuture(message.ChainSvc, &message.GetBestBlock{}, time.Second*2).Result()
	if err != nil {
		mp.Error().Err(err).Msg("failed to get best block")
		panic("Mempool AfterStart Failed")
	}
	bestblock := rsp.(message.GetBestBlockRsp).Block
	mp.setStateDB(bestblock) // nolint: errcheck

	mp.wg.Add(1)
	go mp.monitor()
}

// BeforeStop handles clean-up for mempool service
func (mp *MemPool) BeforeStop() {
	if mp.verifier != nil {
		mp.verifier.GracefulStop()
	}
	mp.dumpTxsToFile()
	mp.quit <- true
	mp.wg.Wait()
}

func (mp *MemPool) monitor() {
	defer mp.wg.Done()

	evict := time.NewTicker(evictInterval)
	defer evict.Stop()

	showmetric := time.NewTicker(metricInterval)
	defer showmetric.Stop()

	for {
		select {
		// Log current counts on mempool
		case <-showmetric.C:
			if mp.cfg.Mempool.ShowMetrics {
				l, o := mp.Size()
				mp.Info().Int("len", l).Int("orphan", o).Int("acc", len(mp.pool)).Msg("mempool metrics")
			}
			// Evict old enough transactions
		case <-evict.C:
			if mp.cfg.Mempool.EnableFadeout {
				mp.evictTransactions()
			}

			// Graceful quit
		case <-mp.quit:
			return
		}
	}

}
func (mp *MemPool) evictTransactions() {
	//startTime := time.Now()
	//expireTimer := time.NewTimer(evictWorkTimeout)
	mp.Lock()
	defer mp.Unlock()

	eTime := time.Now().Add(-1 * evictPeriod)
	workTO := time.NewTimer(evictWorkTimeout)
	total := 0
L:
	for acc, list := range mp.pool {
		// break evictLoop not to hold locks long time
		select {
		case <-workTO.C:
			break L
		default:
		}

		if list.GetLastModifiedTime().After(eTime) {
			continue
		}
		txs := list.GetAll()
		total += len(txs)
		orphan := len(txs) - list.Len()

		for _, tx := range txs {
			mp.cache.Delete(types.ToTxID(tx.GetHash()))
			mp.length--
		}

		mp.orphan -= orphan
		delete(mp.pool, acc)
	}
	if total > 0 {
		mp.Info().Int("num", total).Msg("evict transactions")
	}
}

// Size returns current maintaining number of transactions
// and number of orphan transaction
func (mp *MemPool) Size() (int, int) {
	return mp.length, mp.orphan
}

// Receive handles requested messages from other services
func (mp *MemPool) Receive(context actor.Context) {

	switch msg := context.Message().(type) {
	case *message.MemPoolPut:
		mp.verifier.Request(msg.Tx, context.Sender())
	case *message.MemPoolGet:
		txs, err := mp.get(msg.MaxBlockBodySize)
		context.Respond(&message.MemPoolGetRsp{
			Txs: txs,
			Err: err,
		})
	case *message.MemPoolList:
		txs, more := mp.listHash(msg.Limit)
		context.Respond(&message.MemPoolListRsp{
			Hashes:  txs,
			HasMore: more,
		})
	case *message.MemPoolDel:
		errs := mp.removeOnBlockArrival(msg.Block)
		context.Respond(&message.MemPoolDelRsp{
			Err: errs,
		})
	case *message.MemPoolDelTx:
		mp.Info().Str("txhash", enc.ToString(msg.Tx.GetHash())).Msg("remove tx in mempool")
		err := mp.removeTx(msg.Tx)
		context.Respond(&message.MemPoolDelTxRsp{
			Err: err,
		})
	case *message.MemPoolExist:
		tx := mp.exist(msg.Hash)
		context.Respond(&message.MemPoolExistRsp{
			Tx: tx,
		})
	case *message.MemPoolExistEx:
		txsnum, _ := mp.Size()
		var bucketHash []types.TxHash
		bucketHash = msg.Hashes
		mp.Debug().Int("len", len(bucketHash)).Int("cached", txsnum).Msg("mempool existEx")

		txs := mp.existEx(bucketHash)
		context.Respond(&message.MemPoolExistExRsp{Txs: txs})

	case *message.MemPoolSetWhitelist:
		mp.whitelist.SetWhitelist(msg.Accounts)
	case *message.MemPoolEnableWhitelist:
		mp.whitelist.Enable(msg.On)

	case *message.MemPoolTxStat:
		b, err := json.Marshal(mp.getUnconfirmed(nil, true))
		if err != nil {
			mp.Error().Err(err).Msg("failed to marshal mempool transactions stats")
		}
		context.Respond(&message.MemPoolTxStatRsp{Data: b})

	case *message.MemPoolTx:
		b, err := json.Marshal(mp.getUnconfirmed(msg.Accounts, false))
		if err != nil {
			mp.Error().Err(err).Msg("failed to marshal mempool transactions")
		}
		context.Respond(&message.MemPoolTxRsp{Data: b})

	case *actor.Started:
		mp.loadTxs() // FIXME :work-around for actor settled

	default:
		//mp.Debug().Str("type", reflect.TypeOf(msg).String()).Msg("unhandled message")
	}
}

func (mp *MemPool) Statistics() *map[string]interface{} {
	ret := map[string]interface{}{
		"total":  mp.length,
		"orphan": mp.orphan,
		"dead":   mp.deadtx,
		"config": mp.cfg.Mempool,
	}
	if !mp.isPublic {
		ret["whitelist"] = mp.whitelist.GetWhitelist()
		ret["whitelist_on"] = mp.whitelist.GetOn()
	}
	return &ret
}

func (mp *MemPool) get(maxBlockBodySize uint32) ([]types.Transaction, error) {
	start := time.Now()
	mp.RLock()
	defer mp.RUnlock()
	count := 0
	size := 0
	txs := make([]types.Transaction, 0)
Gather:
	for _, list := range mp.pool {
		for _, tx := range list.Get() {
			if size += proto.Size(tx.GetTx()); uint32(size) > maxBlockBodySize {
				break Gather
			}
			txs = append(txs, tx)
			count++
		}
	}
	elapsed := time.Since(start)
	mp.Debug().Str("elapsed", elapsed.String()).Int("len", mp.length).Int("orphan", mp.orphan).Int("count", count).Msg("total tx returned")
	return txs, nil
}

// check existence.
// validate
// add pool if possible, else pendings
func (mp *MemPool) put(tx types.Transaction) error {
	id := types.ToTxID(tx.GetHash())
	acc := tx.GetBody().GetAccount()
	if tx.HasVerifedAccount() {
		acc = tx.GetVerifedAccount()
	}

	if _, ok := mp.cache.Load(id); ok {
		return types.ErrTxAlreadyInMempool
	}
	/*
		err := mp.verifyTx(tx)
		if err != nil {
			return err
		}
	*/
	err := mp.validateTx(tx, acc)
	if err != nil && err != types.ErrTxNonceToohigh {
		return err
	}
	mp.Lock()
	defer mp.Unlock()

	list, err := mp.acquireMemPoolList(acc)
	if err != nil {
		return err
	}
	defer mp.releaseMemPoolList(list)
	diff, err := list.Put(tx)
	if err != nil {
		mp.Error().Err(err).Msg("fail to put at a mempool list")
		return err
	}

	mp.orphan -= diff
	mp.cache.Store(id, tx)
	mp.length++
	mp.Trace().Object("tx", types.LogTx{Tx: tx.GetTx()}).Msg("tx added")

	if !mp.testConfig {
		mp.notifyNewTx(tx)
	}
	return nil
}

func (mp *MemPool) puts(txs ...types.Transaction) []error {
	errs := make([]error, len(txs))
	for i, tx := range txs {
		errs[i] = mp.put(tx)
	}
	return errs
}

func (mp *MemPool) listHash(maxTxSize int) ([]types.TxID, bool) {
	start := time.Now()
	mp.RLock()
	defer mp.RUnlock()
	size := 0
	hasMore := false
	ids := make([]types.TxID, 0, maxTxSize)
Gather:
	for _, list := range mp.pool {
		toGet := list.Len()
		if toGet > (maxTxSize - size) {
			toGet = maxTxSize - size
		}
		for _, tx := range list.Get() {
			if len(ids) >= maxTxSize {
				hasMore = true
				break Gather
			}
			ids = append(ids, types.ToTxID(tx.GetHash()))
		}
	}
	elapsed := time.Since(start)
	mp.Debug().Str("elapsed", elapsed.String()).Int("len", mp.length).Int("orphan", mp.orphan).Int("count", size).Msg("tx hashes returned")
	return ids, hasMore
}

func (mp *MemPool) setStateDB(block *types.Block) (bool, bool) {
	if mp.testConfig {
		return true, false
	}

	newBlockID := types.ToBlockID(block.BlockHash())
	parentBlockID := types.ToBlockID(block.GetHeader().GetPrevBlockHash())
	reorged := true
	forked := false

	if types.HashID(newBlockID).Compare(types.HashID(mp.bestBlockID)) != 0 {
		if types.HashID(parentBlockID).Compare(types.HashID(mp.bestBlockID)) != 0 {
			reorged = false //reorg case
		}
		mp.bestBlockID = newBlockID
		mp.bestBlockInfo = types.NewBlockHeaderInfo(block)
		mp.acceptChainIdHash = common.Hasher(types.MakeChainId(block.GetHeader().GetChainID(), mp.nextBlockVersion()))
		stateRoot := block.GetHeader().GetBlocksRootHash()
		if mp.stateDB == nil {
			mp.stateDB = mp.sdb.OpenNewStateDB(stateRoot)
			cid := types.NewChainID()
			if err := cid.Read(block.GetHeader().GetChainID()); err != nil {
				mp.Error().Err(err).Msg("failed to read chain ID")
			} else {
				mp.isPublic = cid.PublicNet
				if !mp.isPublic {
					conf, err := enterprise.GetConf(mp.stateDB, enterprise.AccountWhite)
					if err != nil {
						mp.Warn().Err(err).Msg("failed to init whitelist")
					}
					mp.whitelist = newWhitelistConf(mp, conf.GetValues(), conf.GetOn())
				}
			}
			mp.Debug().Str("Hash", newBlockID.String()).
				Str("StateRoot", types.ToHashID(stateRoot).String()).
				Str("chainidhash", enc.ToString(mp.bestChainIdHash)).
				Str("next chainidhash", enc.ToString(mp.acceptChainIdHash)).
				Msg("new StateDB opened")
		} else if !bytes.Equal(mp.stateDB.GetRoot(), stateRoot) {
			if err := mp.stateDB.SetRoot(stateRoot); err != nil {
				mp.Error().Err(err).Msg("failed to set root of StateDB")
			}
		}

		givenId := common.Hasher(block.GetHeader().GetChainID())
		if !bytes.Equal(mp.bestChainIdHash, givenId) {
			mp.bestChainIdHash = givenId
			forked = true
		}
	}
	return reorged, forked
}

func (mp *MemPool) resetAll() {
	mp.orphan = 0
	mp.length = 0
	mp.pool = map[types.AccountID]*txList{}
	mp.cache = sync.Map{}
}

// input tx based ? or pool based?
// concurrency consideration,
func (mp *MemPool) removeOnBlockArrival(block *types.Block) error {
	var ag [2]time.Duration
	start := time.Now()
	mp.Lock()
	defer mp.Unlock()

	check := 0
	dirty := map[types.AccountID]bool{}
	reorg, fork := mp.setStateDB(block)
	if fork {
		mp.Debug().Msg("reset mempool on fork")
		mp.resetAll()
		return nil
	}

	// non-reorg case only look through account related to given block
	if reorg == false {
		for _, tx := range block.GetBody().GetTxs() {
			account := tx.GetBody().GetAccount()
			recipient := tx.GetBody().GetRecipient()
			if tx.HasNameAccount() {
				account = mp.getOwner(account) // it's for the case that tx sender is named smart contract
			}
			if tx.HasNameRecipient() {
				recipient = mp.getAddress(recipient)
			}
			dirty[types.ToAccountID(account)] = true
			dirty[types.ToAccountID(recipient)] = true
		}
	}

	ag[0] = time.Since(start)
	start = time.Now()
	for acc, list := range mp.pool {
		if !reorg && dirty[acc] == false {
			continue
		}
		ns, err := mp.getAccountState(list.GetAccount())
		if err != nil {
			mp.Error().Err(err).Msg("getting Account status failed during removal")
			// TODO : ????
			continue
		}
		diff, delTxs := list.FilterByState(ns)
		mp.orphan -= diff
		for _, tx := range delTxs {
			mp.cache.Delete(types.ToTxID(tx.GetHash()))
			mp.length--
		}
		if len(delTxs) > 0 {
			mp.Trace().Array("txs", types.LogTrsactions{TXs: delTxs, Limit: 5}).Msg("transactions were filtered by state")
		}
		mp.releaseMemPoolList(list)
		check++
	}

	ag[1] = time.Since(start)
	mp.Debug().Int("given", len(block.GetBody().GetTxs())).
		Int("check", check).
		Str("elapse1", ag[0].String()).
		Str("elapse2", ag[1].String()).
		Msg("delete txs on block")
	return nil
}

// signiture verification
func (mp *MemPool) verifyTx(tx types.Transaction) error {
	err := tx.Validate(mp.acceptChainIdHash, mp.isPublic)
	if err != nil {
		return err
	}
	if !tx.GetTx().NeedNameVerify() {
		err = key.VerifyTx(tx.GetTx())
		if err != nil {
			return err
		}
	} else {
		mp.RLock()
		account := mp.getAddress(tx.GetBody().GetAccount())
		mp.RUnlock()
		err = key.VerifyTxWithAddress(tx.GetTx(), account)
		if err != nil {
			return err
		}
		if !tx.SetVerifedAccount(account) {
			mp.Warn().Str("account", string(account)).Msg("could not set verified account")
		}
	}
	return nil
}

func (mp *MemPool) getAddress(account []byte) []byte {
	return mp.getNameDest(account, false)
}

func (mp *MemPool) getOwner(account []byte) []byte {
	return mp.getNameDest(account, true)
}

func (mp *MemPool) getNameDest(account []byte, owner bool) []byte {
	if mp.testConfig {
		return account
	}

	if string(account) == string(types.AergoVault) {
		return account
	}

	nameState, err := mp.getAccountState([]byte(types.AergoName))
	if err != nil {
		mp.Error().Str("for name", string(account)).Msgf("failed to get state %s", types.AergoName)
		return nil
	}
	scs, err := mp.stateDB.OpenContractState(types.ToAccountID([]byte(types.AergoName)), nameState)
	if err != nil {
		mp.Error().Str("for name", string(account)).Msgf("failed to open contract %s", types.AergoName)
		return nil
	}
	if owner {
		return name.GetOwner(scs, account)
	}
	return name.GetAddress(scs, account)
}

func (mp *MemPool) nextBlockVersion() int32 {
	return mp.cfg.Hardfork.Version(mp.bestBlockInfo.No + 1)
}

// check tx sanity
// check if sender has enough balance
// check if recipient is valid name
// check tx account is lower than known value
func (mp *MemPool) validateTx(tx types.Transaction, account types.Address) error {
	if !mp.whitelist.Check(types.EncodeAddress(account)) {
		return types.ErrTxNotAllowedAccount
	}
	ns, err := mp.getAccountState(account)
	if err != nil {
		return err
	}
	err = tx.ValidateWithSenderState(ns, system.GetGasPrice(), mp.nextBlockVersion())
	if err != nil && err != types.ErrTxNonceToohigh {
		return err
	}

	//NOTE: don't overwrite err, if err == ErrTxNonceToohigh
	//because err should be ErrNonceToohigh if following validation has passed
	//this will be refactored soon

	switch tx.GetBody().GetType() {
	case types.TxType_REDEPLOY:
		if chain.IsPublic() {
			return types.ErrTxInvalidType
		}
		if tx.GetBody().GetRecipient() == nil {
			return types.ErrTxInvalidRecipient
		}
		fallthrough
	case types.TxType_NORMAL, types.TxType_TRANSFER, types.TxType_CALL:
		// checking recipient address
		// FIXME make more general code to classify address type; normal(b58 pubkey), special account, name or invalid
		if !types.IsQuirkTx(tx.GetHash()) {
			recipient := tx.GetBody().GetRecipient()
			if tx.GetTx().HasNameRecipient() || types.IsSpecialAccount(recipient) {
				// it will search account directly
			} else {
				if len(recipient) != types.AddressLength {
					return types.ErrTxInvalidRecipient
				}
			}
			recipientAddr := mp.getAddress(recipient)
			if recipientAddr == nil {
				return types.ErrTxInvalidRecipient
			}
		}
	case types.TxType_DEPLOY:
		if tx.GetBody().GetRecipient() != nil {
			return types.ErrTxInvalidRecipient
		}
	case types.TxType_GOVERNANCE:
		aergoState, err := mp.getAccountState(tx.GetBody().GetRecipient())
		if err != nil {
			return err
		}
		aid := types.ToAccountID(tx.GetBody().GetRecipient())
		scs, err := mp.stateDB.OpenContractState(aid, aergoState)
		if err != nil {
			return err
		}
		switch string(tx.GetBody().GetRecipient()) {
		case types.AergoSystem:
			sender, err := mp.stateDB.GetAccountStateV(account)
			if err != nil {
				return err
			}
			nextBlockInfo := types.BlockHeaderInfo{
				No:          mp.bestBlockInfo.No + 1,
				ForkVersion: mp.nextBlockVersion(),
			}
			if _, err := system.ValidateSystemTx(account, tx.GetBody(), sender, scs, &nextBlockInfo); err != nil {
				return err
			}
		case types.AergoName:
			systemcs, err := mp.stateDB.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
			if err != nil {
				return err
			}
			sender, err := mp.stateDB.GetAccountStateV(account)
			if err != nil {
				return err
			}
			if _, err := name.ValidateNameTx(tx.GetBody(), sender, scs, systemcs); err != nil {
				return err
			}
		case types.AergoEnterprise:
			enterprisecs, err := mp.stateDB.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoEnterprise)))
			if err != nil {
				return err
			}
			sender, err := mp.stateDB.GetAccountStateV(account)
			if err != nil {
				return err
			}
			if _, err := enterprise.ValidateEnterpriseTx(tx.GetBody(), sender, enterprisecs, mp.bestBlockInfo.No+1); err != nil {
				return err
			}
		}
	case types.TxType_FEEDELEGATION:
		var recipient []byte

		recipient = tx.GetBody().GetRecipient()
		if recipient == nil {
			return types.ErrTxInvalidRecipient
		}
		if tx.GetTx().HasNameRecipient() {
			recipient = mp.getAddress(recipient)
			if recipient == nil {
				return types.ErrTxInvalidRecipient
			}
		}
		aergoState, err := mp.getAccountState(recipient)
		if err != nil {
			return err
		}
		bal := aergoState.GetBalanceBigInt()
		fee, err := tx.GetMaxFee(bal, system.GetGasPrice(), mp.nextBlockVersion())
		if err != nil {
			return err
		}
		if fee.Cmp(bal) > 0 {
			return types.ErrInsufficientBalance
		}
		txBody := tx.GetBody()
		rsp, err := mp.RequestToFuture(message.ChainSvc,
			&message.CheckFeeDelegation{Payload: txBody.GetPayload(), Contract: recipient,
				Sender: txBody.GetAccount(), TxHash: tx.GetHash(), Amount: txBody.GetAmount()},
			time.Second).Result()
		if err != nil {
			mp.Error().Err(err).Msg("failed to checkFeeDelegation")
			return err
		}
		err = rsp.(message.CheckFeeDelegationRsp).Err
		if err != nil {
			mp.Error().Err(err).Msg("failed to checkFeeDelegation")
			return err
		}
	}
	return err
}

func (mp *MemPool) exist(hash []byte) *types.Tx {
	v := make([]types.TxHash, 1)
	v[0] = hash
	txs := mp.existEx(v)
	return txs[0]
}
func (mp *MemPool) existEx(hashes []types.TxHash) []*types.Tx {

	if len(hashes) > message.MaxReqestHashes {
		mp.Error().Int("size", len(hashes)).
			Msg("request exceeds max hash length")
		return nil
	}

	ret := make([]*types.Tx, len(hashes))
	for i, h := range hashes {
		if v, ok := mp.cache.Load(types.ToTxID(h)); ok {
			ret[i] = v.(types.Transaction).GetTx()
		}
	}
	return ret
}

func (mp *MemPool) acquireMemPoolList(acc []byte) (*txList, error) {
	list := mp.getMemPoolList(acc)
	if list != nil {
		return list, nil
	}
	ns, err := mp.getAccountState(acc)
	if err != nil {
		return nil, err
	}
	id := types.ToAccountID(acc)
	mp.pool[id] = newTxList(acc, ns, mp)
	return mp.pool[id], nil
}

func (mp *MemPool) releaseMemPoolList(list *txList) {
	if list.Empty() {
		id := types.ToAccountID(list.account)
		delete(mp.pool, id)
	}
}

func (mp *MemPool) getMemPoolList(acc []byte) *txList {
	id := types.ToAccountID(acc)
	return mp.pool[id]
}

func (mp *MemPool) getAccountState(acc []byte) (*types.State, error) {
	if mp.testConfig {
		aid := types.ToAccountID(acc)
		strAcc := aid.String()
		bal := getBalanceByAccMock(strAcc)
		nonce := getNonceByAccMock(strAcc)
		//mp.Error().Str("acc:", strAcc).Int("nonce", int(nonce)).Msg("")
		return &types.State{Balance: new(big.Int).SetUint64(bal).Bytes(), Nonce: nonce}, nil
	}

	state, err := mp.stateDB.GetAccountState(types.ToAccountID(acc))

	if err != nil {
		mp.Fatal().Err(err).Str("sroot", enc.ToString(mp.stateDB.GetRoot())).Msg("failed to get state")

		//FIXME PANIC?
		//mp.Fatal().Err(err).Msg("failed to get state")
		return nil, err
	}
	/*
		if state.Balance == 0 {
			strAcc := types.EncodeAddress(acc)
			mp.Info().Str("address", strAcc).Msg("w t f")

		}
	*/
	return state, nil
}

func (mp *MemPool) notifyNewTx(tx types.Transaction) {
	mp.RequestTo(message.P2PSvc, &message.NotifyNewTransactions{
		Txs: []*types.Tx{tx.GetTx()},
	})
}

func (mp *MemPool) isRunning() bool {
	if atomic.LoadInt32(&mp.status) != running {
		mp.Info().Msg("skip to dump txs because mempool is not running yet")
		return false
	}
	return true
}

func (mp *MemPool) loadTxs() {
	time.Sleep(time.Second) // FIXME
	if !atomic.CompareAndSwapInt32(&mp.status, initial, loading) {
		return
	}
	defer atomic.StoreInt32(&mp.status, running)
	mp.Debug().Msg("staring to load mempool dump")
	file, err := os.Open(mp.dumpPath)
	if err != nil {
		if !os.IsNotExist(err) {
			mp.Error().Err(err).Msg("Unable to open dump file")
		}
		return
	}

	defer file.Close() // nolint: errcheck

	reader := bufio.NewReader(file)

	var count int

	for {
		buf := types.Tx{}
		byteInt := make([]byte, 4)
		_, err := io.ReadFull(reader, byteInt)
		if err != nil {
			if err != io.EOF {
				mp.Error().Err(err).Msg("err on read file during loading")
			}
			break
		}

		reclen := binary.LittleEndian.Uint32(byteInt)
		buffer := make([]byte, int(reclen))
		_, err = io.ReadFull(reader, buffer)
		if err != nil {
			if err != io.EOF {
				mp.Error().Err(err).Msg("err on read file during loading")
			}
			break
		}

		err = proto.Unmarshal(buffer, &buf)
		if err != nil {
			mp.Error().Err(err).Msg("errr on unmarshalling tx during loading")
			continue
		}
		count++
		mp.put(types.NewTransaction(&buf)) // nolint: errcheck
	}

	mp.Info().Int("try", count).
		Int("drop", count-mp.length-mp.orphan).
		Int("suceed", mp.length).
		Int("orphan", mp.orphan).
		Msg("loading mempool done")
}

func (mp *MemPool) dumpTxsToFile() {
	if !mp.isRunning() {
		return
	}
	mp.Info().Msg("start mempool dump")

	file, err := os.Create(mp.dumpPath)
	if err != nil {
		mp.Error().Err(err).Msg("Unable to create file")
		return
	}
	defer file.Close() // nolint: errcheck

	writer := bufio.NewWriter(file)
	defer writer.Flush() //nolint: errcheck
	mp.Lock()
	defer mp.Unlock()

	var ag [2]time.Duration
	count := 0

Dump:
	for _, list := range mp.pool {
		for _, v := range list.GetAll() {

			var total_data []byte
			start := time.Now()
			data, err := proto.Marshal(v.GetTx())
			if err != nil {
				mp.Error().Err(err).Msg("Marshal failed")
				continue
			}

			byteInt := make([]byte, 4)
			binary.LittleEndian.PutUint32(byteInt, uint32(len(data)))
			total_data = append(total_data, byteInt...)
			total_data = append(total_data, data...)

			ag[0] += time.Since(start)
			start = time.Now()

			length := len(total_data)
			for {
				size, err := writer.Write(total_data)
				if err != nil {
					mp.Error().Err(err).Msg("writing encoded tx fail")
					break Dump
				}
				if length != size {
					total_data = total_data[size:]
					length -= size
				} else {
					break
				}
			}
			count++
			ag[1] += time.Since(start)
		}
	}

	mp.Info().Int("count", count).Str("path", mp.dumpPath).Str("marshal", ag[0].String()).
		Str("write", ag[1].String()).Msg("dump txs")

}

func (mp *MemPool) removeTx(tx *types.Tx) error {
	mp.Lock()
	defer mp.Unlock()

	if mp.exist(tx.GetHash()) == nil {
		mp.Warn().Str("txhash", enc.ToString(tx.GetHash())).Msg("could not find tx to remove")
		return types.ErrTxNotFound
	}
	acc := tx.GetBody().GetAccount()
	list, err := mp.acquireMemPoolList(acc)
	if err != nil {
		return err
	}
	newOrphan, removed := list.RemoveTx(tx)
	if removed == nil {
		mp.Error().Str("txhash", enc.ToString(tx.GetHash())).Msg("already removed tx")
	}
	mp.orphan += newOrphan
	mp.releaseMemPoolList(list)

	mp.cache.Delete(types.ToTxID(tx.GetHash()))
	mp.length--
	mp.Trace().Object("tx", types.LogTx{Tx: tx}).Msg("removed tx")
	return nil
}

type txIdList struct {
	Count int      `json:"count"`
	IDs   []string `json:"id,omitempty"`
}

type unconfirmedTxs struct {
	Address  string     `json:"address"`
	Expire   *time.Time `json:"expire,omitempty"`
	Pooled   txIdList   `json:"pooled"`
	Orphaned txIdList   `json:"orphaned"`
}

func newUnconfirmedTxs(acc []byte, eTime *time.Time, pooled, orphaned int) *unconfirmedTxs {
	return &unconfirmedTxs{
		Address: types.EncodeAddress(types.Address(acc)),
		Expire:  eTime,
		Pooled: txIdList{
			Count: pooled,
		},
		Orphaned: txIdList{
			Count: orphaned,
		},
	}
}

func (u *unconfirmedTxs) setPooled(txs []types.Transaction) {
	u.Pooled.IDs = txs2ids(txs)
}

func (u *unconfirmedTxs) setOrphaned(txs []types.Transaction) {
	u.Orphaned.IDs = txs2ids(txs)
}

func txs2ids(txs []types.Transaction) []string {
	ids := make([]string, len(txs))
	for i, tx := range txs {
		ids[i] = types.ToTxID(tx.GetHash()).String()
	}
	return ids
}

// getUnconfirmed returns the information of the unconfirmed transactions.
func (mp *MemPool) getUnconfirmed(accounts []types.Address, countOnly bool) []*unconfirmedTxs {
	mp.RLock()
	defer mp.RUnlock()

	getTxList := func(acc types.Address) (*txList, *time.Time) {
		eTime := func(tl *txList) *time.Time {
			if evictPeriod == 0 {
				return nil
			}
			t := tl.GetLastModifiedTime().Add(evictPeriod)
			return &t
		}
		if tl, err := mp.acquireMemPoolList([]byte(acc)); err == nil {
			return tl, eTime(tl)
		}
		return nil, nil
	}

	getAccounts := func(accounts []types.Address) []types.Address {
		if len(accounts) > 0 {
			return accounts
		}

		accounts = make([]types.Address, 0)
		for _, a := range mp.pool {
			accounts = append(accounts, a.account)
		}
		return accounts
	}

	accounts = getAccounts(accounts)
	utxs := make([]*unconfirmedTxs, len(accounts))
	for i, addr := range accounts {
		l, eTime := getTxList(addr)

		utxs[i] = newUnconfirmedTxs(addr, eTime, l.ready, len(l.list)-l.ready)
		if countOnly {
			continue
		}
		utxs[i].setPooled(l.pooled())
		utxs[i].setOrphaned(l.orphaned())
	}

	return utxs
}
