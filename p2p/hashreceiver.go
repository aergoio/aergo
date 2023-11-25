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

// BlockHashesReceiver is send p2p GetHashesRequest to target peer and receive p2p responses till all requested hashes are received
// It will send response actor message if all hashes are received or failed to receive, but not send response if timeout expired.
type BlockHashesReceiver struct {
	syncerSeq uint64
	requestID p2pcommon.MsgID

	peer  p2pcommon.RemotePeer
	actor p2pcommon.ActorService

	prevBlock *types.BlockInfo
	count     int
	timeout   time.Time
	finished  bool
	status    receiverStatus

	reqCnt         int
	got            []message.BlockHash
	senderFinished chan interface{}
}

func NewBlockHashesReceiver(actor p2pcommon.ActorService, peer p2pcommon.RemotePeer, seq uint64, req *message.GetHashes, ttl time.Duration) *BlockHashesReceiver {
	timeout := time.Now().Add(ttl)
	return &BlockHashesReceiver{syncerSeq: seq, actor: actor, peer: peer, prevBlock: req.PrevInfo, count: int(req.Count), timeout: timeout, reqCnt: int(req.Count), got: make([]message.BlockHash, 0, int(req.Count))}
}

func (br *BlockHashesReceiver) StartGet() {
	// create message data
	req := &types.GetHashesRequest{PrevHash: br.prevBlock.Hash, PrevNumber: br.prevBlock.No, Size: uint64(br.count)}
	mo := br.peer.MF().NewMsgRequestOrderWithReceiver(br.ReceiveResp, p2pcommon.GetHashesRequest, req)
	br.peer.SendMessage(mo)
	br.requestID = mo.GetMsgID()
}

// ReceiveResp must be called just in read go routine
func (br *BlockHashesReceiver) ReceiveResp(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) (ret bool) {
	// TODO this code is exact copy of BlocksChunkReceiver, so be lots of other codes in this file. consider refactoring
	ret = true
	switch br.status {
	case receiverStatusWaiting:
		br.handleInWaiting(msg, msgBody)
	case receiverStatusCanceled:
		br.ignoreMsg(msg, msgBody)
		return
	case receiverStatusFinished:
		fallthrough
	default:
		return
	}
	return
}

func (br *BlockHashesReceiver) handleInWaiting(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	// consuming request id when timeout, no more resp expected (i.e. hasNext == false ) or malformed body.
	// timeout
	if br.timeout.Before(time.Now()) {
		// silently ignore already status job
		br.finishReceiver()
		return
	}
	// malformed responses means that later responses will be also malformed..
	respBody, ok := msgBody.(types.ResponseMessage)
	if !ok || respBody.GetStatus() != types.ResultStatus_OK {
		br.cancelReceiving(message.RemotePeerFailError, false)
		return
	}

	// remote peer response failure
	body, ok := msgBody.(*types.GetHashesResponse)
	if !ok || len(body.Hashes) == 0 {
		br.cancelReceiving(message.MissingHashError, false)
		return
	}

	// add to Got
	for _, bHash := range body.Hashes {
		if len(bHash) != types.HashIDLength {
			br.cancelReceiving(message.WrongBlockHashError, body.HasNext)
			return
		}
		// It also error that response has more hashes than expected(=requested).
		if len(br.got) >= br.reqCnt {
			br.cancelReceiving(message.TooManyBlocksError, body.HasNext)
			return
		}
		br.got = append(br.got, bHash)
	}
	// remote peer hopefully sent last part
	if !body.HasNext {
		br.actor.TellRequest(message.SyncerSvc, &message.GetHashesRsp{Seq: br.syncerSeq, Hashes: br.got, PrevInfo: br.prevBlock, Count: uint64(len(br.got))})
		br.finishReceiver()
	}
	return
}

// cancelReceiving is cancel wait for receiving and send syncer the failure result.
// not all part of response is received, it wait remaining (and useless) response. It is assumed canceling is not frequently occur
func (br *BlockHashesReceiver) cancelReceiving(err error, hasNext bool) {
	br.status = receiverStatusCanceled
	br.actor.TellRequest(message.SyncerSvc,
		&message.GetHashesRsp{Seq: br.syncerSeq, PrevInfo: br.prevBlock, Err: err})

	// check time again. since negative duration of timer will not fire channel.
	interval := br.timeout.Sub(time.Now())
	if !hasNext || interval <= 0 {
		// if remote peer will not send partial response anymore. it it actually same as finish.
		br.finishReceiver()
	} else {
		// canceling in the middle of responses
		br.senderFinished = make(chan interface{})
		go func() {
			timer := time.NewTimer(interval)
			select {
			case <-timer.C:
				break
			case <-br.senderFinished:
				break
			}
			br.peer.ConsumeRequest(br.requestID)
		}()
	}
}

// finishReceiver is to cancel works, assuming cancellations are not frequently occur
func (br *BlockHashesReceiver) finishReceiver() {
	br.status = receiverStatusFinished
	br.peer.ConsumeRequest(br.requestID)
}

// ignoreMsg is silently ignore following responses, which is not useless anymore.
func (br *BlockHashesReceiver) ignoreMsg(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	body, ok := msgBody.(*types.GetBlockResponse)
	if !ok {
		return
	}
	if !body.HasNext {
		// really status from remote peer
		select {
		case br.senderFinished <- struct{}{}:
		default:
		}
	}
}
