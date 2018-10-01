package types

import (
	"bytes"

	"reflect"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/internal/enc"
)

// HashID is a fixed size bytes
type HashID [32]byte

// BlockID is a HashID to identify a block
type BlockID HashID

// AccountID is a HashID to identify an account
type AccountID HashID

// TxID is a HashID to identify a transaction
type TxID HashID

// GetHashID make a HashID from hash of bytes
func GetHashID(bytes []byte) HashID {
	hash := common.Hasher(bytes)
	return ToHashID(hash)
}

// ToHashID make a HashID from bytes
func ToHashID(hash []byte) HashID {
	buf := HashID{}
	copy(buf[:], hash)
	return HashID(buf)
}
func (id HashID) String() string {
	return enc.ToString(id[:])
}

// Compare returns an integer comparing two HashIDs as byte slices.
func (id HashID) Compare(alt HashID) int {
	return bytes.Compare(id[:], alt[:])
}

// Equal returns a boolean comparing two HashIDs as byte slices.
func (id HashID) Equal(alt HashID) bool {
	return bytes.Equal(id[:], alt[:])
}

// ToBlockID make a BlockID from bytes
func ToBlockID(blockHash []byte) BlockID {
	return BlockID(ToHashID(blockHash))
}
func (id BlockID) String() string {
	return HashID(id).String()
}

// ToTxID make a TxID from bytes
func ToTxID(txHash []byte) TxID {
	return TxID(ToHashID(txHash))
}
func (id TxID) String() string {
	return HashID(id).String()
}

// ToAccountID make a AccountHash from bytes
func ToAccountID(account []byte) AccountID {
	return AccountID(GetHashID(account))
}
func (id AccountID) String() string {
	return HashID(id).String()
}

// NewState returns an instance of account state
func NewState() *State {
	return &State{
		Nonce:   0,
		Balance: 0,
	}
}

// func (st *State) IsEmpty() bool {
// 	return st.Nonce == 0 && st.Balance == 0
// }

// func (st *State) GetHash() []byte {
// 	digest := sha256.New()
// 	binary.Write(digest, binary.LittleEndian, st.Nonce)
// 	binary.Write(digest, binary.LittleEndian, st.Balance)
// 	return digest.Sum(nil)
// }

// func (st *State) Clone() *State {
// 	if st == nil {
// 		return nil
// 	}
// 	return &State{
// 		Nonce:       st.Nonce,
// 		Balance:     st.Balance,
// 		CodeHash:    st.CodeHash,
// 		StorageRoot: st.StorageRoot,
// 	}
// }

func Clone(i interface{}) interface{} {
	if i == nil {
		return nil
	}
	return reflect.Indirect(reflect.ValueOf(i)).Interface()
}

type BlockInfo struct {
	BlockNo   BlockNo
	BlockHash BlockID
	PrevHash  BlockID
	StateRoot HashID
}
type BlockState struct {
	BlockInfo
	accounts  map[AccountID]*State
	Undo      undoStates
	receiptTx db.Transaction
}
type undoStates struct {
	StateRoot HashID
	Accounts  map[AccountID]*State
}

// NewBlockInfo create new blockInfo contains blockNo, blockHash and blockHash of previous block
func NewBlockInfo(blockNo BlockNo, blockHash, prevHash BlockID) *BlockInfo {
	return &BlockInfo{
		BlockNo:   blockNo,
		BlockHash: blockHash,
		PrevHash:  prevHash,
	}
}

// NewBlockState create new blockState contains blockInfo, account states and undo states
func NewBlockState(blockInfo *BlockInfo, rTx db.Transaction) *BlockState {
	return &BlockState{
		BlockInfo: *blockInfo,
		accounts:  make(map[AccountID]*State),
		Undo: undoStates{
			Accounts: make(map[AccountID]*State),
		},
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

// GetAccount gets account state from blockState
func (bs *BlockState) GetAccount(aid AccountID) (*State, bool) {
	state, ok := bs.accounts[aid]
	return state, ok
}

// GetAccountStates gets account states from blockState
func (bs *BlockState) GetAccountStates() map[AccountID]*State {
	return bs.accounts
}

// PutAccount sets before and changed state to blockState
func (bs *BlockState) PutAccount(aid AccountID, stateBefore, stateChanged *State) {
	if _, ok := bs.Undo.Accounts[aid]; !ok {
		bs.Undo.Accounts[aid] = stateBefore
	}
	bs.accounts[aid] = stateChanged
}

// SetBlockHash sets bs.BlockInfo.BlockHash to blockHash
func (bs *BlockState) SetBlockHash(blockHash BlockID) {
	if bs == nil {
		return
	}
	bs.BlockInfo.BlockHash = blockHash
}
