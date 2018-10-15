/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"time"

	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	protobufCodec "github.com/multiformats/go-multicodec/protobuf"
	"github.com/satori/go.uuid"
)

// ClientVersion is the version of p2p protocol to which this codes are built
// FIXME version should be defined in more general ways
const ClientVersion = "0.1.0"

type pbMessage interface {
	proto.Message
}

type pbMessageOrder struct {
	// reqID means that this message is response of the request of ID. Set empty if the messge is request.
	request         bool
	expecteResponse bool
	gossip          bool
	needSign        bool
	protocolID      SubProtocol // protocolName and msg struct type MUST be matched.

	message *types.P2PMessage
}

var _ msgOrder = (*pbMessageOrder)(nil)


func setupMessageData(md *types.MsgHeader, reqID string, gossip bool, version string, ts int64) {
	md.Id = reqID
	md.Gossip = gossip
	md.ClientVersion = version
	md.Timestamp = ts
}

func (pr *pbMessageOrder) GetMsgID() string {
	return pr.message.Header.Id
}

func (pr *pbMessageOrder) Timestamp() int64 {
	return pr.message.Header.Timestamp
}

func (pr *pbMessageOrder) IsRequest() bool {
	return pr.request
}
func (pr *pbMessageOrder) ResponseExpected() bool {
	return pr.expecteResponse
}

func (pr *pbMessageOrder) IsGossip() bool {
	return pr.gossip
}

func (pr *pbMessageOrder) IsNeedSign() bool {
	return pr.needSign
}

func (pr *pbMessageOrder) GetProtocolID() SubProtocol {
	return pr.protocolID
}

func (pr *pbMessageOrder) SendTo(p *remotePeerImpl) bool {
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID()).Err(err).Msg("fail to SendTo")
		return false
	}
	p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).
		Str(LogMsgID, pr.GetMsgID()).Msg("Send message")

	return true
}

type pbRequestOrder struct {
	pbMessageOrder
}

func (pr *pbRequestOrder) SendTo(p *remotePeerImpl) bool {
	if pr.pbMessageOrder.SendTo(p) {
		p.requests[pr.GetMsgID()] = pr
		return true
	}
	return false
}

type pbBlkNoticeOrder struct {
	pbMessageOrder
	blkHash []byte
}

func (pr *pbBlkNoticeOrder) SendTo(p *remotePeerImpl) bool {
	var blkhash BlockHash
	copy(blkhash[:], pr.blkHash)
	if ok, _ := p.blkHashCache.ContainsOrAdd(blkhash, cachePlaceHolder); ok {
		// the remote peer already know this block hash. skip it
		// too many not-insteresting log,
		// p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).
		// 	Str(LogMsgID, pr.GetMsgID()).Msg("Cancel sending blk notice. peer knows this block")
		return false
	}
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID()).Err(err).Msg("fail to SendTo")
		return false
	}
	return true
}

type pbTxNoticeOrder struct {
	pbMessageOrder
	txHashes [][]byte
}

func (pr *pbTxNoticeOrder) SendTo(p *remotePeerImpl) bool {
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID()).Err(err).Msg("fail to SendTo")
		return false
	}
	if p.logger.IsDebugEnabled() {
		p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).
		Str(LogMsgID, pr.GetMsgID()).Int("hash_cnt", len(pr.txHashes)).Str("hashes",bytesArrToString(pr.txHashes)).Msg("Sent tx notice")
	}
	return true
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

func marshalMessage(message proto.Message) ([]byte, error) {
	return proto.Marshal(message)
}

func unmarshalMessage(data []byte, msgData proto.Message) error {
	return proto.Unmarshal(data, msgData)
}

func unmarshalAndReturn(data []byte, msgData proto.Message) (proto.Message, error) {
	return msgData, proto.Unmarshal(data, msgData)
}

func newP2PMessage(msgID string, gossip bool, protocolID SubProtocol, message pbMessage) *types.P2PMessage {
	p2pmsg := &types.P2PMessage{Header: &types.MsgHeader{}}

	bytes, err := marshalMessage(message)
	if err != nil {
		return nil
	}
	p2pmsg.Data = bytes
	if len(msgID) == 0 {
		msgID = uuid.Must(uuid.NewV4()).String()
	}
	setupMessageData(p2pmsg.Header, msgID, gossip, ClientVersion, time.Now().Unix())
	p2pmsg.Header.Subprotocol = protocolID.Uint32()
	return p2pmsg
}


type pbMOFactory struct {
	signer msgSigner
}

func (mf *pbMOFactory) newMsgRequestOrder(expecteResponse bool, protocolID SubProtocol, message pbMessage) msgOrder {
	rmo := &pbRequestOrder{}
	reqID := uuid.Must(uuid.NewV4()).String()
	if newPbMsgOrder(&rmo.pbMessageOrder, reqID, expecteResponse, false, protocolID, message, mf.signer) {
		return rmo
	}
	return nil
}

func (mf *pbMOFactory) newMsgResponseOrder(reqID string, protocolID SubProtocol, message pbMessage) msgOrder {
	rmo := &pbMessageOrder{}
	if newPbMsgOrder(rmo, reqID, false, false, protocolID, message, mf.signer) {
		return rmo
	}
	return nil
}

func (mf *pbMOFactory) newMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice) msgOrder {
	rmo := &pbBlkNoticeOrder{}
	reqID := uuid.Must(uuid.NewV4()).String()
	if newPbMsgOrder(&rmo.pbMessageOrder, reqID, false, true, NewBlockNotice, noticeMsg, mf.signer) {
		rmo.blkHash = noticeMsg.BlockHash
		return rmo
	}
	return nil
}

func (mf *pbMOFactory) newMsgTxBroadcastOrder(message *types.NewTransactionsNotice) msgOrder {
	rmo := &pbTxNoticeOrder{}
	reqID := uuid.Must(uuid.NewV4()).String()
	if newPbMsgOrder(&rmo.pbMessageOrder, reqID, false, true, NewTxNotice, message, mf.signer) {
		rmo.txHashes = message.TxHashes
		return rmo
	}
	return nil
}

// newPbMsgOrder is base form of making sendrequest struct
// TODO: It seems to have redundant parameter. reqID, expecteResponse and gossip param seems to be compacted to one or two parameters.
func newPbMsgOrder(mo *pbMessageOrder, reqID string, expecteResponse bool, gossip bool, protocolID SubProtocol, message pbMessage, signer msgSigner) bool {
	bytes, err := marshalMessage(message)
	if err != nil {
		return false
	}

	p2pmsg := &types.P2PMessage{Header: &types.MsgHeader{}}
	p2pmsg.Data = bytes
	request := false
	setupMessageData(p2pmsg.Header, reqID, gossip, ClientVersion, time.Now().Unix())
	p2pmsg.Header.Length = uint32(len(bytes))
	p2pmsg.Header.Subprotocol = protocolID.Uint32()
	// pubKey and peerID will be set soon before signing process
	// expecteResponse is only applied when message is request and not a gossip.
	if request == false || gossip {
		expecteResponse = false
	}
	err = signer.signMsg(p2pmsg)
	if err != nil {
		panic("Failed to sign data " + err.Error())
		return false
	}

	mo.request = request
	mo.protocolID = protocolID
	mo.expecteResponse = expecteResponse
	mo.gossip = gossip
	mo.needSign = true
	mo.message = p2pmsg

	return true
}
