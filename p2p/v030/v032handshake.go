/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package v030

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

// V032Handshaker exchange status data over protocol version .0.3.1
// it
type V032Handshaker struct {
	V030Handshaker
	localGenesisHash []byte
}

var _ p2pcommon.VersionedHandshaker = (*V032Handshaker)(nil)

func (h *V032Handshaker) GetMsgRW() p2pcommon.MsgReadWriter {
	return h.msgRW
}

func NewV032VersionedHS(pm p2pcommon.PeerManager, actor p2pcommon.ActorService, log *log.Logger, chainID *types.ChainID, peerID types.PeerID, rwc io.ReadWriteCloser, genesis []byte) *V032Handshaker {
	h := &V032Handshaker{V030Handshaker{pm: pm, actor: actor, logger: log, chainID: chainID, peerID: peerID},
		genesis}
	h.msgRW = NewV030MsgPipe(rwc)
	return h
}

func (h *V032Handshaker) checkRemoteStatus(remotePeerStatus *types.Status) error {
	// v0.3.2 just added genesis hash
	if err := h.V030Handshaker.checkRemoteStatus(remotePeerStatus); err != nil {
		return err
	}
	// do additional check for genesises are identical
	genHash := h.localGenesisHash
	if !bytes.Equal(genHash, remotePeerStatus.Genesis) {
		h.sendGoAway("different genesis block")
		return fmt.Errorf("different genesis block local: %v , remote %v", enc.ToString(genHash), enc.ToString(remotePeerStatus.Genesis))
	}

	return nil
}

func (h *V032Handshaker) DoForOutbound(ctx context.Context) (*p2pcommon.HandshakeResult, error) {
	h.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(h.peerID)).Msg("Starting versioned handshake for outbound peer connection")

	bestBlock, err := h.actor.GetChainAccessor().GetBestBlock()
	if err != nil {
		return nil, err
	}

	// outbound: send, receive and check
	localStatus, err := createLocalStatus(h.pm, h.chainID, bestBlock, h.localGenesisHash)
	if err != nil {
		h.logger.Warn().Err(err).Msg("Failed to create status message.")
		h.sendGoAway("internal error")
		return nil, err
	}
	err = h.sendLocalStatus(ctx, localStatus)
	if err != nil {
		return nil, err
	}

	remotePeerStatus, err := h.receiveRemoteStatus(ctx)
	if err != nil {
		return nil, err
	}

	if err = h.checkRemoteStatus(remotePeerStatus); err != nil {
		return nil, err
	} else {
		hsResult := &p2pcommon.HandshakeResult{Meta: h.remoteMeta, BestBlockHash: h.remoteHash, BestBlockNo: h.remoteNo, MsgRW: h.msgRW, Hidden: remotePeerStatus.NoExpose}
		return hsResult, nil
	}

}

func (h *V032Handshaker) DoForInbound(ctx context.Context) (*p2pcommon.HandshakeResult, error) {
	h.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(h.peerID)).Msg("Starting versioned handshake for inbound peer connection")

	// inbound: receive, check and send
	remotePeerStatus, err := h.receiveRemoteStatus(ctx)
	if err != nil {
		return nil, err
	}
	if err = h.checkRemoteStatus(remotePeerStatus); err != nil {
		return nil, err
	}
	bestBlock, err := h.actor.GetChainAccessor().GetBestBlock()
	if err != nil {
		return nil, err
	}

	// send my status message as response
	localStatus, err := createLocalStatus(h.pm, h.chainID, bestBlock, h.localGenesisHash)
	if err != nil {
		h.logger.Warn().Err(err).Msg("Failed to create status message.")
		h.sendGoAway("internal error")
		return nil, err
	}
	err = h.sendLocalStatus(ctx, localStatus)
	if err != nil {
		return nil, err
	}
	hsResult := &p2pcommon.HandshakeResult{Meta: h.remoteMeta, BestBlockHash: h.remoteHash, BestBlockNo: h.remoteNo, MsgRW: h.msgRW, Hidden: remotePeerStatus.NoExpose}
	return hsResult, nil
}
