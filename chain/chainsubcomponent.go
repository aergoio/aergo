package chain

import (
	"fmt"
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-actor/router"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"reflect"
	"runtime"
)

// SubComponent handles message with Receive(), and requests to other actor services with IComponentRequester
// To use SubComponent, only need to implement Actor interface
type SubComponent struct {
	actor.Actor
	component.IComponentRequester

	name  string
	pid   *actor.PID
	count int
}

const (
	defaultChainWorkerCount = 3
)

func NewSubComponent(subactor actor.Actor, requester component.IComponentRequester, name string, cntWorker int) *SubComponent {
	return &SubComponent{
		Actor:               subactor,
		IComponentRequester: requester,
		name:                name,
		count:               cntWorker}
}

// spawn new subactor
func (sub *SubComponent) Start() {
	sub.pid = actor.Spawn(router.NewRoundRobinPool(sub.count).WithInstance(sub.Actor))

	msg := fmt.Sprintf("%s[%d] started", sub.name, sub.count)
	logger.Info().Msg(msg)
}

// stop subactor
func (sub *SubComponent) Stop() {
	sub.pid.GracefulStop()
	msg := fmt.Sprintf("%s stoped", sub.name)
	logger.Info().Msg(msg)
}

//send message to this subcomponent and reply to actor with pid respondTo
func (sub *SubComponent) Request(message interface{}, respondTo *actor.PID) {
	sub.pid.Request(message, respondTo)
}

type ChainManager struct {
	*SubComponent
	IChainHandler //to use chain APIs
}

type ChainWorker struct {
	*SubComponent
	IChainHandler //to use chain APIs
}

var (
	chainManagerName = "Chain Manager"
	chainWorkerName  = "Chain Worker"
)

func newChainManager(chain *ChainService) *ChainManager {
	chainManager := &ChainManager{IChainHandler: chain}
	chainManager.SubComponent = NewSubComponent(chainManager, chain, chainManagerName, 1)

	return chainManager
}

func newChainWorker(chain *ChainService, cntWorker int) *ChainWorker {
	chainWorker := &ChainWorker{IChainHandler: chain}
	chainWorker.SubComponent = NewSubComponent(chainWorker, chain, chainWorkerName, cntWorker)

	return chainWorker
}

func (cm *ChainManager) Receive(context actor.Context) {
	logger.Debug().Msg("chain manager")
	switch msg := context.Message().(type) {

	case *message.AddBlock:
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		bid := msg.Block.BlockID()
		block := msg.Block
		logger.Debug().Str("hash", msg.Block.ID()).
			Uint64("blockNo", msg.Block.GetHeader().GetBlockNo()).Bool("syncer", msg.IsSync).Msg("add block chainservice")
		_, err := cm.getBlock(bid[:])
		if err == nil {
			logger.Debug().Str("hash", msg.Block.ID()).Msg("already exist")
			err = ErrBlockExist
		} else {
			var bstate *state.BlockState
			if msg.Bstate != nil {
				bstate = msg.Bstate.(*state.BlockState)
			}
			err = cm.addBlock(block, bstate, msg.PeerID)
			if err != nil && err != ErrBlockOrphan {
				logger.Error().Err(err).Str("hash", msg.Block.ID()).Msg("failed add block")
			}
		}

		rsp := message.AddBlockRsp{
			BlockNo:   block.GetHeader().GetBlockNo(),
			BlockHash: block.BlockHash(),
			Err:       err,
		}

		if msg.IsSync {
			//TODO change Syncer AddBlock request to use Request()
			cm.RequestTo(message.SyncerSvc, &rsp)
		} else {
			context.Respond(rsp)
		}

		cm.TellTo(message.RPCSvc, block)
	default:
		debug := fmt.Sprintf("[%s] Missed message. (%v) %s", cm.name, reflect.TypeOf(msg), msg)
		logger.Debug().Msg(debug)
	}
}

func (cw *ChainWorker) Receive(context actor.Context) {
	logger.Debug().Msg("chain worker")
	switch msg := context.Message().(type) {
	/*
		case *message.GetMissing:
			stopHash := msg.StopHash
			hashes := msg.Hashes
			mhashes, mnos := cw.handleMissing(stopHash, hashes)
			context.Respond(message.GetMissingRsp{
				Hashes:   mhashes,
				Blocknos: mnos,
			})
		case *message.SyncBlockState:
			cw.checkBlockHandshake(msg.PeerID, msg.BlockNo, msg.BlockHash)
	*/

	default:
		debug := fmt.Sprintf("[%s] Missed message. (%v) %s", cw.name, reflect.TypeOf(msg), msg)
		logger.Debug().Msg(debug)
	}
}
