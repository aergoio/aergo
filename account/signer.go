package account

import (
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
)

type Signer struct {
	keystore *key.Store
}

func NewSigner(s *key.Store) *Signer {
	return &Signer{keystore: s}
}

//Receive actor message
func (s *Signer) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *types.Tx:
		err := s.keystore.SignTx(msg)
		defer context.Self().Stop()
		if err != nil {
			context.Respond(&message.SignTxRsp{Tx: nil, Err: err})
		} else {
			//context.Tell(context.Sender(), &message.SignTxRsp{Tx: msg, Err: nil})
			context.Respond(&message.SignTxRsp{Tx: msg, Err: nil})
		}
	}
}
