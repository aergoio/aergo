package p2pcommon

import (
	"github.com/aergoio/aergo/types"
)

// PeerAccessor is an interface for a another actor module to get info of peers
type PeerAccessor interface {
	// SelfMeta returns peerid, ipaddress and port of server itself
	SelfMeta() PeerMeta

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

