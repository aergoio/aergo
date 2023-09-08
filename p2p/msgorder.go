/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"time"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/p2p/raftsupport"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/etcd/raft/raftpb"
)

// ClientVersion is the version of p2p protocol to which this codes are built

type pbMessageOrder struct {
	// reqID means that this message is response of the request of ID. Set empty if the message is request.
	request    bool
	needSign   bool
	trace      bool
	protocolID p2pcommon.SubProtocol // protocolName and msg struct type MUST be matched.

	message p2pcommon.Message
}

var _ p2pcommon.MsgOrder = (*pbRequestOrder)(nil)
var _ p2pcommon.MsgOrder = (*pbResponseOrder)(nil)
var _ p2pcommon.MsgOrder = (*pbBlkNoticeOrder)(nil)
var _ p2pcommon.MsgOrder = (*pbTxNoticeOrder)(nil)
var _ p2pcommon.MsgOrder = (*pbRaftMsgOrder)(nil)

func (pr *pbMessageOrder) GetMsgID() p2pcommon.MsgID {
	return pr.message.ID()
}

func (pr *pbMessageOrder) Timestamp() int64 {
	return pr.message.Timestamp()
}

func (pr *pbMessageOrder) IsRequest() bool {
	return pr.request
}

func (pr *pbMessageOrder) IsNeedSign() bool {
	return pr.needSign
}

func (pr *pbMessageOrder) GetProtocolID() p2pcommon.SubProtocol {
	return pr.protocolID
}

func (pr *pbMessageOrder) CancelSend(pi p2pcommon.RemotePeer) {
}

type pbRequestOrder struct {
	pbMessageOrder
	respReceiver p2pcommon.ResponseReceiver
}

func (pr *pbRequestOrder) SendTo(pi p2pcommon.RemotePeer) error {
	p := pi.(*remotePeerImpl)
	p.reqMutex.Lock()
	p.requests[pr.message.ID()] = &requestInfo{cTime: time.Now(), reqMO: pr, receiver: pr.respReceiver}
	p.reqMutex.Unlock()
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		p.reqMutex.Lock()
		delete(p.requests, pr.message.ID())
		p.reqMutex.Unlock()
		return err
	}

	if pr.trace {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).
			Str(p2putil.LogMsgID, pr.GetMsgID().String()).Msg("Send request message")
	}
	return nil
}

type pbResponseOrder struct {
	pbMessageOrder
}

func (pr *pbResponseOrder) SendTo(pi p2pcommon.RemotePeer) error {
	p := pi.(*remotePeerImpl)
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	if pr.trace {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).
			Str(p2putil.LogMsgID, pr.GetMsgID().String()).Str(p2putil.LogOrgReqID, pr.message.OriginalID().String()).Msg("Send response message")
	}

	return nil
}

type pbBlkNoticeOrder struct {
	pbMessageOrder
	blkHash []byte
	blkNo   uint64
}

func (pr *pbBlkNoticeOrder) SendTo(pi p2pcommon.RemotePeer) error {
	p := pi.(*remotePeerImpl)
	var blkHash = types.ToBlockID(pr.blkHash)
	passedTime := time.Now().Sub(p.lastBlkNoticeTime)
	skipNotice := false
	if p.LastStatus().BlockNumber >= pr.blkNo {
		heightDiff := p.LastStatus().BlockNumber - pr.blkNo
		switch {
		case heightDiff >= GapToSkipAll:
			skipNotice = true
		case heightDiff >= GapToSkipHourly:
			skipNotice = p.skipCnt < GapToSkipHourly
		default:
			skipNotice = p.skipCnt < GapToSkip5Min
		}
	}
	if skipNotice || passedTime < MinNewBlkNoticeInterval {
		p.skipCnt++
		if p.skipCnt&0x03ff == 0 && pr.trace {
			p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).Int32("skip_cnt", p.skipCnt).Msg("Skipped NewBlockNotice ")

		}
		return nil
	}

	if ok, _ := p.blkHashCache.ContainsOrAdd(blkHash, cachePlaceHolder); ok {
		// the remote peer already know this block hash. skip too many not-interesting log,
		// p.logger.Debug().Str(LogPeerName,p.Name()).Str(LogProtoID, pr.GetProtocolID().String()).
		// 	Str(LogMsgID, pr.GetMsgID()).Msg("Cancel sending blk notice. peer knows this block")
		return nil
	}
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	p.lastBlkNoticeTime = time.Now()
	if p.skipCnt > 100 && pr.trace {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).Int32("skip_cnt", p.skipCnt).Msg("Send NewBlockNotice after long skip")
	}
	p.skipCnt = 0
	return nil
}

type pbBpNoticeOrder struct {
	pbMessageOrder
	block *types.Block
}

func (pr *pbBpNoticeOrder) SendTo(pi p2pcommon.RemotePeer) error {
	p := pi.(*remotePeerImpl)
	var blkHash = types.ToBlockID(pr.block.Hash)
	p.blkHashCache.ContainsOrAdd(blkHash, cachePlaceHolder)
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	if pr.trace {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).
			Str(p2putil.LogMsgID, pr.GetMsgID().String()).Str(p2putil.LogBlkHash, enc.ToString(pr.block.Hash)).Msg("Notify block produced")
	}
	return nil
}

type pbTxNoticeOrder struct {
	pbMessageOrder
	txHashes []types.TxID
}

func (pr *pbTxNoticeOrder) SendTo(pi p2pcommon.RemotePeer) error {
	p := pi.(*remotePeerImpl)

	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	if p.logger.IsDebugEnabled() && pr.trace {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).
			Str(p2putil.LogMsgID, pr.GetMsgID().String()).Int("hash_cnt", len(pr.txHashes)).Array("hashes", types.NewLogTxIDsMarshaller(pr.txHashes, 10)).Msg("Sent tx notice")
	}
	return nil
}

func (pr *pbTxNoticeOrder) CancelSend(pi p2pcommon.RemotePeer) {
}

type pbRaftMsgOrder struct {
	pbMessageOrder
	raftAcc consensus.AergoRaftAccessor
	msg     *raftpb.Message
}

func (pr *pbRaftMsgOrder) SendTo(pi p2pcommon.RemotePeer) error {
	p := pi.(*remotePeerImpl)

	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Err(err).Object("raftMsg", raftsupport.RaftMsgMarshaller{Message: pr.msg}).Msg("fail to Send raft message")
		pr.raftAcc.ReportUnreachable(pi.ID())
		return err
	}
	if pr.trace && p.logger.IsDebugEnabled() {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Object("raftMsg", raftsupport.RaftMsgMarshaller{Message: pr.msg}).Msg("Sent raft message")
	}
	return nil
}

func (pr *pbRaftMsgOrder) CancelSend(pi p2pcommon.RemotePeer) {
	// TODO test more whether to uncomment or to delete code below
	//pr.raftAcc.ReportUnreachable(pi.ID())
}

type pbTossOrder struct {
	pbMessageOrder
}

func (pr *pbTossOrder) SendTo(pi p2pcommon.RemotePeer) error {
	p := pi.(*remotePeerImpl)
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to toss")
		return err
	}

	if pr.trace {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, pr.GetProtocolID()).
			Str(p2putil.LogMsgID, pr.GetMsgID().String()).Msg("toss message")
	}
	return nil
}
