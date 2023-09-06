/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

// GetTxsReceiver is send p2p getTXsRequest to target peer and receive p2p responses till all requests transactions are received
// syncer actor already dropped wait before.
type GetTxsReceiver struct {
	requestID p2pcommon.MsgID
	logger    *log.Logger

	peer  p2pcommon.RemotePeer
	actor p2pcommon.ActorService
	sm    p2pcommon.SyncManager

	ids    []types.TxID
	hashes [][]byte
	missed [][]byte

	timeout  time.Time
	finished bool
	status   receiverStatus

	inOffset       int
	offset         int
	sent           int
	senderFinished chan interface{}
}

func NewGetTxsReceiver(actor p2pcommon.ActorService, peer p2pcommon.RemotePeer, sm p2pcommon.SyncManager, logger *log.Logger, txIDs []types.TxID, ttl time.Duration) *GetTxsReceiver {
	timeout := time.Now().Add(ttl)
	ids := make([]types.TxID, len(txIDs))
	hashes := make([][]byte, len(txIDs))
	for i := range txIDs {
		ids[i] = txIDs[i]
		hashes[i] = ids[i][:]
	}
	return &GetTxsReceiver{actor: actor, peer: peer, sm: sm, ids: ids, hashes: hashes, timeout: timeout, logger: logger}
}

func (br *GetTxsReceiver) StartGet() {
	// create message data
	req := &types.GetTransactionsRequest{Hashes: br.hashes}
	mo := br.peer.MF().NewMsgRequestOrderWithReceiver(br.ReceiveResp, p2pcommon.GetTXsRequest, req)
	br.peer.SendMessage(mo)
	br.requestID = mo.GetMsgID()
}

// ReceiveResp must be called just in read go routine
func (br *GetTxsReceiver) ReceiveResp(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) (ret bool) {
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

func (br *GetTxsReceiver) handleInWaiting(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
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
		if respBody.GetStatus() == types.ResultStatus_RESOURCE_EXHAUSTED {
			br.sm.RetryGetTx(br.peer, br.hashes)
		}
		return
	}
	// remote peer response malformed data.
	body, ok := msgBody.(*types.GetTransactionsResponse)
	if !ok || len(body.Txs) == 0 {
		br.cancelReceiving(message.MissingHashError, false)
		return
	}

	// add to Got
	for _, tx := range body.Txs {
		// It also error that response has more blocks than expected(=requested).
		if br.offset >= len(br.hashes) {
			br.cancelReceiving(message.TooManyBlocksError, body.HasNext)
			return
		}
		// missing tx
		for !bytes.Equal(br.hashes[br.offset], tx.Hash) {
			br.logger.Trace().Str("expect", enc.ToString(br.hashes[br.offset])).Str("received", enc.ToString(tx.Hash)).Int("offset", br.offset).Msg("expected hash was missing")
			br.missed = append(br.missed, tx.Hash)
			br.offset++
			if br.offset >= len(br.hashes) {
				br.cancelReceiving(message.UnexpectedBlockError, body.HasNext)
				return
			}
		}
		br.actor.SendRequest(message.MemPoolSvc, &message.MemPoolPut{Tx: tx})
		br.sent++
		br.offset++
	}
	// remote peer hopefully sent last chunk
	if !body.HasNext {
		br.finishReceiver()
	}
	return
}

// cancelReceiving is cancel wait for receiving and send syncer the failure result.
// not all part of response is received, it wait remaining (and useless) response. It is assumed cancelling is not frequently occur
func (br *GetTxsReceiver) cancelReceiving(err error, hasNext bool) {
	br.status = receiverStatusCanceled
	br.logger.Info().Str(p2putil.LogOrgReqID, br.requestID.String()).Err(err).Msg("tx receiver canceled by error")
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
func (br *GetTxsReceiver) finishReceiver() {
	br.status = receiverStatusFinished
	br.peer.ConsumeRequest(br.requestID)
	br.logger.Debug().Int("mpSent", br.sent).Int("missed", len(br.missed)).Str(p2putil.LogOrgReqID, br.requestID.String()).Msg("tx receiver finished")
}

// ignoreMsg is silently ignore following responses, which is not useless anymore.
func (br *GetTxsReceiver) ignoreMsg(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	body, ok := msgBody.(*types.GetTransactionsResponse)
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
