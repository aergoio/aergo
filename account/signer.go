package account

import (
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/v2/account/key"
	"github.com/aergoio/aergo/v2/message"
)

type Signer struct {
	keystore *key.Store
}

func NewSigner(s *key.Store) *Signer {
	return &Signer{keystore: s}
}

// Receive actor message
func (s *Signer) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *message.SignTx:
		err := s.keystore.SignTx(msg.Tx, msg.Requester)
		defer context.Self().Stop()
		if err != nil {
			context.Respond(&message.SignTxRsp{Tx: nil, Err: err})
		} else {
			//context.Tell(context.Sender(), &message.SignTxRsp{Tx: msg, Err: nil})
			context.Respond(&message.SignTxRsp{Tx: msg.Tx, Err: nil})
		}
	}
}
