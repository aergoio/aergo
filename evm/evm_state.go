package evm

import (
	"math/big"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/ethdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

var _ vm.StateDB = (*StateDB)(nil)

type StateDB struct {
	bs *state.BlockState
	*ethdb.StateDB
}

func NewStateDB(bs *state.BlockState) *StateDB {
	return &StateDB{
		bs:      bs,
		StateDB: bs.EthStateDB,
	}
}

func (s *StateDB) CreateAccount(addr common.Address) {
	acc, err := state.GetAccountStateEth(addr, s.bs)
	if err != nil {
		panic(err)
	}
	err = acc.PutState()
	if err != nil {
		panic(err)
	}
}

func (s *StateDB) SubBalance(addr common.Address, amount *big.Int) {
	acc, err := state.GetAccountStateEth(addr, s.bs)
	if err != nil {
		panic(err)
	}
	acc.SubBalance(amount)
	err = acc.PutState()
	if err != nil {
		panic(err)
	}
}

func (s *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	acc, err := state.GetAccountStateEth(addr, s.bs)
	if err != nil {
		panic(err)
	}
	acc.AddBalance(amount)
	err = acc.PutState()
	if err != nil {
		panic(err)
	}
}

func (s *StateDB) SetNonce(addr common.Address, nonce uint64) {
	acc, err := state.GetAccountStateEth(addr, s.bs)
	if err != nil {
		panic(err)
	}
	acc.SetNonce(nonce)
	err = acc.PutState()
	if err != nil {
		panic(err)
	}
}
func (s *StateDB) SetCode(addr common.Address, code []byte) {
	s.StateDB.SetCode(addr, code)
	acc, err := state.GetAccountStateEth(addr, s.bs)
	if err != nil {
		panic(err)
	}
	err = acc.PutState()
	if err != nil {
		panic(err)
	}
}
