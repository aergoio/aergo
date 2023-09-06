/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"bytes"
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/contract"
	"github.com/aergoio/aergo/v2/contract/name"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/protobuf/proto"
)

var (
	ErrorNoAncestor      = errors.New("not found ancestor")
	ErrBlockOrphan       = errors.New("block is ohphan, so not connected in chain")
	ErrBlockCachedErrLRU = errors.New("block is in errored blocks cache")
	ErrStateNoMarker     = errors.New("statedb marker of block is not exists")

	errBlockStale       = errors.New("produced block becomes stale")
	errBlockInvalidFork = errors.New("invalid fork occured")
	errBlockTimestamp   = errors.New("invalid timestamp")

	InAddBlock      = make(chan struct{}, 1)
	SendBlockReward = sendRewardCoinbase
)

type BlockRewardFn = func(*state.BlockState, []byte) error

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

	r, err := cs.cdb.getReceipt(block.BlockHash(), block.GetHeader().BlockNo, i.Idx, cs.cfg.Hardfork)
	if err != nil {
		return r, err
	}
	r.ContractAddress = types.AddressOrigin(r.ContractAddress)
	r.From = tx.GetBody().GetAccount()
	r.To = tx.GetBody().GetRecipient()
	return r, nil
}

func (cs *ChainService) getEvents(events *[]*types.Event, blkNo types.BlockNo, filter *types.FilterInfo,
	argFilter []types.ArgFilter) uint64 {
	blkHash, err := cs.cdb.getHashByNo(blkNo)
	if err != nil {
		return 0
	}
	receipts, err := cs.cdb.getReceipts(blkHash, blkNo, cs.cfg.Hardfork)
	if err != nil {
		return 0
	}
	if receipts.BloomFilter(filter) == false {
		return 0
	}
	var totalSize uint64
	for idx, r := range receipts.Get() {
		if r.BloomFilter(filter) == false {
			continue
		}
		for _, e := range r.Events {
			if e.Filter(filter, argFilter) {
				e.SetMemoryInfo(r, blkHash, blkNo, int32(idx))
				*events = append(*events, e)
				totalSize += uint64(proto.Size(e))
			}
		}
	}
	return totalSize
}

const MaxEventSize = 4 * 1024 * 1024

func (cs *ChainService) listEvents(filter *types.FilterInfo) ([]*types.Event, error) {
	from := filter.Blockfrom
	to := filter.Blockto

	if filter.RecentBlockCnt > 0 {
		to = cs.cdb.getBestBlockNo()
		if to <= uint64(filter.RecentBlockCnt) {
			from = 0
		} else {
			from = to - uint64(filter.RecentBlockCnt)
		}
	} else {
		if to == 0 {
			to = cs.cdb.getBestBlockNo()
		}
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
	var totalSize uint64
	if filter.Desc {
		for i := to; i >= from && i != 0; i-- {
			totalSize += cs.getEvents(&events, types.BlockNo(i), filter, argFilter)
			if totalSize > MaxEventSize {
				return nil, errors.New(fmt.Sprintf("too large size of event (%v)", totalSize))
			}
		}
	} else {
		for i := from; i <= to; i++ {
			totalSize += cs.getEvents(&events, types.BlockNo(i), filter, argFilter)
			if totalSize > MaxEventSize {
				return nil, errors.New(fmt.Sprintf("too large size of event (%v)", totalSize))
			}
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
		isByBP:       state != nil,
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

				// Remove a block depending on blk from the orphan cache.
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

	if err := cp.cdb.addBlock(dbTx, blk); err != nil {
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
		logger.Debug().Msg("notify block from other bp")
		cp.notifyBlock(block, false)
	}
}

func checkDebugSleep(isBP bool) {
	if isBP {
		_ = TestDebugger.Check(DEBUG_CHAIN_BP_SLEEP, 0, nil)
	} else {
		_ = TestDebugger.Check(DEBUG_CHAIN_OTHER_SLEEP, 0, nil)
	}
}

func (cp *chainProcessor) executeBlock(block *types.Block) error {
	checkDebugSleep(cp.isByBP)

	err := cp.ChainService.executeBlock(cp.state, block)
	cp.state = nil
	return err
}

func (cp *chainProcessor) execute(block *types.Block) error {
	if !cp.isMainChain {
		return nil
	}

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
	if _, err = cp.connectToChain(block); err != nil {
		return err
	}

	cp.notifyBlockByOther(block)

	return nil
}

func (cp *chainProcessor) connectToChain(block *types.Block) (types.BlockNo, error) {
	dbTx := cp.cdb.store.NewTx()
	defer dbTx.Discard()

	// skip to add hash/block if wal of block is already written
	oldLatest := cp.cdb.connectToChain(dbTx, block, cp.isByBP && cp.HasWAL())
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
		err := cp.reorg(cp.lastBlock, nil)
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

func (cs *ChainService) addBlockInternal(newBlock *types.Block, usedBstate *state.BlockState, peerID types.PeerID) (err error, cache bool) {
	if !cs.VerifyTimestamp(newBlock) {
		return &ErrBlock{
			err: errBlockTimestamp,
			block: &types.BlockInfo{
				Hash: newBlock.BlockHash(),
				No:   newBlock.BlockNo(),
			},
		}, false
	}

	var (
		bestBlock  *types.Block
		savedBlock *types.Block
	)

	if bestBlock, err = cs.cdb.GetBestBlock(); err != nil {
		return err, false
	}

	// The newly produced block becomes stale because the more block(s) are
	// connected to the blockchain so that the best block is cha/nged. In this
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

	//Fork should never occur in raft.
	checkFork := func(block *types.Block) error {
		if cs.IsForkEnable() {
			return nil
		}
		if usedBstate != nil {
			return nil
		}

		savedBlock, err = cs.getBlockByNo(newBlock.GetHeader().GetBlockNo())
		if err == nil {
			/* TODO change to error after testing */
			logger.Fatal().Str("newblock", newBlock.ID()).Str("savedblock", savedBlock.ID()).Msg("drop block making invalid fork")
			return &ErrBlock{
				err:   errBlockInvalidFork,
				block: &types.BlockInfo{Hash: newBlock.BlockHash(), No: newBlock.BlockNo()},
			}
		}

		return nil
	}

	if err := checkFork(newBlock); err != nil {
		return err, false
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
			return nil, false
		}

		return err, false
	}

	// try to acquire lock
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

	logger.Info().Uint64("best", cs.cdb.getBestBlockNo()).Str("hash", newBlock.ID()).Msg("block added successfully")

	return nil, true
}

func (cs *ChainService) addBlock(newBlock *types.Block, usedBstate *state.BlockState, peerID types.PeerID) error {
	hashID := types.ToHashID(newBlock.BlockHash())

	if cs.errBlocks.Contains(hashID) {
		return ErrBlockCachedErrLRU
	}

	var err error

	if cs.IsConnectedBlock(newBlock) {
		logger.Warn().Str("hash", newBlock.ID()).Uint64("no", newBlock.BlockNo()).Msg("block is already connected")
		return nil
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
	verifyOnly       bool
	validateSignWait ValidateSignWaitFn
	bi               *types.BlockHeaderInfo
}

func newBlockExecutor(cs *ChainService, bState *state.BlockState, block *types.Block, verifyOnly bool) (*blockExecutor, error) {
	var exec TxExecFn
	var validateSignWait ValidateSignWaitFn
	var bi *types.BlockHeaderInfo

	commitOnly := false

	// The DPoS block factory executes transactions during block generation. In
	// such a case it send block with block state so that bState != nil. On the
	// contrary, the block propagated from the network is not half-executed.
	// Hence we need a new block state and tx executor (execTx).
	if bState == nil {
		if err := cs.validator.ValidateBlock(block); err != nil {
			return nil, err
		}

		bState = state.NewBlockState(
			cs.sdb.OpenNewStateDB(cs.sdb.GetRoot()),
			state.SetPrevBlockHash(block.GetHeader().GetPrevBlockHash()),
		)
		bi = types.NewBlockHeaderInfo(block)
		exec = NewTxExecutor(cs.ChainConsensus, cs.cdb, bi, contract.ChainService)

		validateSignWait = func() error {
			return cs.validator.WaitVerifyDone()
		}
	} else {
		logger.Debug().Uint64("block no", block.BlockNo()).Msg("received block from block factory")
		// In this case (bState != nil), the transactions has already been
		// executed by the block factory.
		commitOnly = true
	}
	bState.SetGasPrice(system.GetGasPriceFromState(bState))
	bState.Receipts().SetHardFork(cs.cfg.Hardfork, block.BlockNo())

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
		verifyOnly:       verifyOnly,
		validateSignWait: validateSignWait,
		bi:               bi,
	}, nil
}

// NewTxExecutor returns a new TxExecFn.
func NewTxExecutor(ccc consensus.ChainConsensusCluster, cdb contract.ChainAccessor, bi *types.BlockHeaderInfo, preloadService int) TxExecFn {
	return func(bState *state.BlockState, tx types.Transaction) error {
		if bState == nil {
			logger.Error().Msg("bstate is nil in txexec")
			return ErrGatherChain
		}
		if bi.ForkVersion < 0 {
			logger.Error().Err(ErrInvalidBlockHeader).Msgf("ChainID.ForkVersion = %d", bi.ForkVersion)
			return ErrInvalidBlockHeader
		}
		blockSnap := bState.Snapshot()

		err := executeTx(ccc, cdb, bState, tx, bi, preloadService)
		if err != nil {
			logger.Error().Err(err).Str("hash", enc.ToString(tx.GetHash())).Msg("tx failed")
			if err2 := bState.Rollback(blockSnap); err2 != nil {
				logger.Panic().Err(err).Msg("failed to rollback block state")
			}

			return err
		}
		return nil
	}
}

func (e *blockExecutor) execute() error {
	// Receipt must be committed unconditionally.
	if !e.commitOnly {
		defer contract.CloseDatabase()
		var preloadTx *types.Tx
		numTxs := len(e.txs)
		for i, tx := range e.txs {
			// if tx is not the last one, preload the next tx
			if i != numTxs-1 {
				preloadTx = e.txs[i+1]
				contract.RequestPreload(e.BlockState, e.bi, preloadTx, tx, contract.ChainService)
			}
			// execute the transaction
			if err := e.execTx(e.BlockState, types.NewTransaction(tx)); err != nil {
				//FIXME maybe system error. restart or panic
				// all txs have executed successfully in BP node
				return err
			}
			// mark the next preload tx to be executed
			contract.SetPreloadTx(preloadTx, contract.ChainService)
		}

		if e.validateSignWait != nil {
			if err := e.validateSignWait(); err != nil {
				return err
			}
		}

		//TODO check result of verifing txs
		if err := SendBlockReward(e.BlockState, e.coinbaseAcccount); err != nil {
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

	if !e.verifyOnly {
		if err := e.commit(); err != nil {
			return err
		}
	}

	logger.Debug().Msg("block executor finished")
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
	logger.Debug().Str("hash", block.ID()).Uint64("no", block.GetHeader().BlockNo).Msg("start to execute")

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
	bstate = bstate.SetPrevBlockHash(block.GetHeader().GetPrevBlockHash())
	// TODO refactoring: receive execute function as argument (executeBlock or executeBlockReco)
	ex, err := newBlockExecutor(cs, bstate, block, false)
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

	cs.notifyEvents(block, ex.BlockState)

	cs.Update(block)

	logger.Debug().Uint64("no", block.GetHeader().BlockNo).Msg("end to execute")

	return nil
}

// verifyBlock execute block and verify state root but doesn't save data to database.
// ChainVerifier use this function.
func (cs *ChainService) verifyBlock(block *types.Block) error {
	var (
		err error
		ex  *blockExecutor
	)

	// Caution: block must belong to the main chain.
	logger.Debug().Str("hash", block.ID()).Uint64("no", block.GetHeader().BlockNo).Msg("start to verify")

	ex, err = newBlockExecutor(cs, nil, block, true)
	if err != nil {
		return err
	}

	// contract & state DB update is done during execution.
	if err = ex.execute(); err != nil {
		return err
	}

	// set root of sdb to block root hash
	if err = cs.sdb.SetRoot(block.GetHeader().GetBlocksRootHash()); err != nil {
		return fmt.Errorf("failed to set root of sdb(no=%d,hash=%v)", block.BlockNo(), block.ID())
	}

	logger.Debug().Uint64("no", block.GetHeader().BlockNo).Msg("end verify")

	return nil
}

// TODO: Refactoring: batch
func (cs *ChainService) executeBlockReco(_ *state.BlockState, block *types.Block) error {
	// Caution: block must belong to the main chain.
	logger.Debug().Str("hash", block.ID()).Uint64("no", block.GetHeader().BlockNo).Msg("start to execute for reco")

	var (
		bestBlock *types.Block
		err       error
	)

	if bestBlock, err = cs.cdb.GetBestBlock(); err != nil {
		return err
	}

	// Check consensus info validity
	// TODO remove bestblock
	if err = cs.IsBlockValid(block, bestBlock); err != nil {
		return err
	}

	if !cs.sdb.GetStateDB().HasMarker(block.GetHeader().GetBlocksRootHash()) {
		logger.Error().Str("hash", block.ID()).Uint64("no", block.GetHeader().GetBlockNo()).Msg("state marker does not exist")
		return ErrStateNoMarker
	}

	// move stateroot
	if err := cs.sdb.SetRoot(block.GetHeader().GetBlocksRootHash()); err != nil {
		return fmt.Errorf("failed to set sdb(branchRoot:no=%d,hash=%v)", block.GetHeader().GetBlockNo(),
			block.ID())
	}

	cs.Update(block)

	logger.Debug().Uint64("no", block.GetHeader().BlockNo).Msg("end to execute for reco")

	return nil
}

func (cs *ChainService) notifyEvents(block *types.Block, bstate *state.BlockState) {
	blkNo := block.GetHeader().GetBlockNo()
	blkHash := block.BlockHash()

	logger.Debug().Uint64("no", blkNo).Msg("add event from executed block")

	cs.RequestTo(message.MemPoolSvc, &message.MemPoolDel{
		Block: block,
	})

	cs.TellTo(message.RPCSvc, block)

	events := []*types.Event{}
	for idx, receipt := range bstate.Receipts().Get() {
		for _, e := range receipt.Events {
			e.SetMemoryInfo(receipt, blkHash, blkNo, int32(idx))
			events = append(events, e)
		}
	}

	if len(events) != 0 {
		cs.TellTo(message.RPCSvc, events)
	}
}

const maxRetSize = 1024

func adjustRv(ret string) string {
	if len(ret) > maxRetSize {
		modified, _ := json.Marshal(ret[:maxRetSize-4] + " ...")

		return string(modified)
	}
	return ret
}

func resetAccount(account *state.V, fee *big.Int, nonce *uint64) error {
	account.Reset()
	if fee != nil {
		if account.Balance().Cmp(fee) < 0 {
			return &types.InternalError{Reason: "fee is greater than balance"}
		}
		account.SubBalance(fee)
	}
	if nonce != nil {
		account.SetNonce(*nonce)
	}
	return account.PutState()
}

func executeTx(
	ccc consensus.ChainConsensusCluster,
	cdb contract.ChainAccessor,
	bs *state.BlockState,
	tx types.Transaction,
	bi *types.BlockHeaderInfo,
	preloadService int,
) error {
	var (
		txBody    = tx.GetBody()
		isQuirkTx = types.IsQuirkTx(tx.GetHash())
		account   []byte
		recipient []byte
		err       error
	)

	if account, err = name.Resolve(bs, txBody.GetAccount(), isQuirkTx); err != nil {
		return err
	}

	if tx.HasVerifedAccount() {
		txAcc := tx.GetVerifedAccount()
		tx.RemoveVerifedAccount()
		if !bytes.Equal(txAcc, account) {
			return types.ErrSignNotMatch
		}
	}

	err = tx.Validate(bi.ChainIdHash(), IsPublic())
	if err != nil {
		return err
	}

	sender, err := bs.GetAccountStateV(account)
	if err != nil {
		return err
	}

	err = tx.ValidateWithSenderState(sender.State(), bs.GasPrice, bi.ForkVersion)
	if err != nil {
		return err
	}

	if recipient, err = name.Resolve(bs, txBody.Recipient, isQuirkTx); err != nil {
		return err
	}
	var receiver *state.V
	status := "SUCCESS"
	if len(recipient) > 0 {
		receiver, err = bs.GetAccountStateV(recipient)
		if receiver != nil && txBody.Type == types.TxType_REDEPLOY {
			status = "RECREATED"
			receiver.SetRedeploy()
		}
	} else {
		receiver, err = bs.CreateAccountStateV(contract.CreateContractID(txBody.Account, txBody.Nonce))
		status = "CREATED"
	}
	if err != nil {
		return err
	}

	var txFee *big.Int
	var rv string
	var events []*types.Event
	switch txBody.Type {
	case types.TxType_NORMAL, types.TxType_REDEPLOY, types.TxType_TRANSFER, types.TxType_CALL, types.TxType_DEPLOY:
		rv, events, txFee, err = contract.Execute(bs, cdb, tx.GetTx(), sender, receiver, bi, preloadService, false)
		sender.SubBalance(txFee)
	case types.TxType_GOVERNANCE:
		txFee = new(big.Int).SetUint64(0)
		events, err = executeGovernanceTx(ccc, bs, txBody, sender, receiver, bi)
		if err != nil {
			logger.Warn().Err(err).Str("txhash", enc.ToString(tx.GetHash())).Msg("governance tx Error")
		}
	case types.TxType_FEEDELEGATION:
		balance := receiver.Balance()
		var fee *big.Int
		fee, err = tx.GetMaxFee(balance, bs.GasPrice, bi.ForkVersion)
		if err != nil {
			return err
		}
		if fee.Cmp(balance) > 0 {
			return types.ErrInsufficientBalance
		}
		var contractState *state.ContractState
		contractState, err = bs.OpenContractState(receiver.AccountID(), receiver.State())
		if err != nil {
			return err
		}
		err = contract.CheckFeeDelegation(recipient, bs, bi, cdb, contractState, txBody.GetPayload(),
			tx.GetHash(), txBody.GetAccount(), txBody.GetAmount())
		if err != nil {
			if err != types.ErrNotAllowedFeeDelegation {
				logger.Warn().Err(err).Str("txhash", enc.ToString(tx.GetHash())).Msg("checkFeeDelegation Error")
				return err
			}
			return types.ErrNotAllowedFeeDelegation
		}
		rv, events, txFee, err = contract.Execute(bs, cdb, tx.GetTx(), sender, receiver, bi, preloadService, true)
		receiver.SubBalance(txFee)
	}

	if err != nil {
		// Reset events on error
		if bi.ForkVersion >= 3 {
			events = nil
		}

		if !contract.IsRuntimeError(err) {
			return err
		}
		if txBody.Type != types.TxType_FEEDELEGATION || sender.AccountID() == receiver.AccountID() {
			sErr := resetAccount(sender, txFee, &txBody.Nonce)
			if sErr != nil {
				return sErr
			}
		} else {
			sErr := resetAccount(sender, nil, &txBody.Nonce)
			if sErr != nil {
				return sErr
			}
			sErr = resetAccount(receiver, txFee, nil)
			if sErr != nil {
				return sErr
			}
		}
		status = "ERROR"
		rv = err.Error()
	} else {
		if txBody.Type != types.TxType_FEEDELEGATION {
			if sender.Balance().Sign() < 0 {
				return &types.InternalError{Reason: "fee is greater than balance"}
			}
		} else {
			if receiver.Balance().Sign() < 0 {
				return &types.InternalError{Reason: "fee is greater than balance"}
			}
		}
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
		rv = adjustRv(rv)
	}
	bs.BpReward.Add(&bs.BpReward, txFee)

	receipt := types.NewReceipt(receiver.ID(), status, rv)
	receipt.FeeUsed = txFee.Bytes()
	receipt.TxHash = tx.GetHash()
	receipt.Events = events
	receipt.FeeDelegation = txBody.Type == types.TxType_FEEDELEGATION
	receipt.GasUsed = contract.GasUsed(txFee, bs.GasPrice, txBody.Type, bi.ForkVersion)

	return bs.AddReceipt(receipt)
}

func DecorateBlockRewardFn(fn BlockRewardFn) {
	SendBlockReward = func(bState *state.BlockState, coinbaseAccount []byte) error {
		if err := fn(bState, coinbaseAccount); err != nil {
			return err
		}

		return sendRewardCoinbase(bState, coinbaseAccount)
	}
}

func sendRewardCoinbase(bState *state.BlockState, coinbaseAccount []byte) error {
	bpReward := &bState.BpReward
	if bpReward.Cmp(new(big.Int).SetUint64(0)) <= 0 || coinbaseAccount == nil {
		logger.Debug().Str("reward", bpReward.String()).Msg("coinbase is skipped")
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

	orphanBlock := orphan.Block

	if (block.GetHeader().GetBlockNo() + 1) != orphanBlock.GetHeader().GetBlockNo() {
		return nil, fmt.Errorf("invalid orphan block no (p=%d, c=%d)", block.GetHeader().GetBlockNo(),
			orphanBlock.GetHeader().GetBlockNo())
	}

	logger.Info().Str("parent", block.ID()).
		Str("orphan", orphanBlock.ID()).
		Msg("connect orphan")

	if err := cs.op.removeOrphan(orphanID); err != nil {
		return nil, err
	}

	return orphanBlock, nil
}

func (cs *ChainService) isOrphan(block *types.Block) bool {
	prevhash := block.Header.PrevBlockHash
	_, err := cs.getBlock(prevhash)

	return err != nil
}

func (cs *ChainService) handleOrphan(block *types.Block, bestBlock *types.Block, peerID types.PeerID) error {
	err := cs.addOrphan(block)
	if err != nil {
		logger.Error().Err(err).Str("hash", block.ID()).Msg("add orphan block failed")

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
			mainblock = nil
			continue
		}
		// get main hash with same block height
		mainhash, err = cs.cdb.getHashByNo(
			types.BlockNo(mainblock.GetHeader().GetBlockNo()))
		if err != nil {
			mainblock = nil
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

func (cs *ChainService) setSkipMempool(isSync bool) {
	//don't use mempool if sync is in progress
	cs.validator.signVerifier.SetSkipMempool(isSync)
}
