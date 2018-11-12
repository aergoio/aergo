package syncer

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type StubSyncer struct {
	ctx *types.SyncContext

	testhub *StubHub

	hf *HashFetcher
	bf *BlockFetcher

	bfCh chan *HashSet

	remoteChain *StubBlockChain

	isStopHf bool
}

type StubBlockChain struct {
	best   int
	hashes []([]byte)
	blocks []*types.Block

	topBlock *types.Block
}

var (
	ErrNotExitHash = errors.New("not exist hash")
)

func NewStubBlockChain() *StubBlockChain {
	tchain := &StubBlockChain{best: -1}

	tchain.hashes = make([][]byte, 1024)
	tchain.blocks = make([]*types.Block, 1024)

	return tchain
}

func (tchain *StubBlockChain) genAddBlock() {
	newBlock := types.NewBlock(tchain.topBlock, nil, nil, nil, nil, time.Now().UnixNano())
	tchain.addBlock(newBlock)

	time.Sleep(time.Nanosecond * 3)
}

func (tchain *StubBlockChain) addBlock(newBlock *types.Block) {
	tchain.best += 1
	tchain.hashes[tchain.best] = newBlock.BlockHash()
	tchain.blocks[tchain.best] = newBlock
}

func (tchain *StubBlockChain) GetHashes(prevInfo *types.BlockInfo, count uint64) ([][]byte, error) {
	if tchain.best < int(prevInfo.No+count) {
		return nil, ErrNotExitHash
	}

	start := prevInfo.No + 1
	resHashes := tchain.hashes[start : start+count]

	return resHashes, nil
}

func (tchain *StubBlockChain) GetBlockInfo(no uint64) *types.BlockInfo {
	return &types.BlockInfo{tchain.hashes[no], no}
}

func initStubBlockChain(size int) *StubBlockChain {
	chain := NewStubBlockChain()

	//load initial blocks
	for i := 0; i < size; i++ {
		chain.genAddBlock()
	}

	return chain
}

var (
	MaxTestReqSize = uint64(3)
)

func NewStubSyncer(ctx *types.SyncContext, useHashFetcher bool, useBlockFetcher bool, remoteChain *StubBlockChain) *StubSyncer {
	syncer := &StubSyncer{ctx: ctx}
	syncer.bfCh = make(chan *HashSet)
	syncer.testhub = NewStubHub()

	syncer.hf = newHashFetcher(ctx, syncer.testhub, syncer.bfCh, MaxTestReqSize)
	syncer.remoteChain = remoteChain

	return syncer
}

func (tsyncer *StubSyncer) handleMessage(t *testing.T, msg interface{}, responseErr error) {
	switch inmsg := msg.(type) {
	//p2p role
	case *message.GetHashes:
		hashes, _ := tsyncer.remoteChain.GetHashes(inmsg.PrevInfo, inmsg.Count)

		assert.Equal(t, len(hashes), int(inmsg.Count))
		blkHashes := make([]message.BlockHash, 0)
		for _, hash := range hashes {
			blkHashes = append(blkHashes, hash)
		}
		rsp := &message.GetHashesRsp{inmsg.PrevInfo, blkHashes, uint64(len(hashes)), responseErr}

		tsyncer.hf.GetHahsesRsp(rsp)
	case *message.SyncStop:
		tsyncer.hf.stop()
		tsyncer.isStopHf = true
	case *message.CloseFetcher:
		if inmsg.FromWho == NameHashFetcher {
			tsyncer.hf.stop()
		} else if inmsg.FromWho == NameBlockFetcher {
			tsyncer.bf.stop()
		} else {
			logger.Error().Msg("invalid closing module message to syncer")
		}
	default:
		t.Error("invalid syncer message")
	}
}

func (tsyncer *StubSyncer) getResultFromHashFetcher() *HashSet {
	hashSet := <-tsyncer.hf.resultCh
	return hashSet
}

func (tsyncer *StubSyncer) isStoppedHashFetcher() bool {
	return tsyncer.isStopHf
}

func TestHashFetcher_normal(t *testing.T) {
	//make remoteBlockChain
	remoteChain := initStubBlockChain(10)

	ancestor := remoteChain.GetBlockInfo(0)

	ctx := types.NewSyncCtx("p1", 5, 1)
	ctx.SetAncestor(ancestor)

	stubSyncer := NewStubSyncer(ctx, true, false, remoteChain)
	stubSyncer.hf.Start()

	//hashset 1~3, 4~5
	//receive GetHash message
	msg := stubSyncer.testhub.GetMessage()
	assert.IsTypef(t, &message.GetHashes{}, msg, "invalid message from hf")
	stubSyncer.handleMessage(t, msg, nil)

	//when pop result msg, hashfetcher send new request
	resHashSet := stubSyncer.getResultFromHashFetcher()
	assert.Equal(t, int(MaxTestReqSize), resHashSet.Count)

	msg = stubSyncer.testhub.GetMessage()
	stubSyncer.handleMessage(t, msg, nil)

	//when pop result msg, hashfetcher send new request
	resHashSet = stubSyncer.getResultFromHashFetcher()
	assert.Equal(t, 2, resHashSet.Count)

	//receive close hashfetcher message
	msg = stubSyncer.testhub.GetMessage()
	assert.IsTypef(t, &message.CloseFetcher{}, msg, "need syncer close hashfetcher msg")
	stubSyncer.handleMessage(t, msg, nil)
	assert.True(t, stubSyncer.hf.finished, "hashfetcher finished")
}

func TestHashFetcher_ResponseError(t *testing.T) {
	//make remoteBlockChain
	remoteChain := initStubBlockChain(10)

	ancestor := remoteChain.GetBlockInfo(0)

	ctx := types.NewSyncCtx("p1", 5, 1)
	ctx.SetAncestor(ancestor)

	stubSyncer := NewStubSyncer(ctx, true, false, remoteChain)
	stubSyncer.hf.Start()

	//hashset 2~4, 5~7, 8~9
	//receive GetHash message
	msg := stubSyncer.testhub.GetMessage()
	assert.IsTypef(t, &message.GetHashes{}, msg, "invalid message from hf")
	stubSyncer.handleMessage(t, msg, ErrGetHashesRspError)

	//stop
	msg = stubSyncer.testhub.GetMessage()
	assert.IsTypef(t, &message.SyncStop{}, msg, "need syncer stop msg")
	stubSyncer.handleMessage(t, msg, nil)

	assert.True(t, stubSyncer.isStoppedHashFetcher(), "hashfetcher finished")
}
