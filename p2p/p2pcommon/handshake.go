/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"context"
	"encoding/binary"
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

type HSHeader struct {
	Magic   uint32
	Version uint32
}

func (h HSHeader) Marshal() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b, h.Magic)
	binary.BigEndian.PutUint32(b[4:], h.Version)
	return b
}

func (h *HSHeader) Unmarshal(b []byte) {
	h.Magic = binary.BigEndian.Uint32(b)
	h.Version = binary.BigEndian.Uint32(b[4:])
}

// VersionedHandshaker do handshake work and msgreadwriter for a protocol version
type VersionedHandshaker interface {
	DoForOutbound(ctx context.Context) (*types.Status, error)
	DoForInbound(ctx context.Context) (*types.Status, error)
	GetMsgRW() MsgReadWriter
}

