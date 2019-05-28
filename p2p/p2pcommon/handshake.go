/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */
//go:generate mockgen -source=handshake.go  -package=p2pmock -destination=../p2pmock/mock_handshake.go
package p2pcommon

import (
	"context"
	"encoding/binary"
	"github.com/aergoio/aergo/types"
	"io"
	"time"
)


// HSHandlerFactory is creator of HSHandler
type HSHandlerFactory interface {
	CreateHSHandler(p2pVersion P2PVersion, outbound bool, pid types.PeerID) HSHandler
}

// HSHandler handles whole process of connect, handshake, create of remote Peerseer
type HSHandler interface {
	// Handle peer handshake till ttl, and return msgrw for this connection, and status of remote peer.
	Handle(r io.Reader, w io.Writer, ttl time.Duration) (MsgReadWriter, *types.Status, error)
}

type VersionedManager interface {
	FindBestP2PVersion(versions []P2PVersion) P2PVersion
	GetVersionedHandshaker(version P2PVersion, peerID types.PeerID, r io.Reader, w io.Writer) (VersionedHandshaker, error)

	InjectHandlers(version P2PVersion, peer RemotePeer)
}

// VersionedHandshaker do handshake related to chain, and return msgreadwriter for a protocol version
type VersionedHandshaker interface {
	DoForOutbound(ctx context.Context) (*types.Status, error)
	DoForInbound(ctx context.Context) (*types.Status, error)
	GetMsgRW() MsgReadWriter
}

// HSHeader is legacy type of data which peer send first to listening peer in wire handshake
type HSHeader struct {
	Magic   uint32
	Version P2PVersion
}

func (h HSHeader) Marshal() []byte {
	b := make([]byte, V030HSHeaderLength)
	binary.BigEndian.PutUint32(b, h.Magic)
	binary.BigEndian.PutUint32(b[4:], uint32(h.Version))
	return b
}

func (h *HSHeader) Unmarshal(b []byte) {
	h.Magic = binary.BigEndian.Uint32(b)
	h.Version = P2PVersion(binary.BigEndian.Uint32(b[4:]))
}

// HSHeadReq is data which peer send first to listening peer in wire handshake
type HSHeadReq struct {
	Magic uint32
	// Versions are p2p versions which the connecting peer can support.
	Versions []P2PVersion
}

func (h HSHeadReq) Marshal() []byte {
	verCount := len(h.Versions)
	b := make([]byte, HSMagicLength+HSVerCntLength+HSVersionLength*verCount)
	offset := 0
	binary.BigEndian.PutUint32(b[offset:], h.Magic)
	offset += HSMagicLength
	binary.BigEndian.PutUint32(b[offset:], uint32(verCount))
	offset += HSVerCntLength
	for _, version := range h.Versions {
		binary.BigEndian.PutUint32(b[offset:], version.Uint32())
		offset += HSVersionLength
	}
	return b
}

// HSHeadResp is data which listening peer send back to connecting peer as response
type HSHeadResp struct {
	// Magic will be same as the magic in HSHeadReq if wire handshake is successful, or 0 if not.
	Magic    uint32
	// RespCode is different meaning by value of Magic. It is p2p version which listening peer will use, if wire handshake is succesful, or errCode otherwise.
	RespCode uint32
}

func (h HSHeadResp) Marshal() []byte {
	b := make([]byte, V030HSHeaderLength)
	binary.BigEndian.PutUint32(b, h.Magic)
	binary.BigEndian.PutUint32(b[4:], h.RespCode)
	return b
}

func (h *HSHeadResp) Unmarshal(b []byte) {
	h.Magic = binary.BigEndian.Uint32(b)
	h.RespCode = binary.BigEndian.Uint32(b[4:])
}
