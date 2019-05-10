/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	v030 "github.com/aergoio/aergo/p2p/v030"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
	"io"
	"time"
)

// CurrentSupported is list of versions this aergosvr supports. The first is the best recommended version.
var CurrentSupported = []p2pcommon.P2PVersion{p2pcommon.P2PVersion031, p2pcommon.P2PVersion030}

// baseWireHandshaker works to handshake to just connected peer, it detect chain networks
// and protocol versions, and then select InnerHandshaker for that protocol version.
type baseWireHandshaker struct {
	pm     p2pcommon.PeerManager
	actor  p2pcommon.ActorService
	verM   p2pcommon.VersionedManager
	logger *log.Logger
	peerID peer.ID
	// check if is it adhoc
	localChainID *types.ChainID

	remoteStatus *types.Status
}

type InboundWireHandshaker struct {
	baseWireHandshaker
}

func NewInbountHSHandler(pm p2pcommon.PeerManager, actor p2pcommon.ActorService, verManager p2pcommon.VersionedManager, log *log.Logger, chainID *types.ChainID, peerID peer.ID) p2pcommon.HSHandler {
	return &InboundWireHandshaker{baseWireHandshaker{pm: pm, actor: actor, verM:verManager, logger: log, localChainID: chainID, peerID: peerID}}
}

func (h *InboundWireHandshaker) Handle(r io.Reader, w io.Writer, ttl time.Duration) (p2pcommon.MsgReadWriter, *types.Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()
	bufReader, bufWriter := bufio.NewReader(r), bufio.NewWriter(w)
	return h.handleInboundPeer(ctx, bufReader, bufWriter)
}

func (h *InboundWireHandshaker) handleInboundPeer(ctx context.Context, rd io.Reader, wr p2pcommon.FlushableWriter) (p2pcommon.MsgReadWriter, *types.Status, error) {
	// wait initial hsmessage
	hsReq, err := h.readWireHSRequest(rd)
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		// go on
	}
	if err != nil {
		return h.writeErrAndReturn(err, p2pcommon.ErrWrongHSReq, wr)
	}
	// check magic
	if hsReq.Magic != p2pcommon.MAGICMain {
		return h.writeErrAndReturn(fmt.Errorf("wrong magic %v",hsReq.Magic), p2pcommon.ErrWrongHSReq, wr)
	}

	// continue to handshake with VersionedHandshaker
	bestVer := h.verM.FindBestP2PVersion(hsReq.Versions)
	if bestVer == p2pcommon.P2PVersionUnknown {
		return h.writeErrAndReturn(fmt.Errorf("no matchied p2p version for %v", hsReq.Versions), p2pcommon.ErrNoMatchedVersion,wr)
	} else {
		resp := p2pcommon.HSHeadResp{hsReq.Magic, bestVer.Uint32()}
		err = h.writeWireHSResponse(resp, wr)
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
			// go on
		}
		if err != nil {
			return nil, nil, err
		}
	}
	innerHS, err := h.verM.GetVersionedHandshaker(bestVer, h.peerID, rd, wr)
	if err != nil {
		return nil, nil, err
	}
	status, err := innerHS.DoForInbound(ctx)
	// send hsresponse
	h.remoteStatus = status
	return innerHS.GetMsgRW(), status, err
}

type OutboundWireHandshaker struct {
	baseWireHandshaker
}

func NewOutbountHSHandler(pm p2pcommon.PeerManager, actor p2pcommon.ActorService, verManager p2pcommon.VersionedManager, log *log.Logger, chainID *types.ChainID, peerID peer.ID) p2pcommon.HSHandler {
	return &OutboundWireHandshaker{baseWireHandshaker{pm: pm, actor: actor, verM:verManager, logger: log, localChainID: chainID, peerID: peerID}}
}

func (h *OutboundWireHandshaker) Handle(r io.Reader, w io.Writer, ttl time.Duration) (p2pcommon.MsgReadWriter, *types.Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()
	bufReader, bufWriter := bufio.NewReader(r), bufio.NewWriter(w)
	return h.handleOutboundPeer(ctx, bufReader, bufWriter)
}

func (h *OutboundWireHandshaker) handleOutboundPeer(ctx context.Context, bufReader io.Reader, bufWriter p2pcommon.FlushableWriter) (p2pcommon.MsgReadWriter, *types.Status, error) {
	// send initial hsmessage
	versions := []p2pcommon.P2PVersion{
		p2pcommon.P2PVersion031,
		p2pcommon.P2PVersion030,
	}
	hsHeader := p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: versions}
	err := h.writeWireHSRequest(hsHeader, bufWriter)
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		// go on
	}
	if err != nil {
		return nil, nil, err
	}

	// read response
	respHeader, err := h.readWireHSResp(bufReader)
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		// go on
	}
	if err != nil {
		return nil, nil, err
	}
	// check response
	if respHeader.Magic != hsHeader.Magic {
		return nil, nil, fmt.Errorf("remote peer failed: %v", respHeader.RespCode)
	}
	bestVersion := p2pcommon.P2PVersion(respHeader.RespCode)
	// continue to handshake with VersionedHandshaker
	innerHS, err := h.verM.GetVersionedHandshaker(bestVersion, h.peerID, bufReader, bufWriter)
	if err != nil {
		return nil, nil, err
	}
	status, err := innerHS.DoForOutbound(ctx)
	h.remoteStatus = status
	return innerHS.GetMsgRW(), status, err
}

func (h *baseWireHandshaker) writeWireHSRequest(hsHeader p2pcommon.HSHeadReq, wr p2pcommon.FlushableWriter) (err error) {
	bytes := hsHeader.Marshal()
	sent, err := wr.Write(bytes)
	if err != nil {
		return
	}
	err = wr.Flush()
	if err != nil {
		return
	}
	if sent != len(bytes) {
		return fmt.Errorf("wrong sent size")
	}
	return
}

func (h *baseWireHandshaker) readWireHSRequest(rd io.Reader) (header p2pcommon.HSHeadReq, err error) {
	buf := make([]byte, p2pcommon.HSMagicLength)
	readn, err := p2putil.ReadToLen(rd, buf[:p2pcommon.HSMagicLength])
	if err != nil {
		return
	}
	if readn != p2pcommon.HSMagicLength {
		err = fmt.Errorf("transport error")
		return
	}
	header.Magic = binary.BigEndian.Uint32(buf)
	readn, err = p2putil.ReadToLen(rd, buf[:p2pcommon.HSVerCntLength])
	if err != nil {
		return
	}
	if readn != p2pcommon.HSVerCntLength {
		err = fmt.Errorf("transport error")
		return
	}
	verCount := int(binary.BigEndian.Uint32(buf))
	if verCount <= 0 || verCount > p2pcommon.HSMaxVersionCnt {
		err = fmt.Errorf("invalid version count: %d", verCount)
		return
	}
	versions := make([]p2pcommon.P2PVersion, verCount)
	for i := 0; i < verCount; i++ {
		readn, err = p2putil.ReadToLen(rd, buf[:p2pcommon.HSVersionLength])
		if err != nil {
			return
		}
		if readn != p2pcommon.HSVersionLength {
			err = fmt.Errorf("transport error")
			return
		}
		versions[i] = p2pcommon.P2PVersion(binary.BigEndian.Uint32(buf))
	}
	header.Versions = versions
	return
}

func (h *baseWireHandshaker) writeWireHSResponse(hsHeader p2pcommon.HSHeadResp, wr p2pcommon.FlushableWriter) (err error) {
	bytes := hsHeader.Marshal()
	sent, err := wr.Write(bytes)
	if err != nil {
		return
	}
	err = wr.Flush()
	if err != nil {
		return
	}
	if sent != len(bytes) {
		return fmt.Errorf("wrong sent size")
	}
	return
}

func (h *baseWireHandshaker) writeErrAndReturn(err error, errCode uint32, wr p2pcommon.FlushableWriter) (p2pcommon.MsgReadWriter, *types.Status, error) {
	errResp := p2pcommon.HSHeadResp{p2pcommon.HSError, errCode}
	_ = h.writeWireHSResponse(errResp, wr)
	return nil, nil, err
}
func (h *baseWireHandshaker) readWireHSResp(rd io.Reader) (header p2pcommon.HSHeadResp, err error) {
	bytebuf := make([]byte, p2pcommon.HSMagicLength)
	readn, err := p2putil.ReadToLen(rd, bytebuf[:p2pcommon.HSMagicLength])
	if err != nil {
		return
	}
	if readn != p2pcommon.HSMagicLength {
		err = fmt.Errorf("transport error")
		return
	}
	header.Magic = binary.BigEndian.Uint32(bytebuf)
	readn, err = p2putil.ReadToLen(rd, bytebuf[:p2pcommon.HSVersionLength])
	if err != nil {
		return
	}
	if readn != p2pcommon.HSVersionLength {
		err = fmt.Errorf("transport error")
		return
	}
	header.RespCode = binary.BigEndian.Uint32(bytebuf)
	return
}

func (h *baseWireHandshaker) selectVersionedHandshaker(version p2pcommon.P2PVersion, r io.Reader, w io.Writer) (p2pcommon.VersionedHandshaker, error) {
	switch version {
	// p2p version 0.3.1 is just changed in initial connecting and version verification. chain verification is same as 0.3.0
	case p2pcommon.P2PVersion031:
		// TODO:
		v030hs := v030.NewV030StateHS(h.pm, h.actor, h.logger, h.localChainID, h.peerID, r, w)
		return v030hs, nil
	case p2pcommon.P2PVersion030:
		v030hs := v030.NewV030StateHS(h.pm, h.actor, h.logger, h.localChainID, h.peerID, r, w)
		return v030hs, nil
	default:
		return nil, fmt.Errorf("not supported version")
	}
}

func (h *baseWireHandshaker) findBestProtocolVersion(versions []p2pcommon.P2PVersion) p2pcommon.P2PVersion {
	for _, suppored := range CurrentSupported {
		for _, reqVer := range versions {
			if suppored == reqVer {
				return reqVer
			}
		}
	}
	return p2pcommon.P2PVersionUnknown
}
