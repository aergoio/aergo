/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"strconv"
	"sync"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc/proto"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/message"
	"github.com/pkg/errors"
)

// ConcurrentClusterInfoReceiver is a struct that manages concurrent requests for cluster information from peers.
// It sends p2p GetClusterInfo requests to connected peers and collects responses until either:
//   - A successful response is received
//   - The required number of responses are collected
//   - The operation times out
//
// The initial implementation processes requests and responses sequentially (one by one).
type ConcurrentClusterInfoReceiver struct {
	logger *log.Logger            // Logger for recording events and errors
	mf     p2pcommon.MoFactory    // Factory for creating message objects
	peers  []p2pcommon.RemotePeer // List of connected peers to query

	mutex sync.Mutex // Mutex to protect concurrent access to the struct

	sent    map[p2pcommon.MsgID]p2pcommon.RemotePeer // Tracks sent requests by message ID
	sentCnt int                                      // Count of total sent requests

	req     *message.GetCluster // The cluster information request message
	ttl     time.Duration       // Time-to-live duration for the operation
	timeout time.Time           // Absolute time when the operation will timeout

	respCnt      int // Count of received responses
	requiredResp int // Minimum required responses before completing
	// Successful responses by peer ID
	succResps map[types.PeerID]*types.GetClusterInfoResponse

	status   receiverStatus // Current state of the receiver
	finished chan bool      // Channel to signal operation completion
}

func NewConcClusterInfoReceiver(actor p2pcommon.ActorService, mf p2pcommon.MoFactory, peers []p2pcommon.RemotePeer, ttl time.Duration, req *message.GetCluster, logger *log.Logger) *ConcurrentClusterInfoReceiver {
	// TODO the value requiredResp of can cause trouble.
	// Only the members of cluster can give the cluster information. There is a possibility of calculating the quorum because it sends requests to all connected peers regardless of membership. There is another problem. There is no cluster information in node when it send a request, so it is difficult to get the exact quorum because it also has no number of members.
	r := &ConcurrentClusterInfoReceiver{logger: logger, mf: mf, peers: peers, ttl: ttl, req: req,
		requiredResp: len(peers)/2 + 1,
		succResps:    make(map[types.PeerID]*types.GetClusterInfoResponse),
		sent:         make(map[p2pcommon.MsgID]p2pcommon.RemotePeer), finished: make(chan bool)}

	return r
}

func (r *ConcurrentClusterInfoReceiver) StartGet() {
	r.timeout = time.Now().Add(r.ttl)
	// create message data
	// send message to first peer
	go func() {
		r.mutex.Lock()
		if !r.trySendAllPeers() {
			r.cancelReceiving(errors.New("no live peers"), false)
			r.mutex.Unlock()
			return
		}
		r.mutex.Unlock()
		r.runExpireTimer()
	}()
}

func (r *ConcurrentClusterInfoReceiver) runExpireTimer() {
	t := time.NewTimer(r.ttl)
	select {
	case <-t.C:
		// time is up. check or collect mid result.
		r.mutex.Lock()
		defer r.mutex.Unlock()
		if r.status == receiverStatusWaiting {
			r.finishReceiver(nil)
		}
	case <-r.finished:
	}
	r.logger.Debug().Msg("expire timer finished")
}

func (r *ConcurrentClusterInfoReceiver) trySendAllPeers() bool {
	r.logger.Debug().Array("peers", p2putil.NewLogPeersMarshaller(r.peers, 10)).Msg("sending get cluster request to connected peers")
	req := &types.GetClusterInfoRequest{BestBlockHash: r.req.BestBlockHash}
	for _, peer := range r.peers {
		if peer.State() == types.RUNNING {
			mo := r.mf.NewMsgRequestOrderWithReceiver(r.ReceiveResp, p2pcommon.GetClusterRequest, req)
			peer.SendMessage(mo)
			r.sent[mo.GetMsgID()] = peer
			r.sentCnt++
		}
	}
	r.logger.Debug().Int("sent", r.sentCnt).Msg("sent get cluster requests")
	return r.sentCnt >= r.requiredResp
}

// ReceiveResp must be called just in read go routine
func (r *ConcurrentClusterInfoReceiver) ReceiveResp(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) (ret bool) {
	// cases in waiting
	//   normal not status => wait
	//   normal status (last response)  => finish
	//   abnormal resp (no following resp expected): hasNext is true => cancel
	//   abnormal resp (following resp expected): hasNext is false, or invalid resp data type (maybe remote peer is totally broken) => cancel finish
	// case in status or status
	ret = true
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// consuming request id at first
	peer, exist := r.sent[msg.OriginalID()]
	if exist {
		delete(r.sent, msg.OriginalID())
		peer.ConsumeRequest(msg.OriginalID())
	} else {
		// TODO report unknown message?
		return
	}

	status := r.status
	switch status {
	case receiverStatusWaiting:
		r.handleInWaiting(peer, msg, msgBody)
		r.respCnt++
		if r.respCnt >= r.sentCnt {
			r.finishReceiver(nil)
		}
	case receiverStatusCanceled:
		fallthrough
	case receiverStatusFinished:
		fallthrough
	default:
		r.ignoreMsg(msg, msgBody)
		return
	}
	return
}

func (r *ConcurrentClusterInfoReceiver) handleInWaiting(peer p2pcommon.RemotePeer, msg p2pcommon.Message, msgBody proto.Message) {
	// timeout: either runExpireTimer() expire or this test is called just once in the case of timeout
	if r.timeout.Before(time.Now()) {
		// silently ignore already finished job
		r.finishReceiver(nil)
		return
	}

	// remote peer response malformed data.
	body, ok := msgBody.(*types.GetClusterInfoResponse)
	if !ok {
		r.logger.Debug().Str(p2putil.LogPeerName, peer.Name()).Stringer(p2putil.LogMsgID, msg.ID()).Msg("get cluster invalid response data")
		return
	} else if len(body.MbrAttrs) == 0 || body.Error != "" {
		r.logger.Debug().Str(p2putil.LogPeerName, peer.Name()).Stringer(p2putil.LogMsgID, msg.ID()).Err(errors.New(body.Error)).Msg("get cluster response empty member")
		return
	}

	r.logger.Debug().Str(p2putil.LogPeerName, peer.Name()).Stringer(p2putil.LogMsgID, msg.ID()).Object("resp", body).Msg("received get cluster response")
	// return the result
	if len(body.Error) != 0 {
		r.logger.Debug().Str(p2putil.LogPeerName, peer.Name()).Stringer(p2putil.LogMsgID, msg.ID()).Err(errors.New(body.Error)).Msg("get cluster response error")
		return
	}
	r.succResps[peer.ID()] = body
}

// cancelReceiving is cancel wait for receiving and return the failure result.
// it wait remaining (and useless) response. It is assumed cancellations are not frequently occur
func (r *ConcurrentClusterInfoReceiver) cancelReceiving(err error, hasNext bool) {
	r.status = receiverStatusCanceled
	r.finishReceiver(err)
}

// finishReceiver is to cancel works, assuming cancellations are not frequently occur
func (r *ConcurrentClusterInfoReceiver) finishReceiver(err error) {
	if r.status == receiverStatusFinished {
		r.logger.Warn().Msg("redundant finish call")
		return
	}
	r.status = receiverStatusFinished
	r.logger.Debug().Msg("finishing receiver")
	r.req.ReplyC <- r.calculate(err)
	close(r.req.ReplyC)
	close(r.finished)
}

// ignoreMsg is silently ignore following responses, which is not useless anymore.
func (r *ConcurrentClusterInfoReceiver) ignoreMsg(msg p2pcommon.Message, msgBody proto.Message) {
	// nothing to do for now
}

func (r *ConcurrentClusterInfoReceiver) calculate(err error) *message.GetClusterRsp {
	rsp := &message.GetClusterRsp{}
	if err != nil {
		rsp.Err = err
	} else if len(r.succResps) < r.requiredResp {
		rsp.Err = errors.New("too few responses: " + strconv.Itoa(len(r.succResps)) + " , required " + strconv.Itoa(r.requiredResp))
	} else {
		r.logger.Debug().Int("respCnt", len(r.succResps)).Msg("calculating collected responses")
		var bestRsp *types.GetClusterInfoResponse = nil
		var bestPid types.PeerID
		for peerId, rsp := range r.succResps {
			if bestRsp == nil || rsp.BestBlockNo > bestRsp.BestBlockNo {
				bestRsp = rsp
				bestPid = peerId
			}
		}
		if bestRsp != nil {
			r.logger.Debug().Stringer(p2putil.LogPeerID, types.LogPeerShort(bestPid)).Object("resp", bestRsp).Msg("chose best response")
			rsp.ClusterID = bestRsp.GetClusterID()
			rsp.ChainID = bestRsp.GetChainID()
			rsp.Members = bestRsp.GetMbrAttrs()
			rsp.HardStateInfo = bestRsp.HardStateInfo
		} else {
			rsp.Err = errors.New("no successful responses")
		}
	}
	return rsp
}
