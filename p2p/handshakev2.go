/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

// baseWireHandshaker works to handshake to just connected peer, it detect chain networks
// and protocol versions, and then select InnerHandshaker for that protocol version.
type baseWireHandshaker struct {
	pm     p2pcommon.PeerManager
	actor  p2pcommon.ActorService
	verM   p2pcommon.VersionedManager
	logger *log.Logger
	peerID types.PeerID
	// check if is it ad hoc
	localChainID *types.ChainID

	remoteStatus *types.Status

	result     p2pcommon.HandshakeResult
	remoteMeta p2pcommon.PeerMeta
	remoteCert []*p2pcommon.AgentCertificateV1
}

type InboundWireHandshaker struct {
	baseWireHandshaker
}

func NewInboundHSHandler(pm p2pcommon.PeerManager, actor p2pcommon.ActorService, verManager p2pcommon.VersionedManager, log *log.Logger, chainID *types.ChainID, peerID types.PeerID) p2pcommon.HSHandler {
	return &InboundWireHandshaker{baseWireHandshaker{pm: pm, actor: actor, verM: verManager, logger: log, localChainID: chainID, peerID: peerID}}
}

func (h *InboundWireHandshaker) Handle(s io.ReadWriteCloser, ttl time.Duration) (*p2pcommon.HandshakeResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()
	return h.handleInboundPeer(ctx, s)
}

func (h *InboundWireHandshaker) handleInboundPeer(ctx context.Context, rwc io.ReadWriteCloser) (*p2pcommon.HandshakeResult, error) {
	// wait initial hs message
	hsReq, err := h.readWireHSRequest(rwc)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// go on
	}
	if err != nil {
		return h.writeErrAndReturn(err, p2pcommon.HSCodeWrongHSReq, rwc)
	}
	// check magic
	if hsReq.Magic != p2pcommon.MAGICMain {
		return h.writeErrAndReturn(fmt.Errorf("wrong magic %v", hsReq.Magic), p2pcommon.HSCodeWrongHSReq, rwc)
	}

	// continue to handshake with VersionedHandshaker
	bestVer := h.verM.FindBestP2PVersion(hsReq.Versions)
	if bestVer == p2pcommon.P2PVersionUnknown {
		return h.writeErrAndReturn(fmt.Errorf("no matched p2p version for %v", hsReq.Versions), p2pcommon.HSCodeNoMatchedVersion, rwc)
	} else {
		h.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(h.peerID)).Str("version", bestVer.String()).Msg("Responding best p2p version")
		resp := p2pcommon.HSHeadResp{Magic: hsReq.Magic, RespCode: bestVer.Uint32()}
		err = h.writeWireHSResponse(resp, rwc)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// go on
		}
		if err != nil {
			return nil, err
		}
	}
	innerHS, err := h.verM.GetVersionedHandshaker(bestVer, h.peerID, rwc)
	if err != nil {
		return nil, err
	}
	return innerHS.DoForInbound(ctx)
}

type OutboundWireHandshaker struct {
	baseWireHandshaker
}

func NewOutboundHSHandler(pm p2pcommon.PeerManager, actor p2pcommon.ActorService, verManager p2pcommon.VersionedManager, log *log.Logger, chainID *types.ChainID, peerID types.PeerID) p2pcommon.HSHandler {
	return &OutboundWireHandshaker{baseWireHandshaker{pm: pm, actor: actor, verM: verManager, logger: log, localChainID: chainID, peerID: peerID}}
}

func (h *OutboundWireHandshaker) Handle(s io.ReadWriteCloser, ttl time.Duration) (*p2pcommon.HandshakeResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()
	return h.handleOutboundPeer(ctx, s)
}

func (h *OutboundWireHandshaker) handleOutboundPeer(ctx context.Context, rwc io.ReadWriteCloser) (*p2pcommon.HandshakeResult, error) {
	// send initial hs message
	versions := p2pcommon.AttemptingOutboundVersions

	hsHeader := p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: versions}
	err := h.writeWireHSRequest(hsHeader, rwc)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// go on
	}
	if err != nil {
		return nil, err
	}

	// read response
	respHeader, err := h.readWireHSResp(rwc)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// go on
	}
	if err != nil {
		return nil, err
	}
	// check response
	if respHeader.Magic != hsHeader.Magic {
		return nil, fmt.Errorf("remote peer failed: %v", respHeader.RespCode)
	}
	bestVersion := p2pcommon.P2PVersion(respHeader.RespCode)
	h.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(h.peerID)).Str("version", bestVersion.String()).Msg("Responded best p2p version")
	// continue to handshake with VersionedHandshaker
	innerHS, err := h.verM.GetVersionedHandshaker(bestVersion, h.peerID, rwc)
	if err != nil {
		return nil, err
	}
	return innerHS.DoForOutbound(ctx)
}

func (h *baseWireHandshaker) writeWireHSRequest(hsHeader p2pcommon.HSHeadReq, wr io.Writer) (err error) {
	bytes := hsHeader.Marshal()
	sent, err := wr.Write(bytes)
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

func (h *baseWireHandshaker) writeWireHSResponse(hsHeader p2pcommon.HSHeadResp, wr io.Writer) (err error) {
	bytes := hsHeader.Marshal()
	sent, err := wr.Write(bytes)
	if err != nil {
		return
	}
	if sent != len(bytes) {
		return fmt.Errorf("wrong sent size")
	}
	return
}

func (h *baseWireHandshaker) writeErrAndReturn(err error, errCode uint32, wr io.Writer) (*p2pcommon.HandshakeResult, error) {
	errResp := p2pcommon.HSHeadResp{Magic: p2pcommon.HSError, RespCode: errCode}
	_ = h.writeWireHSResponse(errResp, wr)
	return nil, err
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
