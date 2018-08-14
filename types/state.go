package types

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"reflect"

	proto "github.com/golang/protobuf/proto"
)

// HashID is a fixed size bytes
type HashID [sha256.Size]byte

// BlockID is a HashID to identify a block
type BlockID HashID

// AccountID is a HashID to identify an account
type AccountID HashID

// StateID is a HashID to identify a state
type StateID HashID

// TransactionID is a HashID to identify a transactions
type TransactionID HashID

func ToHashID(hash []byte) HashID {
	buf := HashID{}
	copy(buf[:], hash)
	return HashID(buf)
}
func (id HashID) String() string {
	return base64.StdEncoding.EncodeToString(id[:])
}

func ToBlockID(blockHash []byte) BlockID {
	return BlockID(ToHashID(blockHash))
}
func (id BlockID) String() string {
	return HashID(id).String()
}

func ToTransactionID(txHash []byte) TransactionID {
	return TransactionID(ToHashID(txHash))
}
func (id TransactionID) String() string {
	return HashID(id).String()
}

func ToAccountID(account []byte) AccountID {
	return AccountID(sha256.Sum256(account))
}
func (id AccountID) String() string {
	return HashID(id).String()
}

func ToStateIDPb(state *State) StateID {
	if state == nil {
		return StateID{}
	}
	bytes, err := proto.Marshal(state)
	if err != nil {
		return StateID{}
	}
	return ToStateID(bytes)
}
func ToStateID(state []byte) StateID {
	return StateID(sha256.Sum256(state))
}
func (id StateID) String() string {
	return HashID(id).String()
}

func NewState(id AccountID) *State {
	return &State{
		Account: id[:],
		Nonce:   0,
		Balance: 0,
	}
}

func (st *State) IsEmpty() bool {
	return st.Nonce == 0 && st.Balance == 0
}

func (st *State) GetHash() []byte {
	digest := sha256.New()
	digest.Write(st.Account)
	binary.Write(digest, binary.LittleEndian, st.Nonce)
	binary.Write(digest, binary.LittleEndian, st.Balance)
	return digest.Sum(nil)
}

func (st *State) Clone() *State {
	if st == nil {
		return nil
	}
	return &State{
		Account: Clone(st.Account).([]byte),
		Nonce:   st.Nonce,
		Balance: st.Balance,
	}
}

func Clone(i interface{}) interface{} {
	if i == nil {
		return nil
	}
	return reflect.Indirect(reflect.ValueOf(i)).Interface()
}
