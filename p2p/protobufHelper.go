/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"time"

	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	protobufCodec "github.com/multiformats/go-multicodec/protobuf"
)

// ClientVersion is the version of p2p protocol to which this codes are built
// FIXME version should be defined in more general ways
const ClientVersion = "0.2.0"

type pbMessage interface {
	proto.Message
}

type pbMessageOrder struct {
	// reqID means that this message is response of the request of ID. Set empty if the messge is request.
	request    bool
	needSign   bool
	protocolID p2pcommon.SubProtocol // protocolName and msg struct type MUST be matched.

	message p2pcommon.Message
}

var _ msgOrder = (*pbRequestOrder)(nil)
var _ msgOrder = (*pbResponseOrder)(nil)
var _ msgOrder = (*pbBlkNoticeOrder)(nil)
var _ msgOrder = (*pbTxNoticeOrder)(nil)


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
	respReceiver ResponseReceiver
}

func (pr *pbRequestOrder) SendTo(p *remotePeerImpl) error {
	p.reqMutex.Lock()
	p.requests[pr.message.ID()] = &requestInfo{cTime:time.Now(), reqMO:pr, receiver: pr.respReceiver}
	p.reqMutex.Unlock()
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(LogPeerName, p.Name()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		p.reqMutex.Lock()
		delete(p.requests, pr.message.ID())
		p.reqMutex.Unlock()
		return err
	}


	p.logger.Debug().Str(LogPeerName, p.Name()).Str(LogProtoID, pr.GetProtocolID().String()).
		Str(LogMsgID, pr.GetMsgID().String()).Msg("Send request message")

	return nil
}

type pbResponseOrder struct {
	pbMessageOrder
}

func (pr *pbResponseOrder) SendTo(p *remotePeerImpl) error {
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(LogPeerName, p.Name()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	p.logger.Debug().Str(LogPeerName, p.Name()).Str(LogProtoID, pr.GetProtocolID().String()).
		Str(LogMsgID, pr.GetMsgID().String()).Str("req_id", pr.message.OriginalID().String()).Msg("Send response message")

	return nil
}

type pbBlkNoticeOrder struct {
	pbMessageOrder
	blkHash []byte
}

func (pr *pbBlkNoticeOrder) SendTo(p *remotePeerImpl) error {
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
		p.logger.Warn().Str(LogPeerName,p.Name()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	return nil
}


type pbBpNoticeOrder struct {
	pbMessageOrder
	block *types.Block
}

func (pr *pbBpNoticeOrder) SendTo(p *remotePeerImpl) error {
	var blkhash = types.ToBlockID(pr.block.Hash)
	p.blkHashCache.ContainsOrAdd(blkhash, cachePlaceHolder)
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(LogPeerName,p.Name()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	p.logger.Debug().Str(LogPeerName,p.Name()).Str(LogProtoID, pr.GetProtocolID().String()).
		Str(LogMsgID, pr.GetMsgID().String()).Str(LogBlkHash,enc.ToString(pr.block.Hash)).Msg("Notify block produced")
	return nil
}

type pbTxNoticeOrder struct {
	pbMessageOrder
	txHashes [][]byte
}

func (pr *pbTxNoticeOrder) SendTo(p *remotePeerImpl) error {
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(LogPeerName,p.Name()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	if p.logger.IsDebugEnabled() {
		p.logger.Debug().Str(LogPeerName,p.Name()).Str(LogProtoID, pr.GetProtocolID().String()).
		Str(LogMsgID, pr.GetMsgID().String()).Int("hash_cnt", len(pr.txHashes)).Str("hashes",bytesArrToString(pr.txHashes)).Msg("Sent tx notice")
	}
	return nil
}

// SendProtoMessage send proto.Message data over stream
func SendProtoMessage(data proto.Message, rw *bufio.Writer) error {
	enc := protobufCodec.Multicodec(nil).Encoder(rw)
	err := enc.Encode(data)
	if err != nil {
		return err
	}
	rw.Flush()
	return nil
}

func MarshalMessage(message proto.Message) ([]byte, error) {
	return proto.Marshal(message)
}

func UnmarshalMessage(data []byte, msgData proto.Message) error {
	return proto.Unmarshal(data, msgData)
}

func unmarshalAndReturn(data []byte, msgData proto.Message) (proto.Message, error) {
	return msgData, proto.Unmarshal(data, msgData)
}
