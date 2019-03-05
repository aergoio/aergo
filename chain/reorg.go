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

type reorganizer struct {
	//input info
	cs         *ChainService
	dbtx       *db.Transaction
	brTopBlock *types.Block //branch top block

	//collected info from chain
	brStartBlock *types.Block
	newBlocks    []*types.Block //roll forward target blocks
	oldBlocks    []*types.Block //roll back target blocks
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
func (cs *ChainService) reorg(topBlock *types.Block) error {
	reorgtx := cs.cdb.store.NewTx()

	logger.Info().Uint64("blockNo", topBlock.GetHeader().GetBlockNo()).Str("hash", topBlock.ID()).
		Msg("reorg started")

	reorg := &reorganizer{
		cs:         cs,
		dbtx:       &reorgtx,
		brTopBlock: topBlock,
		newBlocks:  make([]*types.Block, 0, initBlkCount),
		oldBlocks:  make([]*types.Block, 0, initBlkCount),
	}

	err := reorg.gatherChainInfo()
	if err != nil {
		return err
	}

	if !cs.NeedReorganization(reorg.brStartBlock.BlockNo()) {
		return consensus.ErrorConsensus{Msg: "reorganization rejected by consensus"}
	}

	err = reorg.rollbackChain()
	if err != nil {
		return err
	}

	//it's possible to occur error while executing branch block (forgery)
	if err := reorg.rollforwardChain(); err != nil {
		return err
	}

	if err := reorg.swapChain(); err != nil {
		return err
	}

	logger.Info().Msg("reorg end")

	return nil
}

// swap oldchain to newchain oneshot (best effort)
//  - chain height mapping
//  - tx mapping
//  - best block
func (reorg *reorganizer) swapChain() error {
	cs := reorg.cs

	cs.cdb.swapChain(reorg.newBlocks)

	reorg.swapTxMapping()

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
	if count > 0 {
		//txs := make([]*types.Tx, 0, count)
		logger.Debug().Int("tx count", count).Msg("tx add to mempool")

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

func (reorg *reorganizer) rollbackChain() error {
	brStartBlock := reorg.brStartBlock
	brStartBlockNo := brStartBlock.GetHeader().GetBlockNo()

	if err := reorg.cs.sdb.Rollback(brStartBlock.GetHeader().GetBlocksRootHash()); err != nil {
		return fmt.Errorf("failed to rollback sdb(branchRoot:no=%d,hash=%v)", brStartBlockNo,
			brStartBlock.ID())
	}

	reorg.cs.Update(brStartBlock)

	dbTx := reorg.cs.cdb.NewTx()
	for _, blk := range reorg.oldBlocks {
		reorg.cs.cdb.deleteReceipts(&dbTx, blk.GetHash(), blk.BlockNo())
	}
	dbTx.Commit()

	return nil
}

/*
	rollforward
		rollforwardBlock
		add oldTxs to mempool
*/
func (reorg *reorganizer) rollforwardChain() error {
	cs := reorg.cs

	for i := len(reorg.newBlocks) - 1; i >= 0; i-- {
		newBlock := reorg.newBlocks[i]
		newBlockNo := newBlock.GetHeader().GetBlockNo()

		logger.Debug().Str("hash", enc.ToString(newBlock.Hash)).Uint64("blockNo", newBlockNo).
			Msg("rollforward block")

		if err := cs.executeBlock(nil, newBlock); err != nil {
			return err
		}
	}

	return nil
}
