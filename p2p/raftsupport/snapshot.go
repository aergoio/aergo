/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"io"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/etcd/snap"
)

// SnapshotIOFactory create SnapshotSender or SnapshotReceiver for a peer
type SnapshotIOFactory interface {
	NewSnapshotSender(peer p2pcommon.RemotePeer) SnapshotSender
	NewSnapshotReceiver(peer p2pcommon.RemotePeer, rwc io.ReadWriteCloser) SnapshotReceiver
}

type SnapshotSender interface {
	// Send send snapshot data to target peer and always return the result to snapMsg (i.e. call Message.CloseWithErr() )
	Send(snapMsg *snap.Message)
}

type SnapshotReceiver interface {
	Receive()
}
