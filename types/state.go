package types

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"reflect"

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

// ToHashID make a HashID from bytes
func ToHashID(hash []byte) HashID {
	buf := HashID{}
	copy(buf[:], hash)
	return HashID(buf)
}
func (id HashID) String() string {
	return enc.ToString(id[:])
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
	accountHash := TrieHasher(account)
	return AccountID(ToHashID(accountHash))
}
func (id AccountID) String() string {
	return HashID(id).String()
}

// TrieHasher exports default hash function for trie
var TrieHasher = func(data ...[]byte) []byte {
	hasher := sha512.New512_256()
	for i := 0; i < len(data); i++ {
		hasher.Write(data[i])
	}
	return hasher.Sum(nil)
}

func NewState() *State {
	return &State{
		Nonce:   0,
		Balance: 0,
	}
}

func (st *State) IsEmpty() bool {
	return st.Nonce == 0 && st.Balance == 0
}

func (st *State) GetHash() []byte {
	digest := sha256.New()
	binary.Write(digest, binary.LittleEndian, st.Nonce)
	binary.Write(digest, binary.LittleEndian, st.Balance)
	return digest.Sum(nil)
}

// func (st *State) ToBytes() []byte {
// 	buf, _ := proto.Marshal(st)
// 	return buf
// }
// func (st *State) FromBytes(buf []byte) {
// 	if st == nil {
// 		st = &State{}
// 	}
// 	_ = proto.Unmarshal(buf, st)
// }

func (st *State) Clone() *State {
	if st == nil {
		return nil
	}
	return &State{
		Nonce:       st.Nonce,
		Balance:     st.Balance,
		CodeHash:    st.CodeHash,
		StorageRoot: st.StorageRoot,
	}
}

func Clone(i interface{}) interface{} {
	if i == nil {
		return nil
	}
	return reflect.Indirect(reflect.ValueOf(i)).Interface()
}
