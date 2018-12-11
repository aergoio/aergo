/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

type DummySyncManager struct {

}

func (DummySyncManager) HandleNewBlockNotice(peer RemotePeer, hash BlkHash, data *types.NewBlockNotice) {
	// do nothing
}

func (DummySyncManager) HandleGetBlockResponse(peer RemotePeer, msg Message, resp *types.GetBlockResponse) {
	// do nothing
}

func (DummySyncManager) HandleNewTxNotice(peer RemotePeer, hashes []TxHash, data *types.NewTransactionsNotice) {
	// do nothing
}

func (DummySyncManager) DoSync(peer RemotePeer, hashes []message.BlockHash, stopHash message.BlockHash) {
	// do nothing
}

type DummyReconnectManager struct {

}

func (DummyReconnectManager) AddJob(meta PeerMeta) {
	// do nothing
}

func (DummyReconnectManager) CancelJob(pid peer.ID) {
	// do nothing
}

func (DummyReconnectManager) jobFinished(pid peer.ID) {
	// do nothing
}

func (DummyReconnectManager) Stop() {
	// do nothing
}


