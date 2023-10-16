package mempool

import (
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/types"
)

type TxVerifier struct {
	mp *MemPool
}

func NewTxVerifier(p *MemPool) *TxVerifier {
	return &TxVerifier{mp: p}
}

// Receive actor message
func (s *TxVerifier) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *types.Tx:
		var err error
		if s.mp.exist(msg.GetHash()) != nil {
			// it's very common cases.
			err = types.ErrTxAlreadyInMempool
			s.mp.Logger.Trace().Object("tx", types.LogTxHash{Tx: msg}).Msg("tx already exist")
		} else {
			tx := types.NewTransaction(msg)
			err = s.mp.verifyTx(tx)
			if err == nil {
				err = s.mp.put(tx)
			}
			if err != nil {
				s.mp.Logger.Info().Err(err).Str("txID", enc.ToString(msg.GetHash())).Msg("tx verification failed")
			}
		}
		context.Respond(&message.MemPoolPutRsp{Err: err})
	}
}
