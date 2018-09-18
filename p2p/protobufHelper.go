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
	crypto "github.com/libp2p/go-libp2p-crypto"
	protobufCodec "github.com/multiformats/go-multicodec/protobuf"
	uuid "github.com/satori/go.uuid"
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

func setupMessageData(md *types.MsgHeader, reqID string, gossip bool, version string, ts int64) {
	md.Id = reqID
	md.Gossip = gossip
	md.ClientVersion = version
	md.Timestamp = ts
}

// newPbMsgRequestOrder make send order for p2p request
func newPbMsgRequestOrder(expecteResponse bool, protocolID SubProtocol, message pbMessage, signer msgSigner) msgOrder {
	rmo := &pbRequestOrder{}
	reqID := uuid.Must(uuid.NewV4()).String()
	if newPbMsgOrder(&rmo.pbMessageOrder, reqID, expecteResponse, false, protocolID, message, signer) {
		return rmo
	}
	return nil
}

// newPbMsgResponseOrder make send order for p2p response
func newPbMsgResponseOrder(reqID string, protocolID SubProtocol, message pbMessage, signer msgSigner) msgOrder {
	rmo := &pbMessageOrder{}
	if newPbMsgOrder(rmo, reqID, false, false, protocolID, message, signer) {
		return rmo
	}
	return nil
}

// newPbMsgBroadcastOrder make send order for p2p broadcast,
// which will be fanouted and doesn't expect response of receiving peer
func newPbMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice, signer msgSigner) msgOrder {
	rmo := &pbBlkNoticeOrder{}
	reqID := uuid.Must(uuid.NewV4()).String()
	if newPbMsgOrder(&rmo.pbMessageOrder, reqID, false, true, newBlockNotice, noticeMsg, signer) {
		rmo.blkHash = noticeMsg.BlockHash
		return rmo
	}
	return nil
}

func newPbMsgTxBroadcastOrder(message *types.NewTransactionsNotice, signer msgSigner) msgOrder {
	rmo := &pbTxNoticeOrder{}
	reqID := uuid.Must(uuid.NewV4()).String()
	if newPbMsgOrder(&rmo.pbMessageOrder, reqID, false, true, newTxNotice, message, signer) {
		rmo.txHashes = message.TxHashes
		return rmo
	}
	return nil
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

func (pr *pbMessageOrder) Skippable() bool {
	return false
}

func (pr *pbMessageOrder) SendTo(p *RemotePeer) bool {
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

func (pr *pbRequestOrder) SendTo(p *RemotePeer) bool {
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

func (pr *pbBlkNoticeOrder) SendTo(p *RemotePeer) bool {
	var blkhash [blkhashLen]byte
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

func (pr *pbBlkNoticeOrder) Skippable() bool {
	return true
}

type pbTxNoticeOrder struct {
	pbMessageOrder
	txHashes [][]byte
}

func (pr *pbTxNoticeOrder) SendTo(p *RemotePeer) bool {
	var txHash [txhashLen]byte
	send, skip := 0, 0
	for _, h := range pr.txHashes {
		copy(txHash[:], h)
		if ok, _ := p.txHashCache.ContainsOrAdd(txHash, cachePlaceHolder); ok {
			skip++
		} else {
			send++
		}
	}
	if skip == len(pr.txHashes) {
		p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).
			Str(LogMsgID, pr.GetMsgID()).Msg("Cancel sending tx notice. peer knows all hashes")
		return false
	}
	p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).
		Str(LogMsgID, pr.GetMsgID()).Msg("Sending tx notice. peer knows all hashes")
	err := p.rw.WriteMsg(pr.message)
	if err != nil {
		p.logger.Warn().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, pr.GetProtocolID().String()).Str(LogMsgID, pr.GetMsgID()).Err(err).Msg("fail to SendTo")
		return false
	}
	return true
}

func (pr *pbTxNoticeOrder) Skippable() bool {
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

// SignProtoMessage sign protocol buffer messge by privKey
func SignProtoMessage(message proto.Message, privKey crypto.PrivKey) ([]byte, error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return SignData(data, privKey)
}

// SignData sign binary data using the local node's private key
func SignData(data []byte, privKey crypto.PrivKey) ([]byte, error) {
	res, err := privKey.Sign(data)
	return res, err
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
