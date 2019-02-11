/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	"time"
)

// BlocksChunkReceiver is send p2p getBlocksRequest to target peer and receive p2p responses till all requestes blocks are received
// It will send response actor message if all blocks are received or failed to receive, but not send response if timeout expired.
type BlockHashByNoReceiver struct {
	requestID p2pcommon.MsgID

	peer  RemotePeer
	actor ActorService

	blockNo  types.BlockNo
	timeout  time.Time
	finished bool

	got message.BlockHash
}

func NewBlockHashByNoReceiver(actor ActorService, peer RemotePeer, blockNo types.BlockNo, ttl time.Duration) *BlockHashByNoReceiver {
	timeout := time.Now().Add(ttl)
	return &BlockHashByNoReceiver{actor: actor, peer: peer, blockNo: blockNo, timeout: timeout}
}

func (br *BlockHashByNoReceiver) StartGet() {
	// create message data
	req := &types.GetHashByNo{BlockNo: br.blockNo}
	mo := br.peer.MF().newMsgBlockRequestOrder(br.ReceiveResp, GetHashByNoRequest, req)
	br.requestID = mo.GetMsgID()
	br.peer.sendMessage(mo)
}

// ReceiveResp must be called just in read go routine
func (br *BlockHashByNoReceiver) ReceiveResp(msg p2pcommon.Message, msgBody proto.Message) (ret bool) {
	ret = true
	// timeout
	if br.finished || br.timeout.Before(time.Now()) {
		// silently ignore already finished job
		//br.actor.TellRequest(message.SyncerSvc,&message.GetBlockChunksRsp{ToWhom:br.peer.ID(), Err:message.RemotePeerFailError})
		br.finished = true
		br.peer.consumeRequest(br.requestID)
		return
	}
	// remote peer response failure
	body := msgBody.(*types.GetHashByNoResponse)
	if body.Status != types.ResultStatus_OK {
		br.actor.TellRequest(message.SyncerSvc, &message.GetHashByNoRsp{BlockHash: nil, Err: message.RemotePeerFailError})
		br.finished = true
		br.peer.consumeRequest(br.requestID)
		return
	}
	br.got = body.BlockHash
	br.actor.TellRequest(message.SyncerSvc, &message.GetHashByNoRsp{BlockHash: br.got})
	br.finished = true
	br.peer.consumeRequest(br.requestID)
	return
}
