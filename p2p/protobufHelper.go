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
const ClientVersion = "0.2.0"

type pbMessage interface {
	proto.Message
}

type pbMessageOrder struct {
	// reqID means that this message is response of the request of ID. Set empty if the messge is request.
	request         bool
	needSign        bool
	protocolID      SubProtocol // protocolName and msg struct type MUST be matched.

	message Message
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

func (pr *pbMessageOrder) GetMsgID() MsgID {
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

func (pr *pbMessageOrder) GetProtocolID() SubProtocol {
	return pr.protocolID
}

type pbRequestOrder struct {
	pbMessageOrder
	respReceiver ResponseReceiver
}

func (pr *pbRequestOrder) SendTo(p *remotePeerImpl) error {
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}

	p.reqMutex.Lock()
	p.requests[pr.message.ID()] = &requestInfo{cTime:time.Now(), reqMO:pr, receiver: pr.respReceiver}
	p.reqMutex.Unlock()

	p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).
		Str(LogMsgID, pr.GetMsgID().String()).Msg("Send request message")

	return nil
}

type pbResponseOrder struct {
	pbMessageOrder
}

func (pr *pbResponseOrder) SendTo(p *remotePeerImpl) error {
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).
		Str(LogMsgID, pr.GetMsgID().String()).Str("req_id", pr.message.OriginalID().String()).Msg("Send response message")

	return nil
}

type pbBlkNoticeOrder struct {
	pbMessageOrder
	blkHash []byte
}

func (pr *pbBlkNoticeOrder) SendTo(p *remotePeerImpl) error {
	var blkhash BlkHash
	copy(blkhash[:], pr.blkHash)
	if ok, _ := p.blkHashCache.ContainsOrAdd(blkhash, cachePlaceHolder); ok {
		// the remote peer already know this block hash. skip it
		// too many not-insteresting log,
		// p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).
		// 	Str(LogMsgID, pr.GetMsgID()).Msg("Cancel sending blk notice. peer knows this block")
		return nil
	}
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	return nil
}

type pbTxNoticeOrder struct {
	pbMessageOrder
	txHashes [][]byte
}

func (pr *pbTxNoticeOrder) SendTo(p *remotePeerImpl) error {
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID().String()).Err(err).Msg("fail to SendTo")
		return err
	}
	if p.logger.IsDebugEnabled() {
		p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).
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

func marshalMessage(message proto.Message) ([]byte, error) {
	return proto.Marshal(message)
}

func unmarshalMessage(data []byte, msgData proto.Message) error {
	return proto.Unmarshal(data, msgData)
}

func unmarshalAndReturn(data []byte, msgData proto.Message) (proto.Message, error) {
	return msgData, proto.Unmarshal(data, msgData)
}


type pbMOFactory struct {
	signer msgSigner
}

func (mf *pbMOFactory) newMsgRequestOrder(expecteResponse bool, protocolID SubProtocol, message pbMessage) msgOrder {
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4()).String()
	if newPbMsgOrder(&rmo.pbMessageOrder, msgID, "", protocolID, message, mf.signer) {
		return rmo
	}
	return nil
}

func (mf *pbMOFactory) newMsgBlockRequestOrder(respReceiver ResponseReceiver, protocolID SubProtocol, message pbMessage) msgOrder {
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4()).String()
	if newPbMsgOrder(&rmo.pbMessageOrder, msgID, "", protocolID, message, mf.signer) {
		rmo.respReceiver = respReceiver
		return rmo
	}
	return nil
}

func (mf *pbMOFactory) newMsgResponseOrder(reqID MsgID, protocolID SubProtocol, message pbMessage) msgOrder {
	rmo := &pbResponseOrder{}
	msgID := uuid.Must(uuid.NewV4()).String()
	if newPbMsgOrder(&rmo.pbMessageOrder, msgID, reqID.String(), protocolID, message, mf.signer) {
		return rmo
	}
	return nil
}

func (mf *pbMOFactory) newMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice) msgOrder {
	rmo := &pbBlkNoticeOrder{}
	reqID := uuid.Must(uuid.NewV4()).String()
	if newPbMsgOrder(&rmo.pbMessageOrder, reqID, "", NewBlockNotice, noticeMsg, mf.signer) {
		rmo.blkHash = noticeMsg.BlockHash
		return rmo
	}
	return nil
}

func (mf *pbMOFactory) newMsgTxBroadcastOrder(message *types.NewTransactionsNotice) msgOrder {
	rmo := &pbTxNoticeOrder{}
	reqID := uuid.Must(uuid.NewV4()).String()
	if newPbMsgOrder(&rmo.pbMessageOrder, reqID, "", NewTxNotice, message, mf.signer) {
		rmo.txHashes = message.TxHashes
		return rmo
	}
	return nil
}

func (mf *pbMOFactory) newHandshakeMessage(protocolID SubProtocol, message pbMessage) Message {
	// TODO define handshake specific datatype
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, protocolID, message) {
		return rmo.message
	}
	return nil
}

// newPbMsgOrder is base form of making sendrequest struct
// TODO: It seems to have redundant parameter. reqID, expecteResponse and gossip param seems to be compacted to one or two parameters.
func newPbMsgOrder(mo *pbMessageOrder, reqID string, orgID string, protocolID SubProtocol, message pbMessage, signer msgSigner) bool {
	bytes, err := marshalMessage(message)
	if err != nil {
		return false
	}

	p2pmsg := &types.P2PMessage{Header: &types.MsgHeader{}}
	p2pmsg.Data = bytes
	request := false
	setupMessageData(p2pmsg.Header, reqID, ClientVersion, time.Now().Unix())
	p2pmsg.Header.Length = uint32(len(bytes))
	p2pmsg.Header.Subprotocol = protocolID.Uint32()
	// pubKey and peerID will be set soon before signing process
	err = signer.signMsg(p2pmsg)
	if err != nil {
		panic("Failed to sign data " + err.Error())
		return false
	}

	mo.request = request
	mo.protocolID = protocolID
	mo.needSign = true
	mo.message = NewV020Wrapper(p2pmsg, orgID)

	return true
}
