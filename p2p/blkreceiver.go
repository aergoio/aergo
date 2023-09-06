/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"time"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
)

// BlocksChunkReceiver is send p2p getBlocksRequest to target peer and receive p2p responses till all requests blocks are received
// It will send response actor message if all blocks are received or failed to receive, but not send response if timeout expired, since
// syncer actor already dropped wait before.
type BlocksChunkReceiver struct {
	syncerSeq uint64
	requestID p2pcommon.MsgID

	peer  p2pcommon.RemotePeer
	actor p2pcommon.ActorService

	blockHashes []message.BlockHash
	timeout     time.Time
	finished    bool
	status      receiverStatus

	got            []*types.Block
	offset         int
	senderFinished chan interface{}
}

type receiverStatus int32

const (
	receiverStatusWaiting receiverStatus = iota
	receiverStatusCanceled
	receiverStatusFinished
)

func NewBlockReceiver(actor p2pcommon.ActorService, peer p2pcommon.RemotePeer, seq uint64, blockHashes []message.BlockHash, ttl time.Duration) *BlocksChunkReceiver {
	timeout := time.Now().Add(ttl)
	return &BlocksChunkReceiver{syncerSeq: seq, actor: actor, peer: peer, blockHashes: blockHashes, timeout: timeout, got: make([]*types.Block, len(blockHashes))}
}

func (br *BlocksChunkReceiver) StartGet() {
	hashes := make([][]byte, len(br.blockHashes))
	for i, hash := range br.blockHashes {
		hashes[i] = ([]byte)(hash)
	}
	// create message data
	req := &types.GetBlockRequest{Hashes: hashes}
	mo := br.peer.MF().NewMsgRequestOrderWithReceiver(br.ReceiveResp, p2pcommon.GetBlocksRequest, req)
	br.peer.SendMessage(mo)
	br.requestID = mo.GetMsgID()
}

// ReceiveResp must be called just in read go routine
func (br *BlocksChunkReceiver) ReceiveResp(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) (ret bool) {
	// cases in waiting
	//   normal not status => wait
	//   normal status (last response)  => finish
	//   abnormal resp (no following resp expected): hasNext is true => cancel
	//   abnormal resp (following resp expected): hasNext is false, or invalid resp data type (maybe remote peer is totally broken) => cancel finish
	// case in status or status
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

func (br *BlocksChunkReceiver) handleInWaiting(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
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
	// remote peer response malformed data.
	body, ok := msgBody.(*types.GetBlockResponse)
	if !ok || len(body.Blocks) == 0 {
		br.cancelReceiving(message.MissingHashError, false)
		return
	}

	// add to Got
	for _, block := range body.Blocks {
		// It also error that response has more blocks than expected(=requested).
		if br.offset >= len(br.got) {
			br.cancelReceiving(message.TooManyBlocksError, body.HasNext)
			return
		}
		// unexpected block
		if !bytes.Equal(br.blockHashes[br.offset], block.Hash) {
			br.cancelReceiving(message.UnexpectedBlockError, body.HasNext)
			return
		}
		if block.Size() > int(chain.MaxBlockSize()) {
			br.cancelReceiving(message.TooBigBlockError, body.HasNext)
			return
		}
		br.got[br.offset] = block
		br.offset++
	}
	// remote peer hopefully sent last chunk
	if !body.HasNext {
		if br.offset < len(br.got) {
			// not all blocks were filled. this is error
			br.cancelReceiving(message.TooFewBlocksError, body.HasNext)
		} else {
			br.actor.TellRequest(message.SyncerSvc, &message.GetBlockChunksRsp{Seq: br.syncerSeq, ToWhom: br.peer.ID(), Blocks: br.got, Err: nil})
			br.finishReceiver()
		}
	}
	return
}

// cancelReceiving is cancel wait for receiving and send syncer the failure result.
// not all part of response is received, it wait remaining (and useless) response. It is assumed cancelling is not frequently occur
func (br *BlocksChunkReceiver) cancelReceiving(err error, hasNext bool) {
	br.status = receiverStatusCanceled
	br.actor.TellRequest(message.SyncerSvc,
		&message.GetBlockChunksRsp{Seq: br.syncerSeq, ToWhom: br.peer.ID(), Err: err})

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
func (br *BlocksChunkReceiver) finishReceiver() {
	br.status = receiverStatusFinished
	br.peer.ConsumeRequest(br.requestID)
}

// ignoreMsg is silently ignore following responses, which is not useless anymore.
func (br *BlocksChunkReceiver) ignoreMsg(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
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
