package types

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"

	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc"
)

const (
	HashIDLength = 32
)

// HashID is a fixed size bytes
type HashID [HashIDLength]byte

// BlockID is a HashID to identify a block
type BlockID HashID

// AccountID is a HashID to identify an account
type AccountID HashID

// TxID is a HashID to identify a transaction
type TxID HashID

// ImplHashID is a object has HashID
type ImplHashID interface {
	HashID() HashID
}

// ImplHashBytes is a object supports Hash
type ImplHashBytes interface {
	Hash() []byte
}

// ImplMarshal is a object has marshal interface
type ImplMarshal interface {
	Marshal() ([]byte, error)
}

var (
	emptyHashID = HashID{}
)

// GetHashID make a HashID from hash of bytes
func GetHashID(bytes ...[]byte) HashID {
	hash := common.Hasher(bytes...)
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

// Bytes make a byte slice from id
func (id HashID) Bytes() []byte {
	if id == emptyHashID {
		return nil
	}
	return id[:]
}

// Compare returns an integer comparing two HashIDs as byte slices.
func (id HashID) Compare(alt HashID) int {
	return bytes.Compare(id.Bytes(), alt.Bytes())
}

// Equal returns a boolean comparing two HashIDs as byte slices.
func (id HashID) Equal(alt HashID) bool {
	return bytes.Equal(id.Bytes(), alt.Bytes())
}

// ParseToTxID parse BlockID from bytes. it return error if length of parameter is not fit.
func ParseToBlockID(blockHash []byte) (BlockID, error) {
	var hash BlockID
	if len(blockHash) != HashIDLength {
		return hash, fmt.Errorf("parse error: invalid length")
	}
	copy(hash[:], blockHash)
	return hash, nil
}

// MustParseBlockID parse, it panics if parsing is failed.
func MustParseBlockID(blockHash []byte) BlockID {
	hash, err := ParseToBlockID(blockHash)
	if err != nil {
		panic(err)
	}
	return hash
}

// ToBlockID make a BlockID from bytes
func ToBlockID(blockHash []byte) BlockID {
	return BlockID(ToHashID(blockHash))
}
func (id BlockID) String() string {
	return HashID(id).String()
}

// ParseToTxID parse TxID from bytes. it return error if length of parameter is not fit.
func ParseToTxID(txHash []byte) (TxID, error) {
	var hash TxID
	if len(txHash) != HashIDLength {
		return hash, fmt.Errorf("parse error: invalid length")
	}
	return ToTxID(txHash), nil
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
		Nonce:            0,
		Balance:          []byte{0},
		SqlRecoveryPoint: uint64(1),
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

func (st *State) GetBalanceBigInt() *big.Int {
	if st == nil {
		return new(big.Int).SetUint64(0)
	}
	return new(big.Int).SetBytes(st.Balance)
}
