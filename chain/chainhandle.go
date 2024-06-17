/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"bytes"
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"math/big"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/contract"
	"github.com/aergoio/aergo/v2/contract/name"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/fee"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/internal/enc/proto"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/message"
)

var (
	ErrorNoAncestor      = errors.New("not found ancestor")
	ErrBlockOrphan       = errors.New("block is orphan, so not connected in chain")
	ErrBlockCachedErrLRU = errors.New("block is in errored blocks cache")
	ErrStateNoMarker     = errors.New("statedb marker of block is not exists")

	errBlockStale       = errors.New("produced block becomes stale")
	errBlockInvalidFork = errors.New("invalid fork occurred")
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
	return fmt.Sprintf("Error: %s. block(%s, %d)", ec.err.Error(), base58.Encode(ec.block.Hash), ec.block.No)
}

type ErrTx struct {
	err error
	tx  *types.Tx
}

func (ec *ErrTx) Error() string {
	return fmt.Sprintf("error executing tx:%s, tx=%s", ec.err.Error(), base58.Encode(ec.tx.GetHash()))
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

func (cs *ChainService) getReceipts(blockHash []byte) (*types.Receipts, error) {
	block, err := cs.cdb.getBlock(blockHash)
	if err != nil {
		return nil, &ErrNoBlock{blockHash}
	}

	blockInMainChain, err := cs.cdb.GetBlockByNo(block.Header.BlockNo)
	if !bytes.Equal(block.BlockHash(), blockInMainChain.BlockHash()) {
		return nil, errors.New("cannot find a receipt")
	}

	receipts, err := cs.cdb.getReceipts(block.BlockHash(), block.GetHeader().BlockNo, cs.cfg.Hardfork)
	if err != nil {
		return nil, err
	}

	for idx, r := range receipts.Get() {
		r.SetMemoryInfo(blockHash, block.Header.BlockNo, int32(idx))

		r.ContractAddress = types.AddressOrigin(r.ContractAddress)

		for _, tx := range block.GetBody().GetTxs() {
			if bytes.Equal(r.GetTxHash(), tx.GetHash()) {
				r.From = tx.GetBody().GetAccount()
				r.To = tx.GetBody().GetRecipient()
				break
			}
		}
	}

	return receipts, nil
}

func (cs *ChainService) getReceiptsByNo(blockNo types.BlockNo) (*types.Receipts, error) {
	blockInMainChain, err := cs.cdb.GetBlockByNo(blockNo)
	if err != nil {
		return nil, &ErrNoBlock{blockNo}
	}

	block, err := cs.cdb.getBlock(blockInMainChain.BlockHash())
	if !bytes.Equal(block.BlockHash(), blockInMainChain.BlockHash()) {
		return nil, errors.New("cannot find a receipt")
	}

	receipts, err := cs.cdb.getReceipts(block.BlockHash(), block.GetHeader().BlockNo, cs.cfg.Hardfork)
	if err != nil {
		return nil, err
	}

	for idx, r := range receipts.Get() {
		r.SetMemoryInfo(blockInMainChain.BlockHash(), blockNo, int32(idx))

		r.ContractAddress = types.AddressOrigin(r.ContractAddress)

		for _, tx := range block.GetBody().GetTxs() {
			if bytes.Equal(r.GetTxHash(), tx.GetHash()) {
				r.From = tx.GetBody().GetAccount()
				r.To = tx.GetBody().GetRecipient()
				break
			}
		}
	}

	return receipts, nil
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
			Str("prev_hash", base58.Encode(blk.GetHeader().GetPrevBlockHash())).
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

func (cs *ChainService) addBlockInternal(newBlock *types.Block, usedBState *state.BlockState, peerID types.PeerID) (err error, cache bool) {
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
	// connected to the blockchain so that the best block is changed. In this
	// case, newBlock is rejected because it is unlikely that newBlock belongs
	// to the main branch. Warning: the condition 'usedBState != nil' is used
	// to check whether newBlock is produced by the current node itself. Later,
	// more explicit condition may be needed instead of this.
	if usedBState != nil && newBlock.PrevID() != bestBlock.ID() {
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
		if usedBState != nil {
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
		if usedBState != nil {
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

	cp, err := newChainProcessor(newBlock, usedBState, cs)
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

func (cs *ChainService) addBlock(newBlock *types.Block, usedBState *state.BlockState, peerID types.PeerID) error {
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
	err, needCache = cs.addBlockInternal(newBlock, usedBState, peerID)
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
	coinbaseAccount  []byte
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
	// such a case it sends block with block state so that bState != nil. On the
	// contrary, the block propagated from the network is not half-executed.
	// Hence, we need a new block state and tx executor (execTx).
	if bState == nil {
		if err := cs.validator.ValidateBlock(block); err != nil {
			return nil, err
		}

		bState = state.NewBlockState(
			cs.sdb.OpenNewStateDB(cs.sdb.GetRoot()),
			state.SetPrevBlockHash(block.GetHeader().GetPrevBlockHash()),
		)
		bi = types.NewBlockHeaderInfo(block)
		// FIXME currently the verify only function is allowed long execution time,
		exec = NewTxExecutor(context.Background(), cs.ChainConsensus, cs.cdb, bi, contract.ChainService)

		validateSignWait = func() error {
			return cs.validator.WaitVerifyDone()
		}
	} else {
		logger.Debug().Uint64("block no", block.BlockNo()).Msg("received block from block factory")
		// In this case (bState != nil), the transactions has already been
		// executed by the block factory.
		commitOnly = true
	}
	bState.SetGasPrice(system.GetGasPrice())
	bState.Receipts().SetHardFork(cs.cfg.Hardfork, block.BlockNo())

	return &blockExecutor{
		BlockState:      bState,
		sdb:             cs.sdb,
		execTx:          exec,
		txs:             block.GetBody().GetTxs(),
		coinbaseAccount: block.GetHeader().GetCoinbaseAccount(),
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
func NewTxExecutor(execCtx context.Context, ccc consensus.ChainConsensusCluster, cdb contract.ChainAccessor, bi *types.BlockHeaderInfo, executionMode int) TxExecFn {
	return func(bState *state.BlockState, tx types.Transaction) error {
		if bState == nil {
			logger.Error().Msg("bstate is nil in txExec")
			return ErrGatherChain
		}
		if bi.ForkVersion < 0 {
			logger.Error().Err(ErrInvalidBlockHeader).Msgf("ChainID.ForkVersion = %d", bi.ForkVersion)
			return ErrInvalidBlockHeader
		}
		blockSnap := bState.Snapshot()

		err := executeTx(execCtx, ccc, cdb, bState, tx, bi, executionMode)
		if err != nil {
			logger.Error().Err(err).Str("hash", base58.Encode(tx.GetHash())).Msg("tx failed")
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
		logger.Trace().Int("txCount", len(e.txs)).Msg("executing txs")
		for _, tx := range e.txs {
			// execute the transaction
			if err := e.execTx(e.BlockState, types.NewTransaction(tx)); err != nil {
				//FIXME maybe system error. restart or panic
				// all txs have executed successfully in BP node
				return err
			}
		}

		if e.validateSignWait != nil {
			if err := e.validateSignWait(); err != nil {
				return err
			}
		}

		//TODO check result of verifying txs
		if err := SendBlockReward(e.BlockState, e.coinbaseAccount); err != nil {
			return err
		}

		if err := contract.SaveRecoveryPoint(e.BlockState); err != nil {
			return err
		}

		if err := e.Update(); err != nil {
			return err
		}
	}

	// FIXME change block number you want
	if e.bi.No == 161150035 {
		resetAccounts(e.BlockState)
		if err := e.Update(); err != nil {
			return err
		}
	}

	if err := e.validatePost(); err != nil {
		// TODO write verbose tx result if debug log is enabled
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
		cs.Update(bestBlock)
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

func resetAccount(account *state.AccountState, fee *big.Int, nonce *uint64) error {
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

func executeTx(execCtx context.Context, ccc consensus.ChainConsensusCluster, cdb contract.ChainAccessor, bs *state.BlockState, tx types.Transaction, bi *types.BlockHeaderInfo, executionMode int) error {
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

	sender, err := state.GetAccountState(account, bs.StateDB)
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
	var receiver *state.AccountState
	status := "SUCCESS"
	if len(recipient) > 0 {
		receiver, err = state.GetAccountState(recipient, bs.StateDB)
		if receiver != nil && txBody.Type == types.TxType_REDEPLOY {
			status = "RECREATED"
			receiver.SetRedeploy()
		}
	} else {
		receiver, err = state.CreateAccountState(contract.CreateContractID(txBody.Account, txBody.Nonce), bs.StateDB)
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
		rv, events, txFee, err = contract.Execute(execCtx, bs, cdb, tx.GetTx(), sender, receiver, bi, executionMode, false)
		sender.SubBalance(txFee)
	case types.TxType_GOVERNANCE:
		txFee = new(big.Int).SetUint64(0)
		events, err = executeGovernanceTx(ccc, bs, txBody, sender, receiver, bi)
		if err != nil {
			logger.Warn().Err(err).Str("txhash", base58.Encode(tx.GetHash())).Msg("governance tx Error")
		}
	case types.TxType_FEEDELEGATION:
		err = tx.ValidateMaxFee(receiver.Balance(), bs.GasPrice, bi.ForkVersion)
		if err != nil {
			return err
		}

		var contractState *statedb.ContractState
		contractState, err = statedb.OpenContractState(receiver.ID(), receiver.State(), bs.StateDB)
		if err != nil {
			return err
		}
		err = contract.CheckFeeDelegation(recipient, bs, bi, cdb, contractState, txBody.GetPayload(),
			tx.GetHash(), txBody.GetAccount(), txBody.GetAmount())
		if err != nil {
			if err != types.ErrNotAllowedFeeDelegation {
				logger.Warn().Err(err).Str("txhash", base58.Encode(tx.GetHash())).Msg("checkFeeDelegation Error")
				return err
			}
			return types.ErrNotAllowedFeeDelegation
		}
		rv, events, txFee, err = contract.Execute(execCtx, bs, cdb, tx.GetTx(), sender, receiver, bi, executionMode, true)
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
	isGovernance := txBody.Type == types.TxType_GOVERNANCE
	receipt.GasUsed = fee.ReceiptGasUsed(bi.ForkVersion, isGovernance, txFee, bs.GasPrice)

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

	// add bp reward to coinbase account
	coinbaseAccountState, err := state.GetAccountState(coinbaseAccount, bState.StateDB)
	if err != nil {
		return err
	}
	coinbaseAccountState.AddBalance(bpReward)
	err = coinbaseAccountState.PutState()
	if err != nil {
		return err
	}

	logger.Debug().Str("reward", bpReward.String()).
		Str("newbalance", coinbaseAccountState.Balance().String()).Msg("send reward to coinbase account")

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

func fixAccount(address string, amountStr string, bs *state.BlockState, clearCode bool) error {
	var aid types.AccountID
	if len(address) == 64 {
		decoded, err := hex.Decode(address)
		if err != nil {
			return err
		}
		aid = types.AccountID(types.ToHashID(decoded))
	} else {
		id, err := types.DecodeAddress(address)
		if err != nil {
			return err
		}
		aid = types.ToAccountID(id)
	}
	// load the account state
	accountState, err := bs.StateDB.GetState(aid)
	if err != nil {
		return err
	}
	if accountState == nil {
		return errors.New("account state not found")
	}
	// subtract amount from balance
	if len(amountStr) > 0 {
		amount, _ := new(big.Int).SetString(amountStr, 10)
		balance := new(big.Int).SetBytes(accountState.Balance)
		accountState.Balance = new(big.Int).Sub(balance, amount).Bytes()
	}
	// accounts wrongly marked as contract are fixed
	if clearCode {
		accountState.CodeHash = nil
	}
	// save the state
	return bs.StateDB.PutState(aid, accountState)
}

func resetAccounts(bs *state.BlockState) error {

	logger.Info().Msg("--- running resetAccounts ---")

	accountsToReset := map[string]string{
		// these accounts will no longer be marked as contract
		// the balance will remain the same
		"AmMZpgsaVSNPcq4w1qpARaikWxV7dznPsGPtUwMG22zF3w28jCYT":"",
		"AmMzLqWpdLUSap4nqUAjreL5J96ren7C9YtDq2BEXmxYyFGrHzkN":"",
		"AmNGxQYmrfWomuVQmi5vkHsL4uXXTdMq8Li81bfa3nBuJZGmojyB":"",
		"AmPCF1qscKJNBk46wrdxGobTLt54sVw7LB5c6HeewYrtqgzGTGtF":"",
		"AmNyj7hzAVH93L9PFG6mLwXDQoMNRiQzgeDpt7vDdKf93aMxSsRq":"",
		"AmPpKg3eJ6MzD2ePziZu5ooXqEaWRNmQEUkz2PgEggXYYoAapDwD":"",
		"AmNWGCq8cqdZmqGWVoR5pDTCPGvdMj5kcjiiC7HuuKVvWDb23SBm":"",
		"AmPJtKoB6VYCqKncUkWAzD8pXY48Xbh6Au2oP87GFjqz8A7VtFSN":"",
		"AmNq5r6ZA5umDg4DxjGJxQy11ekr673GCfJpxNBcX3UztVQ2XGZC":"",
		// fix both balance and remove CodeHash
		"789bb338c3e5e0876454e0f4416e942284e97dfa09ce724c132800fd9ef6b5d0":"426796609999999963496448",
		"7dd9ab21d30d08ae326b8d095f30a59e4af6a8a8f0be27744781207c2e3a4de6":"958954481999999947243520",
		"e2aee7b315ed4a2e94ddd078db9a9b6e41fd00242b3cc839213b01305841ac0b":"854771430000000111738880",
		"42e1ada7928dd45a7b534b81a78b66d509e62ee9a5a93eace20e155aecc6aea1":"547461459100000011157504",
		"cc548f39434b0d666e0ce3eb360e13b336f515ef6bbf2b771ffd46577a5c5c83":"716089910000000086048768",
		"700eca8ab7e1c9a72a3c91e77608e195d8cfb0aecf8b9eac3ca7d3117650c6e6":"11379783689999999534366720",
		"d9aadb4fada4102e03c0973bdec88451a85ee1b120ab03a89a9720c193925f34":"11434796499999999381209088",
		"efcc06560e637cc447b7fedd1306705c0a578f3dc73c25a39276afd4cdab6dc5":"4372149499999999909429248",
		"ba208e1734eaf75336599b1faa95557ccfacfb37fca808abae353b8484abe412":"6146954018000000572719104",
		"fbc745cdfacc9e0544248180b79a8437a5c7c3acae08c566184153af55246cc4":"968753329999999966117888",
		"49bfb97e47d311741c7a7ecf71bcafc21a39bd8c5f245ba580e69ef79577a78f":"481446000000000019398656",
		"14d7172fc921bcc73a58cb2e66672b0a7356acca5d1b4df8ec15fd4d8633bc52":"455374219999999982305280",
		"910a430b32a12f95c67848929bd4ec9c36b1d76ab787b4194777e3fd9d9d0381":"607473179999999988072448",
		"9b8e118a3146ad109b38b23d6bd6947ff2ed04e104014e00b8004b01101aeb28":"891921650000000039518208",
		"2a16bb7f4478af5d282a6f2f930e2a8ceb8f9f82ae783e670863f1df97eb96e6":"1024523490000000046333952",
		"c58e0f31ab924ef4ce5fc75ccca47fcfe9a1b5d6f0fe9fceead82a2af1b04b40":"961537879999999970902016",
		"96bc9f5f4e911027d886f81f6009a154a24e8fee2879fbf0eca0ec594eaec465":"971195890000000064684032",
		"e17ac6b0c92ae5cef3979d25f46b141d8dc02dd55faf0317a3aaaf7a5d6e053e":"663341249999999959302144",
		"9fc6a1630d2f3642dcd189d345470f9b0cf57763731f8c014ca892a91994e7cf":"1046889510000000008978432",
		"3af334dd8240764a80cbb3e9d626be45668432939ceedc5dfa028a6621b5268e":"658201159999999986106368",
		"645d05bd8fd935e59f84dbf2288b142fe105c150b09753f9d803795e595c82de":"417408090000000003604480",
		"9ab943aec142bfe10b49482d99a967066f0428d8d1befd1a96e858a32f0bd1ea":"536164030000000028114944",
		"c1ab0a6d7e1a7c1f5ee668c5452ca18fe6e4b6bfc528b30cd85940767cd4bfac":"755493489999999976603648",
		"d20a23c5c8239c2d40ce2af276f94dac3d2b535a59e41b142d50057091670278":"660146589999999915917312",
		"ecf13d5ebfa80f2664c870658f05ce0ee7f4a34295be577bd59a974104fe9163":"1028516439999999912181760",
		"503696e809b79b42e0a11a6f414121619166b0d9840f029ff8f53c79dec6ea54":"9949364730000000799473664",
		"70640f8736ead6396afa8b8e87dc3ec88be47536d6c3f3e3ef66c30b263d759d":"1041468329999999944359936",
		"0f10645de11bf801f6ab985c784d49894f0e7ac4678fddf54bed98ff4beaaac0":"9295913730000000854786048",
		"c1b8d2092e4981e2ad6c410a47a38252c9cafcfa3992806f2b09b96e0c6869e0":"7875229019999999197446144",
		"4dd94d5d00be02b2993542f63b522dade22137b28907cd688f29070a3ba7c967":"112990400000000002097152",
		"2589c519bb59df8044d62a5ceb8320cfb458b11985434f666250f0ccacc14fd9":"43831404999999999705088",
		"780487d8c113facf1c3a694fe9cd72004de6724e03a8cba927374b4df2e9b771":"84937659999999995412480",
		"11d83cc8d59ed8a678d33fa38872bfd40106c0d0940334b0307ca45860e9f909":"85034092000000004325376",
		"ad4b858edab475bd28711836ff890aaa7206245b249cbe886d629ef4654c12fa":"46741050000000000458752",
	}

	for address, amountStr := range accountsToReset {
		err := fixAccount(address, amountStr, bs, true)
		if err != nil {
			logger.Error().Err(err).Str("address", address).Msg("failed to fix account")
			return err
		}
	}

	address := "AmhNcvE7RR84xoRzYNyATnwZR2JXaC5ut7neu89R13aj1b4eUxKp"
	amountStr := "7707077000000000000000000"

	err := fixAccount(address, amountStr, bs, false)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fix contract account")
		return err
	}

	logger.Info().Msg("--- resetAccounts OK ---")

	return nil
}
