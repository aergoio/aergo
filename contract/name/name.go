package name

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

var nameTable map[string]*Owner
var prefix = []byte("name")

type Owner struct {
	Address []byte
}

func CreateName(scs *state.ContractState, tx *types.TxBody, sender, receiver *state.V) error {
	if err := ValidateNameTx(tx, scs); err != nil {
		return err
	}
	if len(tx.Payload[1:]) != types.NameLength {
		return fmt.Errorf("not supported yet")
	}
	amount := tx.GetAmountBigInt()
	if sender.Balance().Cmp(tx.GetAmountBigInt()) < 0 {
		return types.ErrInsufficientBalance
	}
	sender.SubBalance(amount)
	receiver.AddBalance(amount)

	return createName(scs, tx.Payload[1:], tx.Account)
}

func createName(scs *state.ContractState, name []byte, owner []byte) error {
	return setAddress(scs, name, owner)
}

//UpdateName is avaliable after bid implement
func UpdateName(scs *state.ContractState, tx *types.TxBody, sender, receiver *state.V) error {
	if err := ValidateNameTx(tx, scs); err != nil {
		return err
	}
	name, to := parseUpdatePayload(tx.Payload[1:])
	if len(getAddress(scs, name)) <= types.NameLength {
		return fmt.Errorf("%s is not created yet", string(name))
	}
	amount := tx.GetAmountBigInt()
	if sender.Balance().Cmp(tx.GetAmountBigInt()) < 0 {
		return types.ErrInsufficientBalance
	}
	sender.SubBalance(amount)
	receiver.AddBalance(amount)
	return updateName(scs, name, tx.Account, to)
}

func updateName(scs *state.ContractState, name []byte, from []byte, to []byte) error {
	return setAddress(scs, name, to)
}

//Resolve is resolve name for chain
func Resolve(bs *state.BlockState, name []byte) []byte {
	if len(name) != types.NameLength || bytes.Equal(name, []byte(types.AergoSystem)) {
		return name
	}
	scs, err := openContract(bs)
	if err != nil {
		return name
	}
	defer bs.StateDB.StageContractState(scs)
	return getAddress(scs, name)
}

func openContract(bs *state.BlockState) (*state.ContractState, error) {
	v, err := bs.GetAccountStateV([]byte("aergo.name"))
	if err != nil {
		return nil, err
	}
	scs, err := bs.StateDB.OpenContractState(v.AccountID(), v.State())
	if err != nil {
		return nil, err
	}
	return scs, nil
}

//GetAddress is resolve name for mempool
func GetAddress(scs *state.ContractState, name []byte) []byte {
	if len(name) != types.NameLength || bytes.Equal(name, []byte(types.AergoSystem)) {
		return name
	}
	return getAddress(scs, name)
}

func getAddress(scs *state.ContractState, name []byte) []byte {
	owner := GetOwner(scs, name)
	if owner != nil {
		if len(owner.Address) > types.NameLength {
			return owner.Address
		}
		return getAddress(scs, owner.Address)
	}
	return nil
}

func GetOwner(scs *state.ContractState, name []byte) *Owner {
	return getOwner(scs, name, true)
}

func getOwner(scs *state.ContractState, name []byte, useInitial bool) *Owner {
	lowerCaseName := strings.ToLower(string(name))
	key := append(prefix, lowerCaseName...)
	var err error
	var ownerdata []byte
	if useInitial {
		ownerdata, err = scs.GetInitialData(key)
	} else {
		ownerdata, err = scs.GetData(key)
	}
	if err != nil {
		return nil
	}
	return deserializeOwner(ownerdata)
}

func setAddress(scs *state.ContractState, name []byte, address []byte) error {
	owner := &Owner{Address: address}
	return setOwner(scs, name, owner)
}

func setOwner(scs *state.ContractState, name []byte, owner *Owner) error {
	lowerCaseName := strings.ToLower(string(name))
	key := append(prefix, lowerCaseName...)
	return scs.SetData(key, serializeOwner(owner))
}

func serializeOwner(owner *Owner) []byte {
	if owner != nil {
		return owner.Address
	}
	return nil
}

func deserializeOwner(data []byte) *Owner {
	if data != nil {
		return &Owner{Address: data}
	}
	return nil
}
