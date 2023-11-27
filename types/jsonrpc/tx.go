package jsonrpc

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

func ParseBase58Tx(jsonTx []byte) ([]*types.Tx, error) {
	var inputlist []InOutTx
	err := json.Unmarshal([]byte(jsonTx), &inputlist)
	if err != nil {
		var input InOutTx
		err = json.Unmarshal([]byte(jsonTx), &input)
		if err != nil {
			return nil, err
		}
		inputlist = append(inputlist, input)
	}
	txs := make([]*types.Tx, len(inputlist))
	for i, in := range inputlist {
		tx := &types.Tx{Body: &types.TxBody{}}
		if in.Hash != "" {
			tx.Hash, err = base58.Decode(in.Hash)
			if err != nil {
				return nil, err
			}
		}
		tx.Body, err = ParseTxBody(in.Body)
		if err != nil {
			return nil, err
		}
		txs[i] = tx
	}

	return txs, nil
}

func ParseBase58TxBody(jsonTx []byte) (*types.TxBody, error) {
	in := &InOutTxBody{}
	err := json.Unmarshal(jsonTx, in)
	if err != nil {
		return nil, err
	}

	body, err := ParseTxBody(in)
	if err != nil {
		return nil, err
	}

	return body, nil
}

//-----------------------------------------------------------------------//
//

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
	Hash string       `json:"hash,omitempty"`
	Body *InOutTxBody `json:"body,omitempty"`
}

func (t *InOutTx) String() string {
	return MarshalJSON(t)
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

	if tb.PayloadJson.Name != "" {		
		payload, err := json.Marshal(tb.PayloadJson)

		if err != nil {
			return nil, err
		}else{
			msg.Payload = payload
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

type InOutPayload struct {
	Name string       	`json:"name,omitempty"`
	Arg  []interface{} 	`json:"arg,omitempty"`
}

type InOutTxBody struct {
	Nonce       uint64       	`json:"nonce,omitempty"`
	Account     string       	`json:"account,omitempty"`
	Recipient   string       	`json:"recipient,omitempty"`
	Amount      string      	`json:"amount,omitempty"`
	Payload     string       	`json:"payload,omitempty"`
	PayloadJson	types.CallInfo  `json:"payloadJson,omitempty"`
	GasLimit    uint64       	`json:"gasLimit,omitempty"`
	GasPrice    string       	`json:"gasPrice,omitempty"`
	Type        types.TxType 	`json:"type,omitempty"`
	ChainIdHash string       	`json:"chainIdHash,omitempty"`
	Sign        string       	`json:"sign,omitempty"`
}

func (b *InOutTxBody) String() string {
	return MarshalJSON(b)
}

func ConvTxInBlock(msg *types.TxInBlock, payloadType EncodingType) *InOutTxInBlock {
	tib := &InOutTxInBlock{}

	if msg.GetTxIdx() != nil {
		tib.TxIdx = *ConvTxIdx(msg.GetTxIdx())
	}
	if msg.GetTx() != nil {
		tib.Tx = *ConvTx(msg.GetTx(), payloadType)
	}
	return tib
}

type InOutTxInBlock struct {
	TxIdx InOutTxIdx 	`json:"txIdx"`
	Tx    InOutTx		`json:"tx"`
}

func ConvTxIdx(msg *types.TxIdx) *InOutTxIdx {
	return &InOutTxIdx{
		BlockHash: base58.Encode(msg.GetBlockHash()),
		Idx:       msg.GetIdx(),
	}
}

type InOutTxIdx struct {
	BlockHash string `json:"blockHash,omitempty"`
	Idx       int32  `json:"idx,omitempty"`
}

func (t *InOutTxInBlock) String() string {
	return MarshalJSON(t)
}
