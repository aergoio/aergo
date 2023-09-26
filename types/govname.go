package types

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

type NameVer byte

const (
	NameVer1 NameVer = iota + 1
)

// governance type transaction which has aergo.system in recipient

const (
	SetContractOwner = "v1setOwner"
	NameCreate       = "v1createName"
	NameUpdate       = "v1updateName"

	TxMaxSize = 200 * 1024
)

type NameMap struct {
	Version     NameVer
	Owner       []byte
	Destination []byte
}

func SerializeNameMap(n *NameMap) []byte {
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

func DeserializeNameMap(data []byte) *NameMap {
	if data != nil {
		version := NameVer(data[0])
		if version != NameVer1 {
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

func validateNameTx(tx *TxBody) error {
	var ci CallInfo
	if err := json.Unmarshal(tx.Payload, &ci); err != nil {
		return ErrTxInvalidPayload
	}
	switch ci.Name {
	case NameCreate:
		if err := _validateNameTx(tx, &ci); err != nil {
			return err
		}
		if len(ci.Args) != 1 {
			return fmt.Errorf("invalid arguments in %s", ci)
		}
	case NameUpdate:
		if err := _validateNameTx(tx, &ci); err != nil {
			return err
		}
		if len(ci.Args) != 2 {
			return fmt.Errorf("invalid arguments in %s", ci)
		}
		to, err := DecodeAddress(ci.Args[1].(string))
		if err != nil {
			return fmt.Errorf("invalid receiver in %s", ci)
		}
		if len(to) > AddressLength {
			return fmt.Errorf("too long name %s", string(tx.GetPayload()))
		}
	case SetContractOwner:
		owner, ok := ci.Args[0].(string)
		if !ok {
			return fmt.Errorf("invalid arguments in %s", owner)
		}
		_, err := DecodeAddress(owner)
		if err != nil {
			return fmt.Errorf("invalid new owner %s", err.Error())
		}
	default:
		return ErrTxInvalidPayload
	}
	return nil
}

func _validateNameTx(tx *TxBody, ci *CallInfo) error {
	if len(ci.Args) < 1 {
		return fmt.Errorf("invalid arguments in %s", ci)
	}
	nameParam, ok := ci.Args[0].(string)
	if !ok {
		return fmt.Errorf("invalid arguments in %s", nameParam)
	}

	if len(nameParam) > NameLength {
		return fmt.Errorf("too long name %s", string(tx.GetPayload()))
	}
	if len(nameParam) != NameLength {
		return fmt.Errorf("not supported yet")
	}
	if err := validateAllowedChar([]byte(nameParam)); err != nil {
		return err
	}
	return nil
}
