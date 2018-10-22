package mempool

import (
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
)

type TxVerifier struct {
	mp *MemPool
}

func NewTxVerifier(p *MemPool) *TxVerifier {
	return &TxVerifier{mp: p}
}

//Receive actor message
func (s *TxVerifier) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *types.Tx:
		err := s.mp.verifyTx(msg)
		if err == nil {
			err = s.mp.put(msg)
		}
		context.Respond(&message.MemPoolPutRsp{Err: err})
	}
}
