/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"context"
	"fmt"
	"github.com/aergoio/aergo/p2p/v030"
	"io"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
)

// LegacyInboundHSHandler handshake handler for legacy version
type LegacyInboundHSHandler struct {
	*LegacyWireHandshaker
}

func (ih *LegacyInboundHSHandler) Handle(r io.Reader, w io.Writer, ttl time.Duration) (p2pcommon.MsgReadWriter, *types.Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()
	return ih.handshakeInboundPeer(ctx, r, w)
}

// LegacyOutboundHSHandler handshake handler for legacy version
type LegacyOutboundHSHandler struct {
	*LegacyWireHandshaker
}

func (oh *LegacyOutboundHSHandler) Handle(r io.Reader, w io.Writer, ttl time.Duration) (p2pcommon.MsgReadWriter, *types.Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()
	return oh.handshakeOutboundPeer(ctx, r, w)
}

// LegacyWireHandshaker works to handshake to just connected peer, it detect chain networks
// and protocol versions, and then select InnerHandshaker for that protocol version.
type LegacyWireHandshaker struct {
	pm     p2pcommon.PeerManager
	actor  p2pcommon.ActorService
	logger *log.Logger
	peerID types.PeerID
	// check if is it ad-hoc
	localChainID *types.ChainID

	remoteStatus *types.Status
}

func newHandshaker(pm p2pcommon.PeerManager, actor p2pcommon.ActorService, log *log.Logger, chainID *types.ChainID, peerID types.PeerID) *LegacyWireHandshaker {
	return &LegacyWireHandshaker{pm: pm, actor: actor, logger: log, localChainID: chainID, peerID: peerID}
}

func (h *LegacyWireHandshaker) handshakeOutboundPeer(ctx context.Context, r io.Reader, w io.Writer) (p2pcommon.MsgReadWriter, *types.Status, error) {
	bufReader, bufWriter := bufio.NewReader(r), bufio.NewWriter(w)
	// send initial hsmessage
	hsHeader := p2pcommon.HSHeader{Magic: p2pcommon.MAGICTest, Version: p2pcommon.P2PVersion030}
	sent, err := bufWriter.Write(hsHeader.Marshal())
	if err != nil {
		return nil, nil, err
	}
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		// go on
	}
	if sent != len(hsHeader.Marshal()) {
		return nil, nil, fmt.Errorf("transport error")
	}
	// continue to handshake with VersionedHandshaker
	innerHS, err := h.selectProtocolVersion(hsHeader.Version, bufReader, bufWriter)
	if err != nil {
		return nil, nil, err
	}
	status, err := innerHS.DoForOutbound(ctx)
	h.remoteStatus = status
	return innerHS.GetMsgRW(), status, err
}

func (h *LegacyWireHandshaker) handshakeInboundPeer(ctx context.Context, r io.Reader, w io.Writer) (p2pcommon.MsgReadWriter, *types.Status, error) {
	var hsHeader p2pcommon.HSHeader
	bufReader, bufWriter := bufio.NewReader(r), bufio.NewWriter(w)
	// wait initial hsmessage
	headBuf := make([]byte, p2pcommon.V030HSHeaderLength)
	read, err := h.readToLen(bufReader, headBuf, 8)
	if err != nil {
		return nil, nil, err
	}
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		// go on
	}
	if read != p2pcommon.V030HSHeaderLength {
		return nil, nil, fmt.Errorf("transport error")
	}
	hsHeader.Unmarshal(headBuf)

	// continue to handshake with VersionedHandshaker
	innerHS, err := h.selectProtocolVersion(hsHeader.Version, bufReader, bufWriter)
	if err != nil {
		return nil, nil, err
	}
	status, err := innerHS.DoForInbound(ctx)
	// send hsresponse
	h.remoteStatus = status
	return innerHS.GetMsgRW(), status, err
}

func (h *LegacyWireHandshaker) readToLen(rd io.Reader, bf []byte, max int) (int, error) {
	remain := max
	offset := 0
	for remain > 0 {
		read, err := rd.Read(bf[offset:])
		if err != nil {
			return offset, err
		}
		remain -= read
		offset += read
	}
	return offset, nil
}

func (h *LegacyWireHandshaker) selectProtocolVersion(version p2pcommon.P2PVersion, r *bufio.Reader, w *bufio.Writer) (p2pcommon.VersionedHandshaker, error) {
	switch version {
	case p2pcommon.P2PVersion030:
		v030hs := v030.NewV030StateHS(h.pm, h.actor, h.logger, h.localChainID, h.peerID, r, w)
		return v030hs, nil
	default:
		return nil, fmt.Errorf("not supported version")
	}
}

