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
}

func CreateName(scs *statedb.ContractState, tx *types.TxBody, sender, receiver *state.AccountState, name string) error {
	amount := tx.GetAmountBigInt()
	sender.SubBalance(amount)
	receiver.AddBalance(amount)
	return createName(scs, []byte(name), sender.ID())
}

func createName(scs *statedb.ContractState, name []byte, owner []byte) error {
	//	return setAddress(scs, name, owner)
	return registerOwner(scs, name, owner, owner)
}

// UpdateName is avaliable after bid implement
func UpdateName(bs *state.BlockState, scs *statedb.ContractState, tx *types.TxBody,
	sender, receiver *state.AccountState, name, to string) error {
	if len(getAddress(scs, []byte(name))) <= types.NameLength {
		return fmt.Errorf("%s is not created yet", string(name))
	}
	destination, _ := types.DecodeAddress(to)
	destination = GetAddress(scs, destination)

	amount := tx.GetAmountBigInt()
	sender.SubBalance(amount)
	receiver.AddBalance(amount)
	contract, err := statedb.OpenContractStateAccount(types.ToAccountID(destination), bs.LuaStateDB)
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
	v, err := state.GetAccountState([]byte(types.AergoName), bs.LuaStateDB, bs.EvmStateDB)
	if err != nil {
		return nil, err
	}
	scs, err := statedb.OpenContractState(v.AccountID(), v.State(), bs.LuaStateDB)
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
		ret = append(ret, n.Version)
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(len(n.Owner)))
		ret = append(ret, buf...)
		ret = append(ret, n.Owner...)
		binary.LittleEndian.PutUint64(buf, uint64(len(n.Destination)))
		ret = append(ret, buf...)
		ret = append(ret, n.Destination...)
	}
	return ret
}

func deserializeNameMap(data []byte) *NameMap {
	if data != nil {
		version := data[0]
		if version != 1 {
			panic("could not deserializeOwner, not supported version")
		}
		offset := 1
		next := offset + 8
		sizeOfAddr := binary.LittleEndian.Uint64(data[offset:next])

		offset = next
		next = offset + int(sizeOfAddr)
		owner := data[offset:next]

		offset = next
		next = offset + 8
		sizeOfDest := binary.LittleEndian.Uint64(data[offset:next])

		offset = next
		next = offset + int(sizeOfDest)
		destination := data[offset:next]
		return &NameMap{
			Version:     version,
			Owner:       owner,
			Destination: destination,
		}
	}
	return nil
}
