package p2pcommon

import (
	"io"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/types"
)

// PeerAccessor is an interface for a another actor module to get info of peers
type PeerAccessor interface {
	GetPeerBlockInfos() []types.PeerBlockInfo
	GetPeer(ID types.PeerID) (RemotePeer, bool)
}

// MsgOrder is abstraction of information about the message that will be sent to peer.
// Some type of msgOrder, such as notice mo, should thread-safe and re-entrant
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

type MoFactory interface {
	NewMsgRequestOrder(expectResponse bool, protocolID SubProtocol, message MessageBody) MsgOrder
	NewMsgBlockRequestOrder(respReceiver ResponseReceiver, protocolID SubProtocol, message MessageBody) MsgOrder
	NewMsgResponseOrder(reqID MsgID, protocolID SubProtocol, message MessageBody) MsgOrder
	NewMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice) MsgOrder
	NewMsgTxBroadcastOrder(noticeMsg *types.NewTransactionsNotice) MsgOrder
	NewMsgBPBroadcastOrder(noticeMsg *types.BlockProducedNotice) MsgOrder
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

	// FutureRequest send actor request and get the Future object to get the state and return value of message
	FutureRequest(actor string, msg interface{}, timeout time.Duration) *actor.Future
	// FutureRequestDefaultTimeout is FutureRequest with default timeout
	FutureRequestDefaultTimeout(actor string, msg interface{}) *actor.Future

	GetChainAccessor() types.ChainAccessor
}

// will be changed later
//type PeerID = PeerID

// FlushableWriter is writer which have Flush method, such as bufio.Writer
type FlushableWriter interface {
	io.Writer
	// Flush writes any buffered data to the underlying io.Writer.
	Flush() error
}
