package types

import (
	"encoding/binary"
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
