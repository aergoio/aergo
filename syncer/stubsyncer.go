package syncer

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

type StubSyncer struct {
	realSyncer    *Syncer
	stubRequester *StubRequester

	localChain  *chain.StubBlockChain
	remoteChain *chain.StubBlockChain

	stubPeers []*StubPeer

	t *testing.T

	waitGroup *sync.WaitGroup

	cfg *SyncerConfig

	checkResultFn         TestResultFn
	getAnchorsHookFn      GetAnchorsHookFn
	getSyncAncestorHookFn GetSyncAncestorHookFn
}

type TestResultFn func(stubSyncer *StubSyncer)
type GetAnchorsHookFn func(stubSyncer *StubSyncer)
type GetSyncAncestorHookFn func(stubSyncer *StubSyncer, msg *message.GetSyncAncestor)

var (
	targetPeerID = types.PeerID([]byte(fmt.Sprintf("peer-%d", 0)))
)

func makeStubPeerSet(remoteChains []*chain.StubBlockChain) []*StubPeer {
	stubPeers := make([]*StubPeer, len(remoteChains))

	for i, chain := range remoteChains {
		stubPeers[i] = NewStubPeer(i, uint64(chain.Best), chain)
	}

	return stubPeers
}

func NewTestSyncer(t *testing.T, localChain *chain.StubBlockChain, remoteChain *chain.StubBlockChain, peers []*StubPeer, cfg *SyncerConfig) *StubSyncer {
	syncer := NewSyncer(nil, localChain, cfg)
	testsyncer := &StubSyncer{realSyncer: syncer, localChain: localChain, remoteChain: remoteChain, stubPeers: peers, cfg: cfg, t: t}

	testsyncer.stubRequester = NewStubRequester()

	syncer.SetRequester(testsyncer.stubRequester)

	return testsyncer
}

func (stubSyncer *StubSyncer) start() {
	stubSyncer.waitGroup = &sync.WaitGroup{}
	stubSyncer.waitGroup.Add(1)

	go func() {
		defer stubSyncer.waitGroup.Done()

		for {
			msg := stubSyncer.stubRequester.recvMessage()
			isStop := stubSyncer.handleMessage(msg)
			if isStop {
				return
			}
		}
	}()
}

func (stubSyncer *StubSyncer) waitStop() {
	logger.Info().Msg("test syncer wait to stop")
	stubSyncer.waitGroup.Wait()
	logger.Info().Msg("test syncer stopped")
}

func isOtherActorRequest(msg interface{}) bool {
	switch msg.(type) {
	case *message.GetSyncAncestor:
		return true
	case *message.GetAnchors:
		return true
	case *message.GetAncestor:
		return true
	case *message.GetHashByNo:
		return true
	case *message.GetHashes:
		return true
	case *message.GetPeers:
		return true
	case *message.GetBlockChunks:
		return true
	case *message.AddBlock:
		return true
	}

	return false
}

func (stubSyncer *StubSyncer) handleMessage(msg interface{}) bool {
	//prefix handle
	switch resmsg := msg.(type) {
	case *message.FinderResult:
		if resmsg.Ancestor != nil && resmsg.Err == nil && resmsg.Ancestor.No >= 0 {
			stubSyncer.localChain.Rollback(resmsg.Ancestor)
		}
	case *message.CloseFetcher:
		if resmsg.FromWho == NameHashFetcher {
			if stubSyncer.cfg.debugContext.debugHashFetcher {
				assert.Equal(stubSyncer.t, stubSyncer.realSyncer.hashFetcher.lastBlockInfo.No, stubSyncer.cfg.debugContext.targetNo, "invalid hash target")
			}
		} else {
			assert.Fail(stubSyncer.t, "invalid closefetcher")
		}
	case *message.SyncStop:
		if stubSyncer.cfg.debugContext.expErrResult != nil {
			assert.Equal(stubSyncer.t, stubSyncer.cfg.debugContext.expErrResult, resmsg.Err, "expected syncer stop error")
		}
		//check final result
		if stubSyncer.checkResultFn != nil {
			stubSyncer.checkResultFn(stubSyncer)
		}
	default:
	}

	if isOtherActorRequest(msg) {
		logger.Debug().Msg("msg is for testsyncer")

		stubSyncer.handleActorMsg(msg)
	} else {

		logger.Debug().Msg("msg is for syncer")
		stubSyncer.realSyncer.handleMessage(msg)
	}

	//check stop
	switch resmsg := msg.(type) {
	case *message.SyncStop:
		return true
	case *message.FinderResult:
		if stubSyncer.cfg.debugContext.expAncestor >= 0 {
			assert.Equal(stubSyncer.t, uint64(stubSyncer.cfg.debugContext.expAncestor), resmsg.Ancestor.No, "ancestor mismatch")
		} else if !stubSyncer.realSyncer.isRunning {
			assert.Equal(stubSyncer.t, stubSyncer.cfg.debugContext.expAncestor, -1, "ancestor mismatch")
			return true
		}

		if stubSyncer.cfg.debugContext.debugFinder {
			return true
		}
	case *message.CloseFetcher:
		if stubSyncer.cfg.debugContext.debugHashFetcher {
			return true
		}
	default:
		return false
	}

	return false
}

func (stubSyncer *StubSyncer) handleActorMsg(inmsg interface{}) {
	switch msg := inmsg.(type) {
	case *message.GetAnchors:
		stubSyncer.GetAnchors(msg)
	case *message.GetSyncAncestor:
		stubSyncer.GetSyncAncestor(msg)
	case *message.GetHashByNo:
		stubSyncer.GetHashByNo(msg)

	case *message.GetHashes:
		stubSyncer.GetHashes(msg, nil)

	case *message.GetPeers:
		stubSyncer.GetPeers(msg)

	case *message.GetBlockChunks:
		stubSyncer.GetBlockChunks(msg)

	case *message.AddBlock:
		stubSyncer.AddBlock(msg, nil)

	case *actor.Started, *actor.Stopping, *actor.Stopped, *component.CompStatReq: // donothing

	default:
		str := fmt.Sprintf("Missed message. (%v) %s", reflect.TypeOf(msg), msg)
		stubSyncer.t.Fatal(str)
	}
}

// reply to requestFuture()
func (syncer *StubSyncer) GetAnchors(msg *message.GetAnchors) {
	if syncer.getAnchorsHookFn != nil {
		syncer.getAnchorsHookFn(syncer)
	}

	go func() {
		if syncer.cfg.debugContext.debugFinder {
			if syncer.stubPeers[0].timeDelaySec > 0 {
				logger.Debug().Stringer("peer", types.LogPeerShort(types.PeerID(syncer.stubPeers[0].addr.PeerID))).Msg("slow target peer sleep")
				time.Sleep(syncer.stubPeers[0].timeDelaySec)
				logger.Debug().Stringer("peer", types.LogPeerShort(types.PeerID(syncer.stubPeers[0].addr.PeerID))).Msg("slow target peer wakeup")
			}
		}

		hashes, lastno, err := syncer.localChain.GetAnchors()

		rspMsg := message.GetAnchorsRsp{Seq: msg.Seq, Hashes: hashes, LastNo: lastno, Err: err}
		syncer.stubRequester.sendReply(StubRequestResult{rspMsg, nil})
	}()
}

func (syncer *StubSyncer) GetPeers(msg *message.GetPeers) {
	rspMsg := makePeerReply(syncer.stubPeers)
	syncer.stubRequester.sendReply(StubRequestResult{rspMsg, nil})
}

func (syncer *StubSyncer) SendGetSyncAncestorRsp(msg *message.GetSyncAncestor) {
	//find peer
	stubPeer := syncer.findStubPeer(msg.ToWhom)
	ancestor := stubPeer.blockChain.GetAncestorWithHashes(msg.Hashes)

	rspMsg := &message.GetSyncAncestorRsp{Seq: msg.Seq, Ancestor: ancestor}
	syncer.stubRequester.TellTo(message.SyncerSvc, rspMsg) //TODO refactoring: stubcompRequesterresult
}

func (syncer *StubSyncer) GetSyncAncestor(msg *message.GetSyncAncestor) {
	if syncer.getSyncAncestorHookFn != nil {
		syncer.getSyncAncestorHookFn(syncer, msg)
	}

	syncer.SendGetSyncAncestorRsp(msg)
}

func (syncer *StubSyncer) GetHashByNo(msg *message.GetHashByNo) {
	//targetPeer = 0
	hash, err := syncer.stubPeers[0].blockChain.GetHashByNo(msg.BlockNo)
	rsp := &message.GetHashByNoRsp{Seq: msg.Seq, BlockHash: hash, Err: err}
	syncer.stubRequester.TellTo(message.SyncerSvc, rsp)
}
func (syncer *StubSyncer) GetHashes(msg *message.GetHashes, responseErr error) {
	blkHashes, _ := syncer.remoteChain.GetHashes(msg.PrevInfo, msg.Count)

	assert.Equal(syncer.t, len(blkHashes), int(msg.Count))
	rsp := &message.GetHashesRsp{Seq: msg.Seq, PrevInfo: msg.PrevInfo, Hashes: blkHashes, Count: uint64(len(blkHashes)), Err: responseErr}

	syncer.stubRequester.TellTo(message.SyncerSvc, rsp)
}

func (syncer *StubSyncer) GetBlockChunks(msg *message.GetBlockChunks) {
	stubPeer := syncer.findStubPeer(msg.ToWhom)
	stubPeer.blockFetched = true

	assert.True(syncer.t, stubPeer != nil, "peer exist")

	if stubPeer.HookGetBlockChunkRsp != nil {
		stubPeer.HookGetBlockChunkRsp(msg)
		return
	}
	go func() {
		if stubPeer.timeDelaySec > 0 {
			logger.Debug().Stringer("peer", types.LogPeerShort(types.PeerID(stubPeer.addr.PeerID))).Msg("slow peer sleep")
			time.Sleep(stubPeer.timeDelaySec)
			logger.Debug().Stringer("peer", types.LogPeerShort(types.PeerID(stubPeer.addr.PeerID))).Msg("slow peer wakeup")
		}

		//send reply
		blocks, err := stubPeer.blockChain.GetBlocks(msg.Hashes)

		rsp := &message.GetBlockChunksRsp{Seq: msg.Seq, ToWhom: msg.ToWhom, Blocks: blocks, Err: err}
		syncer.stubRequester.TellTo(message.SyncerSvc, rsp)
	}()
}

// ChainService
func (syncer *StubSyncer) AddBlock(msg *message.AddBlock, responseErr error) {
	err := syncer.localChain.AddBlock(msg.Block)

	rsp := &message.AddBlockRsp{BlockNo: msg.Block.GetHeader().BlockNo, BlockHash: msg.Block.GetHash(), Err: err}
	logger.Debug().Uint64("no", msg.Block.GetHeader().BlockNo).Msg("add block succeed")
	syncer.stubRequester.TellTo(message.SyncerSvc, rsp)
}

func (syncer *StubSyncer) findStubPeer(peerID types.PeerID) *StubPeer {
	for _, tmpPeer := range syncer.stubPeers {
		peerIDStr := string(tmpPeer.addr.PeerID)
		logger.Info().Str("tmp", peerIDStr).Msg("peer is")
		if strings.Compare(peerIDStr, string(peerID)) == 0 {
			return tmpPeer
		}
	}

	logger.Error().Stringer("peer", types.LogPeerShort(peerID)).Msg("can't find peer")
	panic("peer find fail")
}

func makePeerReply(stubPeers []*StubPeer) *message.GetPeersRsp {
	count := len(stubPeers)
	now := time.Now()
	peers := make([]*message.PeerInfo, count)
	for i, p := range stubPeers {
		peerInfo := &message.PeerInfo{
			Addr: p.addr, CheckTime: now, State: p.state, LastBlockHash: p.lastBlk.BlockHash, LastBlockNumber: p.lastBlk.BlockNo,
		}
		peers[i] = peerInfo
	}

	return &message.GetPeersRsp{Peers: peers}
}

// test block fetcher only
func (stubSyncer *StubSyncer) runTestBlockFetcher(ctx *types.SyncContext) {
	stubSyncer.realSyncer.blockFetcher = newBlockFetcher(ctx, stubSyncer.realSyncer.getCompRequester(), stubSyncer.cfg)
	stubSyncer.realSyncer.blockFetcher.Start()
}

func (stubSyncer *StubSyncer) sendHashSetToBlockFetcher(hashSet *HashSet) {
	logger.Debug().Uint64("no", hashSet.StartNo).Msg("test syncer pushed hashset to blockfetcher")

	stubSyncer.realSyncer.blockFetcher.hfCh <- hashSet
}
