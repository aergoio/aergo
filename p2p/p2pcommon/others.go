package p2pcommon

import (
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/types"
)

// PeerAccessor is an interface for a another actor module to get info of peers
type PeerAccessor interface {
	GetPeerBlockInfos() []types.PeerBlockInfo
	GetPeer(ID types.PeerID) (RemotePeer, bool)
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