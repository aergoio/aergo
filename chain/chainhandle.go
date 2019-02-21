/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"math/big"

	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/contract/name"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

var (
	ErrorNoAncestor          = errors.New("not found ancestor")
	ErrBlockOrphan           = errors.New("block is ohphan, so not connected in chain")
	ErrBlockCachedErrLRU     = errors.New("block is in errored blocks cache")
	ErrBlockTooHighSideChain = errors.New("block no is higher than best block, it should have been reorganized")

	errBlockStale     = errors.New("produced block becomes stale")
	errBlockTimestamp = errors.New("invalid timestamp")

	InAddBlock = make(chan struct{}, 1)
)

type ErrReorg struct {
	err error
}

func (ec *ErrReorg) Error() string {
	return fmt.Sprintf("reorg failed. maybe need reconfiguration. error: %s", ec.err.Error())
}

type ErrBlock struct {
	err   error
	block *types.BlockInfo
}

func (ec *ErrBlock) Error() string {
	return fmt.Sprintf("Error: %s. block(%s, %d)", ec.err.Error(), enc.ToString(ec.block.Hash), ec.block.No)
}

type ErrTx struct {
	err error
	tx  *types.Tx
}

func (ec *ErrTx) Error() string {
	return fmt.Sprintf("error executing tx:%s, tx=%s", ec.err.Error(), enc.ToString(ec.tx.GetHash()))
}

func (cs *ChainService) getBestBlockNo() types.BlockNo {
	return cs.cdb.getBestBlockNo()
}

// GetGenesisInfo returns the information on the genesis block.
func (cs *ChainService) GetGenesisInfo() *types.Genesis {
	return cs.cdb.GetGenesisInfo()
}

func (cs *ChainService) GetBestBlock() (*types.Block, error) {
	return cs.cdb.GetBestBlock()
}

func (cs *ChainService) getBlockByNo(blockNo types.BlockNo) (*types.Block, error) {
	return cs.cdb.GetBlockByNo(blockNo)
}

func (cs *ChainService) GetBlock(blockHash []byte) (*types.Block, error) {
	return cs.getBlock(blockHash)
}

func (cs *ChainService) getBlock(blockHash []byte) (*types.Block, error) {
	return cs.cdb.getBlock(blockHash)
}

func (cs *ChainService) GetHashByNo(blockNo types.BlockNo) ([]byte, error) {
	return cs.getHashByNo(blockNo)
}

func (cs *ChainService) getHashByNo(blockNo types.BlockNo) ([]byte, error) {
	return cs.cdb.getHashByNo(blockNo)
}

func (cs *ChainService) getTx(txHash []byte) (*types.Tx, *types.TxIdx, error) {
	tx, txidx, err := cs.cdb.getTx(txHash)
	if err != nil {
		return nil, nil, err
	}
	block, err := cs.cdb.getBlock(txidx.BlockHash)
	blockInMainChain, err := cs.cdb.GetBlockByNo(block.Header.BlockNo)
	if !bytes.Equal(block.BlockHash(), blockInMainChain.BlockHash()) {
		return tx, nil, errors.New("tx is not in the main chain")
	}
	return tx, txidx, err
}

func (cs *ChainService) getReceipt(txHash []byte) (*types.Receipt, error) {
	tx, i, err := cs.cdb.getTx(txHash)
	if err != nil {
		return nil, err
	}

	block, err := cs.cdb.getBlock(i.BlockHash)
	blockInMainChain, err := cs.cdb.GetBlockByNo(block.Header.BlockNo)
	if !bytes.Equal(block.BlockHash(), blockInMainChain.BlockHash()) {
		return nil, errors.New("cannot find a receipt")
	}

	r, err := cs.cdb.getReceipt(block.BlockHash(), block.GetHeader().BlockNo, i.Idx)
	if err != nil {
		return r, err
	}
	r.From = tx.GetBody().GetAccount()
	r.To = tx.GetBody().GetRecipient()
	return r, nil
}

func (cs *ChainService) getEvents(events *[]*types.Event, blkNo types.BlockNo, filter *types.FilterInfo,
	argFilter []types.ArgFilter) {
	blkHash, err := cs.cdb.getHashByNo(blkNo)
	if err != nil {
		return
	}
	receipts, err := cs.cdb.getReceipts(blkHash, blkNo)
	if err != nil {
		return
	}
	if receipts.BloomFilter(filter) == false {
		return
	}
	for idx, r := range receipts.Get() {
		if r.BloomFilter(filter) == false {
			continue
		}
		for _, e := range r.Events {
			if e.Filter(filter, argFilter) {
				e.SetMemoryInfo(r, blkHash, blkNo, int32(idx))
				*events = append(*events, e)
			}
		}
	}
}

func (cs *ChainService) listEvents(filter *types.FilterInfo) ([]*types.Event, error) {
	from := filter.Blockfrom
	to := filter.Blockto

	if to == 0 {
		to = cs.cdb.getBestBlockNo()
	}

	err := filter.ValidateCheck(to)
	if err != nil {
		return nil, err
	}
	argFilter, err := filter.GetExArgFilter()
	if err != nil {
		return nil, err
	}
	events := []*types.Event{}
	if filter.Desc {
		for i := to; i >= from && i != 0; i-- {
			cs.getEvents(&events, types.BlockNo(i), filter, argFilter)
		}
	} else {
		for i := from; i <= to; i++ {
			cs.getEvents(&events, types.BlockNo(i), filter, argFilter)
		}
	}
	return events, nil
}

type chainProcessor struct {
	*ChainService
	block       *types.Block // starting block
	lastBlock   *types.Block
	state       *state.BlockState
	mainChain   *list.List
	isByBP      bool
	isMainChain bool

	add   func(blk *types.Block) error
	apply func(blk *types.Block) error
	run   func() error
}

func newChainProcessor(block *types.Block, state *state.BlockState, cs *ChainService) (*chainProcessor, error) {
	var isMainChain bool
	var err error

	if isMainChain, err = cs.cdb.isMainChain(block); err != nil {
		return nil, err
	}

	cp := &chainProcessor{
		ChainService: cs,
		block:        block,
		state:        state,
		isByBP:       (state != nil),
		isMainChain:  isMainChain,
	}

	if cp.isMainChain {
		cp.apply = cp.execute
	} else {
		cp.apply = cp.addBlock
	}

	if cp.isByBP {
		cp.run = func() error {
			blk := cp.block
			cp.notifyBlockByBP(blk)
			return cp.apply(blk)
		}
	} else {
		cp.run = func() error {
			blk := cp.block

			for blk != nil {
				if err = cp.apply(blk); err != nil {
					return err
				}

				// Remove a block depnding on blk from the orphan cache.
				if blk, err = cp.resolveOrphan(blk); err != nil {
					return err
				}
			}
			return nil
		}
	}

	return cp, nil
}

func (cp *chainProcessor) addBlock(blk *types.Block) error {
	dbTx := cp.cdb.store.NewTx()
	defer dbTx.Discard()

	if err := cp.cdb.addBlock(&dbTx, blk); err != nil {
		return err
	}

	dbTx.Commit()

	if logger.IsDebugEnabled() {
		logger.Debug().Bool("isMainChain", cp.isMainChain).
			Uint64("latest", cp.cdb.getBestBlockNo()).
			Uint64("blockNo", blk.BlockNo()).
			Str("hash", blk.ID()).
			Str("prev_hash", enc.ToString(blk.GetHeader().GetPrevBlockHash())).
			Msg("block added to the block indices")
	}
	cp.lastBlock = blk

	return nil
}

func (cp *chainProcessor) notifyBlockByBP(block *types.Block) {
	if cp.isByBP {
		cp.notifyBlock(block, true)
	}
}

func (cp *chainProcessor) notifyBlockByOther(block *types.Block) {
	if !cp.isByBP {
		cp.notifyBlock(block, false)
	}
}

func (cp *chainProcessor) executeBlock(block *types.Block) error {
	err := cp.ChainService.executeBlock(cp.state, block)
	cp.state = nil
	return err
}

func (cp *chainProcessor) execute(block *types.Block) error {
	if !cp.isMainChain {
		return nil
	}
	logger.Debug().Uint64("no", block.GetHeader().BlockNo).Msg("start to execute")

	var err error

	err = cp.executeBlock(block)
	if err != nil {
		logger.Error().Str("error", err.Error()).Str("hash", block.ID()).
			Msg("failed to execute block")
		return err
	}
	//SyncWithConsensus :ga
	// 	After executing MemPoolDel in the chain service, MemPoolGet must be executed on the consensus.
	// 	To do this, cdb.setLatest() must be executed after MemPoolDel.
	//	In this case, messages of mempool is synchronized in actor message queue.
	var oldLatest types.BlockNo
	if oldLatest, err = cp.connectToChain(block); err != nil {
		return err
	}

	cp.notifyBlockByOther(block)

	blockNo := block.BlockNo()
	if logger.IsDebugEnabled() {
		logger.Debug().
			Uint64("old latest", oldLatest).
			Uint64("new latest", blockNo).
			Str("hash", block.ID()).
			Str("prev_hash", enc.ToString(block.GetHeader().GetPrevBlockHash())).
			Msg("block executed")
	}

	return nil
}

func (cp *chainProcessor) connectToChain(block *types.Block) (types.BlockNo, error) {
	dbTx := cp.cdb.store.NewTx()
	defer dbTx.Discard()

	oldLatest := cp.cdb.connectToChain(&dbTx, block)

	if err := cp.cdb.addTxsOfBlock(&dbTx, block.GetBody().GetTxs(), block.BlockHash()); err != nil {
		return 0, err
	}

	dbTx.Commit()

	return oldLatest, nil
}

func (cp *chainProcessor) reorganize() error {
	// - Reorganize if new bestblock then process Txs
	// - Add block if new bestblock then update context connect next orphan
	if !cp.isMainChain && cp.needReorg(cp.lastBlock) {
		err := cp.reorg(cp.lastBlock)
		if e, ok := err.(consensus.ErrorConsensus); ok {
			logger.Info().Err(e).Msg("reorg stopped by consensus error")
			return nil
		}

		if err != nil {
			logger.Info().Err(err).Msg("reorg stopped by unexpected error")
			return &ErrReorg{err: err}
		}
	}

	return nil
}

func (cs *ChainService) addBlockInternal(newBlock *types.Block, usedBstate *state.BlockState, peerID peer.ID) (err error, cache bool) {
	logger.Debug().Str("hash", newBlock.ID()).Msg("add block")

	if !cs.VerifyTimestamp(newBlock) {
		return &ErrBlock{
			err: errBlockTimestamp,
			block: &types.BlockInfo{
				Hash: newBlock.BlockHash(),
				No:   newBlock.BlockNo(),
			},
		}, false
	}

	var bestBlock *types.Block

	if bestBlock, err = cs.cdb.GetBestBlock(); err != nil {
		return err, false
	}

	// The newly produced block becomes stale because the more block(s) are
	// connected to the blockchain so that the best block is changed. In this
	// case, newBlock is rejected because it is unlikely that newBlock belongs
	// to the main branch. Warning: the condition 'usedBstate != nil' is used
	// to check whether newBlock is produced by the current node itself. Later,
	// more explicit condition may be needed instead of this.
	if usedBstate != nil && newBlock.PrevID() != bestBlock.ID() {
		return &ErrBlock{
			err: errBlockStale,
			block: &types.BlockInfo{
				Hash: newBlock.BlockHash(),
				No:   newBlock.BlockNo(),
			},
		}, false
	}

	if !newBlock.ValidChildOf(bestBlock) {
		return fmt.Errorf("invalid chain id - best: %v, current: %v",
			bestBlock.GetHeader().GetChainID(), newBlock.GetHeader().GetChainID()), false
	}

	if err := cs.VerifySign(newBlock); err != nil {
		return err, true
	}

	// handle orphan
	if cs.isOrphan(newBlock) {
		if usedBstate != nil {
			return fmt.Errorf("block received from BP can not be orphan"), false
		}
		err := cs.handleOrphan(newBlock, bestBlock, peerID)
		if err == nil {
			return ErrBlockOrphan, false
		}

		return err, false
	}

	select {
	case InAddBlock <- struct{}{}:
	}
	defer func() {
		<-InAddBlock
	}()

	cp, err := newChainProcessor(newBlock, usedBstate, cs)
	if err != nil {
		return err, true
	}

	if err := cp.run(); err != nil {
		return err, true
	}

	// TODO: reorganization should be done before chain execution to avoid an
	// unnecessary chain execution & rollback.
	if err := cp.reorganize(); err != nil {
		return err, true
	}

	logger.Info().Uint64("best", cs.cdb.getBestBlockNo()).Msg("Block added successfully")

	return nil, true
}

func (cs *ChainService) addBlock(newBlock *types.Block, usedBstate *state.BlockState, peerID peer.ID) error {
	hashID := types.ToHashID(newBlock.BlockHash())

	if cs.errBlocks.Contains(hashID) {
		return ErrBlockCachedErrLRU
	}

	_, err := cs.getBlock(newBlock.BlockHash())
	if err == nil {
		logger.Debug().Str("hash", newBlock.ID()).Msg("already exist")
		return ErrBlockExist
	}

	var needCache bool
	err, needCache = cs.addBlockInternal(newBlock, usedBstate, peerID)
	if err != nil {
		if needCache {
			evicted := cs.errBlocks.Add(hashID, newBlock)
			logger.Error().Err(err).Bool("evicted", evicted).Uint64("no", newBlock.GetHeader().BlockNo).
				Str("hash", newBlock.ID()).Msg("add errored block to errBlocks lru")
		}
		// err must be returned regardless of the value of needCache.
		return err
	}

	return nil
}

func (cs *ChainService) CountTxsInChain() int {
	var txCount int

	blk, err := cs.GetBestBlock()
	if err != nil {
		return -1
	}

	var no uint64
	for {
		no = blk.GetHeader().GetBlockNo()
		if no == 0 {
			break
		}

		txCount += len(blk.GetBody().GetTxs())

		blk, err = cs.getBlock(blk.GetHeader().GetPrevBlockHash())
		if err != nil {
			txCount = -1
			break
		}
	}

	return txCount
}

type TxExecFn func(bState *state.BlockState, tx types.Transaction) error
type ValidatePostFn func() error
type ValidateSignWaitFn func() error

type blockExecutor struct {
	*state.BlockState
	sdb              *state.ChainStateDB
	execTx           TxExecFn
	txs              []*types.Tx
	validatePost     ValidatePostFn
	coinbaseAcccount []byte
	commitOnly       bool
	validateSignWait ValidateSignWaitFn
}

func newBlockExecutor(cs *ChainService, bState *state.BlockState, block *types.Block) (*blockExecutor, error) {
	var exec TxExecFn
	var validateSignWait ValidateSignWaitFn

	commitOnly := false

	// The DPoS block factory excutes transactions during block generation. In
	// such a case it send block with block state so that bState != nil. On the
	// contrary, the block propagated from the network is not half-executed.
	// Hence we need a new block state and tx executor (execTx).
	if bState == nil {
		if err := cs.validator.ValidateBlock(block); err != nil {
			return nil, err
		}

		bState = state.NewBlockState(cs.sdb.OpenNewStateDB(cs.sdb.GetRoot()))

		exec = NewTxExecutor(block.BlockNo(), block.GetHeader().GetTimestamp(), block.GetHeader().GetPrevBlockHash(), contract.ChainService)

		validateSignWait = func() error {
			return cs.validator.WaitVerifyDone()
		}
	} else {
		logger.Debug().Uint64("block no", block.BlockNo()).Msg("received block from block factory")
		// In this case (bState != nil), the transactions has already been
		// executed by the block factory.
		commitOnly = true
	}

	return &blockExecutor{
		BlockState:       bState,
		sdb:              cs.sdb,
		execTx:           exec,
		txs:              block.GetBody().GetTxs(),
		coinbaseAcccount: block.GetHeader().GetCoinbaseAccount(),
		validatePost: func() error {
			return cs.validator.ValidatePost(bState.GetRoot(), bState.Receipts(), block)
		},
		commitOnly:       commitOnly,
		validateSignWait: validateSignWait,
	}, nil
}

// NewTxExecutor returns a new TxExecFn.
func NewTxExecutor(blockNo types.BlockNo, ts int64, prevBlockHash []byte, preLoadService int) TxExecFn {
	return func(bState *state.BlockState, tx types.Transaction) error {
		if bState == nil {
			logger.Error().Msg("bstate is nil in txexec")
			return ErrGatherChain
		}
		snapshot := bState.Snapshot()

		err := executeTx(bState, tx, blockNo, ts, prevBlockHash, preLoadService)
		if err != nil {
			logger.Error().Err(err).Str("hash", enc.ToString(tx.GetHash())).Msg("tx failed")
			bState.Rollback(snapshot)
			return err
		}
		return nil
	}
}

func (e *blockExecutor) execute() error {
	// Receipt must be committed unconditionally.
	if !e.commitOnly {
		var preLoadTx *types.Tx
		nCand := len(e.txs)
		for i, tx := range e.txs {
			if i != nCand-1 {
				preLoadTx = e.txs[i+1]
				contract.PreLoadRequest(e.BlockState, preLoadTx, contract.ChainService)
			}
			if err := e.execTx(e.BlockState, types.NewTransaction(tx)); err != nil {
				//FIXME maybe system error. restart or panic
				// all txs have executed successfully in BP node
				return err
			}
			contract.SetPreloadTx(preLoadTx, contract.ChainService)
		}

		if e.validateSignWait != nil {
			if err := e.validateSignWait(); err != nil {
				return err
			}
		}

		//TODO check result of verifing txs
		if err := SendRewardCoinbase(e.BlockState, e.coinbaseAcccount); err != nil {
			return err
		}

		if err := contract.SaveRecoveryPoint(e.BlockState); err != nil {
			return err
		}

		if err := e.Update(); err != nil {
			return err
		}
	}

	if err := e.validatePost(); err != nil {
		return err
	}

	// TODO: sync status of bstate and cdb what to do if cdb.commit fails after

	if err := e.commit(); err != nil {
		return err
	}

	logger.Debug().Msg("executed block")
	return nil
}

func (e *blockExecutor) commit() error {
	if err := e.BlockState.Commit(); err != nil {
		return err
	}

	//TODO: after implementing BlockRootHash, remove statedb.lastest
	if err := e.sdb.UpdateRoot(e.BlockState); err != nil {
		return err
	}

	return nil
}

// TODO: Refactoring: batch
func (cs *ChainService) executeBlock(bstate *state.BlockState, block *types.Block) error {
	// Caution: block must belong to the main chain.

	var (
		bestBlock *types.Block
		err       error
	)

	if bestBlock, err = cs.cdb.GetBestBlock(); err != nil {
		return err
	}

	// Check consensus info validity
	if err = cs.IsBlockValid(block, bestBlock); err != nil {
		return err
	}

	ex, err := newBlockExecutor(cs, bstate, block)
	if err != nil {
		return err
	}

	// contract & state DB update is done during execution.
	if err := ex.execute(); err != nil {
		return err
	}

	if len(ex.BlockState.Receipts().Get()) != 0 {
		cs.cdb.writeReceipts(block.BlockHash(), block.BlockNo(), ex.BlockState.Receipts())
	}

	cs.RequestTo(message.MemPoolSvc, &message.MemPoolDel{
		Block: block,
	})

	cs.Update(block)

	return nil
}

func executeTx(bs *state.BlockState, tx types.Transaction, blockNo uint64, ts int64, prevBlockHash []byte, preLoadService int) error {

	txBody := tx.GetBody()

	var account []byte
	if tx.HasVerifedAccount() {
		account = tx.GetVerifedAccount()
		tx.RemoveVerifedAccount()
		resolvedAccount := name.Resolve(bs, txBody.GetAccount())
		if !bytes.Equal(account, resolvedAccount) {
			return types.ErrSignNotMatch
		}
	} else {
		account = name.Resolve(bs, txBody.GetAccount())
	}

	err := tx.Validate()
	if err != nil {
		return err
	}

	sender, err := bs.GetAccountStateV(account)
	if err != nil {
		return err
	}

	err = tx.ValidateWithSenderState(sender.State(), CoinbaseFee())
	if err != nil {
		return err
	}

	recipient := name.Resolve(bs, txBody.Recipient)
	var receiver *state.V
	if len(recipient) > 0 {
		receiver, err = bs.GetAccountStateV(recipient)
	} else {
		receiver, err = bs.CreateAccountStateV(contract.CreateContractID(txBody.Account, txBody.Nonce))
	}
	if err != nil {
		return err
	}

	var txFee *big.Int
	var rv string
	var events []*types.Event
	switch txBody.Type {
	case types.TxType_NORMAL:
		txFee = CoinbaseFee()
		sender.SubBalance(txFee)
		rv, events, err = contract.Execute(bs, tx.GetTx(), blockNo, ts, prevBlockHash, sender, receiver, preLoadService)
	case types.TxType_GOVERNANCE:
		txFee = new(big.Int).SetUint64(0)
		err = executeGovernanceTx(bs, txBody, sender, receiver, blockNo)
		if err != nil {
			logger.Warn().Err(err).Str("txhash", enc.ToString(tx.GetHash())).Msg("governance tx Error")
		}
	}

	if err != nil {
		if !contract.IsRuntimeError(err) {
			return err
		}
		sender.Reset()
		sender.SubBalance(txFee)
		sender.SetNonce(txBody.Nonce)
		sErr := sender.PutState()
		if sErr != nil {
			return sErr
		}
	} else {
		sender.SetNonce(txBody.Nonce)
		err = sender.PutState()
		if err != nil {
			return err
		}
		if sender.AccountID() != receiver.AccountID() {
			err = receiver.PutState()
			if err != nil {
				return err
			}
		}
	}
	bs.BpReward = new(big.Int).Add(new(big.Int).SetBytes(bs.BpReward), txFee).Bytes()

	var receipt *types.Receipt
	if err != nil {
		receipt = types.NewReceipt(receiver.ID(), err.Error(), "")
	} else if receiver.IsNew() && txBody.Recipient == nil {
		receipt = types.NewReceipt(receiver.ID(), "CREATED", rv)
	} else {
		receipt = types.NewReceipt(receiver.ID(), "SUCCESS", rv)
	}
	receipt.FeeUsed = txFee.Bytes()
	receipt.TxHash = tx.GetHash()
	receipt.Events = events

	return bs.AddReceipt(receipt)
}

func SendRewardCoinbase(bState *state.BlockState, coinbaseAccount []byte) error {
	bpReward := new(big.Int).SetBytes(bState.BpReward)
	if bpReward.Cmp(new(big.Int).SetUint64(0)) <= 0 || coinbaseAccount == nil {
		logger.Debug().Str("reward", new(big.Int).SetBytes(bState.BpReward).String()).Msg("coinbase is skipped")
		return nil
	}

	receiverID := types.ToAccountID(coinbaseAccount)
	receiverState, err := bState.GetAccountState(receiverID)
	if err != nil {
		return err
	}

	receiverChange := types.State(*receiverState)
	receiverChange.Balance = new(big.Int).Add(receiverChange.GetBalanceBigInt(), bpReward).Bytes()

	err = bState.PutState(receiverID, &receiverChange)
	if err != nil {
		return err
	}

	logger.Debug().Str("reward", bpReward.String()).
		Str("newbalance", receiverChange.GetBalanceBigInt().String()).Msg("send reward to coinbase account")

	return nil
}

// find an orphan block which is the child of the added block
func (cs *ChainService) resolveOrphan(block *types.Block) (*types.Block, error) {
	hash := block.BlockHash()

	orphanID := types.ToBlockID(hash)
	orphan, exists := cs.op.cache[orphanID]
	if !exists {
		return nil, nil
	}

	orphanBlock := orphan.block

	if (block.GetHeader().GetBlockNo() + 1) != orphanBlock.GetHeader().GetBlockNo() {
		return nil, fmt.Errorf("invalid orphan block no (p=%d, c=%d)", block.GetHeader().GetBlockNo(),
			orphanBlock.GetHeader().GetBlockNo())
	}

	logger.Debug().Str("parentHash=", block.ID()).
		Str("orphanHash=", orphanBlock.ID()).
		Msg("connect orphan")

	cs.op.removeOrphan(orphanID)

	return orphanBlock, nil
}

func (cs *ChainService) isOrphan(block *types.Block) bool {
	prevhash := block.Header.PrevBlockHash
	_, err := cs.getBlock(prevhash)

	return err != nil
}

func (cs *ChainService) handleOrphan(block *types.Block, bestBlock *types.Block, peerID peer.ID) error {
	err := cs.addOrphan(block)
	if err != nil {
		// logging???
		logger.Debug().Str("hash", block.ID()).Msg("add Orphan Block failed")

		return err
	}

	cs.RequestTo(message.SyncerSvc, &message.SyncStart{PeerID: peerID, TargetNo: block.GetHeader().GetBlockNo()})

	return nil
}

func (cs *ChainService) addOrphan(block *types.Block) error {
	return cs.op.addOrphan(block)
}

func (cs *ChainService) findAncestor(Hashes [][]byte) (*types.BlockInfo, error) {
	// 1. check endpoint is on main chain (or, return nil)
	logger.Debug().Int("len", len(Hashes)).Msg("find ancestor")

	var mainhash []byte
	var mainblock *types.Block
	var err error
	// 2. get the highest block of Hashes hash on main chain
	for _, hash := range Hashes {
		// need to be short
		mainblock, err = cs.cdb.getBlock(hash)
		if err != nil {
			continue
		}
		// get main hash with same block height
		mainhash, err = cs.cdb.getHashByNo(
			types.BlockNo(mainblock.GetHeader().GetBlockNo()))
		if err != nil {
			continue
		}

		if bytes.Equal(mainhash, mainblock.BlockHash()) {
			break
		}
		mainblock = nil
	}

	// TODO: handle the case that can't find the hash in main chain
	if mainblock == nil {
		logger.Debug().Msg("Can't search same ancestor")
		return nil, ErrorNoAncestor
	}

	return &types.BlockInfo{Hash: mainblock.BlockHash(), No: mainblock.GetHeader().GetBlockNo()}, nil
}

func (cs *ChainService) setSync(isSync bool) {
	//don't use mempool if sync is in progress
	cs.validator.signVerifier.SetSkipMempool(isSync)
}
