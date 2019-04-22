/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"time"
)

type RemotePeer interface {
	ID() peer.ID
	Meta() PeerMeta
	ManageNumber() uint32
	Name() string

	AddMessageHandler(subProtocol SubProtocol, handler MessageHandler)

	State() types.PeerState
	// LastStatus returns last observed status of remote peer. this value will be changed by notice, or ping
	LastStatus() *types.LastBlockStatus

	RunPeer()
	Stop()

	SendMessage(msg MsgOrder)
	SendAndWaitMessage(msg MsgOrder, ttl time.Duration) error

	PushTxsNotice(txHashes []types.TxID)
	// utility method

	ConsumeRequest(msgID MsgID)
	GetReceiver(id MsgID) ResponseReceiver

	// updateBlkCache add hash to block cache and return true if this hash already exists.
	UpdateBlkCache(blkHash []byte, blkNumber uint64) bool
	// updateTxCache add hashes to transaction cache and return newly added hashes.
	UpdateTxCache(hashes []types.TxID) []types.TxID
	// updateLastNotice change estimate of the last status of remote peer
	UpdateLastNotice(blkHash []byte, blkNumber uint64)

	// TODO
	MF() MoFactory
}
