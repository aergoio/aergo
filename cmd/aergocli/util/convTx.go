package util

import "github.com/aergoio/aergo/v2/types"

type EncodingType int

const (
	Raw EncodingType = 0 + iota
	Base58
)

type InOutTx struct {
	Hash string       `json:",omitempty"`
	Body *InOutTxBody `json:",omitempty"`
}

type InOutTxBody struct {
	Nonce       uint64       `json:",omitempty"`
	Account     string       `json:",omitempty"`
	Recipient   string       `json:",omitempty"`
	Amount      string       `json:",omitempty"`
	Payload     string       `json:",omitempty"`
	GasLimit    uint64       `json:",omitempty"`
	GasPrice    string       `json:",omitempty"`
	Type        types.TxType `json:",omitempty"`
	ChainIdHash string       `json:",omitempty"`
	Sign        string       `json:",omitempty"`
}

type InOutTxIdx struct {
	BlockHash string
	Idx       int32
}

type InOutTxInBlock struct {
	TxIdx *InOutTxIdx
	Tx    *InOutTx
}

func (b *InOutTxBody) String() string {
	return toString(b)
}

func (t *InOutTx) String() string {
	return toString(t)
}

func (t *InOutTxInBlock) String() string {
	return toString(t)
}

func TxConvBase58Addr(tx *types.Tx) string {
	return toString(ConvTx(tx))
}

func TxConvBase58AddrEx(tx *types.Tx, payloadType EncodingType) string {
	switch payloadType {
	case Raw:
		return toString(ConvTxEx(tx, Raw))
	case Base58:
		return toString(ConvTxEx(tx, Base58))
	}
	return ""
}

func TxInBlockConvBase58Addr(txInBlock *types.TxInBlock) string {
	return toString(ConvTxInBlock(txInBlock))
}

func ConvTx(tx *types.Tx) *InOutTx {
	return ConvTxEx(tx, Base58)
}

func ConvTxInBlock(txInBlock *types.TxInBlock) *InOutTxInBlock {
	return ConvTxInBlockEx(txInBlock, Base58)
}
