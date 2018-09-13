package util

import (
	"encoding/base64"
	"encoding/json"

	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
)

type InOutTx struct {
	Hash string
	Body *InOutTxBody
}
type InOutTxBody struct {
	Nonce     uint64
	Account   string
	Recipient string
	Amount    uint64
	Payload   string
	Limit     uint64
	Price     uint64
	Sign      string
	Type      types.TxType
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
		if in.Hash != "" {
			tx.Hash, err = base58.Decode(in.Hash)
			if err != nil {
				return nil, err
			}
		}
		tx.Body.Nonce = in.Body.Nonce
		if in.Body.Account != "" {
			tx.Body.Account, err = base58.Decode(in.Body.Account)
			if err != nil {
				return nil, err
			}
		}
		if in.Body.Recipient != "" {
			tx.Body.Recipient, err = base58.Decode(in.Body.Recipient)
			if err != nil {
				return nil, err
			}
		}
		tx.Body.Amount = in.Body.Amount
		if in.Body.Payload != "" {
			tx.Body.Payload, err = base58.Decode(in.Body.Payload)
			if err != nil {
				return nil, err
			}
		}
		tx.Body.Limit = in.Body.Limit
		tx.Body.Price = in.Body.Price
		if in.Body.Sign != "" {
			tx.Body.Sign, err = base58.Decode(in.Body.Sign)
			if err != nil {
				return nil, err
			}
		}
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
	if in.Account != "" {
		body.Account, err = base58.Decode(in.Account)
		if err != nil {
			return nil, err
		}
	}
	if in.Recipient != "" {
		body.Recipient, err = base58.Decode(in.Recipient)
		if err != nil {
			return nil, err
		}
	}
	body.Amount = in.Amount
	if in.Payload != "" {
		body.Payload, err = base58.Decode(in.Payload)
		if err != nil {
			return nil, err
		}
	}
	body.Limit = in.Limit
	body.Price = in.Price
	if in.Sign != "" {
		body.Sign, err = base58.Decode(in.Sign)
		if err != nil {
			return nil, err
		}
	}
	body.Type = in.Type

	return body, nil
}

func ConvBase58Addr(tx *types.Tx) string {
	out := &InOutTx{Body: &InOutTxBody{}}
	out.Hash = base58.Encode(tx.Hash)
	out.Body.Nonce = tx.Body.Nonce
	out.Body.Account = base58.Encode(tx.Body.Account)
	out.Body.Recipient = base58.Encode(tx.Body.Recipient)
	out.Body.Amount = tx.Body.Amount
	out.Body.Payload = base58.Encode(tx.Body.Payload)
	out.Body.Limit = tx.Body.Limit
	out.Body.Price = tx.Body.Price
	out.Body.Sign = base58.Encode(tx.Body.Sign)
	out.Body.Type = tx.Body.Type
	jsonout, err := json.MarshalIndent(out, "", " ")
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
