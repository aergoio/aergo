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

func (v *AccountState) ID() []byte {
	if len(v.id) < types.AddressLength {
		v.id = types.AddressPadding(v.id)
	}
	return v.id
}

func (v *AccountState) AccountID() types.AccountID {
	return v.aid
}

func (v *AccountState) State() *types.State {
	return v.newV
}

func (v *AccountState) SetNonce(nonce uint64) {
	v.newV.Nonce = nonce
}

func (v *AccountState) Balance() *big.Int {
	return new(big.Int).SetBytes(v.newV.Balance)
}

func (v *AccountState) AddBalance(amount *big.Int) {
	balance := new(big.Int).SetBytes(v.newV.Balance)
	v.newV.Balance = new(big.Int).Add(balance, amount).Bytes()
}

func (v *AccountState) SubBalance(amount *big.Int) {
	balance := new(big.Int).SetBytes(v.newV.Balance)
	v.newV.Balance = new(big.Int).Sub(balance, amount).Bytes()
}

func (v *AccountState) RP() uint64 {
	return v.newV.SqlRecoveryPoint
}

func (v *AccountState) IsNew() bool {
	return v.newOne
}

func (v *AccountState) IsContract() bool {
	return len(v.State().CodeHash) > 0
}

func (v *AccountState) IsDeploy() bool {
	return v.deploy&deployFlag != 0
}

func (v *AccountState) SetRedeploy() {
	v.deploy = deployFlag | redeployFlag
}

func (v *AccountState) IsRedeploy() bool {
	return v.deploy&redeployFlag != 0
}

func (v *AccountState) Reset() {
	v.newV = v.oldV.Clone()
}

func (v *AccountState) PutState() error {
	return v.sdb.PutState(v.aid, v.newV)
}

func (v *AccountState) ClearAid() {
	v.aid = emptyAccountID
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
