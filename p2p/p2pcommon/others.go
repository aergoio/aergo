package p2pcommon

import (
	"fmt"

	"github.com/aergoio/aergo/v2/types"
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
	Summary() map[string]interface{}

	// handle notice from bp
	HandleBlockProducedNotice(peer RemotePeer, block *types.Block)
	// handle notice from other node
	HandleNewBlockNotice(peer RemotePeer, data *types.NewBlockNotice)
	HandleGetBlockResponse(peer RemotePeer, msg Message, resp *types.GetBlockResponse)

	// RegisterTxNotice caching ids of tx that was added to local node.
	RegisterTxNotice(txs []*types.Tx)
	// HandleNewTxNotice handle received tx from remote peer. it caches txIDs.
	HandleNewTxNotice(peer RemotePeer, hashes []types.TxID, data *types.NewTransactionsNotice)
	HandleGetTxReq(peer RemotePeer, msgID MsgID, data *types.GetTransactionsRequest) error
	RetryGetTx(peer RemotePeer, hashes [][]byte)
}

//go:generate sh -c "mockgen github.com/aergoio/aergo/v2/p2p/p2pcommon SyncManager,PeerAccessor | sed -e 's/^package mock_p2pcommon/package p2pmock/g' > ../p2pmock/mock_syncmanager.go"

var SyncManagerBusyError = fmt.Errorf("server is busy")
