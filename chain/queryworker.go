package chain

import (
	ctx "context"
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

// QueryWorker works for contract query only
type QueryWorker struct {
	*SubComponent
	IChainHandler //to use chain APIs
	*Core
}

func newQueryWorker(cs *ChainService, cntWorker int, core *Core) *QueryWorker {
	worker := &QueryWorker{IChainHandler: cs, Core: core}
	worker.SubComponent = NewSubComponent(worker, cs.BaseComponent, chainWorkerName, cntWorker)
	return worker
}

func (qw *QueryWorker) Receive(context actor.Context) {
	defer RecoverExit()
	switch msg := context.Message().(type) {
	case *message.GetQueryNonBlock:
		select {
		case <-msg.Ctx.Done():
			logger.Warn().Str("contract", types.ToAccountID(msg.Contract).String()).Msg("timeout before querying contract")
		default:
			qw.queryContract(msg.Ctx, msg.Contract, msg.QueryInfo, msg.ReturnChannel)
		}
	}
}

func (qw *QueryWorker) queryContract(ctx ctx.Context, qContract []byte, qInfo []byte, returnChannel chan message.GetQueryRsp) {
	var result message.GetQueryRsp
	defer func() {
		select {
		case <-ctx.Done():
			return
		case returnChannel <- result:
		default:
			logger.Debug().Msg("result channel is already closed or deleted")
		}
	}()
	{
		var sdb = qw.sdb.OpenNewStateDB(qw.sdb.GetRoot())
		address, err := getAddressNameResolved(sdb, qContract)
		if err != nil {
			result = message.GetQueryRsp{Result: nil, Err: err}
			return
		}
		ctrState, err := sdb.OpenContractStateAccount(types.ToAccountID(address))
		if err != nil {
			logger.Error().Str("hash", enc.ToString(address)).Err(err).Msg("failed to get state for contract")
			result = message.GetQueryRsp{Result: nil, Err: err}
		} else {
			bs := state.NewBlockState(sdb)
			ret, err := contract.Query(address, bs, qw.cdb, ctrState, qInfo)
			result = message.GetQueryRsp{Result: ret, Err: err}
		}
	}
}
