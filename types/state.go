package types

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"

	proto "github.com/golang/protobuf/proto"
)

type BlockKey Hash
type AccountKey Hash
type StateKey Hash
type TransactionKey Hash

var (
	EmptyBlockKey       = BlockKey{}
	EmptyAccountKey     = AccountKey{}
	EmptyStateKey       = StateKey{}
	EmptyTransactionKey = TransactionKey{}
)

func ToBlockKey(blockHash []byte) BlockKey {
	buf := BlockKey{}
	copy(buf[:], blockHash)
	return BlockKey(buf)
}
func (key BlockKey) String() string {
	return hex.EncodeToString(key[:])
}

func ToTransactionKey(txHash []byte) TransactionKey {
	buf := TransactionKey{}
	copy(buf[:], txHash)
	return TransactionKey(buf)
}
func (key TransactionKey) String() string {
	return hex.EncodeToString(key[:])
}

func ToAccountKey(account []byte) AccountKey {
	buf := sha256.Sum256(account)
	return AccountKey(buf)
}
func (key AccountKey) String() string {
	return hex.EncodeToString(key[:])
}
func ToStateKeyPb(state *State) StateKey {
	if state == nil {
		return EmptyStateKey
	}
	bytes, err := proto.Marshal(state)
	if err != nil {
		return EmptyStateKey
	}
	return ToStateKey(bytes)
}
func ToStateKey(state []byte) StateKey {
	buf := sha256.Sum256(state)
	return StateKey(buf)
}
func (key StateKey) String() string {
	return hex.EncodeToString(key[:])
}

func NewState(akey AccountKey) *State {
	return &State{
		Account: akey[:],
		Nonce:   0,
		Balance: 0,
	}
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
