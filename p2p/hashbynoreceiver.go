/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"time"

	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
)

// BlocksChunkReceiver is send p2p getBlocksRequest to target peer and receive p2p responses till all requests blocks are received
// It will send response actor message if all blocks are received or failed to receive, but not send response if timeout expired.
type BlockHashByNoReceiver struct {
	syncerSeq uint64
	requestID p2pcommon.MsgID

	peer  p2pcommon.RemotePeer
	actor p2pcommon.ActorService

	blockNo  types.BlockNo
	timeout  time.Time
	finished bool

	got message.BlockHash
}

func NewBlockHashByNoReceiver(actor p2pcommon.ActorService, peer p2pcommon.RemotePeer, seq uint64, blockNo types.BlockNo, ttl time.Duration) *BlockHashByNoReceiver {
	timeout := time.Now().Add(ttl)
	return &BlockHashByNoReceiver{syncerSeq: seq, actor: actor, peer: peer, blockNo: blockNo, timeout: timeout}
}

func (br *BlockHashByNoReceiver) StartGet() {
	// create message data
	req := &types.GetHashByNo{BlockNo: br.blockNo}
	mo := br.peer.MF().NewMsgRequestOrderWithReceiver(br.ReceiveResp, p2pcommon.GetHashByNoRequest, req)
	br.requestID = mo.GetMsgID()
	br.peer.SendMessage(mo)
}

// ReceiveResp must be called just in read go routine
func (br *BlockHashByNoReceiver) ReceiveResp(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) (ret bool) {
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
	body := msgBody.(*types.GetHashByNoResponse)
	if body.Status != types.ResultStatus_OK {
		br.actor.TellRequest(message.SyncerSvc, &message.GetHashByNoRsp{Seq: br.syncerSeq, BlockHash: nil, Err: message.RemotePeerFailError})
		br.finished = true
		br.peer.ConsumeRequest(br.requestID)
		return
	}
	br.got = body.BlockHash
	br.actor.TellRequest(message.SyncerSvc, &message.GetHashByNoRsp{Seq: br.syncerSeq, BlockHash: br.got})
	br.finished = true
	br.peer.ConsumeRequest(br.requestID)
	return
}
