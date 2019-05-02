/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/subproto"
	"github.com/aergoio/aergo/types"
	"time"
)

// BlocksChunkReceiver is send p2p getBlocksRequest to target peer and receive p2p responses till all requestes blocks are received
// It will send response actor message if all blocks are received or failed to receive, but not send response if timeout expired.
type AncestorReceiver struct {
	syncerSeq uint64
	requestID p2pcommon.MsgID

	peer  p2pcommon.RemotePeer
	actor p2pcommon.ActorService

	hashes  [][]byte
	timeout  time.Time
	finished bool
}

func NewAncestorReceiver(actor p2pcommon.ActorService, peer p2pcommon.RemotePeer, seq uint64, hashes [][]byte, ttl time.Duration) *AncestorReceiver {
	timeout := time.Now().Add(ttl)
	return &AncestorReceiver{syncerSeq: seq, actor: actor, peer: peer, hashes: hashes, timeout: timeout}
}

func (br *AncestorReceiver) StartGet() {
	// create message data
	req := &types.GetAncestorRequest{Hashes: br.hashes}
	mo := br.peer.MF().NewMsgBlockRequestOrder(br.ReceiveResp, subproto.GetAncestorRequest, req)
	br.requestID = mo.GetMsgID()
	br.peer.SendMessage(mo)
}

// ReceiveResp must be called just in read go routine
func (br *AncestorReceiver) ReceiveResp(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) (ret bool) {
	ret = true
	// timeout
	if br.finished || br.timeout.Before(time.Now()) {
		br.finished = true
		br.peer.ConsumeRequest(br.requestID)
		return
	}
	// remote peer response failure
	data := msgBody.(*types.GetAncestorResponse)
	if data.Status != types.ResultStatus_OK {
		br.actor.TellRequest(message.SyncerSvc, &message.GetSyncAncestorRsp{Seq:br.syncerSeq, Ancestor: nil})
		br.finished = true
		br.peer.ConsumeRequest(br.requestID)
		return
	}
	ancestor := &types.BlockInfo{Hash: data.AncestorHash, No: data.AncestorNo}

	br.actor.TellRequest(message.SyncerSvc, &message.GetSyncAncestorRsp{Seq:br.syncerSeq, Ancestor: ancestor})
	br.finished = true
	br.peer.ConsumeRequest(br.requestID)
	return
}
