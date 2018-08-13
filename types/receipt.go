package types

import (
	"bytes"
	"github.com/gogo/protobuf/jsonpb"
	"encoding/json"
	"github.com/mr-tron/base58/base58"
)

func NewReceipt(contractAddress []byte, status string) Receipt {
	return Receipt{
		ContractAddress: contractAddress,
		Status:          status,
	}
}

func NewReceiptFromBytes(b []byte) *Receipt {
	r := new(Receipt)
	r.ContractAddress = b[:20]
	r.Status = string(b[20:])
	return r
}

func (r Receipt) Bytes() []byte {
	var b bytes.Buffer
	b.Write(r.ContractAddress)
	b.WriteString(r.Status)
	return b.Bytes()
}

func (r Receipt) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	return json.Marshal(r)
}

func (r Receipt) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(`{"contractAddress":"`)
	b.WriteString(base58.Encode(r.ContractAddress))
	b.WriteString(`","status":"`)
	b.WriteString(r.Status)
	b.WriteString(`"}`)
	return b.Bytes(), nil
}
