package syncer

import (
	"github.com/aergoio/aergo-lib/log"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/pkg/component"

	"fmt"
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/pkg/errors"
	"reflect"
	"sync"
	"time"
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
	blockAdder   *BlockAdder
}

var (
	logger = log.NewLogger("syncer")
)

var (
	ErrorNotFoundAncestor = errors.New("not found ancestor in remote server")
)

type Finder struct {
	hub *component.ComponentHub //for communicate with other service

	lScanCh chan *types.BlockInfo
	fScanCh chan []byte

	doneCh chan *FinderResult
	quitCh chan interface{}

	lastAnchor []byte //point last block during lightscan
	ctx        types.SyncContext

	dfltTimeout time.Duration

	waitGroup *sync.WaitGroup
}

type FinderResult struct {
	ancestor *types.BlockInfo
	err      error
}

type HashFetcher struct {
	actor.Actor
	*SubActor

	ctx *types.SyncContext
}

type BlockFetcher struct {
	actor.Actor
	*SubActor
}

type BlockAdder struct {
	actor.Actor
	*SubActor
}

var (
	FinderName       = "Ancestor Finder"
	HashFetcherName  = "Hash Fetcher"
	BlockFetcherName = "Block Fetcher"
	BlockAdderName   = "Block Adder"
)

func NewSyncer(cfg *cfg.Config, chain types.ChainAccessor) *Syncer {
	syncer := &Syncer{cfg: cfg}

	syncer.BaseComponent = component.NewBaseComponent(message.SyncerSvc, syncer, logger)

	hub := syncer.BaseComponent.Hub()
	syncer.hashFetcher = newHashFetcher(1, hub)
	/*
		syncer.blockFetcher = newBlockFetcher(1, hub)
		syncer.blockAdder = newBlockAdder(1, hub)*/

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
	syncer.hashFetcher.Stop()
	syncer.blockFetcher.Stop()
	syncer.blockAdder.Stop()
}

func (syncer *Syncer) Reset() {
	syncer.finder.stop()
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
			logger.Error().Err(err).Msg("FindAncestorRsp failed")
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
		logger.Debug().Uint64("targetNo", msg.TargetNo).Msg("skipped syncer is startning")
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
	if msg.Ancestor == nil {
		logger.Error().Msg("Find Ancestor failed")
		return ErrorNotFoundAncestor
	}

	//set ancestor in types.SyncContext
	ancestor := msg.Ancestor
	syncer.ctx.CommonAncestor = ancestor
	syncer.ctx.TotalCnt = (syncer.ctx.TargetNo - syncer.ctx.CommonAncestor.No)
	syncer.ctx.RemainCnt = syncer.ctx.TotalCnt

	logger.Info().Str("hash", enc.ToString(ancestor.Hash)).Uint64("no", ancestor.No).Msg("sync found ancestor")

	syncer.hashFetcher.ctx = syncer.ctx
	//request hash download
	syncer.hashFetcher.Tell(&message.StartFetch{})

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

	syncer.hashFetcher.ctx = syncer.ctx
	//request hash download
	syncer.hashFetcher.Tell(&message.StartFetch{})

	return nil
}

func (syncer *Syncer) Statistics() *map[string]interface{} {
	return &map[string]interface{}{
		"startning": syncer.isstartning,
		"total":     syncer.ctx.TotalCnt,
		"remain":    syncer.ctx.RemainCnt,
	}
}

func newHashFetcher(cntWorker int, hub *component.ComponentHub) *HashFetcher {
	HashFetcher := &HashFetcher{}
	HashFetcher.SubActor = newSubActor(HashFetcher, HashFetcherName, cntWorker, hub)

	return HashFetcher
}

func newBlockFetcher(cntWorker int, hub *component.ComponentHub) *BlockFetcher {
	BlockFetcher := &BlockFetcher{}
	BlockFetcher.SubActor = newSubActor(BlockFetcher, BlockFetcherName, cntWorker, hub)

	return BlockFetcher
}

/*
func newBlockAdder(cntWorker int, hub *component.ComponentHub) *BlockAdder {
	blockAdder := &BlockAdder{}
	blockAdder.SubActor = newSubActor(blockAdder, BlockAdderName, cntWorker, hub)
	return blockAdder
}*/

func (hdl *HashFetcher) Receive(context actor.Context) {
	logger.Debug().Msg("HashFetcher")
	switch msg := context.Message().(type) {
	case *message.StartFetch:
		if hdl.ctx == nil {
			panic("Hash downloader context is nil")
		}
		hdl.StartFetch(msg, context)
	case actor.Started:
		logger.Debug().Str("name", hdl.name).Msg("actor started")

	case actor.SystemMessage,
		actor.AutoReceiveMessage,
		actor.NotInfluenceReceiveTimeout:
		str := fmt.Sprintf("Received message. (%v) %s", reflect.TypeOf(msg), msg)
		logger.Debug().Msg(str)

	default:
		debug := fmt.Sprintf("[%s] Missed message. (%v) %s", hdl.name, reflect.TypeOf(msg), msg)
		logger.Debug().Msg(debug)
	}
}

func (hdl *HashFetcher) StartFetch(msg *message.StartFetch, context actor.Context) {
	panic("not implemented")
}

func (bdl *BlockFetcher) Receive(context actor.Context) {
	logger.Debug().Msg("BlockFetcher")

	switch msg := context.Message().(type) {
	case actor.Started:
		logger.Debug().Str("name", bdl.name).Msg("actor started")

	case actor.SystemMessage,
		actor.AutoReceiveMessage,
		actor.NotInfluenceReceiveTimeout:
		str := fmt.Sprintf("Received message. (%v) %s", reflect.TypeOf(msg), msg)
		logger.Debug().Msg(str)

	default:
		debug := fmt.Sprintf("[%s] Missed message. (%v) %s", bdl.name, reflect.TypeOf(msg), msg)
		logger.Debug().Msg(debug)
	}
}
