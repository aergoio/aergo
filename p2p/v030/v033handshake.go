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
	"github.com/aergoio/aergo/v2/internal/network"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

// V033Handshaker exchange status data over protocol version .0.3.1
// it
type V033Handshaker struct {
	V032Handshaker
	vm p2pcommon.VersionedManager
}

var _ p2pcommon.VersionedHandshaker = (*V033Handshaker)(nil)

func (h *V033Handshaker) GetMsgRW() p2pcommon.MsgReadWriter {
	return h.msgRW
}

func NewV033VersionedHS(pm p2pcommon.PeerManager, actor p2pcommon.ActorService, log *log.Logger, vm p2pcommon.VersionedManager, peerID types.PeerID, rwc io.ReadWriteCloser, genesis []byte) *V033Handshaker {
	v032 := NewV032VersionedHS(pm, actor, log, vm.GetChainID(0), peerID, rwc, genesis)
	h := &V033Handshaker{V032Handshaker: *v032, vm: vm}

	return h
}

func (h *V033Handshaker) checkRemoteStatus(remotePeerStatus *types.Status) error {
	// v030 checking
	// check if chainID is same or not
	remoteChainID := types.NewChainID()
	err := remoteChainID.Read(remotePeerStatus.ChainID)
	if err != nil {
		h.sendGoAway("wrong status")
		return err
	}
	localID := h.vm.GetChainID(remotePeerStatus.BestHeight)
	if !localID.Equals(remoteChainID) {
		h.sendGoAway("different chainID")
		return fmt.Errorf("different chainID : local is %s, remote is %s at bloco no %d", p2putil.PrintChainID(localID), p2putil.PrintChainID(remoteChainID), remotePeerStatus.BestHeight)
	}

	// handshake v0.3.x don't check format of block hash
	h.remoteHash, _ = types.ParseToBlockID(remotePeerStatus.BestBlockHash)
	h.remoteNo = remotePeerStatus.BestHeight

	peerAddress := remotePeerStatus.Sender
	if peerAddress == nil || network.CheckAddressType(peerAddress.Address) == network.AddressTypeError {
		h.sendGoAway("invalid peer address")
		return fmt.Errorf("invalid peer address : %s", peerAddress)
	}

	rMeta := p2pcommon.NewMetaFromStatus(remotePeerStatus)
	if rMeta.ID != h.peerID {
		h.logger.Debug().Str("received_peer_id", rMeta.ID.Pretty()).Stringer(p2putil.LogPeerID, types.LogPeerShort(h.peerID)).Msg("Inconsistent peerID")
		h.sendGoAway("Inconsistent peerID")
		return fmt.Errorf("inconsistent peerID")
	}
	h.remoteMeta = rMeta

	// check if genesis hashes are identical
	genHash := h.localGenesisHash
	if !bytes.Equal(genHash, remotePeerStatus.Genesis) {
		h.sendGoAway("different genesis block")
		return fmt.Errorf("different genesis block local: %v , remote %v", enc.ToString(genHash), enc.ToString(remotePeerStatus.Genesis))
	}

	return nil
}

func (h *V033Handshaker) DoForOutbound(ctx context.Context) (*p2pcommon.HandshakeResult, error) {
	h.logger.Debug().Stringer(p2putil.LogPeerID, types.LogPeerShort(h.peerID)).Msg("Starting versioned handshake for outbound peer connection")

	// find my best block
	bestBlock, err := h.actor.GetChainAccessor().GetBestBlock()
	if err != nil {
		return nil, err
	}
	localID := h.vm.GetChainID(bestBlock.Header.BlockNo)

	// outbound: send, receive and check
	localStatus, err := createLocalStatus(h.pm, localID, bestBlock, h.localGenesisHash)
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

func (h *V033Handshaker) DoForInbound(ctx context.Context) (*p2pcommon.HandshakeResult, error) {
	h.logger.Debug().Stringer(p2putil.LogPeerID, types.LogPeerShort(h.peerID)).Msg("Starting versioned handshake for inbound peer connection")

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
	localID := h.vm.GetChainID(bestBlock.Header.BlockNo)

	// send my status message as response
	localStatus, err := createLocalStatus(h.pm, localID, bestBlock, h.localGenesisHash)
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
