package types

import (
	"bytes"
	"encoding/json"
	"github.com/gogo/protobuf/jsonpb"
	"strings"
)

func NewReceipt(contractAddress []byte, status string, jsonRet string) Receipt {
	return Receipt{
		ContractAddress: contractAddress,
		Status:          status,
		Ret:             jsonRet,
	}
}

func NewReceiptFromBytes(b []byte) *Receipt {
	r := new(Receipt)
	r.ContractAddress = b[:20]
	endIdx := bytes.IndexByte(b[20:], 0x00) + 20
	r.Status = string(b[20:endIdx])
	r.Ret = string(b[endIdx+1:])
	return r
}

func (r Receipt) Bytes() []byte {
	var b bytes.Buffer
	b.Write(r.ContractAddress)
	b.WriteString(r.Status)
	b.WriteByte(0x00)
	b.WriteString(r.Ret)
	return b.Bytes()
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
	b.WriteString(`","ret":"`)
	b.WriteString(strings.Replace(r.Ret, "\"", "'", -1))
	b.WriteString(`"}`)
	return b.Bytes(), nil
}
