package util

import (
	"encoding/base64"
	"encoding/json"

	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
)

type InOutTx struct {
	Hash []byte
	Body *InOutTxBody
}
type InOutTxBody struct {
	Nonce     uint64
	Account   string
	Recipient string
	Amount    uint64
	Payload   []byte
	Limit     uint64
	Price     uint64
	Sign      []byte
	Type      uint64
}

func ParseBase58Tx(jsonTx []byte) ([]*types.Tx, error) {
	//tx := &types.Tx{Body: &types.TxBody{}}
	//in := &InOutTx{Body: &InOutTxBody{}}
	var inputlist []InOutTx
	err := json.Unmarshal([]byte(jsonTx), &inputlist)
	if err != nil {
		return nil, err
	}
	txs := make([]*types.Tx, len(inputlist))
	for i, in := range inputlist {
		tx := &types.Tx{Body: &types.TxBody{}}
		tx.Hash = in.Hash
		tx.Body.Nonce = in.Body.Nonce
		tx.Body.Account, _ = base58.Decode(in.Body.Account)
		tx.Body.Recipient, _ = base58.Decode(in.Body.Recipient)
		tx.Body.Amount = in.Body.Amount
		tx.Body.Payload = in.Body.Payload
		tx.Body.Limit = in.Body.Limit
		tx.Body.Price = in.Body.Price
		tx.Body.Sign = in.Body.Sign
		tx.Body.Type = in.Body.Type
		txs[i] = tx
	}

	return txs, nil
}

func ParseBase58TxBody(jsonTx []byte) (*types.TxBody, error) {
	body := &types.TxBody{}
	in := &InOutTxBody{}

	err := json.Unmarshal([]byte(jsonTx), in)
	if err != nil {
		return nil, err
	}

	body.Nonce = in.Nonce
	body.Account, _ = base58.Decode(in.Account)
	body.Recipient, _ = base58.Decode(in.Recipient)
	body.Amount = in.Amount
	if in.Payload != nil {
		body.Payload = in.Payload
	}
	body.Limit = in.Limit
	body.Price = in.Price
	body.Sign = in.Sign
	body.Type = in.Type

	return body, nil
}

func ConvBase58Addr(tx *types.Tx) string {
	out := &InOutTx{Body: &InOutTxBody{}}
	out.Hash = tx.Hash
	out.Body.Nonce = tx.Body.Nonce
	out.Body.Account = base58.Encode(tx.Body.Account)
	out.Body.Recipient = base58.Encode(tx.Body.Recipient)
	out.Body.Amount = tx.Body.Amount
	out.Body.Payload = tx.Body.Payload
	out.Body.Limit = tx.Body.Limit
	out.Body.Price = tx.Body.Price
	out.Body.Sign = tx.Body.Sign
	out.Body.Type = tx.Body.Type
	jsonout, err := json.Marshal(out)
	if err != nil {
		return ""
	}
	return string(jsonout)
}

//TODO: refactoring util function
func EncodeB64(bs []byte) string {
	return base64.StdEncoding.EncodeToString(bs)
}
func DecodeB64(sb string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(sb)
}
