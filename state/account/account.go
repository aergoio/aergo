package account

import (
	"math/big"

	lua_state "github.com/aergoio/aergo/v2/state"
	lua_types "github.com/aergoio/aergo/v2/types"
	eth_common "github.com/ethereum/go-ethereum/common"
	eth_state "github.com/ethereum/go-ethereum/core/state"
)

// address : compressed public key (33 bytes)
func NewAccount(addr []byte, luaState *lua_state.StateDB, ethState *eth_state.StateDB) (*Account, error) {
	var err error

	acc := &Account{}
	acc.LuaAccount = lua_types.ToAccountID(addr)
	acc.EthAccount = eth_common.BytesToAddress(addr) // FIXME : is compressed key can encode to common.Address? check compressed pubkey

	if acc.LuaAccState, err = luaState.GetAccountState(acc.LuaAccount); err != nil {
		return nil, err
	}

	if balance := acc.LuaAccState.GetBalanceBigInt(); balance.Cmp(ethState.GetBalance(acc.EthAccount)) != 0 {
		panic("impossible") // FIXME : handle exception
	} else {
		acc.Balance = balance
	}

	if nonce := acc.LuaAccState.GetNonce(); nonce != ethState.GetNonce(acc.EthAccount) {
		panic("impossible") // FIXME : handle exception
	} else {
		acc.Nonce = nonce
	}

	return acc, nil
}

type Account struct {
	// common
	Balance *big.Int
	Nonce   uint64

	// lua
	LuaAccount  lua_types.AccountID
	LuaState    *lua_state.StateDB
	LuaAccState *lua_types.State

	// ethereum
	EthAccount eth_common.Address
	EthState   *eth_state.StateDB
}

func (acc *Account) SetNonce(nonce uint64) {
	acc.Nonce = nonce
}

func (acc *Account) GetNonce() uint64 {
	return acc.Nonce
}

func (acc *Account) SetBalance(balance *big.Int) {
	acc.Balance = new(big.Int).Set(balance)
}

func (acc *Account) AddBalance(balance *big.Int) {
	acc.Balance.Add(acc.Balance, balance)
}

func (acc *Account) SubBalance(balance *big.Int) {
	acc.Balance.Sub(acc.Balance, balance)
}

func (acc *Account) GetBalance() *big.Int {
	return acc.Balance
}

//-------------------------------------------------------------------------//
// commit

func (acc *Account) CommitLua() error {
	acc.LuaAccState.Balance = acc.Balance.Bytes()
	acc.LuaAccState.Nonce = acc.Nonce
	err := acc.LuaState.PutState(acc.LuaAccount, acc.LuaAccState)
	if err != nil {
		return err
	}
	return nil
}

func (acc *Account) CommitEth() error {
	acc.EthState.SetBalance(acc.EthAccount, acc.Balance)
	acc.EthState.SetNonce(acc.EthAccount, acc.Nonce)
	return nil
}
