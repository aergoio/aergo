package state

import (
	"fmt"
	"math/big"

	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
)

type AccountState struct {
	sdb      *statedb.StateDB
	id       []byte
	aid      types.AccountID
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

func (as *AccountState) State() *types.State {
	return as.newState
}

func (as *AccountState) SetNonce(nonce uint64) {
	as.newState.Nonce = nonce
}

func (as *AccountState) Nonce() uint64 {
	if as.newState == nil {
		return 0
	}
	return as.newState.Nonce
}

func (as *AccountState) Balance() *big.Int {
	if as.newState == nil {
		return big.NewInt(0)
	}
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

func (as *AccountState) SetRP() uint64 {
	return as.newState.SqlRecoveryPoint
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
	return as.sdb.PutState(as.aid, as.newState)
}

func (as *AccountState) ClearAid() {
	as.aid = statedb.EmptyAccountID
}

//----------------------------------------------------------------------------------------------//
// global functions

func CreateAccountState(id []byte, sdb *statedb.StateDB) (*AccountState, error) {
	v, err := GetAccountState(id, sdb)
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

func GetAccountState(id []byte, states *statedb.StateDB) (*AccountState, error) {
	aid := types.ToAccountID(id)
	st, err := states.GetState(aid)
	if err != nil {
		return nil, err
	}
	if st == nil {
		if states.Testmode {
			amount := new(big.Int).Add(types.StakingMinimum, types.StakingMinimum)
			return &AccountState{
				sdb:      states,
				id:       id,
				aid:      aid,
				oldState: &types.State{Balance: amount.Bytes()},
				newState: &types.State{Balance: amount.Bytes()},
				newOne:   true,
			}, nil
		}
		return &AccountState{
			sdb:      states,
			id:       id,
			aid:      aid,
			oldState: &types.State{},
			newState: &types.State{},
			newOne:   true,
		}, nil
	}
	return &AccountState{
		sdb:      states,
		id:       id,
		aid:      aid,
		oldState: st,
		newState: st.Clone(),
	}, nil
}

func InitAccountState(id []byte, sdb *statedb.StateDB, stOld, stNew *types.State) *AccountState {
	return &AccountState{
		sdb:      sdb,
		id:       id,
		aid:      types.ToAccountID(id),
		oldState: stOld,
		newState: stNew,
	}
}

func SendBalance(sender, receiver *AccountState, amount *big.Int) error {
	if sender == receiver {
		return nil
	}
	if sender.Balance().Cmp(amount) < 0 {
		return fmt.Errorf("insufficient balance")
	}
	sender.SubBalance(amount)
	receiver.AddBalance(amount)
	return nil
}
