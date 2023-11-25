/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

//go:generate mockgen -source=remotepeer.go  -package=p2pmock -destination=../p2pmock/mock_remotepeer.go
package p2pcommon

import (
	"time"

	"github.com/aergoio/aergo/v2/types"
)

type PeerFactory interface {
	CreateRemotePeer(remoteInfo RemoteInfo, seq uint32, rw MsgReadWriter) RemotePeer
}

// PeerTask is function to exec in remote peer's own goroutine, and should not consume lots of time to process.
type PeerTask func(p RemotePeer)

type RemotePeer interface {
	ID() types.PeerID
	RemoteInfo() RemoteInfo
	Meta() PeerMeta
	ManageNumber() uint32
	Name() string
	Version() string
	AcceptedRole() types.PeerRole
	ChangeRole(role types.PeerRole)

	AddMessageHandler(subProtocol SubProtocol, handler MessageHandler)

	State() types.PeerState
	// LastStatus returns last observed status of remote peer. this value will be changed by notice, or ping
	LastStatus() *types.LastBlockStatus

	RunPeer()
	Stop()

	SendMessage(msg MsgOrder)
	// TrySendMessage try to send message with check. It will not send message and
	// return false if io write buffer is full or prev tx query was not responed.
	TrySendMessage(msg MsgOrder) bool
	SendAndWaitMessage(msg MsgOrder, ttl time.Duration) error

	PushTxsNotice(txHashes []types.TxID)
	// utility method

	ConsumeRequest(msgID MsgID) MsgOrder
	GetReceiver(id MsgID) ResponseReceiver

	// updateBlkCache add hash to block cache and return true if this hash already exists.
	UpdateBlkCache(blkHash types.BlockID, blkNumber types.BlockNo) bool
	// updateTxCache add hashes to transaction cache and return newly added hashes.
	UpdateTxCache(hashes []types.TxID) []types.TxID
	// updateLastNotice change estimate of the last status of remote peer
	UpdateLastNotice(blkHash types.BlockID, blkNumber types.BlockNo)

	MF() MoFactory

	// AddCertificate add to my certificate list
	AddCertificate(cert *AgentCertificateV1)

	// DoTask execute task in remote peer's own goroutine, it should not consume lots of time to process.
	DoTask(task PeerTask) bool
}
