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
	"github.com/libp2p/go-libp2p-peer"
)

type InboundHSHandler struct {
	*PeerHandshaker
}

func (ih *InboundHSHandler) Handle(r io.Reader, w io.Writer, ttl time.Duration) (p2pcommon.MsgReadWriter, *types.Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()
	return ih.handshakeInboundPeer(ctx, r, w)
}

type OutboundHSHandler struct {
	*PeerHandshaker
}

func (oh *OutboundHSHandler) Handle(r io.Reader, w io.Writer, ttl time.Duration) (p2pcommon.MsgReadWriter, *types.Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()
	return oh.handshakeOutboundPeer(ctx, r, w)
}

// PeerHandshaker works to handshake to just connected peer, it detect chain networks
// and protocol versions, and then select InnerHandshaker for that protocol version.
type PeerHandshaker struct {
	pm        p2pcommon.PeerManager
	actorServ p2pcommon.ActorService
	logger    *log.Logger
	peerID    peer.ID
	// check if is it adhoc
	localChainID *types.ChainID

	remoteStatus *types.Status
}

type hsResult struct {
	rw        p2pcommon.MsgReadWriter
	statusMsg *types.Status
	err       error
}

func newHandshaker(pm p2pcommon.PeerManager, actor p2pcommon.ActorService, log *log.Logger, chainID *types.ChainID, peerID peer.ID) *PeerHandshaker {
	return &PeerHandshaker{pm: pm, actorServ: actor, logger: log, localChainID: chainID, peerID: peerID}
}

func (h *PeerHandshaker) handshakeOutboundPeer(ctx context.Context, r io.Reader, w io.Writer) (p2pcommon.MsgReadWriter, *types.Status, error) {
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
	innerHS, err := h.selectProtocolVersion(hsHeader, bufReader, bufWriter)
	if err != nil {
		return nil, nil, err
	}
	status, err := innerHS.DoForOutbound(ctx)
	h.remoteStatus = status
	return innerHS.GetMsgRW(), status, err
}

func (h *PeerHandshaker) handshakeInboundPeer(ctx context.Context, r io.Reader, w io.Writer) (p2pcommon.MsgReadWriter, *types.Status, error) {
	var hsHeader p2pcommon.HSHeader
	bufReader, bufWriter := bufio.NewReader(r), bufio.NewWriter(w)
	// wait initial hsmessage
	headBuf := make([]byte, 8)
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
	if read != 8 {
		return nil, nil, fmt.Errorf("transport error")
	}
	hsHeader.Unmarshal(headBuf)

	// continue to handshake with VersionedHandshaker
	innerHS, err := h.selectProtocolVersion(hsHeader, bufReader, bufWriter)
	if err != nil {
		return nil, nil, err
	}
	status, err := innerHS.DoForInbound(ctx)
	// send hsresponse
	h.remoteStatus = status
	return innerHS.GetMsgRW(), status, err
}

func (h *PeerHandshaker) readToLen(rd io.Reader, bf []byte, max int) (int, error) {
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

func (h *PeerHandshaker) selectProtocolVersion(head p2pcommon.HSHeader, r *bufio.Reader, w *bufio.Writer) (p2pcommon.VersionedHandshaker, error) {
	switch head.Version {
	case p2pcommon.P2PVersion030:
		v030hs := v030.NewV030StateHS(h.pm, h.actorServ, h.logger, h.localChainID, h.peerID, r, w)
		return v030hs, nil
	default:
		return nil, fmt.Errorf("not supported version")
	}
}

func (h *PeerHandshaker) checkProtocolVersion(versionStr string) error {
	// TODO modify interface and put check code here
	return nil
}

