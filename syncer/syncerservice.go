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
	"time"
)

type Syncer struct {
	*component.BaseComponent

	cfg   *cfg.Config
	chain types.ChainAccessor

	isRunning bool
	ctx       *types.SyncContext

	finder       *Finder
	hashFetcher  *HashFetcher
	blockFetcher *BlockFetcher
	blockAdder   *BlockAdder
}

var (
	logger = log.NewLogger("syncer")
)

type Finder struct {
	actor.Actor
	*SubActor

	ctx *types.SyncContext
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
	syncer.finder = newFinder(1, hub)
	/*
		syncer.hashFetcher = newHashFetcher(1, hub)
		syncer.blockFetcher = newBlockFetcher(1, hub)
		syncer.blockAdder = newBlockAdder(1, hub)
	*/

	syncer.chain = chain

	logger.Info().Msg("Syncer started")

	return syncer
}

// BeforeStart initialize chain database and generate empty genesis block if necessary
func (syncer *Syncer) BeforeStart() {
}

// AfterStart ... do nothing
func (syncer *Syncer) AfterStart() {
	syncer.finder.start()
}

func (syncer *Syncer) BeforeStop() {
	syncer.finder.Stop()
	/*
		syncer.hashFetcher.Stop()
		syncer.blockFetcher.Stop()
		syncer.blockAdder.Stop()
	*/
}

func (syncer *Syncer) Reset() {
	syncer.isRunning = false
	syncer.ctx = nil
}

// Receive actor message
func (syncer *Syncer) Receive(context actor.Context) {

	switch msg := context.Message().(type) {
	case message.SyncRequest:
		err := syncer.handleSyncRequest(msg)
		if err != nil {
			logger.Error().Err(err).Msg("SyncRequest failed")
		}

	case message.FindAncestorRsp:
		err := syncer.handleFindAncestorRsp(msg)
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

func (syncer *Syncer) handleSyncRequest(msg message.SyncRequest) error {
	var err error
	var bestBlock *types.Block

	if syncer.isRunning {
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

	if msg.TargetNo <= bestBlockNo+1 {
		logger.Debug().Uint64("targetNo", msg.TargetNo).Uint64("bestNo", bestBlockNo).
			Msg("skipped syncer. requested no is too low")
		return nil
	}

	logger.Info().Uint64("targetNo", msg.TargetNo).Uint64("bestNo", bestBlockNo).Msg("sync started")

	//TODO BP stop
	syncer.ctx = &types.SyncContext{PeerID: msg.PeerID, TargetNo: msg.TargetNo, BestNo: bestBlockNo}
	syncer.isRunning = true

	syncer.finder.Tell(&message.FindAncestor{Ctx: syncer.ctx})

	return nil
}

func (syncer *Syncer) handleFindAncestorRsp(msg message.FindAncestorRsp) error {
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
		"running": syncer.isRunning,
		"total":   syncer.ctx.TotalCnt,
		"remain":  syncer.ctx.RemainCnt,
	}
}

func newFinder(cntWorker int, hub *component.ComponentHub) *Finder {
	finder := &Finder{}
	finder.SubActor = newSubActor(finder, FinderName, cntWorker, hub)

	return finder
}

/*
func newBlockAdder(cntWorker int, hub *component.ComponentHub) *BlockAdder {
	blockAdder := &BlockAdder{}
	blockAdder.SubActor = newSubActor(BlockAdderName, cntWorker, hub)
	return blockAdder
}


func newHashFetcher(cntWorker int, hub *component.ComponentHub) *HashFetcher {
	HashFetcher := &HashFetcher{}
	HashFetcher.SubActor = newSubActor(HashFetcherName, cntWorker, hub)

	return HashFetcher
}

func newBlockFetcher(cntWorker int, hub *component.ComponentHub) *BlockFetcher {
	BlockFetcher := &BlockFetcher{}
	BlockFetcher.SubActor = newSubActor(BlockFetcherName, cntWorker, hub)

	return BlockFetcher
}
*/

func (finder *Finder) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *message.FindAncestor:
		finder.handleFindAncestor(msg, context)

	case actor.Started:
		logger.Debug().Msg("actor[Common Ancestor Finder] started")

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

func (finder *Finder) handleFindAncestor(msg *message.FindAncestor, context actor.Context) {
	var ancestor *types.BlockInfo
	var err error

	finder.ctx = msg.Ctx

	//1. light sync
	//   gather summary of my chain nodes, request searching ancestor to remote node
	ancestor, err = finder.lightscan()

	//2. heavy sync
	//	 full binary search in my chain
	if ancestor == nil && err == nil {
		ancestor, err = finder.fullscan()
	}

	context.Respond(message.FindAncestorRsp{
		Ancestor: ancestor,
		Err:      err,
	})
}

func (finder *Finder) lightscan() (*types.BlockInfo, error) {
	result, err := finder.hub.RequestFuture(message.ChainSvc, message.GetAnchors{}, time.Second*100,
		"syncer/finder/handleFindAncestor").Result()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get anchors")
		return nil, err
	}

	//	send remote Peer
	result, err = finder.hub.RequestFuture(message.P2PSvc,
		message.GetAncestor{ToWhom: finder.ctx.PeerID, Hashes: result.(message.GetAnchorsRsp).Hashes},
		time.Second*300, "syncer/finder/handleFindAncestor").Result()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get common ancestor")
		return nil, err
	}

	ancestor := &types.BlockInfo{Hash: result.(message.GetAncestorRsp).Hash, No: result.(message.GetAncestorRsp).No}
	return ancestor, nil

}

//FIXME XXX no commit
func (finder *Finder) fullscan() (*types.BlockInfo, error) {
	return nil, nil
}

func (hdl *HashFetcher) Receive(context actor.Context) {
	logger.Debug().Msg("HashFetcher")
	switch msg := context.Message().(type) {
	case message.StartFetch:
		if hdl.ctx == nil {
			panic("Hash downloader context is nil")
		}
		//TODO
	default:
		debug := fmt.Sprintf("[%s] Missed message. (%v) %s", hdl.name, reflect.TypeOf(msg), msg)
		logger.Debug().Msg(debug)
	}
}

func (bdl *BlockFetcher) Receive(context actor.Context) {
	logger.Debug().Msg("BlockFetcher")

	switch msg := context.Message().(type) {
	default:
		debug := fmt.Sprintf("[%s] Missed message. (%v) %s", bdl.name, reflect.TypeOf(msg), msg)
		logger.Debug().Msg(debug)
	}
}
