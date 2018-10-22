package state

import (
	"github.com/aergoio/aergo-lib/db"
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
	receiptTx db.Transaction

	BpReward uint64 //final bp reward, increment when tx executes
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
func NewBlockState(states *StateDB, rTx db.Transaction) *BlockState {
	return &BlockState{
		StateDB:   *states,
		receiptTx: rTx,
	}
}

// ReceiptTx return bs.receiptTx.
func (bs *BlockState) ReceiptTx() db.Transaction {
	return bs.receiptTx
}

// CommitReceipt commit bs.receiptTx.
func (bs *BlockState) CommitReceipt() {
	if bs.receiptTx != nil {
		bs.receiptTx.Commit()
	}
}

// // GetAccount gets account state from blockState
// func (bs *BlockState) GetAccount(aid types.AccountID) (*types.State, bool) {
// 	state, ok := bs.accounts[aid]
// 	return state, ok
// }

// // GetAccountStates gets account states from blockState
// func (bs *BlockState) GetAccountStates() map[types.AccountID]*types.State {
// 	return bs.accounts
// }

// // PutAccount sets before and changed state to blockState
// func (bs *BlockState) PutAccount(aid types.AccountID, stateChanged *types.State) {
// 	bs.accounts[aid] = stateChanged
// }

// Commit writes statedb and mapping information about block hash and state root
func (bs *BlockState) Commit() error {
	if err := bs.StateDB.Commit(); err != nil {
		return err
	}
	return nil
}
