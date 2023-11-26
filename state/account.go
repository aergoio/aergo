package state

import (
	"fmt"
	"math/big"

	"github.com/aergoio/aergo/v2/types"
)

type AccountState struct {
	sdb    *StateDB
	id     []byte
	aid    types.AccountID
	oldV   *types.State
	newV   *types.State
	newOne bool
	deploy int8
}

const (
	deployFlag = 0x01 << iota
	redeployFlag
)

func (as *AccountState) ID() []byte {
	if len(as.id) < types.AddressLength {
		as.id = types.AddressPadding(as.id)
	}
	return as.id
}

func (as *AccountState) AccountID() types.AccountID {
	return as.aid
}

func (as *AccountState) State() *types.State {
	return as.newV
}

func (as *AccountState) SetNonce(nonce uint64) {
	as.newV.Nonce = nonce
}

func (as *AccountState) Balance() *big.Int {
	return new(big.Int).SetBytes(as.newV.Balance)
}

func (as *AccountState) AddBalance(amount *big.Int) {
	balance := new(big.Int).SetBytes(as.newV.Balance)
	as.newV.Balance = new(big.Int).Add(balance, amount).Bytes()
}

func (as *AccountState) SubBalance(amount *big.Int) {
	balance := new(big.Int).SetBytes(as.newV.Balance)
	as.newV.Balance = new(big.Int).Sub(balance, amount).Bytes()
}

func (as *AccountState) RP() uint64 {
	return as.newV.SqlRecoveryPoint
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
	as.newV = as.oldV.Clone()
}

func (as *AccountState) PutState() error {
	return as.sdb.PutState(as.aid, as.newV)
}

func (as *AccountState) ClearAid() {
	as.aid = emptyAccountID
}

//----------------------------------------------------------------------------------------------//
//

func CreateAccountState(id []byte, sdb *StateDB) (*AccountState, error) {
	v, err := GetAccountState(id, sdb)
	if err != nil {
		return nil, err
	}
	if !v.newOne {
		return nil, fmt.Errorf("account(%s) aleardy exists", types.EncodeAddress(v.ID()))
	}
	v.newV.SqlRecoveryPoint = 1
	v.deploy = deployFlag
	return v, nil
}

func GetAccountState(id []byte, states *StateDB) (*AccountState, error) {
	aid := types.ToAccountID(id)
	st, err := states.GetState(aid)
	if err != nil {
		return nil, err
	}
	if st == nil {
		if states.testmode {
			amount := new(big.Int).Add(types.StakingMinimum, types.StakingMinimum)
			return &AccountState{
				sdb:    states,
				id:     id,
				aid:    aid,
				oldV:   &types.State{Balance: amount.Bytes()},
				newV:   &types.State{Balance: amount.Bytes()},
				newOne: true,
			}, nil
		}
		return &AccountState{
			sdb:    states,
			id:     id,
			aid:    aid,
			oldV:   &types.State{},
			newV:   &types.State{},
			newOne: true,
		}, nil
	}
	return &AccountState{
		sdb:  states,
		id:   id,
		aid:  aid,
		oldV: st,
		newV: st.Clone(),
	}, nil
}

func InitAccountState(id []byte, sdb *StateDB, old *types.State, new *types.State) *AccountState {
	return &AccountState{
		sdb:  sdb,
		id:   id,
		aid:  types.ToAccountID(id),
		oldV: old,
		newV: new,
	}
}
