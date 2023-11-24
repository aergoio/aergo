package account

import (
	"math/big"

	lua_state "github.com/aergoio/aergo/v2/state"
	lua_types "github.com/aergoio/aergo/v2/types"
	eth_common "github.com/ethereum/go-ethereum/common"
	eth_state "github.com/ethereum/go-ethereum/core/state"
	eth_types "github.com/ethereum/go-ethereum/core/types"
)

// address : compressed public key (33 bytes)
func NewAccount(addr []byte, nonce uint64,
	luaDB *lua_state.StateDB, luaState *lua_types.State,
	ethDB *eth_state.StateDB, ethState *eth_types.StateAccount,
) *Account {
	acc := &Account{}
	acc.OldBalance = new(big.Int).SetBytes(luaState.Balance)
	acc.Nonce = nonce

	acc.LuaAccount = lua_types.ToAccountID(addr)
	acc.LuaDB = luaDB
	acc.LuaState = luaState

	// FIXME : is compressed key can encode to common.Address?
	// 안되면 uncompressed pubkey 로 conv 후 address 제작
	acc.EthAccount = eth_common.BytesToAddress(addr)
	acc.EthState = ethState
	acc.EthDB = ethDB

	return acc
}

type Account struct {
	// common
	OldBalance *big.Int
	NewBalance *big.Int
	Nonce      uint64

	// lua
	LuaAccount lua_types.AccountID
	LuaDB      *lua_state.StateDB
	LuaState   *lua_types.State

	// ethereum
	EthAccount eth_common.Address
	EthState   *eth_types.StateAccount
	EthDB      *eth_state.StateDB
}

func (acc *Account) SetNonce(nonce uint64) {
	acc.Nonce = nonce
}

func (acc *Account) GetNonce() uint64 {
	return acc.Nonce
}

func (acc *Account) SetBalance(balance *big.Int) {
	acc.NewBalance = new(big.Int).Set(balance)
}

func (acc *Account) GetBalanceOld() *big.Int {
	return acc.OldBalance
}

func (acc *Account) GetBalanceNew() *big.Int {
	return acc.NewBalance
}

//-------------------------------------------------------------------------//
// commit

func (acc *Account) Commit() {
	acc.OldBalance = new(big.Int).Set(acc.NewBalance)
	acc.Nonce++
}

// eth tx 를 통해 변경된 balance 를 lua 에 동일하게 적용
func (acc *Account) CommitLua() error {
	acc.LuaState.Balance = acc.OldBalance.Bytes()
	acc.LuaState.Nonce = acc.Nonce
	err := acc.LuaDB.PutState(acc.LuaAccount, acc.LuaState)
	if err != nil {
		return err
	}
	return nil
}

// lua tx 를 통해 변경된 balance 를 eth 에 동일하게 적용
func (acc *Account) CommitEth() error {
	acc.EthState.Balance = acc.OldBalance
	acc.EthState.Nonce = acc.Nonce
	acc.EthDB.SetBalance(acc.EthAccount, acc.OldBalance)
	acc.EthDB.SetNonce(acc.EthAccount, acc.Nonce)
	return nil
}
