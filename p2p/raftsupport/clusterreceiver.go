/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"fmt"
	"sync"
	"time"

	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// ClusterInfoReceiver is send p2p getClusterInfo to connected peers and Receive p2p responses one of peers return successful response
// The first version will be simplified version. it send and Receive one by one.
type ClusterInfoReceiver struct {
	mf p2pcommon.MoFactory

	peers  []p2pcommon.RemotePeer
	mutex  sync.Mutex
	sents  map[p2pcommon.MsgID]p2pcommon.RemotePeer
	offset int

	req   *message.GetCluster
	actor p2pcommon.ActorService

	ttl      time.Duration
	timeout  time.Time
	finished bool
	status   receiverStatus

	got            []*types.Block
	senderFinished chan interface{}
}

type receiverStatus int32

const (
	receiverStatusWaiting receiverStatus = iota
	receiverStatusCanceled
	receiverStatusFinished
)

func NewClusterInfoReceiver(actor p2pcommon.ActorService, mf p2pcommon.MoFactory, peers []p2pcommon.RemotePeer, ttl time.Duration, req *message.GetCluster) *ClusterInfoReceiver {
	return &ClusterInfoReceiver{actor: actor, mf: mf, peers: peers, ttl: ttl, req: req, sents: make(map[p2pcommon.MsgID]p2pcommon.RemotePeer)}
}

func (br *ClusterInfoReceiver) StartGet() {
	br.timeout = time.Now().Add(br.ttl)
	// create message data
	// send message to first peer
	go func() {
		br.mutex.Lock()
		defer br.mutex.Unlock()
		if !br.trySendNextPeer() {
			br.cancelReceiving(errors.New("no live peers"), false)
		}
	}()
}

func (br *ClusterInfoReceiver) trySendNextPeer() bool {
	for ; br.offset < len(br.peers); br.offset++ {
		peer := br.peers[br.offset]
		if peer.State() == types.RUNNING {
			br.offset++
			mo := br.mf.NewMsgRequestOrderWithReceiver(br.ReceiveResp, p2pcommon.GetClusterRequest, &types.GetClusterInfoRequest{BestBlockHash: br.req.BestBlockHash})
			peer.SendMessage(mo)
			br.sents[mo.GetMsgID()] = peer
			return true
		}
	}
	return false
}

// ReceiveResp must be called just in read go routine
func (br *ClusterInfoReceiver) ReceiveResp(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) (ret bool) {
	// cases in waiting
	//   normal not status => wait
	//   normal status (last response)  => finish
	//   abnormal resp (no following resp expected): hasNext is true => cancel
	//   abnormal resp (following resp expected): hasNext is false, or invalid resp data type (maybe remote peer is totally broken) => cancel finish
	// case in status or status
	ret = true
	br.mutex.Lock()
	defer br.mutex.Unlock()
	// consuming request id at first
	peer, exist := br.sents[msg.OriginalID()]
	if exist {
		delete(br.sents, msg.OriginalID())
		peer.ConsumeRequest(msg.OriginalID())
	}

	status := br.status
	switch status {
	case receiverStatusWaiting:
		br.handleInWaiting(msg, msgBody)
	case receiverStatusCanceled:
		fallthrough
	case receiverStatusFinished:
		fallthrough
	default:
		br.ignoreMsg(msg, msgBody)
		return
	}
	return
}

func (br *ClusterInfoReceiver) handleInWaiting(msg p2pcommon.Message, msgBody proto.Message) {
	// timeout
	if br.timeout.Before(time.Now()) {
		// silently ignore already finished job
		br.finishReceiver()
		return
	}

	// remote peer response malformed data.
	body, ok := msgBody.(*types.GetClusterInfoResponse)
	if !ok || len(body.MbrAttrs) == 0 || body.Error != "" {
		// TODO log fail reason?
		if !br.trySendNextPeer() {
			br.cancelReceiving(errors.New("no live peers"), false)
		}
		return
	}

	// return the result
	var err error
	br.finishReceiver()
	if len(body.Error) != 0 {
		err = fmt.Errorf("get cluster info error: %s", body.Error)
	}
	result := &message.GetClusterRsp{ClusterID: body.GetClusterID(), ChainID: body.GetChainID(), Members: body.GetMbrAttrs(),
		Err: err, HardStateInfo: body.HardStateInfo}
	br.req.ReplyC <- result
	close(br.req.ReplyC)
	return
}

// cancelReceiving is cancel wait for receiving and return the failure result.
// it wait remaining (and useless) response. It is assumed cancellations are not frequently occur
func (br *ClusterInfoReceiver) cancelReceiving(err error, hasNext bool) {
	br.status = receiverStatusCanceled
	result := &message.GetClusterRsp{Err: err}
	br.req.ReplyC <- result
	close(br.req.ReplyC)
	br.finishReceiver()
}

// finishReceiver is to cancel works, assuming cancellations are not frequently occur
func (br *ClusterInfoReceiver) finishReceiver() {
	br.status = receiverStatusFinished
}

// ignoreMsg is silently ignore following responses, which is not useless anymore.
func (br *ClusterInfoReceiver) ignoreMsg(msg p2pcommon.Message, msgBody proto.Message) {
	// nothing to do for now
}
