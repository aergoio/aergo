package syncer

import (
	"fmt"
	"time"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/types"
)

type StubPeer struct {
	addr    *types.PeerAddress
	lastBlk *types.NewBlockNotice
	state   types.PeerState

	blockChain *chain.StubBlockChain

	blockFetched bool //check if called while testing

	timeDelaySec time.Duration

	HookGetBlockChunkRsp func(msgReq *message.GetBlockChunks)
}

var (
	TestMaxBlockFetchSize = 2
	TestMaxHashReqSize    = uint64(3)
)

func NewStubPeer(idx int, lastNo uint64, blockChain *chain.StubBlockChain) *StubPeer {
	stubPeer := &StubPeer{}

	peerIDBytes := []byte(fmt.Sprintf("peer-%d", idx))
	stubPeer.addr = &types.PeerAddress{PeerID: peerIDBytes}
	stubPeer.lastBlk = &types.NewBlockNotice{BlockNo: lastNo}
	stubPeer.state = types.RUNNING

	stubPeer.blockChain = blockChain

	return stubPeer
}
