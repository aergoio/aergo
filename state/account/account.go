package account

import (
	"math/big"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethstate "github.com/ethereum/go-ethereum/core/state"
)

// address : compressed public key (33 bytes)
func NewAccount(addr []byte, luaState *state.StateDB, ethState *ethstate.StateDB) (*Account, error) {
	var err error

	acc := &Account{}
	acc.LuaAccount = types.ToAccountID(addr)
	acc.EthAccount = ethcommon.BytesToAddress(addr) // FIXME : is compressed key can encode to common.Address? check compressed pubkey

	if acc.LuaAccState, err = luaState.GetAccountState(acc.LuaAccount); err != nil {
		return nil, err
	}

	if b := acc.LuaAccState.GetBalanceBigInt(); b.Cmp(ethState.GetBalance(acc.EthAccount)) != 0 {
		panic("impossible") // FIXME : handle exception
	} else {
		acc.Balance = b
	}

	if n := acc.LuaAccState.GetNonce(); n != ethState.GetNonce(acc.EthAccount) {
		panic("impossible") // FIXME : handle exception
	} else {
		acc.Nonce = n
	}

	return acc, nil
}

type Account struct {
	// common
	Balance *big.Int
	Nonce   uint64

	// lua
	LuaAccount  types.AccountID
	LuaState    *state.StateDB
	LuaAccState *types.State

	// ethereum
	EthAccount ethcommon.Address
	EthState   *ethstate.StateDB
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
