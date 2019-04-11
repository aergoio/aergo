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
		var err error
		if s.mp.exist(msg.GetHash()) != nil {
			err = types.ErrTxAlreadyInMempool
		} else {
			tx := types.NewTransaction(msg)
			err = s.mp.verifyTx(tx)
			if err == nil {
				err = s.mp.put(tx)
			}
		}
		context.Respond(&message.MemPoolPutRsp{Err: err})
	}
}
