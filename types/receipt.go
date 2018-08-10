package types

import (
	"bytes"
	)

func NewReceipt(contractAddress []byte, status string) Receipt {
	return Receipt{
		ContractAddress: contractAddress,
		Status:          status,
	}
}

func NewReceiptFromBytes(b []byte) *Receipt {
	r := new(Receipt)
	r.ContractAddress = b[:32]
	r.Status = string(b[32:])
	return r
}

func (r Receipt) Bytes() []byte {
	var b bytes.Buffer
	b.Write(r.ContractAddress)
	b.WriteString(r.Status)
	return b.Bytes()
}
