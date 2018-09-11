package blockchain

import (
	"bytes"
	"fmt"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
)

type reorgBlock struct {
	BlockNo types.BlockNo
	Hash    []byte
}

const (
	initBlkCount = 20
)

type reorganizer struct {
	//input info
	cs         *ChainService
	dbtx       *db.Transaction
	brTopBlock *types.Block //branch top block

	//collected info from chain
	brRootBlock *types.Block
	rfBlocks    []*reorgBlock //roll forward target blocks
	rbBlocks    []*reorgBlock //roll back target blocks

	rbTxs map[types.TxID]*types.Tx //rollbacked txs from rollback target blocks
}

func (cs *ChainService) needReorg(block *types.Block) bool {
	cdb := cs.cdb
	blockNo := types.BlockNo(block.GetHeader().GetBlockNo())

	isNeed := cdb.latest < blockNo

	if isNeed {
		logger.Debug().Uint64("blockNo", blockNo).Uint64("latestNo", cdb.latest).
			Str("prev", block.ID()).Msg("need reorganizing")
	}

	return isNeed
}

//TODO: on booting, retry reorganizing
//TODO: on booting, delete played tx of block. because deleting txs from mempool is done after commit
//TODO: gather delete request of played tx (1 msg)
func (cs *ChainService) reorg(topBlock *types.Block) error {
	reorgtx := cs.cdb.store.NewTx(true)

	logger.Info().Uint64("blockNo", topBlock.GetHeader().GetBlockNo()).Str("hash", topBlock.ID()).
		Msg("reorg started")

	reorg := &reorganizer{
		cs:         cs,
		dbtx:       &reorgtx,
		brTopBlock: topBlock,
		rfBlocks:   make([]*reorgBlock, 0, initBlkCount),
		rbBlocks:   make([]*reorgBlock, 0, initBlkCount),
		rbTxs:      make(map[types.TxID]*types.Tx),
	}

	err := reorg.gatherChainInfo()
	if err != nil {
		return err
	}

	/* XXX */
	//reorg.dumpRbBlocks()

	err = reorg.rollbackChain()
	if err != nil {
		return err
	}

	if err := reorg.rollforwardChain(); err != nil {
		return err
	}

	logger.Info().Msg("reorg end")

	reorgtx.Commit()

	return nil
}

func (reorg *reorganizer) dumpRbBlocks() {
	for _, rbBlock := range reorg.rbBlocks {
		logger.Debug().Str("hash", enc.ToString(rbBlock.Hash)).Uint64("blockNo", rbBlock.BlockNo).
			Msg("dump rollback block")
	}
}

// find branch root
// gather rollforard/rollback target blocks
func (reorg *reorganizer) gatherChainInfo() error {
	//find branch root block , gather rollforward Target block
	cdb := reorg.cs.cdb

	brBlock := reorg.brTopBlock
	brBlockNo := brBlock.GetHeader().GetBlockNo()
	brBlockHash := brBlock.BlockHash()

	latestNo := cdb.latest

	for {
		mainBlockHash, err := cdb.getHashByNo(brBlockNo)

		if latestNo < brBlockNo {
			//must not exist (no, hash) record
			if err == nil {
				return fmt.Errorf("block of main chain can't be higher than latest. no=%d, latest=%d",
					brBlockNo, latestNo)
			}
		} else {
			//must exist
			if err != nil {
				return err
			}

			if bytes.Equal(brBlock.Hash, mainBlockHash) {
				if latestNo == brBlockNo {
					return fmt.Errorf("best block can't be branch root block")
				}
				reorg.brRootBlock = brBlock

				logger.Debug().Str("hash", brBlock.ID()).Uint64("blockNo", brBlockNo).
					Msg("found branch root block")

				return nil
			}

			//gather rollback target

			logger.Debug().Str("hash", enc.ToString(mainBlockHash)).Uint64("blockNo", brBlockNo).
				Msg("gather rollback target")
			reorg.rbBlocks = append(reorg.rbBlocks, &reorgBlock{brBlockNo, mainBlockHash})
		}

		if brBlockNo <= 0 {
			break
		}

		//gather rollforward target
		logger.Debug().Str("hash", enc.ToString(brBlockHash)).Uint64("blockNo", brBlockNo).
			Msg("gather rollforward target")
		reorg.rfBlocks = append(reorg.rfBlocks, &reorgBlock{brBlockNo, brBlockHash})

		//get prev block from branch
		if brBlock, err = cdb.getBlock(brBlock.GetHeader().GetPrevBlockHash()); err != nil {
			return err
		}

		prevBrBlockNo := brBlock.GetHeader().GetBlockNo()
		if brBlockNo-1 != prevBrBlockNo {
			return fmt.Errorf("rollback target is not valid. block(%v), blockno(exp=%d,res=%d)",
				brBlock.ID(), brBlockNo-1, prevBrBlockNo)
		}
		brBlockNo = prevBrBlockNo
		brBlockHash = brBlock.BlockHash()
	}

	return fmt.Errorf("branch root block(%v) doesn't exist", reorg.brTopBlock.ID())
}

/*
	rollbackBlock
	rollback overall stateDB
*/
func (reorg *reorganizer) rollbackChain() error {
	cdb := reorg.cs.cdb

	for _, rbBlock := range reorg.rbBlocks {
		logger.Debug().Str("hash", enc.ToString(rbBlock.Hash)).Uint64("blockNo", rbBlock.BlockNo).
			Msg("rollback block")

		//get target block
		targetBlock, err := cdb.getBlock(rbBlock.Hash)
		if err != nil {
			return err
		}

		if targetBlock.GetHeader().GetBlockNo() != rbBlock.BlockNo {
			return fmt.Errorf("invalid rollback target block(%d, %v.err=%s)", rbBlock.BlockNo, rbBlock.Hash,
				err.Error())
		}

		reorg.rollbackBlock(targetBlock)
	}

	//rollback stateDB
	if err := reorg.rollbackChainState(); err != nil {
		logger.Debug().Err(err).Msg("reorganization failed")
		return err
	}

	return nil
}

func (reorg *reorganizer) rollbackChainState() error {
	brRootBlock := reorg.brRootBlock
	brRootBlockNo := brRootBlock.GetHeader().GetBlockNo()

	if err := reorg.cs.sdb.Rollback(brRootBlockNo); err != nil {

		return fmt.Errorf("failed to rollback sdb(branchRoot:no=%d,hash=%v)", brRootBlockNo,
			brRootBlock.ID())
	}

	return nil
}

/*
	rollbackBlock
	- cdb.latest -= - 1
	- gather rollbacked Txs
*/
func (reorg *reorganizer) rollbackBlock(block *types.Block) {
	cdb := reorg.cs.cdb

	blockNo := block.GetHeader().GetBlockNo()

	for _, tx := range block.GetBody().GetTxs() {
		reorg.rbTxs[types.ToTxID(tx.GetHash())] = tx
	}

	cdb.setLatest(blockNo - 1)
}

/*
	rollforward
		rollforwardBlock
		add rbTxs to mempool
*/
func (reorg *reorganizer) rollforwardChain() error {
	cs := reorg.cs
	cdb := cs.cdb

	for i := len(reorg.rfBlocks) - 1; i >= 0; i-- {
		rfBlock := reorg.rfBlocks[i]

		logger.Debug().Str("hash", enc.ToString(rfBlock.Hash)).Uint64("blockNo", rfBlock.BlockNo).
			Msg("rollforward block")

		targetBlock, err := cdb.getBlock(rfBlock.Hash)
		if err != nil {
			return fmt.Errorf("can not find target block(%d, %v)", rfBlock.BlockNo, rfBlock.Hash)
		}

		if targetBlock.GetHeader().GetBlockNo() != rfBlock.BlockNo {
			return fmt.Errorf("invalid target block(%d, %v)", rfBlock.BlockNo, rfBlock.Hash)
		}

		if err := reorg.rollforwardBlock(targetBlock); err != nil {
			return err
		}
	}

	//add rollbacked Tx to mempool (except played tx in roll forward)
	cntRbTxs := len(reorg.rbTxs)
	if cntRbTxs > 0 {
		txs := make([]*types.Tx, 0, cntRbTxs)
		logger.Debug().Int("tx count", cntRbTxs).Msg("tx add to mempool")

		for txID, tx := range reorg.rbTxs {
			logger.Debug().Str("txID", txID.String()).Msg("tx added")
			txs = append(txs, tx)
		}

		cs.RequestTo(message.MemPoolSvc, &message.MemPoolPut{
			Txs: txs,
		})
	}

	return nil
}

/*
	play Tx & update stateDB
	update db (blkNo, hash)
	cdb.latest
	tx delete from rbTxs
*/
func (reorg *reorganizer) rollforwardBlock(block *types.Block) error {
	cs := reorg.cs
	cdb := reorg.cs.cdb

	if err := cs.executeBlock(nil, block); err != nil {
		return err
	}

	if err := cdb.addBlock(reorg.dbtx, block, true, false); err != nil {
		return err
	}

	blockNo := block.GetHeader().GetBlockNo()
	cs.RequestTo(message.MemPoolSvc, &message.MemPoolDel{
		// FIXME: remove legacy
		BlockNo: blockNo,
		Txs:     block.GetBody().GetTxs(),
	})

	//SyncWithConsensus
	cdb.setLatest(blockNo)

	//remove played tx from rbTxs
	reorg.removePlayedTxs(block)

	return nil
}

func (reorg *reorganizer) removePlayedTxs(block *types.Block) {
	blockNo := block.GetHeader().GetBlockNo()
	txs := block.GetBody().GetTxs()

	for _, tx := range txs {
		txID := types.ToTxID(tx.GetHash())

		if _, exists := reorg.rbTxs[txID]; exists {
			logger.Debug().Str("tx", txID.String()).Uint64("blockNo", blockNo).
				Msg("played tx deleted from rollback Tx set")

			delete(reorg.rbTxs, txID)
		}
	}
}
