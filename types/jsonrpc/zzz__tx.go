package jsonrpc

import (
	"errors"
	"math/big"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

type InOutTx struct {
	Hash string       `json:",omitempty"`
	Body *InOutTxBody `json:",omitempty"`
}

func (t *InOutTx) FromProto(msg *types.Tx, payloadType EncodingType) {
	t.Body = &InOutTxBody{}
	if msg == nil {
		return
	}
	t.Hash = base58.Encode(msg.Hash)
	if msg.Body != nil {
		t.Body.FromProto(msg.Body, payloadType)
	}
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

func (tb *InOutTxBody) FromProto(msg *types.TxBody, payloadType EncodingType) {
	tb.Nonce = msg.Nonce
	if msg.Account != nil {
		tb.Account = types.EncodeAddress(msg.Account)
	}
	if msg.Recipient != nil {
		tb.Recipient = types.EncodeAddress(msg.Recipient)
	}
	if msg.Amount != nil {
		tb.Amount = new(big.Int).SetBytes(msg.Amount).String()
	}
	switch payloadType {
	case Raw:
		tb.Payload = string(msg.Payload)
	case Base58:
		tb.Payload = base58.Encode(msg.Payload)
	}
	tb.GasLimit = msg.GasLimit
	if msg.GasPrice != nil {
		tb.GasPrice = new(big.Int).SetBytes(msg.GasPrice).String()
	}
	tb.ChainIdHash = base58.Encode(msg.ChainIdHash)
	tb.Sign = base58.Encode(msg.Sign)
	tb.Type = msg.Type
}

func (tb *InOutTxBody) ToProto() (msg *types.TxBody, err error) {
	if tb == nil {
		return nil, errors.New("tx body is empty")
	}

	msg.Nonce = tb.Nonce
	if tb.Account != "" {
		msg.Account, err = types.DecodeAddress(tb.Account)
		if err != nil {
			return nil, err
		}
	}
	if tb.Recipient != "" {
		msg.Recipient, err = types.DecodeAddress(tb.Recipient)
		if err != nil {
			return nil, err
		}
	}
	if tb.Amount != "" {
		amount, err := ParseUnit(tb.Amount)
		if err != nil {
			return nil, err
		}
		msg.Amount = amount.Bytes()
	}
	if tb.Payload != "" {
		msg.Payload, err = base58.Decode(tb.Payload)
		if err != nil {
			return nil, err
		}
	}
	msg.GasLimit = tb.GasLimit
	if tb.GasPrice != "" {
		price, err := ParseUnit(tb.GasPrice)
		if err != nil {
			return nil, err
		}
		msg.GasPrice = price.Bytes()
	}
	if tb.ChainIdHash != "" {
		msg.ChainIdHash, err = base58.Decode(tb.ChainIdHash)
		if err != nil {
			return nil, err
		}
	}
	if tb.Sign != "" {
		msg.Sign, err = base58.Decode(tb.Sign)
		if err != nil {
			return nil, err
		}
	}
	msg.Type = tb.Type
	return msg, nil
}

type InOutTxInBlock struct {
	TxIdx *InOutTxIdx
	Tx    *InOutTx
}

func (tib *InOutTxInBlock) FromProto(txInBlock *types.TxInBlock, payloadType EncodingType) {
	tib.TxIdx = &InOutTxIdx{}
	tib.Tx = &InOutTx{}

	if txInBlock.GetTxIdx() != nil {
		tib.TxIdx.FromProto(txInBlock.GetTxIdx())
	}
	if txInBlock.GetTx() != nil {
		tib.Tx.FromProto(txInBlock.GetTx(), payloadType)
	}
}

type InOutTxIdx struct {
	BlockHash string
	Idx       int32
}

func (ti *InOutTxIdx) FromProto(msg *types.TxIdx) {
	ti.BlockHash = base58.Encode(msg.GetBlockHash())
	ti.Idx = msg.GetIdx()
}
