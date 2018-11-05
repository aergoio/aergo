package syncer

import (
	"github.com/aergoio/aergo-lib/log"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/pkg/component"

	"fmt"
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"reflect"
)

type Syncer struct {
	*component.BaseComponent

	cfg   *cfg.Config
	chain types.ChainAccessor

	isstartning bool
	ctx         *types.SyncContext

	finder       *Finder
	hashFetcher  *HashFetcher
	blockFetcher *BlockFetcher
}

var (
	logger = log.NewLogger("syncer")
)

func NewSyncer(cfg *cfg.Config, chain types.ChainAccessor) *Syncer {
	syncer := &Syncer{cfg: cfg}

	syncer.BaseComponent = component.NewBaseComponent(message.SyncerSvc, syncer, logger)

	syncer.chain = chain

	logger.Info().Msg("Syncer started")

	return syncer
}

// BeforeStart initialize chain database and generate empty genesis block if necessary
func (syncer *Syncer) BeforeStart() {
}

// AfterStart ... do nothing
func (syncer *Syncer) AfterStart() {

}

func (syncer *Syncer) BeforeStop() {
}

func (syncer *Syncer) Reset() {
	syncer.finder.stop()
	syncer.hashFetcher.stop()

	syncer.finder = nil
	syncer.isstartning = false
	syncer.ctx = nil
}

// Receive actor message
func (syncer *Syncer) Receive(context actor.Context) {

	switch msg := context.Message().(type) {
	case *message.SyncStart:
		err := syncer.handleSyncStart(msg)
		if err != nil {
			logger.Error().Err(err).Msg("SyncStart failed")
		}
	case *message.GetSyncAncestorRsp:
		err := syncer.handleAncestorRsp(msg)
		if err != nil {
			syncer.Reset()
			logger.Error().Err(err).Msg("FindAncestorRsp failed")
		}
	case *message.FinderResult:
		err := syncer.handleFinderResult(msg)
		if err != nil {
			syncer.Reset()
			logger.Error().Err(err).Msg("FinderResult failed")
		}
	case *message.GetHashesRsp:
		err := syncer.hashFetcher.setResult(msg)
		if err != nil {
			syncer.Reset()
			logger.Error().Err(err).Msg("GetHashes failed")
		}
	case *message.GetBlockChunksRsp:
		err := fmt.Errorf("to imple")
		if err != nil {
			syncer.Reset()
			logger.Error().Err(err).Msg("GetHashes failed")
		}
	case actor.SystemMessage,
		actor.AutoReceiveMessage,
		actor.NotInfluenceReceiveTimeout:
		str := fmt.Sprintf("Received message. (%v) %s", reflect.TypeOf(msg), msg)
		logger.Debug().Msg(str)

	default:
		str := fmt.Sprintf("Missed message. (%v) %s", reflect.TypeOf(msg), msg)
		logger.Debug().Msg(str)
	}
}

func (syncer *Syncer) handleSyncStart(msg *message.SyncStart) error {
	var err error
	var bestBlock *types.Block

	logger.Debug().Uint64("targetNo", msg.TargetNo).Msg("sync requested")

	if syncer.isstartning {
		logger.Debug().Uint64("targetNo", msg.TargetNo).Msg("skipped syncer is running")
		return nil
	}

	//TODO skip sync in reorgnizing
	bestBlock, _ = syncer.chain.GetBestBlock()
	if err != nil {
		logger.Error().Err(err).Msg("error getting block in syncer")
		return err
	}

	bestBlockNo := bestBlock.GetHeader().BlockNo

	if msg.TargetNo <= bestBlockNo {
		logger.Debug().Uint64("targetNo", msg.TargetNo).Uint64("bestNo", bestBlockNo).
			Msg("skipped syncer. requested no is too low")
		return nil
	}

	logger.Info().Uint64("targetNo", msg.TargetNo).Uint64("bestNo", bestBlockNo).Msg("sync started")

	//TODO BP stop
	syncer.ctx = &types.SyncContext{PeerID: msg.PeerID, TargetNo: msg.TargetNo, BestNo: bestBlockNo}
	syncer.isstartning = true

	syncer.finder = newFinder(syncer.ctx, syncer.Hub())

	return err
}

func (syncer *Syncer) handleAncestorRsp(msg *message.GetSyncAncestorRsp) error {
	logger.Info().Msg("syncer received ancestor response")

	//set ancestor in types.SyncContext
	syncer.finder.lScanCh <- msg.Ancestor

	return nil
}

func (syncer *Syncer) handleFinderResult(msg *message.FinderResult) error {
	if msg.Err != nil {
		logger.Error().Err(msg.Err).Msg("Find Ancestor failed")
		syncer.Reset()
		return nil
	}

	//set ancestor in types.SyncContext
	syncer.ctx.CommonAncestor = msg.Ancestor
	syncer.ctx.TotalCnt = (syncer.ctx.TargetNo - syncer.ctx.CommonAncestor.No)
	syncer.ctx.RemainCnt = syncer.ctx.TotalCnt

	syncer.blockFetcher = newBlockFetcher(syncer.ctx, syncer.Hub())
	syncer.hashFetcher = newHashFetcher(syncer.ctx, syncer.Hub(), syncer.blockFetcher.hfCh)

	return nil
}

func (syncer *Syncer) Statistics() *map[string]interface{} {
	return &map[string]interface{}{
		"startning": syncer.isstartning,
		"total":     syncer.ctx.TotalCnt,
		"remain":    syncer.ctx.RemainCnt,
	}
}
