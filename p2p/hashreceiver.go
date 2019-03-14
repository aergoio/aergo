/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"time"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/subproto"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

// BlocksChunkReceiver is send p2p getBlocksRequest to target peer and receive p2p responses till all requestes blocks are received
// It will send response actor message if all blocks are received or failed to receive, but not send response if timeout expired.
type BlockHashesReceiver struct {
	requestID p2pcommon.MsgID

	peer  p2pcommon.RemotePeer
	actor p2pcommon.ActorService

	prevBlock *types.BlockInfo
	count     int
	timeout   time.Time
	finished  bool

	got    []message.BlockHash
	offset int
}

func NewBlockHashesReceiver(actor p2pcommon.ActorService, peer p2pcommon.RemotePeer, req *message.GetHashes, ttl time.Duration) *BlockHashesReceiver {
	timeout := time.Now().Add(ttl)
	return &BlockHashesReceiver{actor: actor, peer: peer, prevBlock: req.PrevInfo, count: int(req.Count), timeout: timeout, got: make([]message.BlockHash, 0, int(req.Count))}
}

func (br *BlockHashesReceiver) StartGet() {
	// create message data
	req := &types.GetHashesRequest{PrevHash: br.prevBlock.Hash, PrevNumber: br.prevBlock.No, Size: uint64(br.count)}
	mo := br.peer.MF().NewMsgBlockRequestOrder(br.ReceiveResp, subproto.GetHashesRequest, req)
	br.peer.SendMessage(mo)
	br.requestID = mo.GetMsgID()
}

// ReceiveResp must be called just in read go routine
func (br *BlockHashesReceiver) ReceiveResp(msg p2pcommon.Message, msgBody proto.Message) (ret bool) {
	ret = true
	// timeout
	if br.finished || br.timeout.Before(time.Now()) {
		// silently ignore already finished job
		//br.actor.TellRequest(message.SyncerSvc,&message.GetBlockChunksRsp{ToWhom:br.peer.ID(), Err:message.RemotePeerFailError})
		br.finished = true
		br.peer.ConsumeRequest(br.requestID)
		return
	}
	// remote peer response failure
	body := msgBody.(*types.GetHashesResponse)
	if body.Status != types.ResultStatus_OK || len(body.Hashes) == 0 {
		br.actor.TellRequest(message.SyncerSvc, &message.GetHashesRsp{Hashes: nil, PrevInfo: br.prevBlock, Count: 0, Err: message.RemotePeerFailError})
		br.finished = true
		br.peer.ConsumeRequest(br.requestID)
		return
	}

	// add to Got
	for _, block := range body.Hashes {
		// unexpected block
		br.got = append(br.got, block)
		br.offset++
		// check overflow
		if br.offset >= int(br.count) {
			br.actor.TellRequest(message.SyncerSvc, &message.GetHashesRsp{Hashes: br.got, PrevInfo: br.prevBlock, Count: uint64(br.offset)})
			br.finished = true
			br.peer.ConsumeRequest(br.requestID)
			return
		}
	}
	// is it end?
	if !body.HasNext {
		if br.offset < br.count {
			br.actor.TellRequest(message.SyncerSvc, &message.GetHashesRsp{Hashes: br.got, PrevInfo: br.prevBlock, Count: 0, Err: message.MissingHashError})
			// not all blocks were filled. this is error
		} else {
			br.actor.TellRequest(message.SyncerSvc, &message.GetHashesRsp{Hashes: br.got, PrevInfo: br.prevBlock, Count: uint64(len(br.got))})
		}
		br.finished = true
		br.peer.ConsumeRequest(br.requestID)
	}
	return
}
