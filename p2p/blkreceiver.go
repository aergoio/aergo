/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	"time"
)

// BlocksChunkReceiver is send p2p getBlocksRequest to target peer and receive p2p responses till all requestes blocks are received
// It will send response actor message if all blocks are received or failed to receive, but not send response if timeout expired.
type BlocksChunkReceiver struct {
	context actor.Context
	requestID MsgID

	peer RemotePeer
	actor ActorService

	blockHashes []message.BlockHash
	timeout time.Time
	finished bool

	got []*types.Block
	offset int
}

func NewBlockReceiver(context actor.Context, peer RemotePeer, blockHashes []message.BlockHash, ttl time.Duration) *BlocksChunkReceiver {
	timeout := time.Now().Add(ttl)
	return &BlocksChunkReceiver{context: context, peer:peer, blockHashes:blockHashes, timeout:timeout, got:make([]*types.Block, len(blockHashes))}
}

func (br *BlocksChunkReceiver) StartGet() {
	hashes := make([][]byte, len(br.blockHashes))
	for i, hash := range br.blockHashes {
		hashes[i] = ([]byte)(hash)
	}
	// create message data
	req := &types.GetBlockRequest{Hashes: hashes}
	mo := br.peer.MF().newMsgBlockRequestOrder(br.ReceiveResp, GetBlocksRequest, req)
	br.peer.sendMessage(mo)
	br.requestID = mo.GetMsgID()
}

// ReceiveResp must be called just in read go routine
func (br *BlocksChunkReceiver) ReceiveResp(msg Message, msgBody proto.Message) (ret bool) {
	ret = true
	// timeout
	if br.finished || br.timeout.Before(time.Now()) {
		// silently ignore already finished job
		//br.context.Respond(&message.GetBlockChunksRsp{ToWhom:br.peer.ID(), Err:message.RemotePeerFailError})
		br.finished = true
		br.peer.consumeRequest(br.requestID)
		return
	}
	// remote peer response failure
	body :=  msgBody.(*types.GetBlockResponse)
	if body.Status != types.ResultStatus_OK || len(body.Blocks) == 0 {
		br.context.Respond(&message.GetBlockChunksRsp{ToWhom:br.peer.ID(), Err:message.RemotePeerFailError})
		br.finished = true
		br.peer.consumeRequest(br.requestID)
		return
	}

	// add to Got
	for _, block := range body.Blocks {
		// unexpected block
		if !bytes.Equal(br.blockHashes[br.offset], block.Hash) {
			br.context.Respond(&message.GetBlockChunksRsp{ToWhom:br.peer.ID(), Err:message.UnexpectedBlockError})
			br.finished = true
			br.peer.consumeRequest(br.requestID)
			return
		}
		br.got[br.offset] = block
		br.offset++
		// check overflow
		if br.offset >= len(br.got) {
			br.context.Respond(&message.GetBlockChunksRsp{ToWhom:br.peer.ID(), Blocks:br.got, Err:nil})
			br.finished = true
			br.peer.consumeRequest(br.requestID)
			return
		}
	}
	// is it end?
	if !body.HasNext {
		if br.offset < len(br.got) {
			br.context.Respond(&message.GetBlockChunksRsp{ToWhom:br.peer.ID(), Err:message.MissingHashError})
			// not all blocks were filled. this is error
		} else {
			br.context.Respond(&message.GetBlockChunksRsp{ToWhom:br.peer.ID(), Blocks:br.got, Err:nil})
		}
		br.finished = true
		br.peer.consumeRequest(br.requestID)
	}
	return
}