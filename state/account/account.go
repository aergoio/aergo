package account

import (
	"math/big"

	key "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethstate "github.com/ethereum/go-ethereum/core/state"
)

// address : compressed public key (33 bytes)
func NewAccount(address []byte, luaState *statedb.StateDB, ethState *ethstate.StateDB) (*Account, error) {
	var err error

	acc := &Account{}
	acc.LuaAccount = types.EncodeAddress(address)
	acc.LuaAid = types.ToAccountID(address)
	if unCompressed := key.ConvAddressUncompressed(address); unCompressed != nil {
		acc.EthAccount = ethcommon.BytesToAddress(unCompressed)
	}

	if luaState == nil || ethState == nil { // FIXME : handle exception
		return acc, nil
	}

	if acc.LuaAccState, err = luaState.GetAccountState(acc.LuaAid); err != nil {
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
	LuaAccount  string
	LuaAid      types.AccountID
	LuaState    *statedb.StateDB
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
	err := acc.LuaState.PutState(acc.LuaAid, acc.LuaAccState)
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
