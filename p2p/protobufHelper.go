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
func newPbMsgOrder(reqID string, expecteResponse bool, gossip bool, protocolID SubProtocol, message pbMessage, signer msgSigner) *pbMessageOrder {
	bytes, err := marshalMessage(message)
	if err != nil {
		return nil
	}

	p2pmsg := &types.P2PMessage{Header: &types.MsgHeader{}}
	p2pmsg.Data = bytes
	request := false
	if len(reqID) == 0 {
		reqID = uuid.Must(uuid.NewV4()).String()
		request = true
	}
	setupMessageData(p2pmsg.Header, reqID, gossip, ClientVersion, time.Now().Unix())
	p2pmsg.Header.Subprotocol = protocolID.Uint32()
	// pubKey and peerID will be set soon before signing process
	// expecteResponse is only applied when message is request and not a gossip.
	if request == false || gossip {
		expecteResponse = false
	}
	err = signer.signMsg(p2pmsg)
	if err != nil {
		panic("Failed to sign data " + err.Error())
		return nil
	}
	return &pbMessageOrder{request: request, protocolID: protocolID, expecteResponse: expecteResponse, gossip: gossip, needSign: true, message: p2pmsg}
}

func setupMessageData(md *types.MsgHeader, reqID string, gossip bool, version string, ts int64) {
	md.Id = reqID
	md.Gossip = gossip
	md.ClientVersion = version
	md.Timestamp = ts
}

// newPbMsgRequestOrder make send order for p2p request
func newPbMsgRequestOrder(expecteResponse bool, protocolID SubProtocol, message pbMessage, signer msgSigner) *pbMessageOrder {
	return newPbMsgOrder("", expecteResponse, false, protocolID, message, signer)
}

// newPbMsgResponseOrder make send order for p2p response
func newPbMsgResponseOrder(reqID string, protocolID SubProtocol, message pbMessage, signer msgSigner) *pbMessageOrder {
	return newPbMsgOrder(reqID, false, true, protocolID, message, signer)
}

// newPbMsgBroadcastOrder make send order for p2p broadcast,
// which will be fanouted and doesn't expect response of receiving peer
func newPbMsgBroadcastOrder(protocolID SubProtocol, message pbMessage, signer msgSigner) *pbMessageOrder {
	return newPbMsgOrder("", false, true, protocolID, message, signer)
}

func (pr *pbMessageOrder) GetRequestID() string {
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

// SendOver is send itself over the writer rw.
func (pr *pbMessageOrder) SendOver(w MsgWriter) error {
	return w.WriteMsg(pr.message)
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
