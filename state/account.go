package state

import (
	"fmt"
	"math/big"

	"github.com/aergoio/aergo/v2/state/ethdb"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/ethereum/go-ethereum/common"
)

type AccountState struct {
	luaStates *statedb.StateDB
	ethStates *ethdb.StateDB

	id    []byte
	aid   types.AccountID
	ethId common.Address

	oldState *types.State
	newState *types.State
	newOne   bool
	deploy   int8
}

const (
	deployFlag = 0x01 << iota
	redeployFlag
)

func (as *AccountState) ID() []byte {
	if len(as.id) < types.AddressLength {
		return types.AddressPadding(as.id)
	}
	return as.IDNoPadding()
}

func (as *AccountState) IDNoPadding() []byte {
	return as.id
}

func (as *AccountState) AccountID() types.AccountID {
	return as.aid
}

func (as *AccountState) EthID() common.Address {
	return as.ethId
}

func (as *AccountState) State() *types.State {
	return as.newState
}

func (as *AccountState) SetNonce(nonce uint64) {
	as.newState.Nonce = nonce
}

func (as *AccountState) Nonce() uint64 {
	return as.newState.Nonce
}

func (as *AccountState) Balance() *big.Int {
	return new(big.Int).SetBytes(as.newState.Balance)
}

func (as *AccountState) AddBalance(amount *big.Int) {
	balance := new(big.Int).SetBytes(as.newState.Balance)
	as.newState.Balance = new(big.Int).Add(balance, amount).Bytes()
}

func (as *AccountState) SubBalance(amount *big.Int) {
	balance := new(big.Int).SetBytes(as.newState.Balance)
	as.newState.Balance = new(big.Int).Sub(balance, amount).Bytes()
}

func (as *AccountState) SetCodeHash(codeHash []byte) {
	as.newState.CodeHash = codeHash
}

func (as *AccountState) CodeHash() []byte {
	if as.newState == nil {
		return nil
	}
	return as.newState.CodeHash
}

func (as *AccountState) SetRP(sqlRecoveryPoint uint64) {
	as.newState.SqlRecoveryPoint = sqlRecoveryPoint
}

func (as *AccountState) RP() uint64 {
	if as.newState == nil {
		return 0
	}
	return as.newState.SqlRecoveryPoint
}

func (as *AccountState) SetStorageRoot(storageRoot []byte) {
	as.newState.StorageRoot = storageRoot
}

func (as *AccountState) StorageRoot() []byte {
	if as.newState == nil {
		return nil
	}
	return as.newState.StorageRoot
}

func (as *AccountState) IsNew() bool {
	return as.newOne
}

func (as *AccountState) IsContract() bool {
	return len(as.State().CodeHash) > 0
}

func (as *AccountState) IsDeploy() bool {
	return as.deploy&deployFlag != 0
}

func (as *AccountState) SetRedeploy() {
	as.deploy = deployFlag | redeployFlag
}

func (as *AccountState) IsRedeploy() bool {
	return as.deploy&redeployFlag != 0
}

func (as *AccountState) Reset() {
	as.newState = as.oldState.Clone()
}

func (as *AccountState) PutState() error {
	if len(as.id) > 0 {
		if err := as.luaStates.PutState(as.aid, as.newState); err != nil {
			return err
		}
	}
	if as.ethStates != nil {
		as.ethStates.PutId(as.ethId, as.id)
		as.ethStates.PutState(as.ethId, as.newState)
	}
	return nil
}

//----------------------------------------------------------------------------------------------//
// global functions

func CreateAccountState(id []byte, bs *BlockState) (*AccountState, error) {
	v, err := GetAccountState(id, bs)
	if err != nil {
		return nil, err
	}
	if !v.newOne {
		return nil, fmt.Errorf("account(%s) aleardy exists", types.EncodeAddress(v.ID()))
	}
	v.newState.SqlRecoveryPoint = 1
	v.deploy = deployFlag
	return v, nil
}

func GetAccountState(id []byte, bs *BlockState) (*AccountState, error) {
	var st *types.State
	var err error
	var newOne bool

	aid := types.ToAccountID(id)
	ethId := ethdb.GetAddressEth(id)

	st, err = bs.LuaStateDB.GetState(aid)
	if err != nil {
		return nil, err
	}
	if st == nil {
		newOne = true // new address
		if bs.LuaStateDB.Testmode {
			amount := new(big.Int).Add(types.StakingMinimum, types.StakingMinimum)
			st = &types.State{Balance: amount.Bytes()}
		} else if bs.EthStateDB != nil {
			// if not exist in lua state db, check eth state db
			st = bs.EthStateDB.GetState(ethId)
		}
		if st == nil {
			st = &types.State{}
		}
	}
	return &AccountState{
		luaStates: bs.LuaStateDB,
		ethStates: bs.EthStateDB,
		id:        id,
		aid:       aid,
		ethId:     ethId,
		oldState:  st,
		newState:  st.Clone(),
		newOne:    newOne,
	}, nil
}

func GetAccountStateEth(ethId common.Address, bs *BlockState) (*AccountState, error) {
	id := bs.EthStateDB.GetId(ethId)
	if len(id) > 0 {
		aid := types.ToAccountID(id)
		st, err := bs.LuaStateDB.GetState(aid)
		if err != nil {
			return nil, err
		}
		return &AccountState{
			luaStates: bs.LuaStateDB,
			ethStates: bs.EthStateDB,
			id:        id,
			aid:       aid,
			ethId:     ethId,
			oldState:  st,
			newState:  st.Clone(),
			newOne:    true,
		}, nil
	}
	st := bs.EthStateDB.GetState(ethId)
	if st == nil {
		st = &types.State{}
	}
	return &AccountState{
		luaStates: bs.LuaStateDB,
		ethStates: bs.EthStateDB,
		id:        nil,
		aid:       types.AccountID{},
		ethId:     ethId,
		oldState:  st,
		newState:  st.Clone(),
	}, nil
}

func InitAccountState(id []byte, states *statedb.StateDB, ethstates *ethdb.StateDB, stOld, stNew *types.State) *AccountState {
	return &AccountState{
		luaStates: states,
		ethStates: ethstates,
		id:        id,
		aid:       types.ToAccountID(id),
		oldState:  stOld,
		newState:  stNew,
	}
}

func SendBalance(sender, receiver *AccountState, amount *big.Int) error {
	if len(sender.id) > 0 && sender.AccountID() == receiver.AccountID() {
		return nil
	}
	if sender.Balance().Cmp(amount) < 0 {
		return types.ErrInsufficientBalance
	}
	sender.SubBalance(amount)
	receiver.AddBalance(amount)
	return nil
}
