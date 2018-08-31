/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"fmt"
	"time"

	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	protobufCodec "github.com/multiformats/go-multicodec/protobuf"
	uuid "github.com/satori/go.uuid"
)

// ClientVersion is the version of p2p protocol to which this codes are built
// FIXME version should be defined in more general ways
const ClientVersion = "0.1.0"

type pbMessage interface {
	proto.Message
	GetMessageData() *types.MessageData
}

type pbMessageOrder struct {
	// reqID means that this message is response of the request of ID. Set empty if the messge is request.
	request         bool
	expecteResponse bool
	gossip          bool
	needSign        bool
	protocolID      SubProtocol // protocolName and msg struct type MUST be matched.

	message pbMessage
}

var _ msgOrder = (*pbMessageOrder)(nil)

// newPbMsgOrder is base form of making sendrequest struct
// TODO: It seems to have redundant parameter. reqID, expecteResponse and gossip param seems to be compacted to one or two parameters.
func newPbMsgOrder(reqID string, expecteResponse bool, gossip bool, sign bool, protocolID SubProtocol, message pbMessage) *pbMessageOrder {
	bytes, err := marshalMessage(message)
	if err != nil {
		return nil
	}

	p2pmsg := &types.P2PMessage{Header: &types.MessageData{}}
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
	return &pbMessageOrder{request: request, protocolID: protocolID, expecteResponse: expecteResponse, gossip: gossip, needSign: sign, message: p2pmsg}
}

func setupMessageData(md *types.MessageData, reqID string, gossip bool, version string, ts int64) {
	md.Id = reqID
	md.Gossip = gossip
	md.ClientVersion = version
	md.Timestamp = ts
}

// newPbMsgRequestOrder make send order for p2p request
func newPbMsgRequestOrder(expecteResponse bool, sign bool, protocolID SubProtocol, message pbMessage) *pbMessageOrder {
	return newPbMsgOrder("", expecteResponse, false, sign, protocolID, message)
}

// newPbMsgResponseOrder make send order for p2p response
func newPbMsgResponseOrder(reqID string, sign bool, protocolID SubProtocol, message pbMessage) *pbMessageOrder {
	return newPbMsgOrder(reqID, false, true, sign, protocolID, message)
}

// newPbMsgBroadcastOrder make send order for p2p broadcast,
// which will be fanouted and doesn't expect response of receiving peer
func newPbMsgBroadcastOrder(sign bool, protocolID SubProtocol, message pbMessage) *pbMessageOrder {
	return newPbMsgOrder("", false, true, sign, protocolID, message)
}

func (pr *pbMessageOrder) GetRequestID() string {
	return pr.message.GetMessageData().Id
}

func (pr *pbMessageOrder) Timestamp() int64 {
	return pr.message.GetMessageData().Timestamp
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
func (pr *pbMessageOrder) SignWith(ps PeerManager) error {
	messageData := pr.message.GetMessageData()
	messageData.PeerID = peer.IDB58Encode(ps.SelfNodeID())
	messageData.NodePubKey, _ = ps.PublicKey().Bytes()
	signature, err := ps.SignProtoMessage(pr.message)
	if err != nil {
		return err
	}
	// TOCO check if this string conversion is safe. This conversion will corrupt data if []byte is binary data.
	messageData.Sign = signature
	return nil

}

// SendOver is send itself over the writer rw.
func (pr *pbMessageOrder) SendOver(rw *bufio.ReadWriter) error {
	err := SendProtoMessage(pr.message, rw)
	if err == nil {
		rw.Flush()
	}
	return err
}

// NewMessageData is helper method - generate message data shared between all node's p2p protocols
// messageId: unique for requests, copied from request for responses
func NewMessageData(pubKeyBytes []byte, peerID peer.ID, messageID string, gossip bool) *types.MessageData {
	// Add protobufs bin data for message author public key
	// this is useful for authenticating  messages forwarded by a node authored by another node

	return &types.MessageData{ClientVersion: "0.1.0",
		Id:         messageID,
		NodePubKey: pubKeyBytes,
		Timestamp:  time.Now().Unix(),
		PeerID:     peer.IDB58Encode(peerID),
		Gossip:     gossip}
}

// SendProtoMessage send proto.Message data over stream
func SendProtoMessage(data proto.Message, rw *bufio.ReadWriter) error {
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

// VerifyData Verifies incoming p2p message data integrity
// data: data to verify
// signature: author signature provided in the message payload
// peerID: author peer peer.ID from the message payload
// pubKeyData: author public key from the message payload
func VerifyData(data []byte, signature []byte, peerID peer.ID, pubKeyData []byte) error {
	key, err := crypto.UnmarshalPublicKey(pubKeyData)
	if err != nil {
		return err
	}

	// extract node peer.ID from the provided public key
	idFromKey, err := peer.IDFromPublicKey(key)
	if err != nil {
		return err
	}

	// verify that message author node peer.ID matches the provided node public key
	if idFromKey != peerID {
		return fmt.Errorf("PeerID mismatch")
	}

	res, err := key.Verify(data, signature)
	if err != nil {
		return err
	}
	if !res {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

func marshalMessage(message proto.Message) ([]byte, error) {
	return proto.Marshal(message)
}

func unmarshalMessage(data []byte, msgData proto.Message) error {
	return proto.Unmarshal(data, msgData)
}

func newP2PMessage(msgID string, gossip bool, protocolID SubProtocol, message pbMessage) *types.P2PMessage {
	p2pmsg := &types.P2PMessage{Header: &types.MessageData{}}

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
