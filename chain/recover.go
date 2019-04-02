package chain

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
	"os"
	"runtime"
	"runtime/debug"
)

var (
	ErrInvalidPrevHash = errors.New("no of previous hash block is invalid")
	ErrRecoInvalidBest = errors.New("best block is not equal to old chain")
)

func RecoverExit() {
	if r := recover(); r != nil {
		logger.Error().Str("callstack", string(debug.Stack())).Msg("panic occurred in chain manager")
		os.Exit(10)
	}
}

// Recover has 2 situation
// 1. normal recovery
//    normal recovery recovers error that has occures while adding single block
// 2. reorg recovery
//    reorg recovery recovers error that has occures while executing reorg
func (cs *ChainService) Recover() error {
	defer RecoverExit()

	// check if reorg marker exists
	marker, err := cs.cdb.getReorgMarker()
	if err != nil {
		return err
	}

	if marker == nil {
		// normal recover
		// TODO check state root maker of bestblock
		if err := cs.recoverNormal(); err != nil {
			return err
		}
		return nil
	}

	logger.Info().Str("reorg marker", marker.toString()).Msg("chain recovery started")

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	best, err := cs.GetBestBlock()
	if err != nil {
		return err
	}

	// check status of chain
	if !bytes.Equal(best.BlockHash(), marker.BrBestHash) {
		logger.Error().Msg("best block is not equal to old chain")
		return ErrRecoInvalidBest
	}

	if err = cs.recoverReorg(marker); err != nil {
		return err
	}

	cs.recovered = true

	return nil
}

// recover from normal
// set stateroot for bestblock
// TODO panic , shutdown server force.
// when panic occured, memory state of server may not be consistent.
// so restart server when panic in chainservice
func (cs *ChainService) recoverNormal() error {
	best, err := cs.GetBestBlock()
	if err != nil {
		return err
	}

	stateDB := cs.sdb.GetStateDB()
	if !stateDB.HasMarker(best.GetHeader().GetBlocksRootHash()) {
		return ErrRecoNoBestStateRoot
	}

	if !bytes.Equal(cs.sdb.GetStateDB().GetRoot(), best.GetHeader().GetBlocksRootHash()) {
		logger.Info().Str("besthash", best.ID()).Uint64("no", best.GetHeader().GetBlockNo()).
			Str("sroot", enc.ToString(best.GetHeader().GetBlocksRootHash())).Msg("set root of stateDB force for crash recovery")
		if err := stateDB.SetRoot(best.GetHeader().GetBlocksRootHash()); err != nil {
			return err
		}
	}

	return nil
}

// recoverReorg redo task that need to be performed after swapping chain meta
// 1. delete receipts of rollbacked blocks
// 2. swap tx mapping
func (cs *ChainService) recoverReorg(marker *ReorgMarker) error {
	// build reorgnizer from reorg marker
	topBlock, err := cs.GetBlock(marker.BrTopHash)
	if err != nil {
		return err
	}

	if err = cs.reorg(topBlock, marker); err != nil {
		logger.Error().Err(err).Msg("failed to retry reorg")
		return err
	}

	logger.Info().Msg("recovery succeeded")
	return nil
}

type ReorgMarker struct {
	BrStartHash []byte
	BrStartNo   types.BlockNo
	BrBestHash  []byte
	BrBestNo    types.BlockNo
	BrTopHash   []byte
	BrTopNo     types.BlockNo
}

func NewReorgMarker(reorg *reorganizer) *ReorgMarker {
	return &ReorgMarker{
		BrStartHash: reorg.brStartBlock.BlockHash(),
		BrStartNo:   reorg.brStartBlock.GetHeader().GetBlockNo(),
		BrBestHash:  reorg.bestBlock.BlockHash(),
		BrBestNo:    reorg.bestBlock.GetHeader().GetBlockNo(),
		BrTopHash:   reorg.brTopBlock.BlockHash(),
		BrTopNo:     reorg.brTopBlock.GetHeader().GetBlockNo(),
	}
}

func (rm *ReorgMarker) toBytes() ([]byte, error) {
	var val bytes.Buffer
	gob := gob.NewEncoder(&val)
	if err := gob.Encode(rm); err != nil {
		return nil, err
	}

	return val.Bytes(), nil
}

func (rm *ReorgMarker) toString() string {
	buf := ""

	if len(rm.BrStartHash) != 0 {
		buf = buf + fmt.Sprintf("branch root=(%d,%s)", rm.BrStartNo, enc.ToString(rm.BrStartHash))
	}
	if len(rm.BrTopHash) != 0 {
		buf = buf + fmt.Sprintf("branch top=(%d,%s)", rm.BrTopNo, enc.ToString(rm.BrTopHash))
	}
	if len(rm.BrBestHash) != 0 {
		buf = buf + fmt.Sprintf("org best=(%d,%s)", rm.BrTopNo, enc.ToString(rm.BrTopHash))
	}

	return buf
}

// RecoverChainMapping rollback chain (no/hash) mapping to old chain of reorg.
// it is required for LIB loading
func (rm *ReorgMarker) RecoverChainMapping(cdb *ChainDB) error {
	best, err := cdb.GetBestBlock()
	if err != nil {
		return err
	}

	if bytes.Equal(best.BlockHash(), rm.BrBestHash) {
		return nil
	}

	logger.Info().Str("marker", rm.toString()).Str("curbest", best.ID()).Uint64("curbestno", best.GetHeader().GetBlockNo()).Msg("start to recover chain mapping")

	bestBlock, err := cdb.getBlock(rm.BrBestHash)
	if err != nil {
		return err
	}

	bulk := cdb.store.NewBulk()
	defer bulk.DiscardLast()

	var tmpBlkNo types.BlockNo
	var tmpBlk *types.Block

	// remove unnecessary chain mapping of new chain
	for tmpBlkNo = rm.BrTopNo; tmpBlkNo > rm.BrBestNo; tmpBlkNo-- {
		logger.Debug().Uint64("no", tmpBlkNo).Msg("delete chain mapping of new chain")
		bulk.Delete(types.BlockNoToBytes(tmpBlkNo))
	}

	tmpBlk = bestBlock
	tmpBlkNo = tmpBlk.GetHeader().GetBlockNo()

	for tmpBlkNo > rm.BrStartNo {
		logger.Debug().Str("hash", tmpBlk.ID()).Uint64("no", tmpBlk.GetHeader().GetBlockNo()).Msg("update chain mapping to old chain")

		bulk.Set(types.BlockNoToBytes(tmpBlkNo), tmpBlk.BlockHash())

		if tmpBlk, err = cdb.getBlock(tmpBlk.GetHeader().GetPrevBlockHash()); err != nil {
			return err
		}

		if tmpBlkNo != tmpBlk.GetHeader().GetBlockNo()+1 {
			return ErrInvalidPrevHash
		}
		tmpBlkNo = tmpBlk.GetHeader().GetBlockNo()
	}

	logger.Info().Uint64("bestno", rm.BrBestNo).Msg("update best block")

	bulk.Set(latestKey, types.BlockNoToBytes(rm.BrBestNo))
	bulk.Flush()

	cdb.setLatest(bestBlock)

	logger.Info().Msg("succeed to recover chain mapping")
	return nil
}
