package chain

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

const (
	initBlkCount = 20
)

var (
	reorgKeyStr = "_reorg_marker_"
	reorgKey    = []byte(reorgKeyStr)
)

var (
	ErrInvalidReorgMarker = errors.New("reorg marker is invalid")
	ErrMarkerNil          = errors.New("reorg marker is nil")
)

type reorganizer struct {
	//input info
	cs         *ChainService
	bestBlock  *types.Block
	brTopBlock *types.Block //branch top block

	//collected info from chain
	brStartBlock *types.Block
	newBlocks    []*types.Block //roll forward target blocks
	oldBlocks    []*types.Block //roll back target blocks

	marker *ReorgMarker

	recover bool

	gatherFn       func() error
	gatherPostFn   func()
	executeBlockFn func(bstate *state.BlockState, block *types.Block) error
}

type ErrReorgBlock struct {
	msg string

	blockNo   uint64
	blockHash []byte
}

func (ec *ErrReorgBlock) Error() string {
	if ec.blockHash != nil {
		return fmt.Sprintf("%s, block:%d,%s", ec.msg, ec.blockNo, enc.ToString(ec.blockHash))
	} else if ec.blockNo != 0 {
		return fmt.Sprintf("%s, block:%d", ec.msg, ec.blockNo)
	} else {
		return fmt.Sprintf("%s", ec.msg)
	}
}

var (
	ErrInvalidBranchRoot  = errors.New("best block can't be branch root block")
	ErrGatherChain        = errors.New("new/old blocks must exist")
	ErrNotExistBranchRoot = errors.New("branch root block doesn't exist")
	ErrInvalidSwapChain   = errors.New("New chain is not longer than old chain")
	ErrInvalidBlockHeader = errors.New("invalid block header")

	errMsgNoBlock         = "block not found in the chain DB"
	errMsgInvalidOldBlock = "rollback target is not valid"
)

func (cs *ChainService) needReorg(block *types.Block) bool {
	cdb := cs.cdb
	blockNo := block.BlockNo()

	latest := cdb.getBestBlockNo()
	isNeed := latest < blockNo

	if isNeed {
		logger.Debug().
			Uint64("blockNo", blockNo).
			Uint64("latestNo", latest).
			Str("prev", block.ID()).
			Msg("need reorganization")
	}

	return isNeed
}

func newReorganizer(cs *ChainService, topBlock *types.Block, marker *ReorgMarker) (*reorganizer, error) {
	isReco := (marker != nil)

	reorg := &reorganizer{
		cs:         cs,
		brTopBlock: topBlock,
		newBlocks:  make([]*types.Block, 0, initBlkCount),
		oldBlocks:  make([]*types.Block, 0, initBlkCount),
		recover:    isReco,
		marker:     marker,
	}

	if isReco {
		marker.setCDB(reorg.cs.cdb)

		if err := reorg.initRecovery(marker); err != nil {
			return nil, err
		}

		reorg.gatherFn = reorg.gatherReco
		reorg.gatherPostFn = nil
		reorg.executeBlockFn = cs.executeBlockReco
	} else {
		reorg.gatherFn = reorg.gather
		reorg.gatherPostFn = reorg.newMarker
		reorg.executeBlockFn = cs.executeBlock
	}

	TestDebugger.Check(DEBUG_CHAIN_RANDOM_STOP, 0, nil)

	return reorg, nil
}

// TODO: gather delete request of played tx (1 msg)
func (cs *ChainService) reorg(topBlock *types.Block, marker *ReorgMarker) error {
	logger.Info().Uint64("blockNo", topBlock.GetHeader().GetBlockNo()).Str("hash", topBlock.ID()).
		Bool("recovery", (marker != nil)).Msg("reorg started")

	begT := time.Now()

	reorg, err := newReorganizer(cs, topBlock, marker)
	if err != nil {
		logger.Error().Err(err).Msg("new reorganizer failed")
		return err
	}

	err = reorg.gatherFn()
	if err != nil {
		return err
	}

	if reorg.gatherPostFn != nil {
		reorg.gatherPostFn()
	}

	if !cs.NeedReorganization(reorg.brStartBlock.BlockNo()) {
		return consensus.ErrorConsensus{Msg: "reorganization rejected by consensus"}
	}

	err = reorg.rollback()
	if err != nil {
		return err
	}

	//it's possible to occur error while executing branch block (forgery)
	if err := reorg.rollforward(); err != nil {
		return err
	}

	if err := reorg.swapChain(); err != nil {
		switch ec := err.(type) {
		case *ErrDebug:
			return ec
		}
		logger.Fatal().Err(err).Msg("reorg failed while swapping chain, it can't recover")
		return err
	}

	cs.stat.updateEvent(ReorgStat, time.Since(begT), reorg.oldBlocks[0], reorg.newBlocks[0], reorg.brStartBlock)
	systemStateDB, err := cs.SDB().GetSystemAccountState()
	system.InitSystemParams(systemStateDB, system.RESET)
	logger.Info().Msg("reorg end")

	return nil
}

func (reorg *reorganizer) initRecovery(marker *ReorgMarker) error {
	var startBlock, bestBlock, topBlock *types.Block
	var err error

	if marker == nil {
		return ErrMarkerNil
	}

	topBlock = reorg.brTopBlock

	cdb := reorg.cs.cdb

	logger.Info().Str("marker", marker.toString()).Msg("new reorganizer")

	if startBlock, err = cdb.getBlock(marker.BrStartHash); err != nil {
		return err
	}

	if bestBlock, err = cdb.getBlock(marker.BrBestHash); err != nil {
		return err
	}

	if bestBlock.GetHeader().GetBlockNo() >= topBlock.GetHeader().GetBlockNo() ||
		startBlock.GetHeader().GetBlockNo() >= bestBlock.GetHeader().GetBlockNo() ||
		startBlock.GetHeader().GetBlockNo() >= topBlock.GetHeader().GetBlockNo() {
		return ErrInvalidReorgMarker
	}

	reorg.brStartBlock = startBlock
	reorg.bestBlock = bestBlock

	return nil
}

func (reorg *reorganizer) newMarker() {
	if reorg.marker != nil {
		return
	}

	reorg.marker = NewReorgMarker(reorg)
}

// swap oldchain to newchain oneshot (best effort)
//   - chain height mapping
//   - tx mapping
//   - best block
func (reorg *reorganizer) swapChain() error {
	logger.Info().Msg("swap chain to new branch")

	if err := TestDebugger.Check(DEBUG_CHAIN_STOP, 1, nil); err != nil {
		return err
	}

	if err := reorg.marker.write(); err != nil {
		return err
	}

	if err := TestDebugger.Check(DEBUG_CHAIN_STOP, 2, nil); err != nil {
		return err
	}

	reorg.deleteOldReceipts()

	//TODO batch notification of rollforward blocks

	if err := reorg.swapTxMapping(); err != nil {
		return err
	}

	if err := reorg.swapChainMapping(); err != nil {
		return err
	}

	if err := TestDebugger.Check(DEBUG_CHAIN_STOP, 3, nil); err != nil {
		return err
	}

	reorg.marker.delete()

	return nil
}

// swapChainMapping swaps chain meta from org chain to side chain and deleting reorg marker.
// it should be executed by 1 tx to be atomic.
func (reorg *reorganizer) swapChainMapping() error {
	cdb := reorg.cs.cdb

	logger.Info().Msg("swap chain mapping for new branch")

	best, err := cdb.GetBestBlock()
	if err != nil {
		return err
	}

	if reorg.recover && bytes.Equal(best.GetHash(), reorg.brTopBlock.GetHash()) {
		logger.Warn().Msg("swap of chain mapping has already finished")
		return nil
	}

	if err := cdb.swapChainMapping(reorg.newBlocks); err != nil {
		return err
	}

	return nil
}

func (reorg *reorganizer) swapTxMapping() error {
	// newblock/oldblock
	// push mempool (old - tx)
	cs := reorg.cs
	cdb := cs.cdb

	var oldTxs = make(map[types.TxID]*types.Tx)

	for _, oldBlock := range reorg.oldBlocks {
		for _, tx := range oldBlock.GetBody().GetTxs() {
			oldTxs[types.ToTxID(tx.GetHash())] = tx
		}
	}

	var overwrap int

	// insert new tx mapping
	for i := len(reorg.newBlocks) - 1; i >= 0; i-- {
		newBlock := reorg.newBlocks[i]

		for _, tx := range newBlock.GetBody().GetTxs() {
			if _, ok := oldTxs[types.ToTxID(tx.GetHash())]; ok {
				overwrap++
				delete(oldTxs, types.ToTxID(tx.GetHash()))
			}
		}

		dbTx := cs.cdb.store.NewTx()

		if err := cdb.addTxsOfBlock(&dbTx, newBlock.GetBody().GetTxs(), newBlock.BlockHash()); err != nil {
			dbTx.Discard()
			return err
		}

		dbTx.Commit()
	}

	// delete old tx mapping
	bulk := cdb.store.NewBulk()
	defer bulk.DiscardLast()

	for _, oldTx := range oldTxs {
		bulk.Delete(oldTx.Hash)
	}

	bulk.Flush()

	//add rollbacked Tx to mempool (except played tx in roll forward)
	count := len(oldTxs)
	logger.Debug().Int("tx count", count).Int("overwrapped count", overwrap).Msg("tx add to mempool")

	if count > 0 {
		//txs := make([]*types.Tx, 0, count)

		for _, tx := range oldTxs {
			//			logger.Debug().Str("txID", txID.String()).Msg("tx added")
			//			txs = append(txs, tx)
			cs.RequestTo(message.MemPoolSvc, &message.MemPoolPut{
				Tx: tx,
			})
		}
		//	cs.RequestTo(message.MemPoolSvc, &message.MemPoolPut{
		//		Txs: txs,
		//	})
	}
	return nil
}

func (reorg *reorganizer) dumpOldBlocks() {
	for _, block := range reorg.oldBlocks {
		logger.Debug().Str("hash", block.ID()).Uint64("blockNo", block.GetHeader().GetBlockNo()).
			Msg("dump rollback block")
	}
}

// Find branch root and gather rollforward/rollback target blocks
func (reorg *reorganizer) gather() error {
	//find branch root block , gather rollforward Target block
	var err error
	cdb := reorg.cs.cdb

	bestBlock, err := cdb.GetBestBlock()
	if err != nil {
		return err
	}
	reorg.bestBlock = bestBlock

	brBlock := reorg.brTopBlock
	brBlockNo := brBlock.BlockNo()

	curBestNo := cdb.getBestBlockNo()

	for {
		if brBlockNo <= curBestNo {
			mainBlock, err := cdb.GetBlockByNo(brBlockNo)
			// One must be able to look up any main chain block by its block
			// no from the chain DB.
			if err != nil {
				return &ErrReorgBlock{errMsgNoBlock, brBlockNo, nil}
			}

			//found branch root
			if bytes.Equal(brBlock.BlockHash(), mainBlock.BlockHash()) {
				if curBestNo == brBlockNo {
					return ErrInvalidBranchRoot
				}
				if len(reorg.newBlocks) == 0 || len(reorg.oldBlocks) == 0 {
					return ErrGatherChain
				}
				reorg.brStartBlock = brBlock

				logger.Debug().Str("hash", brBlock.ID()).Uint64("blockNo", brBlockNo).
					Msg("found branch root block")

				return nil
			}

			//gather rollback target
			logger.Debug().Str("hash", mainBlock.ID()).Uint64("blockNo", brBlockNo).
				Msg("gather rollback target")
			reorg.oldBlocks = append(reorg.oldBlocks, mainBlock)
		}

		if brBlockNo <= 0 {
			break
		}

		//gather rollforward target
		logger.Debug().Str("hash", brBlock.ID()).Uint64("blockNo", brBlockNo).
			Msg("gather rollforward target")
		reorg.newBlocks = append(reorg.newBlocks, brBlock)

		//get prev block from branch
		if brBlock, err = cdb.getBlock(brBlock.GetHeader().GetPrevBlockHash()); err != nil {
			return err
		}

		prevBrBlockNo := brBlock.GetHeader().GetBlockNo()
		if brBlockNo-1 != prevBrBlockNo {
			return &ErrReorgBlock{errMsgInvalidOldBlock, prevBrBlockNo, brBlock.BlockHash()}
		}
		brBlockNo = brBlock.GetHeader().GetBlockNo()
	}

	return ErrNotExistBranchRoot
}

// build reorg chain info from marker
func (reorg *reorganizer) gatherReco() error {
	var err error

	cdb := reorg.cs.cdb

	startBlock := reorg.brStartBlock
	bestBlock := reorg.bestBlock
	topBlock := reorg.brTopBlock

	reorg.brStartBlock = startBlock
	reorg.bestBlock = bestBlock

	gatherBlocksToStart := func(top *types.Block, stage string) ([]*types.Block, error) {
		blocks := make([]*types.Block, 0)

		for tmpBlk := top; tmpBlk.GetHeader().GetBlockNo() > startBlock.GetHeader().GetBlockNo(); {
			blocks = append(blocks, tmpBlk)

			logger.Debug().Str("stage", stage).Str("hash", tmpBlk.ID()).Uint64("blockNo", tmpBlk.GetHeader().GetBlockNo()).
				Msg("gather target for reco")

			if tmpBlk, err = cdb.getBlock(tmpBlk.GetHeader().GetPrevBlockHash()); err != nil {
				return blocks, err
			}
		}

		return blocks, nil
	}

	reorg.oldBlocks, err = gatherBlocksToStart(bestBlock, "rollback")
	if err != nil {
		return err
	}

	reorg.newBlocks, err = gatherBlocksToStart(topBlock, "rollforward")
	if err != nil {
		return err
	}

	return nil
}

func (reorg *reorganizer) rollback() error {
	brStartBlock := reorg.brStartBlock
	brStartBlockNo := brStartBlock.GetHeader().GetBlockNo()

	logger.Info().Str("hash", brStartBlock.ID()).Uint64("no", brStartBlockNo).Msg("rollback chain to branch start block")

	if err := reorg.cs.sdb.SetRoot(brStartBlock.GetHeader().GetBlocksRootHash()); err != nil {
		return fmt.Errorf("failed to rollback sdb(branchRoot:no=%d,hash=%v)", brStartBlockNo,
			brStartBlock.ID())
	}

	reorg.cs.Update(brStartBlock)

	return nil
}

func (reorg *reorganizer) deleteOldReceipts() {
	dbTx := reorg.cs.cdb.NewTx()
	for _, blk := range reorg.oldBlocks {
		reorg.cs.cdb.deleteReceipts(&dbTx, blk.GetHash(), blk.BlockNo())
	}
	dbTx.Commit()
}

/*
rollforward

	rollforwardBlock
	add oldTxs to mempool
*/
func (reorg *reorganizer) rollforward() error {
	//cs := reorg.cs

	logger.Info().Bool("recover", reorg.recover).Msg("rollforward chain started")

	for i := len(reorg.newBlocks) - 1; i >= 0; i-- {
		newBlock := reorg.newBlocks[i]
		newBlockNo := newBlock.GetHeader().GetBlockNo()

		if err := reorg.executeBlockFn(nil, newBlock); err != nil {
			logger.Error().Bool("recover", reorg.recover).Str("hash", newBlock.ID()).Uint64("no", newBlockNo).
				Msg("failed to execute block in reorg")
			return err
		}
	}

	return nil
}
