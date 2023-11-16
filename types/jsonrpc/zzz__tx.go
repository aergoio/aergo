package jsonrpc

import (
	"errors"
	"math/big"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

func ConvTx(msg *types.Tx, payloadType EncodingType) (tx *InOutTx) {
	tx = &InOutTx{}
	tx.Body = &InOutTxBody{}
	if msg == nil {
		return
	}
	tx.Hash = base58.Encode(msg.Hash)
	if msg.Body != nil {
		tx.Body = ConvTxBody(msg.Body, payloadType)
	}
	return tx
}

type InOutTx struct {
	Hash string       `json:",omitempty"`
	Body *InOutTxBody `json:",omitempty"`
}

func ConvTxBody(msg *types.TxBody, payloadType EncodingType) *InOutTxBody {
	tb := &InOutTxBody{}
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
	return tb
}

func ParseTxBody(tb *InOutTxBody) (msg *types.TxBody, err error) {
	if tb == nil {
		return nil, errors.New("tx body is empty")
	}
	msg = &types.TxBody{}

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

func ConvTxInBlock(msg *types.TxInBlock, payloadType EncodingType) *InOutTxInBlock {
	tib := &InOutTxInBlock{}
	tib.TxIdx = &InOutTxIdx{}
	tib.Tx = &InOutTx{}

	if msg.GetTxIdx() != nil {
		tib.TxIdx = ConvTxIdx(msg.GetTxIdx())
	}
	if msg.GetTx() != nil {
		tib.Tx = ConvTx(msg.GetTx(), payloadType)
	}
	return tib
}

type InOutTxInBlock struct {
	TxIdx *InOutTxIdx
	Tx    *InOutTx
}

func ConvTxIdx(msg *types.TxIdx) *InOutTxIdx {
	ti := &InOutTxIdx{}
	ti.BlockHash = base58.Encode(msg.GetBlockHash())
	ti.Idx = msg.GetIdx()
	return ti
}

type InOutTxIdx struct {
	BlockHash string
	Idx       int32
}
