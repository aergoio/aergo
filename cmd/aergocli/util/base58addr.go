package util

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anaskhan96/base58check"

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
	Type      types.TxType
	Sign      string
}

func FillTxBody(source *InOutTxBody, target *types.TxBody) error {
	var err error
	target.Nonce = source.Nonce
	if source.Account != "" {
		target.Account, err = types.DecodeAddress(source.Account)
		if err != nil {
			return err
		}
	}
	if source.Recipient != "" {
		target.Recipient, err = types.DecodeAddress(source.Recipient)
		if err != nil {
			return err
		}
	}
	target.Amount = source.Amount
	if source.Payload != "" {
		target.Payload, err = base58.Decode(source.Payload)
		if err != nil {
			return err
		}
	}
	target.Limit = source.Limit
	target.Price = source.Price
	if source.Sign != "" {
		target.Sign, err = base58.Decode(source.Sign)
		if err != nil {
			return err
		}
	}
	target.Type = source.Type
	return nil
}
func ParseBase58Tx(jsonTx []byte) ([]*types.Tx, error) {
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
		err = FillTxBody(in.Body, tx.Body)
		if err != nil {
			return nil, err
		}
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

	err = FillTxBody(in, body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func TxConvBase58Addr(tx *types.Tx) string {
	out := &InOutTx{Body: &InOutTxBody{}}
	out.Hash = base58.Encode(tx.Hash)
	out.Body.Nonce = tx.Body.Nonce
	out.Body.Account = types.EncodeAddress(tx.Body.Account)
	out.Body.Recipient = types.EncodeAddress(tx.Body.Recipient)
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

const CodeVersion = 0xC0

func EncodeCode(code []byte) string {
	encoded, _ := base58check.Encode(fmt.Sprintf("%x", CodeVersion), hex.EncodeToString(code))
	return encoded
}

func DecodeCode(encodedCode string) ([]byte, error) {
	decodedString, err := base58check.Decode(encodedCode)
	if err != nil {
		return nil, err
	}
	decodedBytes, err := hex.DecodeString(decodedString)
	if err != nil {
		return nil, err
	}
	version := decodedBytes[0]
	if version != CodeVersion {
		return nil, errors.New("Invalid code version")
	}
	decoded := decodedBytes[1:]
	return decoded, nil
}
