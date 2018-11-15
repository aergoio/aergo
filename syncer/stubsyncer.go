package syncer

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

//StubSyncer receive Syncer, P2P, Chain Service actor message
type StubSyncer struct {
	ctx *types.SyncContext

	testhub *StubHub

	hf *HashFetcher
	bf *BlockFetcher

	bfInputCh chan *HashSet

	remoteChain *StubBlockChain
	stubPeers   []*StubPeer

	isStop bool
}

type StubBlockChain struct {
	best   int
	hashes []([]byte)
	blocks []*types.Block

	bestBlock *types.Block
}

type StubPeer struct {
	addr    *types.PeerAddress
	lastBlk *types.NewBlockNotice
	state   types.PeerState

	blockChain *StubBlockChain
}

var (
	TestMaxFetchSize = 2
	TestMaxTasks     = 2

	TestMaxHashReq = uint64(3)

	TestMaxPeer = 2
)

var (
	ErrNotExistHash  = errors.New("not exist hash")
	ErrNotExistBlock = errors.New("not exist block of the hash")
)

func NewStubBlockChain() *StubBlockChain {
	tchain := &StubBlockChain{best: -1}

	tchain.hashes = make([][]byte, 1024)
	tchain.blocks = make([]*types.Block, 1024)

	return tchain
}

func (tchain *StubBlockChain) genAddBlock() {
	newBlock := types.NewBlock(tchain.bestBlock, nil, nil, nil, nil, time.Now().UnixNano())
	tchain.addBlock(newBlock)

	time.Sleep(time.Nanosecond * 3)
}

func (tchain *StubBlockChain) addBlock(newBlock *types.Block) {
	tchain.best += 1
	tchain.hashes[tchain.best] = newBlock.BlockHash()
	tchain.blocks[tchain.best] = newBlock
	tchain.bestBlock = newBlock
}

func (tchain *StubBlockChain) GetHashes(prevInfo *types.BlockInfo, count uint64) ([]message.BlockHash, error) {
	if tchain.best < int(prevInfo.No+count) {
		return nil, ErrNotExistHash
	}

	start := prevInfo.No + 1
	resHashes := tchain.hashes[start : start+count]

	blkHashes := make([]message.BlockHash, 0)
	for _, hash := range resHashes {
		blkHashes = append(blkHashes, hash)
	}

	return blkHashes, nil
}

func (tchain *StubBlockChain) GetBlockInfo(no uint64) *types.BlockInfo {
	return &types.BlockInfo{tchain.hashes[no], no}
}

func (tchain *StubBlockChain) GetBlocks(hashes []message.BlockHash) ([]*types.Block, error) {
	startNo := -1

	for i, block := range tchain.blocks {
		if bytes.Equal(block.GetHash(), hashes[0]) {
			startNo = i
			break
		}
	}

	if startNo == -1 {
		return nil, ErrNotExistBlock
	}

	resultBlocks := make([]*types.Block, 0)
	i := startNo
	for _, hash := range hashes {
		if !bytes.Equal(tchain.blocks[i].GetHash(), hash) {
			return nil, ErrNotExistBlock
		}

		resultBlocks = append(resultBlocks, tchain.blocks[i])
		i++
	}

	return resultBlocks, nil
}

func initStubBlockChain(size int) *StubBlockChain {
	chain := NewStubBlockChain()

	//load initial blocks
	for i := 0; i < size; i++ {
		chain.genAddBlock()
	}

	return chain
}

func NewStubSyncer(ctx *types.SyncContext, useHashFetcher bool, useBlockFetcher bool, remoteChain *StubBlockChain) *StubSyncer {
	syncer := &StubSyncer{ctx: ctx}
	syncer.bfInputCh = make(chan *HashSet)
	syncer.testhub = NewStubHub()

	if useHashFetcher {
		syncer.hf = newHashFetcher(ctx, syncer.testhub, syncer.bfInputCh, TestMaxHashReq)
	} else if useBlockFetcher {
		syncer.bf = newBlockFetcher(ctx, syncer.testhub, TestMaxFetchSize, TestMaxTasks)

		syncer.bfInputCh = syncer.bf.hfCh
	}

	syncer.remoteChain = remoteChain

	syncer.makeStubPeerSet(TestMaxPeer)

	return syncer
}

func (syncer *StubSyncer) stop(t *testing.T) {
	if !syncer.isStop {
		logger.Debug().Msg("stubsyncer stop")
		syncer.hf.stop()
		syncer.hf = nil
		syncer.bf.stop()
		syncer.bf = nil
		syncer.isStop = true
	}
}

func (syncer *StubSyncer) handleMessage(t *testing.T, inmsg interface{}, responseErr error) {
	switch msg := inmsg.(type) {
	//p2p role
	case *message.GetHashes: //from HashFetcher
		blkHashes, _ := syncer.remoteChain.GetHashes(msg.PrevInfo, msg.Count)

		assert.Equal(t, len(blkHashes), int(msg.Count))
		rsp := &message.GetHashesRsp{msg.PrevInfo, blkHashes, uint64(len(blkHashes)), responseErr}

		syncer.hf.GetHahsesRsp(rsp)

	case *message.GetPeers: //from BlockFetcher
		rspMsg := makePeerReply(syncer.stubPeers)
		syncer.testhub.sendReply(StubHubResult{rspMsg, nil})

	case *message.GetBlockChunks:
		syncer.GetBlockChunks(t, msg)

	case *message.AddBlock:
		syncer.AddBlock(t, msg, responseErr)

	case *message.SyncStop:
		syncer.stop(t)
	case *message.CloseFetcher:
		if msg.FromWho == NameHashFetcher {
			syncer.hf.stop()
			syncer.hf = nil
		} else if msg.FromWho == NameBlockFetcher {
			syncer.bf.stop()
			syncer.bf = nil
		} else {
			logger.Error().Msg("invalid closing module message to syncer")
		}
	default:
		t.Error("invalid syncer message")
	}
}

func (syncer *StubSyncer) findStubPeer(peerID peer.ID) *StubPeer {
	for _, tmpPeer := range syncer.stubPeers {
		peerIDStr := string(tmpPeer.addr.PeerID)
		if strings.Compare(peerIDStr, string(peerID)) == 0 {
			return tmpPeer
		}
	}

	return nil
}

func (syncer *StubSyncer) GetBlockChunks(t *testing.T, msg *message.GetBlockChunks) {
	stubPeer := syncer.findStubPeer(msg.ToWhom)

	assert.True(t, stubPeer != nil, "peer exist")

	//send reply
	blocks, err := stubPeer.blockChain.GetBlocks(msg.Hashes)

	msgRsp := &message.GetBlockChunksRsp{ToWhom: msg.ToWhom, Blocks: blocks, Err: err}

	syncer.bf.handleBlockRsp(msgRsp)
}

func (syncer *StubSyncer) AddBlock(t *testing.T, msg *message.AddBlock, responseError error) {
	msgRsp := &message.AddBlockRsp{BlockNo: msg.Block.GetHeader().BlockNo, BlockHash: msg.Block.GetHash(), Err: responseError}

	syncer.bf.handleBlockRsp(msgRsp)
}

func (syncer *StubSyncer) makeStubPeerSet(count int) {
	syncer.stubPeers = make([]*StubPeer, count)

	for i := 0; i < count; i++ {
		syncer.stubPeers[i] = NewStubPeer(i, uint64(syncer.remoteChain.best), syncer.remoteChain)
	}
}

func NewStubPeer(idx int, lastNo uint64, blockChain *StubBlockChain) *StubPeer {
	stubPeer := &StubPeer{}

	peerIDBytes := []byte(fmt.Sprintf("peer-%d", idx))
	stubPeer.addr = &types.PeerAddress{PeerID: peerIDBytes}
	stubPeer.lastBlk = &types.NewBlockNotice{BlockNo: lastNo}
	stubPeer.state = types.RUNNING

	stubPeer.blockChain = blockChain

	return stubPeer
}

func makePeerReply(stubPeers []*StubPeer) *message.GetPeersRsp {
	count := len(stubPeers)
	peerAddrs := make([]*types.PeerAddress, count)
	blockNotices := make([]*types.NewBlockNotice, count)
	states := make([]types.PeerState, count)

	for i, p := range stubPeers {
		peerAddrs[i] = p.addr
		blockNotices[i] = p.lastBlk
		states[i] = p.state
	}

	return &message.GetPeersRsp{Peers: peerAddrs, LastBlks: blockNotices, States: states}
}

func (syncer *StubSyncer) getResultFromHashFetcher() *HashSet {
	hashSet := <-syncer.hf.resultCh
	return hashSet
}
