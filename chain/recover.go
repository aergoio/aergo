package chain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
	"os"
	"runtime"
	"runtime/debug"
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
	if !bytes.Equal(best.GetHash(), marker.BrBestHash) {
		logger.Info().Str("best", best.ID()).Uint64("no", best.GetHeader().GetBlockNo()).Msg("best block doesn't changed in prev reorg")
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
		buf = buf + fmt.Sprintf("branch root=%s", enc.ToString(rm.BrStartHash))
	}
	if len(rm.BrTopHash) != 0 {
		buf = buf + fmt.Sprintf("branch top=%s", enc.ToString(rm.BrTopHash))
	}

	return buf
}
