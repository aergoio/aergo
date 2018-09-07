package account

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
)

//Signer is submodule of account for signing the transaction
type Signer struct {
	log *log.Logger
	key *aergokey
}

//NewSigner make new instance
func NewSigner(l *log.Logger, k *aergokey) *Signer {
	return &Signer{
		log: l,
		key: k,
	}
}

//Receive actor message
func (s *Signer) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *types.Tx:
		err := s.SignTx(msg)
		defer context.Self().Stop()
		if err != nil {
			context.Respond(&message.SignTxRsp{Tx: nil, Err: err})
		} else {
			//context.Tell(context.Sender(), &message.SignTxRsp{Tx: msg, Err: nil})
			context.Respond(&message.SignTxRsp{Tx: msg, Err: nil})
		}
	}
}

//SignTx sign transaction with key
func (s *Signer) SignTx(tx *types.Tx) error {
	//hash tx
	txbody := tx.Body
	hash := HashWithoutSign(txbody)
	//sign tx
	sign, err := btcec.SignCompact(btcec.S256(), s.key, hash, true)
	if err != nil {
		s.log.Warn().Err(err).Msg("could not sign")
		return err
	}
	txbody.Sign = sign
	//txbody.Sign = sign
	tx.Hash = tx.CalculateTxHash()
	return nil
}

func HashWithoutSign(txBody *types.TxBody) []byte {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, txBody.Nonce)
	h.Write(txBody.Account)
	h.Write(txBody.Recipient)
	binary.Write(h, binary.LittleEndian, txBody.Amount)
	h.Write(txBody.Payload)
	binary.Write(h, binary.LittleEndian, txBody.Limit)
	binary.Write(h, binary.LittleEndian, txBody.Price)
	binary.Write(h, binary.LittleEndian, txBody.Type)
	return h.Sum(nil)
}
