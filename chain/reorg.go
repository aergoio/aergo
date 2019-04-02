package chain

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
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

	recover bool
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

//TODO: on booting, retry reorganizing
//TODO: on booting, delete played tx of block. because deleting txs from mempool is done after commit
//TODO: gather delete request of played tx (1 msg)
func (cs *ChainService) reorg(topBlock *types.Block, marker *ReorgMarker) error {
	isReco := (marker != nil)

	logger.Info().Uint64("blockNo", topBlock.GetHeader().GetBlockNo()).Str("hash", topBlock.ID()).
		Bool("recovery", isReco).Msg("reorg started")

	reorg := &reorganizer{
		cs:         cs,
		brTopBlock: topBlock,
		newBlocks:  make([]*types.Block, 0, initBlkCount),
		oldBlocks:  make([]*types.Block, 0, initBlkCount),
		recover:    isReco,
	}

	var err error

	// TODO gatherChain 전에 marker build한후 공통코드 타게 하자
	if isReco {
		err = reorg.gatherChainInfoReco(marker)
		if err != nil {
			return err
		}
	} else {
		err = reorg.gatherChainInfo()
		if err != nil {
			return err
		}
	}

	if !cs.NeedReorganization(reorg.brStartBlock.BlockNo()) {
		return consensus.ErrorConsensus{Msg: "reorganization rejected by consensus"}
	}

	err = reorg.rollbackChain()
	if err != nil {
		return err
	}

	if isReco {
		// only need  consensus update
		if err := reorg.rollforwardChainReco(); err != nil {
			return err
		}
	} else {
		//it's possible to occur error while executing branch block (forgery)
		if err := reorg.rollforwardChain(); err != nil {
			return err
		}
	}

	if err := reorg.swapChain(); err != nil {
		switch ec := err.(type) {
		case *ErrDebug:
			return ec
		}
		logger.Fatal().Err(err).Msg("reorg failed while swapping chain, it can't recover")
		return err
	}

	cs.stat.updateEvent(ReorgStat, reorg.oldBlocks[0], reorg.newBlocks[0], reorg.brStartBlock)

	logger.Info().Msg("reorg end")

	return nil
}

// swap oldchain to newchain oneshot (best effort)
//  - chain height mapping
//  - tx mapping
//  - best block
func (reorg *reorganizer) swapChain() error {
	cdb := reorg.cs.cdb

	logger.Info().Msg("swap chain to new branch")

	marker := NewReorgMarker(reorg)

	if err := debugger.check(DEBUG_CHAIN_STOP_1); err != nil {
		return err
	}

	if err := cdb.writeReorgMarker(marker); err != nil {
		return err
	}

	if err := debugger.check(DEBUG_CHAIN_STOP_2); err != nil {
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

	if err := debugger.check(DEBUG_CHAIN_STOP_3); err != nil {
		return err
	}

	cdb.deleteReorgMarker()

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

	// insert new tx mapping
	for i := len(reorg.newBlocks) - 1; i >= 0; i-- {
		newBlock := reorg.newBlocks[i]

		for _, tx := range newBlock.GetBody().GetTxs() {
			delete(oldTxs, types.ToTxID(tx.GetHash()))
		}

		dbTx := cs.cdb.store.NewTx()
		defer dbTx.Discard()

		if err := cdb.addTxsOfBlock(&dbTx, newBlock.GetBody().GetTxs(), newBlock.BlockHash()); err != nil {
			return err
		}

		dbTx.Commit()
	}

	// delete old tx mapping
	txCnt := 0
	var dbTx db.Transaction

	for _, oldTx := range oldTxs {
		if dbTx == nil {
			dbTx = cs.cdb.store.NewTx()
		}
		defer dbTx.Discard()

		cdb.deleteTx(&dbTx, oldTx)

		txCnt++

		if txCnt >= TxBatchMax {
			dbTx.Commit()
			dbTx = nil
			txCnt = 0
		}
	}

	//add rollbacked Tx to mempool (except played tx in roll forward)
	count := len(oldTxs)
	logger.Debug().Int("tx count", count).Msg("tx add to mempool")

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

// Find branch root and gather rollforard/rollback target blocks
func (reorg *reorganizer) gatherChainInfo() error {
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
func (reorg *reorganizer) gatherChainInfoReco(marker *ReorgMarker) error {
	var err error
	cdb := reorg.cs.cdb

	var startBlock, bestBlock, topBlock *types.Block

	if startBlock, err = cdb.getBlock(marker.BrStartHash); err != nil {
		return err
	}

	if bestBlock, err = cdb.getBlock(marker.BrBestHash); err != nil {
		return err
	}

	if topBlock, err = cdb.getBlock(marker.BrTopHash); err != nil {
		return err
	}

	if bestBlock.GetHeader().GetBlockNo() >= topBlock.GetHeader().GetBlockNo() ||
		startBlock.GetHeader().GetBlockNo() >= bestBlock.GetHeader().GetBlockNo() ||
		startBlock.GetHeader().GetBlockNo() >= topBlock.GetHeader().GetBlockNo() {
		return ErrInvalidReorgMarker
	}

	reorg.brStartBlock = startBlock
	reorg.bestBlock = bestBlock

	for tmpBlk := bestBlock; tmpBlk.GetHeader().GetBlockNo() > startBlock.GetHeader().GetBlockNo(); {
		reorg.oldBlocks = append(reorg.oldBlocks, tmpBlk)
		logger.Debug().Str("hash", tmpBlk.ID()).Uint64("blockNo", tmpBlk.GetHeader().GetBlockNo()).
			Msg("gather rollback target for reco")

		if tmpBlk, err = cdb.getBlock(tmpBlk.GetHeader().GetPrevBlockHash()); err != nil {
			return err
		}
	}

	for tmpBlk := topBlock; tmpBlk.GetHeader().GetBlockNo() > startBlock.GetHeader().GetBlockNo(); {
		reorg.newBlocks = append(reorg.newBlocks, tmpBlk)

		logger.Debug().Str("hash", tmpBlk.ID()).Uint64("blockNo", tmpBlk.GetHeader().GetBlockNo()).
			Msg("gather rollforward target for reco")

		if tmpBlk, err = cdb.getBlock(tmpBlk.GetHeader().GetPrevBlockHash()); err != nil {
			return err
		}
	}

	return nil
}

func (reorg *reorganizer) rollbackChain() error {
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
func (reorg *reorganizer) rollforwardChain() error {
	cs := reorg.cs

	logger.Info().Msg("rollforward chain started")

	for i := len(reorg.newBlocks) - 1; i >= 0; i-- {
		newBlock := reorg.newBlocks[i]
		newBlockNo := newBlock.GetHeader().GetBlockNo()

		if err := cs.executeBlock(nil, newBlock); err != nil {
			logger.Error().Str("hash", newBlock.ID()).Uint64("no", newBlockNo).
				Msg("failed to execute block in reorg")
			return err
		}
	}

	return nil
}

func (reorg *reorganizer) rollforwardChainReco() error {
	cs := reorg.cs

	logger.Info().Msg("rollforward chain started for reco")

	for i := len(reorg.newBlocks) - 1; i >= 0; i-- {
		newBlock := reorg.newBlocks[i]
		newBlockNo := newBlock.GetHeader().GetBlockNo()

		//TODO cs를 closure로 받으면 refactoring 가능
		if err := cs.executeBlockReco(nil, newBlock); err != nil {
			logger.Error().Str("hash", newBlock.ID()).Uint64("no", newBlockNo).
				Msg("failed to execute block in reorg for reco")
			return err
		}
	}

	return nil
}
