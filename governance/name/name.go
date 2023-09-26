package name

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

var prefix = []byte("name")

// names in memory
type Names struct {
	names map[string]*NameMap
}

func NewNames() *Names {
	return &Names{
		names: map[string]*NameMap{},
	}
}

func (n *Names) Copy() *Names {
	names := &Names{
		names: map[string]*NameMap{},
	}
	for k, v := range n.names {
		names.names[k] = &NameMap{
			Version:     v.Version,
			Owner:       v.Owner,
			Destination: v.Destination,
		}
	}
	return names
}

func (n *Names) GetAddress(name []byte) []byte {
	if nameMap, ok := n.names[string(name)]; ok {
		return nameMap.Destination
	}
	return nil
}

func (n *Names) GetOwner(name []byte) []byte {
	if nameMap, ok := n.names[string(name)]; ok {
		return nameMap.Owner
	}
	return nil
}

func (n *Names) SetOwner(name, owner []byte) {
	n.names[string(name)] = &NameMap{
		Version:     types.NameVer1,
		Owner:       owner,
		Destination: owner,
	}
}

type NameMap struct {
	Version     types.NameVer
	Owner       []byte
	Destination []byte
}

// AccountStateReader is an interface for getting a name account state.
type AccountStateReader interface {
	GetNameAccountState() (*state.ContractState, error)
}

// names in state
// GetAddressFromState is resolve name for mempool
func GetAddressFromState(scs *state.ContractState, name []byte) []byte {
	if len(name) == types.AddressLength ||
		types.IsSpecialAccount(name) {
		return name
	}
	return getAddress(scs, name)
}

// GetAddressLegacy is resolve name for mempool by buggy logic, leaved for backward compatibility
func GetAddressLegacy(scs *state.ContractState, name []byte) []byte {
	if len(name) == types.AddressLength ||
		strings.Contains(string(name), ".") {
		return name
	}
	return getAddress(scs, name)
}

func getAddress(scs *state.ContractState, name []byte) []byte {
	nameMap := getNameMap(scs, name, true)
	if nameMap != nil {
		return nameMap.Destination
	}
	return nil
}

func GetOwnerFromState(scs *state.ContractState, name []byte) []byte {
	return getOwner(scs, name, true)
}

func getOwner(scs *state.ContractState, name []byte, useInitial bool) []byte {
	nameMap := getNameMap(scs, name, useInitial)
	if nameMap != nil {
		return nameMap.Owner
	}
	return nil
}

func getNameMap(scs *state.ContractState, name []byte, useInitial bool) *NameMap {
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
	return deserializeNameMap(ownerdata)
}

func GetNameInfo(r AccountStateReader, name string) (*types.NameInfo, error) {
	scs, err := r.GetNameAccountState()
	if err != nil {
		return nil, err
	}
	owner := getOwner(scs, []byte(name), true)
	return &types.NameInfo{Name: &types.Name{Name: string(name)}, Owner: owner, Destination: GetAddressFromState(scs, []byte(name))}, err
}

func registerOwner(scs *state.ContractState, name, owner, destination []byte) error {
	nameMap := &NameMap{Version: types.NameVer1, Owner: owner, Destination: destination}
	return setNameMap(scs, name, nameMap)
}

func setNameMap(scs *state.ContractState, name []byte, n *NameMap) error {
	lowerCaseName := strings.ToLower(string(name))
	key := append(prefix, lowerCaseName...)
	return scs.SetData(key, serializeNameMap(n))
}

func serializeNameMap(n *NameMap) []byte {
	var ret []byte
	if n != nil {
		ret = append(ret, byte(n.Version))
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
		version := types.NameVer(data[0])
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

func CreateName(scs *state.ContractState, tx *types.TxBody, sender, receiver *state.V, name string) error {
	amount := tx.GetAmountBigInt()
	sender.SubBalance(amount)
	receiver.AddBalance(amount)
	return createName(scs, []byte(name), sender.ID())
}

func createName(scs *state.ContractState, name []byte, owner []byte) error {
	//	return setAddress(scs, name, owner)
	return registerOwner(scs, name, owner, owner)
}

// UpdateName is avaliable after bid implement
func UpdateName(bs *state.BlockState, scs *state.ContractState, tx *types.TxBody,
	sender, receiver *state.V, name, to string) error {
	amount := tx.GetAmountBigInt()
	if len(getAddress(scs, []byte(name))) <= types.NameLength {
		return fmt.Errorf("%s is not created yet", string(name))
	}
	destination, _ := types.DecodeAddress(to)
	destination = GetAddressFromState(scs, destination)
	sender.SubBalance(amount)
	receiver.AddBalance(amount)
	contract, err := bs.StateDB.OpenContractStateAccount(types.ToAccountID(destination))
	if err != nil {
		return types.ErrTxInvalidRecipient
	}
	creator, err := contract.GetData([]byte("Creator"))
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

func updateName(scs *state.ContractState, name []byte, owner []byte, to []byte) error {
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
