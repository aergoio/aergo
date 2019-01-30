package types

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/param"
	"math/big"
	"strings"

	"github.com/aergoio/aergo/internal/merkle"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/minio/sha256-simd"
)

func NewReceipt(contractAddress []byte, status string, jsonRet string) *Receipt {
	return &Receipt{
		ContractAddress: contractAddress[:33],
		Status:          status,
		Ret:             jsonRet,
	}
}

func (r Receipt) marshalBody(b *bytes.Buffer) {
	l := make([]byte, 8)
	b.Write(r.ContractAddress)
	binary.LittleEndian.PutUint16(l[:2], uint16(len(r.Status)))
	b.Write(l[:2])
	b.WriteString(r.Status)
	if param.GetForkConfig().ISAIP1(r.BlockNo) {
		b.WriteByte(0)
		binary.LittleEndian.PutUint32(l[:4], uint32(len(r.Ret)))
		b.Write(l[:4])
		b.WriteString(r.Ret)
		b.Write(r.TxHash)

		binary.LittleEndian.PutUint32(l[:4], uint32(len(r.FeeUsed)))
		b.Write(l[:4])
		b.Write(r.FeeUsed)
		binary.LittleEndian.PutUint32(l[:4], uint32(len(r.CumulativeFeeUsed)))
		b.Write(l[:4])
		b.Write(r.CumulativeFeeUsed)
		binary.LittleEndian.PutUint32(l[:4], uint32(len(r.Data)))
		b.Write(l[:4])
		b.Write(r.Data)
		binary.LittleEndian.PutUint32(l[:4], uint32(len(r.Events)))
		b.Write(l[:4])
	} else {
		b.WriteString(r.Ret)
	}
}

func (r Receipt) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer

	r.marshalBody(&b)
	for _, ev := range r.Events {
		evB, err := ev.MarshalBinary()
		if err != nil {
			return nil, err
		}
		b.Write(evB)
	}

	return b.Bytes(), nil
}

func (r Receipt) MarshalMerkleBinary() ([]byte, error) {
	var b bytes.Buffer

	r.marshalBody(&b)
	for _, ev := range r.Events {
		evB, err := ev.MarshalMerkleBinary()
		if err != nil {
			return nil, err
		}
		b.Write(evB)
	}

	return b.Bytes(), nil
}

func (r *Receipt) UnmarshalBinary(data []byte) error {
	r.ContractAddress = data[:33]
	l := uint32(binary.LittleEndian.Uint16(data[33:]))
	pos := 35 + l
	r.Status = string(data[35:pos])
	if pos == uint32(len(data)) || data[pos] != 0 {
		return nil
	}
	pos += 1
	l = binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	r.Ret = string(data[pos : pos+l])
	pos += l
	r.TxHash = data[pos : pos+32]
	pos += 32
	l = binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	r.FeeUsed = data[pos : pos+l]
	pos += l
	l = binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	r.CumulativeFeeUsed = data[pos : pos+l]
	pos += l
	l = binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	r.Data = data[pos : pos+l]
	pos += l
	evCount := binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	evData := data[pos:]

	r.Events = make([]*Event, evCount)
	var err error
	for i := uint32(0); i < evCount; i++ {
		var ev Event
		evData, err = ev.UnmarshalBinary(evData)
		if err != nil {
			return err
		}
		r.Events[i] = &ev
	}
	return nil
}

func (r Receipt) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	return json.Marshal(r)
}

func (r Receipt) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(`{"contractAddress":"`)
	b.WriteString(EncodeAddress(r.ContractAddress))
	b.WriteString(`","status":"`)
	b.WriteString(strings.Replace(r.Status, "\"", "'", -1))
	if len(r.Ret) == 0 {
		b.WriteString(`","ret": {}`)
	} else {
		b.WriteString(`","ret": `)
		b.WriteString(r.Ret)
	}
	b.WriteString(`,"txHash":"`)
	b.WriteString(enc.ToString(r.TxHash))
	b.WriteString(`","usedFee":`)
	b.WriteString(new(big.Int).SetBytes(r.FeeUsed).String())
	b.WriteString(`,"events":[`)
	for i, ev := range r.Events {
		if i != 0 {
			b.WriteString(`,`)
		}
		byte, err := ev.MarshalJSON()
		if err != nil {
			return nil, err
		}
		b.Write(byte)
	}
	b.WriteString(`]}`)
	return b.Bytes(), nil
}

func (r Receipt) GetHash() []byte {
	h := sha256.New()
	b, _ := r.MarshalMerkleBinary()
	h.Write(b)
	return h.Sum(nil)
}

type Receipts []*Receipt

func (rs Receipts) MerkleRoot() []byte {
	mes := make([]merkle.MerkleEntry, len(rs))
	for i, r := range rs {
		mes[i] = r
	}
	return merkle.CalculateMerkleRoot(mes)
}

func (rs Receipts) SetBLockInfo(blockHash []byte, blockNo uint64) {
	for i, r := range rs {
		r.BlockNo = blockNo
		for _, e := range r.GetEvents() {
			e.TxIndex = int32(i)
			e.TxHash = r.TxHash
			e.BlockHash = blockHash
			e.BlockNo = blockNo
		}
	}
}

func (ev Event) marshalCommonBinary(b *bytes.Buffer) {
	l := make([]byte, 4)
	b.Write(ev.ContractAddress)
	binary.LittleEndian.PutUint32(l, uint32(len(ev.EventName)))
	b.Write(l)
	b.WriteString(ev.EventName)

	binary.LittleEndian.PutUint32(l, uint32(len(ev.JsonArgs)))
	b.Write(l)
	b.WriteString(ev.JsonArgs)

	b.Write(ev.TxHash)

	binary.LittleEndian.PutUint32(l, uint32(ev.EventIdx))
	b.Write(l)
}

func (ev Event) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	l := make([]byte, 8)
	ev.marshalCommonBinary(&b)

	b.Write(ev.BlockHash)
	binary.LittleEndian.PutUint64(l[:], ev.BlockNo)
	b.Write(l)
	binary.LittleEndian.PutUint32(l[:4], uint32(ev.TxIndex))
	b.Write(l[:4])
	return b.Bytes(), nil
}

func (ev Event) MarshalMerkleBinary() ([]byte, error) {
	var b bytes.Buffer
	ev.marshalCommonBinary(&b)
	return b.Bytes(), nil
}

func (ev *Event) UnmarshalBinary(data []byte) ([]byte, error) {
	pos := uint32(33)
	ev.ContractAddress = data[:pos]

	l := binary.LittleEndian.Uint32(data[33:])
	pos += 4
	ev.EventName = string(data[pos : pos+l])
	pos += l

	l = binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	ev.JsonArgs = string(data[pos : pos+l])
	pos += l

	ev.TxHash = data[pos : pos+32]
	pos += 32

	ev.EventIdx = int32(binary.LittleEndian.Uint32(data[pos:]))
	pos += 4

	ev.BlockHash = data[pos : pos+32]
	pos += 32

	ev.BlockNo = binary.LittleEndian.Uint64(data[pos:])
	pos += 8

	ev.TxIndex = int32(binary.LittleEndian.Uint32(data[pos:]))

	return data[pos+4:], nil
}

func (ev Event) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(`{"contractAddress":"`)
	b.WriteString(EncodeAddress(ev.ContractAddress))
	b.WriteString(`","eventName":"`)
	b.WriteString(ev.EventName)
	b.WriteString(`","Args":`)
	b.WriteString(ev.JsonArgs)
	b.WriteString(`,"txHash":"`)
	b.WriteString(enc.ToString(ev.TxHash))
	b.WriteString(`","EventIdx":`)
	b.WriteString(fmt.Sprintf("%d", ev.EventIdx))
	b.WriteString(`,"BlockHash":"`)
	b.WriteString(enc.ToString(ev.BlockHash))
	b.WriteString(`","BlockNo":`)
	b.WriteString(fmt.Sprintf("%d", ev.BlockNo))
	b.WriteString(`,"TxIndex":`)
	b.WriteString(fmt.Sprintf("%d", ev.TxIndex))
	b.WriteString(`}`)
	return b.Bytes(), nil
}

func (ev Event) Filter(filter *FilterInfo) bool {

	if filter.ContractAddress != nil && !bytes.Equal(ev.ContractAddress, filter.ContractAddress) {
		return false
	}
	if len(filter.EventName) != 0 && ev.EventName != filter.EventName {
		return false
	}
	return true
}
