/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"time"

	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"

	"github.com/aergoio/aergo/types"
)

// ClientVersion is the version of p2p protocol to which this codes are built
// FIXME version should be defined in more general ways
const ClientVersion = "0.2.0"

type pbMessageOrder struct {
	// reqID means that this message is response of the request of ID. Set empty if the messge is request.
	request    bool
	needSign   bool
	protocolID p2pcommon.SubProtocol // protocolName and msg struct type MUST be matched.

	message p2pcommon.Message
}

var _ p2pcommon.MsgOrder = (*pbRequestOrder)(nil)
var _ p2pcommon.MsgOrder = (*pbResponseOrder)(nil)
var _ p2pcommon.MsgOrder = (*pbBlkNoticeOrder)(nil)
var _ p2pcommon.MsgOrder = (*pbTxNoticeOrder)(nil)

func setupMessageData(md *types.MsgHeader, reqID string, version string, ts int64) {
	md.Id = reqID
	md.Gossip = false
	md.ClientVersion = version
	md.Timestamp = ts
}

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
		p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, pr.GetProtocolID().String()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		p.reqMutex.Lock()
		delete(p.requests, pr.message.ID())
		p.reqMutex.Unlock()
		return err
	}

	p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, pr.GetProtocolID().String()).
		Str(p2putil.LogMsgID, pr.GetMsgID().String()).Msg("Send request message")

	return nil
}

type pbResponseOrder struct {
	pbMessageOrder
}

func (pr *pbResponseOrder) SendTo(pi p2pcommon.RemotePeer) error {
	p := pi.(*remotePeerImpl)
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, pr.GetProtocolID().String()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, pr.GetProtocolID().String()).
		Str(p2putil.LogMsgID, pr.GetMsgID().String()).Str(p2putil.LogOrgReqID, pr.message.OriginalID().String()).Msg("Send response message")

	return nil
}

type pbBlkNoticeOrder struct {
	pbMessageOrder
	blkHash []byte
}

func (pr *pbBlkNoticeOrder) SendTo(pi p2pcommon.RemotePeer) error {
	p := pi.(*remotePeerImpl)
	var blkhash = types.ToBlockID(pr.blkHash)
	if ok, _ := p.blkHashCache.ContainsOrAdd(blkhash, cachePlaceHolder); ok {
		// the remote peer already know this block hash. skip it
		// too many not-insteresting log,
		// p.logger.Debug().Str(LogPeerName,p.Name()).Str(LogProtoID, pr.GetProtocolID().String()).
		// 	Str(LogMsgID, pr.GetMsgID()).Msg("Cancel sending blk notice. peer knows this block")
		return nil
	}
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, pr.GetProtocolID().String()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	return nil
}

type pbBpNoticeOrder struct {
	pbMessageOrder
	block *types.Block
}

func (pr *pbBpNoticeOrder) SendTo(pi p2pcommon.RemotePeer) error {
	p := pi.(*remotePeerImpl)
	var blkhash = types.ToBlockID(pr.block.Hash)
	p.blkHashCache.ContainsOrAdd(blkhash, cachePlaceHolder)
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, pr.GetProtocolID().String()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, pr.GetProtocolID().String()).
		Str(p2putil.LogMsgID, pr.GetMsgID().String()).Str(p2putil.LogBlkHash, enc.ToString(pr.block.Hash)).Msg("Notify block produced")
	return nil
}

type pbTxNoticeOrder struct {
	pbMessageOrder
	txHashes [][]byte
}

func (pr *pbTxNoticeOrder) SendTo(pi p2pcommon.RemotePeer) error {
	p := pi.(*remotePeerImpl)

	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, pr.GetProtocolID().String()).Str(p2putil.LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	if p.logger.IsDebugEnabled() {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, pr.GetProtocolID().String()).
			Str(p2putil.LogMsgID, pr.GetMsgID().String()).Int("hash_cnt", len(pr.txHashes)).Str("hashes", p2putil.BytesArrToString(pr.txHashes)).Msg("Sent tx notice")
	}
	return nil
}
