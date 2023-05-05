package name

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/dbkey"
)

type NameMap struct {
	Version     byte
	Owner       []byte
	Destination []byte
	Operator    []byte
}

func CreateName(scs *statedb.ContractState, tx *types.TxBody, sender, receiver *state.AccountState, name string) error {
	amount := tx.GetAmountBigInt()
	if err := state.SendBalance(sender, receiver, amount); err != nil {
		return err
	}
	return createName(scs, []byte(name), sender.ID())
}

func createName(scs *statedb.ContractState, name []byte, owner []byte) error {
	//	return setAddress(scs, name, owner)
	return registerOwner(scs, name, owner, owner)
}

//UpdateOperator sets the operator for a given name
func UpdateOperator(bs *state.BlockState, scs *statedb.ContractState, tx *types.TxBody,
	sender, receiver *state.AccountState, name, operator string) error {
	if len(getAddress(scs, []byte(name))) <= types.NameLength {
		return fmt.Errorf("%s is not created yet", string(name))
	}
	var operatorAddr []byte
	if len(operator) > 0 {
		// convert the operator to bytes
		operatorAddr, _ = types.DecodeAddress(operator)
		// if it is a name, resolve it to an address
		operatorAddr = GetAddress(scs, operatorAddr)
	}
	return updateOperator(scs, []byte(name), operatorAddr)
}

func updateOperator(scs *statedb.ContractState, name []byte, operator []byte) error {
	nameMap := getNameMap(scs, name, true)
	nameMap.Operator = operator
	return setNameMap(scs, name, nameMap)
}

//UpdateName changes the destination and the owner
func UpdateName(bs *state.BlockState, scs *statedb.ContractState, tx *types.TxBody,
	sender, receiver *state.AccountState, name, to string) error {
	if len(getAddress(scs, []byte(name))) <= types.NameLength {
		return fmt.Errorf("%s is not created yet", string(name))
	}
	destination, _ := types.DecodeAddress(to)
	// if it is a name, resolve it to an address
	destination = GetAddress(scs, destination)

	amount := tx.GetAmountBigInt()
	if err := state.SendBalance(sender, receiver, amount); err != nil {
		return err
	}
	contract, err := statedb.OpenContractStateAccount(destination, bs.StateDB)
	if err != nil {
		return types.ErrTxInvalidRecipient
	}
	creator, err := contract.GetData(dbkey.CreatorMeta())
	if err != nil {
		return err
	}
	ownerAddr := destination
	if creator != nil {
		ownerAddr, err = types.DecodeAddress(string(creator))
		if err != nil {
			return types.ErrTxInvalidRecipient
		}
	}
	return updateName(scs, []byte(name), ownerAddr, destination)
}

//UpdateName does not save the operator (it is cleared on transfer)
func updateName(scs *statedb.ContractState, name []byte, owner []byte, to []byte) error {
	//return setAddress(scs, name, to)
	return registerOwner(scs, name, owner, to)
}

func isPredefined(name []byte, legacy bool) bool {
	if legacy {
		return len(name) == types.AddressLength || strings.Contains(string(name), ".")
	}
	return len(name) == types.AddressLength || types.IsSpecialAccount(name)
}

// Resolve is resolve name for chain
func Resolve(bs *state.BlockState, name []byte, legacy bool) ([]byte, error) {
	if isPredefined(name, legacy) {
		return name, nil
	}
	scs, err := openContract(bs)
	if err != nil {
		return nil, err
	}
	return getAddress(scs, name), nil
}

func openContract(bs *state.BlockState) (*statedb.ContractState, error) {
	v, err := state.GetAccountState([]byte(types.AergoName), bs.StateDB)
	if err != nil {
		return nil, err
	}
	scs, err := statedb.OpenContractState(v.ID(), v.State(), bs.StateDB)
	if err != nil {
		return nil, err
	}
	return scs, nil
}

// GetAddress is resolve name for mempool
func GetAddress(scs *statedb.ContractState, name []byte) []byte {
	if len(name) == types.AddressLength || types.IsSpecialAccount(name) {
		return name
	}
	return getAddress(scs, name)
}

// GetAddressLegacy is resolve name for mempool by buggy logic, leaved for backward compatibility
func GetAddressLegacy(scs *statedb.ContractState, name []byte) []byte {
	if len(name) == types.AddressLength || strings.Contains(string(name), ".") {
		return name
	}
	return getAddress(scs, name)
}

func getAddress(scs *statedb.ContractState, name []byte) []byte {
	nameMap := getNameMap(scs, name, true)
	if nameMap != nil {
		return nameMap.Destination
	}
	return nil
}

func GetOwner(scs *statedb.ContractState, name []byte) []byte {
	return getOwner(scs, name, true)
}

func getOwner(scs *statedb.ContractState, name []byte, useInitial bool) []byte {
	nameMap := getNameMap(scs, name, useInitial)
	if nameMap != nil {
		return nameMap.Owner
	}
	return nil
}

func GetOperator(scs *statedb.ContractState, name []byte) []byte {
	return getOperator(scs, name)
}

func getOperator(scs *statedb.ContractState, name []byte) []byte {
	nameMap := getNameMap(scs, name, true)
	if nameMap != nil {
		return nameMap.Operator
	}
	return nil
}

func getNameMap(scs *statedb.ContractState, name []byte, useInitial bool) *NameMap {
	var err error
	var ownerdata []byte
	if useInitial {
		ownerdata, err = scs.GetInitialData(dbkey.Name(name))
	} else {
		ownerdata, err = scs.GetData(dbkey.Name(name))
	}
	if err != nil {
		return nil
	}
	return deserializeNameMap(ownerdata)
}

func GetNameInfo(ncs *statedb.ContractState, name string) (*types.NameInfo, error) {
	owner := getOwner(ncs, []byte(name), true)
	return &types.NameInfo{Name: &types.Name{Name: string(name)}, Owner: owner, Destination: GetAddress(ncs, []byte(name))}, nil
}

func registerOwner(scs *statedb.ContractState, name, owner, destination []byte) error {
	nameMap := &NameMap{Version: 1, Owner: owner, Destination: destination}
	return setNameMap(scs, name, nameMap)
}

func setNameMap(scs *statedb.ContractState, name []byte, n *NameMap) error {
	return scs.SetData(dbkey.Name(name), serializeNameMap(n))
}

func serializeNameMap(n *NameMap) []byte {
	var ret []byte
	if n != nil {
		// store version
		ret = append(ret, n.Version)
		buf := make([]byte, 8)
		// store the size of owner
		binary.LittleEndian.PutUint64(buf, uint64(len(n.Owner)))
		ret = append(ret, buf...)
		// store the owner address
		ret = append(ret, n.Owner...)
		// store the size of destination
		binary.LittleEndian.PutUint64(buf, uint64(len(n.Destination)))
		ret = append(ret, buf...)
		// store the destination address
		ret = append(ret, n.Destination...)
		// if there is an operator, store it
		if n.Operator != nil {
			// store the size of operator
			binary.LittleEndian.PutUint64(buf, uint64(len(n.Operator)))
			ret = append(ret, buf...)
			// store the operator address
			ret = append(ret, n.Operator...)
		}
	}
	return ret
}

func deserializeNameMap(data []byte) *NameMap {
	if data != nil {
		var operator []byte

		version := data[0]
		if version != 1 {
			panic("could not deserializeOwner, not supported version")
		}

		// read the size of owner
		offset := 1
		next := offset + 8
		sizeOfAddr := binary.LittleEndian.Uint64(data[offset:next])

		// read the owner address
		offset = next
		next = offset + int(sizeOfAddr)
		owner := data[offset:next]

		// read the size of destination
		offset = next
		next = offset + 8
		sizeOfDest := binary.LittleEndian.Uint64(data[offset:next])

		// read the destination address
		offset = next
		next = offset + int(sizeOfDest)
		destination := data[offset:next]

		// if there are remaining bytes, then the operator is included
		if (len(data) > next) {
			// read the size of operator
			offset = next
			next = offset + 8
			sizeOfOperator := binary.LittleEndian.Uint64(data[offset:next])

			// read the operator address
			offset = next
			next = offset + int(sizeOfOperator)
			operator = data[offset:next]
		}

		return &NameMap{
			Version:     version,
			Owner:       owner,
			Destination: destination,
			Operator:    operator,
		}
	}
	return nil
}
