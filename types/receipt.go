package types

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
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

func (r Receipt) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	l := make([]byte, 2)
	b.Write(r.ContractAddress)
	binary.LittleEndian.PutUint16(l[:], uint16(len(r.Status)))
	b.Write(l)
	b.WriteString(r.Status)
	b.WriteString(r.Ret)
	return b.Bytes(), nil
}

func (r *Receipt) UnmarshalBinary(data []byte) error {
	r.ContractAddress = data[:33]
	l := binary.LittleEndian.Uint16(data[33:])
	r.Status = string(data[35 : 35+l])
	r.Ret = string(data[35+l:])
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
		b.WriteString(`","ret": {}}`)
	} else {
		b.WriteString(`","ret": `)
		b.WriteString(r.Ret)
		b.WriteString(`}`)
	}
	return b.Bytes(), nil
}

func (r Receipt) GetHash() []byte {
	h := sha256.New()
	b, _ := r.MarshalBinary()
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
