package syncer

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type TestSyncer struct {
	ctx *types.SyncContext

	hf   *HashFetcher
	bfCh chan *HashSet
}

type TestBlockChain struct {
	best   int
	hashes []([]byte)
	blocks []*types.Block

	topBlock *types.Block
}

func NewTestBlockChain() *TestBlockChain {
	tchain := &TestBlockChain{best: -1}

	tchain.hashes = make([][]byte, 0)
	tchain.blocks = make([]*types.Block, 0)
}

func (tchain *TestBlockChain) genAddBlock() {
	newBlock := types.NewBlock(tchain.topBlock, nil, nil, nil, nil, time.Now().UnixNano())
	tchain.addBlock(newBlock)

	time.Sleep(time.Nanosecond * 3)
}

func (tchain *TestBlockChain) addBlock(newBlock *types.Block) {
	tchain.best += 1
	tchain.hashes[tchain.best] = newBlock.GetHash()
	tchain.blocks[tchain.best] = newBlock
}

func initTestBlockChain(size int) *TestBlockChain {
	chain := NewTestBlockChain()

	//load initial blocks
	for i := 0; i < size; i++ {
		chain.genAddBlock()
	}

	return chain
}

func NewTestSyncer(ctx *types.SyncContext, useHashFetcher bool, useBlockFetcher bool, remoteChain *TestBlockChain) *TestSyncer {
	syncer := &TestSyncer{ctx: ctx}

	if !useBlockFetcher {
		syncer.bfCh = make(chan *HashSet)
	}

	syncer.hf = newHashFetcher(ctx, nil, syncer.bfCh, syncer.sendReqHashToCh)

	return syncer
}

func (tsyncer *TestSyncer) sendReqHashToCh() {
}

func (tsyncer *TestSyncer) getMessageFromHf() {
}

func (tsyncer *TestSyncer) sendMessageToHf() {
}

/*
func TestHashFetcher_normal(t *testing.T) {
	//make remoteBlockChain
	remoteChain := initTestBlockChain(10)
	ancestor := &types.BlockInfo{}

	ctx := types.NewSyncCtx("p1", 10, 1)
	//ctx의 ancestor set
	ctx.SetAncestor(&types.BlockInfo{ancestor.Hash, ancestor.No})

	testSyncer := NewTestSyncer(ctx, true, false, remoteChain)
	hf := testSyncer.hf

	//receive GetHash message
	msg := testSyncer.getMessageFromHf()

	//reply GetHashRsp
	testSyncer.sendMessageToHf()
}
*/
