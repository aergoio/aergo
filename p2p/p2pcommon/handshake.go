/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"io"
	"time"
)

// HSHandlerFactory is creator of HSHandler
type HSHandlerFactory interface {
	CreateHSHandler(outbound bool, pm PeerManager, actor ActorService, log *log.Logger, pid peer.ID) HSHandler
}

// HSHandler will do handshake with remote peer
type HSHandler interface {
	// Handle peer handshake till ttl, and return msgrw for this connection, and status of remote peer.
	Handle(r io.Reader, w io.Writer, ttl time.Duration) (MsgReadWriter, *types.Status, error)
}

