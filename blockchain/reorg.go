package blockchain

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/types"
)

type reorgElem struct {
	BlockNo types.BlockNo
	Hash    []byte
}

func (cs *ChainService) needReorg(block *types.Block) bool {
	cdb := cs.cdb
	blockNo := types.BlockNo(block.GetHeader().GetBlockNo())

	// assumption: not an orphan
	if blockNo > 0 && blockNo != cdb.latest+1 {
		return false
	}
	prevHash := block.GetHeader().GetPrevBlockHash()
	latestHash, err := cdb.getHashByNo(cdb.getBestBlockNo())

	if err != nil {
		// assertion case
		return false
	}

	isNeed := !bytes.Equal(prevHash, latestHash)
	if isNeed {
		logger.Debug().Uint64("blockNo", blockNo).Uint64("latestNo", cdb.latest).Str("prev", EncodeB64(prevHash)).Str("latest", EncodeB64(latestHash)).
			Msg("need reorg true")
	}

	return isNeed
}

func (cs *ChainService) reorg(block *types.Block) error {
	reorgtx := cs.cdb.store.NewTx(true)

	cdb := cs.cdb
	if cdb.ChainInfo != nil {
		cdb.SetReorganizing()
	}
	logger.Info().Uint64("blockNo", block.GetHeader().GetBlockNo()).Str("hash", block.ID()).
		Msg("reorg started")

	elems, err := cs.rollbackChain(&reorgtx, block)
	if err != nil {
		panic(err)
	}

	if err := cs.rollforwardChain(&reorgtx, elems); err != nil {
		logger.Error().Msg("failed reorg replay")
		return err
	}

	//TODO update cdb.bestblock

	if cdb.ChainInfo != nil {
		cdb.UnsetReorganizing()
	}

	logger.Info().Msg("reorg end")

	reorgtx.Commit()

	return nil
}

// rollback old main chain and collect blocks to process txs in branch(new main chain).
// New best block doesn't included. it will be processed after reorg.
func (cs *ChainService) rollbackChain(dbtx *db.Transaction, newBlock *types.Block) ([]reorgElem, error) {
	cdb := cs.cdb

	/* skip best block */
	brBlock := newBlock
	blockNo := brBlock.GetHeader().GetBlockNo()
	elems := make([]reorgElem, 0)

	var branchRootBlock *types.Block

	//find branch root block & gather target blocks for roll forward
	for {
		// get prev block info
		prevHash := brBlock.GetHeader().GetPrevBlockHash()
		brBlock, _ = cdb.getBlock(prevHash)
		mHash, _ := cdb.getHashByNo(brBlock.GetHeader().GetBlockNo())
		mBlock, _ := cdb.getBlock(mHash)
		blockNo--

		if blockNo != brBlock.GetHeader().GetBlockNo() {
			logger.Fatal().Uint64("blockNo", blockNo).Uint64("TBlockNo", brBlock.GetHeader().GetBlockNo()).
				Msg("failed rollback. invalid blockNo")
			return nil, fmt.Errorf("failed rollback. invalid blockNo(exp=%d, res=%d)", blockNo,
				brBlock.GetHeader().GetBlockNo())
		}

		if bytes.Equal(brBlock.Hash, mHash) {
			// branch root found
			logger.Debug().Uint64("blockNo", blockNo).Str("from", brBlock.ID()).
				Str("to", EncodeB64(mHash)).Msg("found branch root block")

			branchRootBlock = mBlock
			break
		}

		// error: cannot find branch root
		if blockNo == 0 {
			logger.Fatal().Uint64("blockNo", blockNo).Str("hash", brBlock.ID()).
				Msg("Error! blockNo(0) is diffrent in branch")
			break
		}

		cs.rollbackBlock(dbtx, mBlock)

		elems = cs.collectReorgTarget(brBlock, elems)
	}

	if branchRootBlock == nil {
		logger.Fatal().Str("hash", newBlock.ID()).Msg("failed to find branch root block")
		return nil, fmt.Errorf("failed to find branch root block of block(%s,%d)", newBlock.ID(),
			newBlock.GetHeader().GetBlockNo())
	}

	if err := cs.sdb.Rollback(branchRootBlock.GetHeader().GetBlockNo()); err != nil {
		logger.Fatal().Str("hash", newBlock.ID()).Str("root", branchRootBlock.ID()).
			Msg("failed to rollback sdb")
		return nil, err
	}

	return elems, nil
}

func (cs *ChainService) collectReorgTarget(block *types.Block, elems []reorgElem) []reorgElem {
	newElem := reorgElem{
		BlockNo: block.GetHeader().GetBlockNo(),
		Hash:    block.Hash,
	}
	elems = append(elems, newElem)

	logger.Debug().Uint64("blockNo", block.GetHeader().GetBlockNo()).Str("hash", block.ID()).
		Msg("collect reorg target block")
	return elems
}

func (cs *ChainService) rollforwardChain(dbtx *db.Transaction, elems []reorgElem) error {
	cdb := cs.cdb

	for _, elem := range elems {
		logger.Debug().Uint64("blockNo", uint64(elem.BlockNo)).
			Str("hash", EncodeB64(elem.Hash)).Msg("roll forward")

		blockIdx := types.BlockNoToBytes(elem.BlockNo)

		// change main chain info
		(*dbtx).Set(blockIdx, elem.Hash)

		//TODO proces block tx
		block, err := cdb.getBlock(elem.Hash)
		if err != nil {
			return err
		}
		if err := cs.processTxsAndState(dbtx, block); err != nil {
			return err
		}

		if cdb.latest+1 != elem.BlockNo {
			return fmt.Errorf("roll forward failed. invalid latest no(%d), block(%d, %v)",
				cdb.latest, elem.BlockNo, block.ID())
		}

		cdb.setLatest(elem.BlockNo)
	}

	return nil
}
