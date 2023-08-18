package syncer

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"testing"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/chain"
	cfg "github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/pkg/errors"
)

type Syncer struct {
	*component.BaseComponent

	Seq       uint64
	cfg       *cfg.Config
	syncerCfg *SyncerConfig
	chain     types.ChainAccessor

	isRunning bool
	ctx       *types.SyncContext

	finder       *Finder
	hashFetcher  *HashFetcher
	blockFetcher *BlockFetcher

	compRequester component.IComponentRequester //for test
}

type SyncerConfig struct {
	maxHashReqSize   uint64
	maxBlockReqSize  int
	maxPendingConn   int
	maxBlockReqTasks int

	fetchTimeOut time.Duration

	useFullScanOnly bool

	debugContext *SyncerDebug
}
type SyncerDebug struct {
	t            *testing.T
	expAncestor  int
	targetNo     uint64
	expErrResult error

	debugHashFetcher bool
	debugFinder      bool

	BfWaitTime time.Duration

	logAllPeersBad bool
	logBadPeers    map[int]bool //register bad peers for unit test
}

var (
	logger             = log.NewLogger("syncer")
	NameFinder         = "Finder"
	NameHashFetcher    = "HashFetcher"
	NameBlockFetcher   = "BlockFetcher"
	NameBlockProcessor = "BlockProcessor"
	SyncerCfg          = &SyncerConfig{
		maxHashReqSize:   DfltHashReqSize,
		maxBlockReqSize:  DfltBlockFetchSize,
		maxPendingConn:   MaxBlockPendingTasks,
		maxBlockReqTasks: DfltBlockFetchTasks,
		fetchTimeOut:     DfltFetchTimeOut,
		useFullScanOnly:  false}
)

var (
	ErrFinderInternal = errors.New("error finder internal")
	ErrSyncerPanic    = errors.New("syncer panic")
)

type ErrSyncMsg struct {
	msg interface{}
	str string
}

func (ec *ErrSyncMsg) Error() string {
	return fmt.Sprintf("Error sync message: type=%T, desc=%s", ec.msg, ec.str)
}

func NewSyncer(cfg *cfg.Config, chain types.ChainAccessor, syncerCfg *SyncerConfig) *Syncer {
	if syncerCfg == nil {
		syncerCfg = SyncerCfg
	}

	syncer := &Syncer{cfg: cfg, syncerCfg: syncerCfg}

	syncer.BaseComponent = component.NewBaseComponent(message.SyncerSvc, syncer, logger)
	syncer.compRequester = syncer.BaseComponent
	syncer.chain = chain
	syncer.Seq = 1

	logger.Info().Uint64("seq", syncer.Seq).Msg("Syncer started")

	return syncer
}

// BeforeStart initialize chain database and generate empty genesis block if necessary
func (syncer *Syncer) BeforeStart() {
}

// AfterStart ... do nothing
func (syncer *Syncer) AfterStart() {

}

func (syncer *Syncer) BeforeStop() {
	if syncer.isRunning {
		logger.Info().Msg("syncer BeforeStop")
		syncer.Reset(nil)
	}
}

func (syncer *Syncer) Reset(err error) {
	if syncer.isRunning {
		logger.Info().Uint64("targetNo", syncer.ctx.TargetNo).Msg("syncer stop#1")

		syncer.finder.stop()
		syncer.hashFetcher.stop()
		syncer.blockFetcher.stop()

		syncer.finder = nil
		syncer.hashFetcher = nil
		syncer.blockFetcher = nil
		syncer.isRunning = false

		syncer.notifyStop(err)

		syncer.ctx = nil
	}

	logger.Info().Msg("syncer stopped")
}

func (syncer *Syncer) notifyStop(err error) {
	if syncer.ctx == nil || syncer.ctx.NotifyC == nil {
		return
	}

	logger.Info().Err(err).Msg("notify syncer stop")

	select {
	case syncer.ctx.NotifyC <- err:
	default:
		logger.Debug().Msg("failed to notify syncer stop")
	}
}

func (syncer *Syncer) GetSeq() uint64 {
	return syncer.Seq
}

func (syncer *Syncer) IncSeq() uint64 {
	syncer.Seq++
	return syncer.Seq
}

func (syncer *Syncer) getCompRequester() component.IComponentRequester {
	if syncer.compRequester != nil {
		return syncer.compRequester
	} else {
		return syncer.BaseComponent
	}
}

// This api used for test to set stub IComponentRequester
func (syncer *Syncer) SetRequester(stubRequester component.IComponentRequester) {
	syncer.compRequester = stubRequester
}

// Receive actor message
func (syncer *Syncer) Receive(context actor.Context) {
	//drop garbage message
	if !syncer.isRunning {
		switch context.Message().(type) {
		case *message.GetSyncAncestorRsp,
			*message.FinderResult,
			*message.GetHashesRsp,
			*message.GetHashByNoRsp,
			*message.GetBlockChunks,
			*message.GetBlockChunksRsp,
			*message.AddBlockRsp,
			*message.SyncStop,
			*message.CloseFetcher:
			return
		}
	}

	syncer.handleMessage(context.Message())
}

func (syncer *Syncer) verifySeq(inmsg interface{}) bool {
	isMatch := func(seq uint64) bool {
		return syncer.Seq == seq
	}

	var seq uint64
	var match bool

	switch msg := inmsg.(type) {
	case *message.GetAnchorsRsp:
		seq = msg.Seq
		match = isMatch(seq)
	case *message.GetSyncAncestorRsp:
		seq = msg.Seq
		match = isMatch(seq)
	case *message.FinderResult:
		seq = msg.Seq
		match = isMatch(seq)
	case *message.GetHashesRsp:
		seq = msg.Seq
		match = isMatch(seq)
	case *message.GetHashByNoRsp:
		seq = msg.Seq
		match = isMatch(seq)
	case *message.GetBlockChunksRsp:
		seq = msg.Seq
		match = isMatch(seq)
	case *message.SyncStop:
		seq = msg.Seq
		match = isMatch(seq)
	case *message.CloseFetcher:
		seq = msg.Seq
		match = isMatch(seq)
	default:
		match = true
	}

	if !match {
		logger.Debug().Msgf("syncer(seq=%d) message(%T, seq=%d) is dropped", syncer.GetSeq(), inmsg, seq)
	}

	return match
}

func (syncer *Syncer) handleMessage(inmsg interface{}) {
	defer syncer.RecoverSyncerSelf()

	if !syncer.verifySeq(inmsg) {
		return
	}

	switch msg := inmsg.(type) {
	case *message.SyncStart:
		err := syncer.handleSyncStart(msg)
		if err != nil {
			logger.Error().Err(err).Msg("SyncStart failed")
		}
	case *message.GetSyncAncestorRsp:
		syncer.handleAncestorRsp(msg)
	case *message.GetHashByNoRsp:
		syncer.handleGetHashByNoRsp(msg)
	case *message.FinderResult:
		err := syncer.handleFinderResult(msg)
		if err != nil {
			syncer.Reset(err)
			logger.Error().Err(err).Msg("FinderResult failed")
		}
	case *message.GetHashesRsp:
		syncer.hashFetcher.GetHahsesRsp(msg)

	case *message.GetBlockChunksRsp:
		err := syncer.blockFetcher.handleBlockRsp(msg)
		if err != nil {
			syncer.Reset(err)
			logger.Error().Err(err).Msg("GetBlockChunksRsp failed")
		}
	case *message.AddBlockRsp:
		err := syncer.blockFetcher.handleBlockRsp(msg)
		if err != nil {
			syncer.Reset(err)
			logger.Error().Err(err).Msg("AddBlockRsp failed")
		}
	case *message.SyncStop:
		if msg.Err == nil {
			logger.Info().Str("from", msg.FromWho).Msg("syncer try to stop successfully")
		} else {
			logger.Error().Str("from", msg.FromWho).Err(msg.Err).Msg("syncer try to stop by error")
		}
		syncer.Reset(msg.Err)
	case *message.CloseFetcher:
		if msg.FromWho == NameHashFetcher {
			syncer.hashFetcher.stop()
		} else if msg.FromWho == NameBlockFetcher {
			syncer.blockFetcher.stop()
		} else {
			logger.Error().Msg("invalid closing module message to syncer")
		}
	case actor.SystemMessage,
		actor.AutoReceiveMessage,
		actor.NotInfluenceReceiveTimeout:
		str := fmt.Sprintf("Received message. (%v) %s", reflect.TypeOf(msg), msg)
		logger.Debug().Msg(str)
	case *actor.Started, *actor.Stopping, *actor.Stopped, *component.CompStatReq: // donothing
	default:
		str := fmt.Sprintf("Missed message. (%v) %s", reflect.TypeOf(msg), msg)
		logger.Debug().Msg(str)
	}
}

func (syncer *Syncer) handleSyncStart(msg *message.SyncStart) error {
	var err error
	var bestBlock *types.Block

	logger.Debug().Uint64("targetNo", msg.TargetNo).Str("peer", p2putil.ShortForm(msg.PeerID)).Msg("syncer requested")

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

	if msg.TargetNo <= bestBlockNo {
		logger.Debug().Uint64("targetNo", msg.TargetNo).Uint64("bestNo", bestBlockNo).
			Msg("skipped syncer. requested no is too low")
		return nil
	}

	syncer.IncSeq()

	logger.Info().Uint64("seq", syncer.GetSeq()).Uint64("targetNo", msg.TargetNo).Uint64("bestNo", bestBlockNo).Msg("syncer started")

	//TODO BP stop
	syncer.ctx = types.NewSyncCtx(syncer.GetSeq(), msg.PeerID, msg.TargetNo, bestBlockNo, msg.NotifyC)
	syncer.isRunning = true

	syncer.finder = newFinder(syncer.ctx, syncer.getCompRequester(), syncer.chain, syncer.syncerCfg)
	syncer.finder.start()

	return err
}

func (syncer *Syncer) handleAncestorRsp(msg *message.GetSyncAncestorRsp) {
	var ancestorNo uint64

	if msg.Ancestor != nil {
		ancestorNo = msg.Ancestor.No
	}

	logger.Debug().Uint64("no", ancestorNo).Msg("syncer received ancestor response")

	if syncer.finder == nil {
		logger.Debug().Msg("finder already stopped. so drop unexpected AncestorRsp message")
		return
	}

	//set ancestor in types.SyncContext
	select {
	case syncer.finder.lScanCh <- msg.Ancestor:
		logger.Debug().Uint64("seq", msg.Seq).Msg("syncer transfer response to finder")
	default:
		logger.Debug().Uint64("seq", msg.Seq).Msg("syncer dropped response of finder")
	}
}

func (syncer *Syncer) handleGetHashByNoRsp(msg *message.GetHashByNoRsp) {
	logger.Debug().Msg("syncer received gethashbyno response")

	//set ancestor in types.SyncContext
	syncer.finder.GetHashByNoRsp(msg)
}

func (syncer *Syncer) handleFinderResult(msg *message.FinderResult) error {
	logger.Debug().Msg("syncer received finder result message")

	if err := chain.TestDebugger.Check(chain.DEBUG_SYNCER_CRASH, 0, nil); err != nil {
		chain.TestDebugger.Unset(chain.DEBUG_SYNCER_CRASH)
		return err
	}

	if msg.Err != nil || msg.Ancestor == nil {
		logger.Error().Err(msg.Err).Msg("Find Ancestor failed")
		return ErrFinderInternal
	}

	ancestor, err := syncer.chain.GetBlock(msg.Ancestor.Hash)
	if err != nil {
		logger.Error().Err(err).Msg("error getting ancestor block in syncer")
		return err
	}

	//set ancestor in types.SyncContext
	syncer.ctx.SetAncestor(ancestor)

	syncer.finder.stop()
	syncer.finder = nil

	if syncer.syncerCfg.debugContext != nil && syncer.syncerCfg.debugContext.debugFinder {
		return nil
	}

	syncer.blockFetcher = newBlockFetcher(syncer.ctx, syncer.getCompRequester(), syncer.syncerCfg)
	syncer.hashFetcher = newHashFetcher(syncer.ctx, syncer.getCompRequester(), syncer.blockFetcher.hfCh, syncer.syncerCfg)

	syncer.blockFetcher.Start()
	syncer.hashFetcher.Start()

	return nil
}

func (syncer *Syncer) Statistics() *map[string]interface{} {
	var start, end, total, added, blockfetched uint64

	if syncer.ctx != nil {
		end = syncer.ctx.TargetNo
		if syncer.ctx.CommonAncestor != nil {
			total = syncer.ctx.TotalCnt
			start = syncer.ctx.CommonAncestor.BlockNo()
		}
	} else {
		return &map[string]interface{}{
			"running":       syncer.isRunning,
			"total":         total,
			"start":         start,
			"end":           end,
			"block_added":   added,
			"block_fetched": blockfetched,
		}
	}

	if syncer.blockFetcher != nil {
		lastblock := syncer.blockFetcher.stat.getLastAddBlock()
		added = lastblock.BlockNo()
		if syncer.blockFetcher.stat.getMaxChunkRsp() != nil {
			blockfetched = syncer.blockFetcher.stat.getMaxChunkRsp().BlockNo()
		}
	}

	return &map[string]interface{}{
		"running":       syncer.isRunning,
		"total":         total,
		"start":         start,
		"end":           end,
		"block_added":   added,
		"block_fetched": blockfetched,
	}
}

func (syncer *Syncer) RecoverSyncerSelf() {
	if r := recover(); r != nil {
		logger.Error().Str("dest", "SYNCER").Str("callstack", string(debug.Stack())).Msg("syncer recovered it's panic")
		syncer.Reset(ErrSyncerPanic)
	}
}

func stopSyncer(compRequester component.IComponentRequester, seq uint64, who string, err error) {
	logger.Info().Str("who", who).Err(err).Msg("request syncer stop")

	compRequester.TellTo(message.SyncerSvc, &message.SyncStop{Seq: seq, FromWho: who, Err: err})
}

func closeFetcher(compRequester component.IComponentRequester, seq uint64, who string) {
	compRequester.TellTo(message.SyncerSvc, &message.CloseFetcher{Seq: seq, FromWho: who})
}

func RecoverSyncer(name string, seq uint64, compRequester component.IComponentRequester, finalize func()) {
	if r := recover(); r != nil {
		logger.Error().Str("child", name).Str("callstack", string(debug.Stack())).Msg("syncer recovered child panic")
		stopSyncer(compRequester, seq, name, ErrSyncerPanic)
	}

	if finalize != nil {
		finalize()
	}
}
