package p2pcommon

import (
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
)

// msgOrder is abstraction information about the message that will be sent to peer
// some type of msgOrder, such as notice mo, should thread-safe and re-entrant
type MsgOrder interface {
	GetMsgID() MsgID
	// Timestamp is unit time value
	Timestamp() int64
	IsRequest() bool
	IsNeedSign() bool
	GetProtocolID() SubProtocol

	// SendTo send message to remote peer. it return err if write fails, or nil if write is successful or ignored.
	SendTo(p RemotePeer) error
}

type ResponseReceiver func(Message, proto.Message) bool
type PbMessage interface {
	proto.Message
}

type MoFactory interface {
	NewMsgRequestOrder(expecteResponse bool, protocolID SubProtocol, message PbMessage) MsgOrder
	NewMsgBlockRequestOrder(respReceiver ResponseReceiver, protocolID SubProtocol, message PbMessage) MsgOrder
	NewMsgResponseOrder(reqID MsgID, protocolID SubProtocol, message PbMessage) MsgOrder
	NewMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice) MsgOrder
	NewMsgTxBroadcastOrder(noticeMsg *types.NewTransactionsNotice) MsgOrder
	NewMsgBPBroadcastOrder(noticeMsg *types.BlockProducedNotice) MsgOrder
}

// PeerManager is internal service that provide peer management
type PeerManager interface {
	Start() error
	Stop() error

	//NetworkTransport
	SelfMeta() PeerMeta
	SelfNodeID() peer.ID

	AddNewPeer(peer PeerMeta)
	// Remove peer from peer list. Peer dispose relative resources and stop itself, and then call RemovePeer to peermanager
	RemovePeer(peer RemotePeer)
	// NotifyPeerHandshake is called after remote peer is completed handshake and ready to receive or send
	NotifyPeerHandshake(peerID peer.ID)
	NotifyPeerAddressReceived([]PeerMeta)

	// GetPeer return registered(handshaked) remote peer object
	GetPeer(ID peer.ID) (RemotePeer, bool)
	GetPeers() []RemotePeer
	GetPeerAddresses(noHidden bool, showSelf bool) []*message.PeerInfo

	GetPeerBlockInfos() []types.PeerBlockInfo
}
type SyncManager interface {
	// handle notice from bp
	HandleBlockProducedNotice(peer RemotePeer, block *types.Block)
	// handle notice from other node
	HandleNewBlockNotice(peer RemotePeer, data *types.NewBlockNotice)
	HandleGetBlockResponse(peer RemotePeer, msg Message, resp *types.GetBlockResponse)
	HandleNewTxNotice(peer RemotePeer, hashes []types.TxID, data *types.NewTransactionsNotice)
}

// ActorService is collection of helper methods to use actor
// FIXME move to more general package. it used in p2p and rpc
type ActorService interface {
	// TellRequest send actor request, which does not need to get return value, and forget it.
	TellRequest(actor string, msg interface{})
	// SendRequest send actor request, and the response is expected to go back asynchronously.
	SendRequest(actor string, msg interface{})
	// CallRequest send actor request and wait the handling of that message to finished,
	// and get return value.
	CallRequest(actor string, msg interface{}, timeout time.Duration) (interface{}, error)
	// CallRequestDefaultTimeout is CallRequest with default timeout
	CallRequestDefaultTimeout(actor string, msg interface{}) (interface{}, error)

	// FutureRequest send actor reqeust and get the Future object to get the state and return value of message
	FutureRequest(actor string, msg interface{}, timeout time.Duration) *actor.Future
	// FutureRequestDefaultTimeout is FutureRequest with default timeout
	FutureRequestDefaultTimeout(actor string, msg interface{}) *actor.Future

	GetChainAccessor() types.ChainAccessor
}

// MessageHandler handle incoming subprotocol message
type MessageHandler interface {
	ParsePayload([]byte) (proto.Message, error)
	CheckAuth(msgHeader Message, msgBody proto.Message) error
	Handle(msgHeader Message, msgBody proto.Message)
	PreHandle()
	PostHandle(msgHeader Message, msgBody proto.Message)
}

// signHandler sign or verify p2p message
type MsgSigner interface {
	// signMsg calulate signature and fill related fields in msg(peerid, pubkey, signature or etc)
	SignMsg(msg *types.P2PMessage) error
	// verifyMsg check signature is valid
	VerifyMsg(msg *types.P2PMessage, senderID peer.ID) error
}

// NTContainer can provide NetworkTransport interface.
type NTContainer interface {
	GetNetworkTransport() NetworkTransport

	// ChainID return id of current chain.
	ChainID() *types.ChainID
}

// NetworkTransport do manager network connection
// TODO need refactoring. it has other role, pk management of self peer
type NetworkTransport interface {
	host.Host
	Start() error
	Stop() error

	PrivateKey() crypto.PrivKey
	PublicKey() crypto.PubKey
	SelfMeta() PeerMeta
	SelfNodeID() peer.ID

	GetAddressesOfPeer(peerID peer.ID) []string

	// AddStreamHandler wrapper function which call host.SetStreamHandler after transport is initialized, this method is for preventing nil error.
	AddStreamHandler(pid protocol.ID, handler inet.StreamHandler)

	GetOrCreateStream(meta PeerMeta, protocolID protocol.ID) (inet.Stream, error)
	GetOrCreateStreamWithTTL(meta PeerMeta, protocolID protocol.ID, ttl time.Duration) (inet.Stream, error)

	FindPeer(peerID peer.ID) bool
	ClosePeerConnection(peerID peer.ID) bool
}
