package state

import (
	"github.com/aergoio/aergo/types"
)

// BlockInfo contains BlockHash and StateRoot
type BlockInfo struct {
	BlockHash types.BlockID
	StateRoot types.HashID
}

// BlockState contains BlockInfo and statedb for block
type BlockState struct {
	StateDB
	BpReward uint64 //final bp reward, increment when tx executes
	receipts types.Receipts
	CodeMap  map[string][]byte
}

// NewBlockInfo create new blockInfo contains blockNo, blockHash and blockHash of previous block
func NewBlockInfo(blockHash types.BlockID, stateRoot types.HashID) *BlockInfo {
	return &BlockInfo{
		BlockHash: blockHash,
		StateRoot: stateRoot,
	}
}

// GetStateRoot return bytes of bi.StateRoot
func (bi *BlockInfo) GetStateRoot() []byte {
	if bi == nil {
		return nil
	}
	return bi.StateRoot.Bytes()
}

// NewBlockState create new blockState contains blockInfo, account states and undo states
func NewBlockState(states *StateDB) *BlockState {
	return &BlockState{
		StateDB: *states,
		CodeMap: make(map[string][]byte),
	}
}

func (bs *BlockState) AddReceipt(r *types.Receipt) {
	bs.receipts = append(bs.receipts, r)
}

func (bs *BlockState) Receipts() types.Receipts {
	return bs.receipts
}
