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
	Start()
	Stop()

	// handle notice from bp
	HandleBlockProducedNotice(peer RemotePeer, block *types.Block)
	// handle notice from other node
	HandleNewBlockNotice(peer RemotePeer, data *types.NewBlockNotice)
	HandleGetBlockResponse(peer RemotePeer, msg Message, resp *types.GetBlockResponse)

	// RegisterTxNotice caching ids of tx that was added to local node.
	RegisterTxNotice(txIDs []types.TxID)
	// HandleNewTxNotice handle received tx from remote peer. it caches txIDs.
	HandleNewTxNotice(peer RemotePeer, hashes []types.TxID, data *types.NewTransactionsNotice)
}

//go:generate sh -c "mockgen github.com/aergoio/aergo/p2p/p2pcommon SyncManager,PeerAccessor | sed -e 's/^package mock_p2pcommon/package p2pmock/g' > ../p2pmock/mock_syncmanager.go"
