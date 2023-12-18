/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package v030

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/network"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

// V030Handshaker exchange status data over protocol version .0.3.0
type V030Handshaker struct {
	pm      p2pcommon.PeerManager
	actor   p2pcommon.ActorService
	logger  *log.Logger
	peerID  types.PeerID
	chainID *types.ChainID

	msgRW      p2pcommon.MsgReadWriter
	remoteMeta p2pcommon.PeerMeta
	remoteHash types.BlockID
	remoteNo   types.BlockNo
}

var _ p2pcommon.VersionedHandshaker = (*V030Handshaker)(nil)

func (h *V030Handshaker) GetMsgRW() p2pcommon.MsgReadWriter {
	return h.msgRW
}

func NewV030VersionedHS(pm p2pcommon.PeerManager, actor p2pcommon.ActorService, log *log.Logger, chainID *types.ChainID, peerID types.PeerID, rwc io.ReadWriteCloser) *V030Handshaker {
	h := &V030Handshaker{pm: pm, actor: actor, logger: log, chainID: chainID, peerID: peerID}
	h.msgRW = NewV030MsgPipe(rwc)
	return h
}

// handshakeOutboundPeer start handshake with outbound peer
func (h *V030Handshaker) DoForOutbound(ctx context.Context) (*p2pcommon.HandshakeResult, error) {
	h.logger.Debug().Stringer(p2putil.LogPeerID, types.LogPeerShort(h.peerID)).Msg("Starting versioned handshake for outbound peer connection")
	bestBlock, err := h.actor.GetChainAccessor().GetBestBlock()
	if err != nil {
		return nil, err
	}

	status, err := createLocalStatus(h.pm, h.chainID, bestBlock, nil)
	if err != nil {
		h.logger.Warn().Err(err).Msg("Failed to create status message.")
		h.sendGoAway("internal error")
		return nil, err
	}

	err = h.sendLocalStatus(ctx, status)
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

func (h *V030Handshaker) sendLocalStatus(ctx context.Context, hostStatus *types.Status) error {
	var err error
	container := createMessage(p2pcommon.StatusRequest, p2pcommon.NewMsgID(), hostStatus)
	if container == nil {
		h.logger.Warn().Stringer(p2putil.LogPeerID, types.LogPeerShort(h.peerID)).Msg("failed to create p2p message")
		h.sendGoAway("internal error")
		// h.logger.Warn().Str(LogPeerID, ShortForm(peerID)).Err(err).Msg("failed to create p2p message")
		return fmt.Errorf("failed to craete container message")
	}
	if err = h.msgRW.WriteMsg(container); err != nil {
		h.logger.Info().Stringer(p2putil.LogPeerID, types.LogPeerShort(h.peerID)).Err(err).Msg("failed to write local status ")
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// go on
	}
	return nil
}

func (h *V030Handshaker) receiveRemoteStatus(ctx context.Context) (*types.Status, error) {
	// and wait to response status
	data, err := h.msgRW.ReadMsg()
	if err != nil {
		h.sendGoAway("malformed message")
		// h.logger.Info().Err(err).Msg("fail to decode")
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// go on
	}
	if data.Subprotocol() != p2pcommon.StatusRequest {
		if data.Subprotocol() == p2pcommon.GoAway {
			return h.handleGoAway(h.peerID, data)
		} else {
			h.logger.Info().Stringer(p2putil.LogPeerID, types.LogPeerShort(h.peerID)).Str("expected", p2pcommon.StatusRequest.String()).Str("actual", data.Subprotocol().String()).Msg("unexpected message type")
			h.sendGoAway("unexpected message type")
			return nil, fmt.Errorf("unexpected message type")
		}
	}

	remotePeerStatus := &types.Status{}
	err = p2putil.UnmarshalMessageBody(data.Payload(), remotePeerStatus)
	if err != nil {
		h.sendGoAway("malformed status message")
		return nil, err
	}

	// convert old fashioned data structure to current versions.
	//    modify address information to array, and set version to peerAddress
	sender := remotePeerStatus.Sender
	if sender == nil {
		h.sendGoAway("malformed status message")
		return nil, fmt.Errorf("malformed status message")
	}
	if len(sender.Addresses) == 0 {
		ma, err := types.ToMultiAddr(sender.Address, sender.Port)
		if err != nil {
			h.sendGoAway("malformed status message")
			return nil, err
		}
		sender.Addresses = append(sender.Addresses, ma.String())
	}
	sender.Version = remotePeerStatus.Version

	return remotePeerStatus, nil
}

func (h *V030Handshaker) checkRemoteStatus(remotePeerStatus *types.Status) error {
	// check if chainID is same or not
	remoteChainID := types.NewChainID()
	err := remoteChainID.Read(remotePeerStatus.ChainID)
	if err != nil {
		h.sendGoAway("wrong status")
		return err
	}
	if !h.chainID.Equals(remoteChainID) {
		h.sendGoAway("different chainID")
		return fmt.Errorf("different chainID : %s", remoteChainID.ToJSON())
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
		h.logger.Debug().Str("received_peer_id", rMeta.ID.String()).Stringer(p2putil.LogPeerID, types.LogPeerShort(h.peerID)).Msg("Inconsistent peerID")
		h.sendGoAway("Inconsistent peerID")
		return fmt.Errorf("Inconsistent peerID")
	}
	h.remoteMeta = rMeta

	return nil
}

// DoForInbound is handle handshake from inbound peer
func (h *V030Handshaker) DoForInbound(ctx context.Context) (*p2pcommon.HandshakeResult, error) {
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

	// send my localStatus message as response
	localStatus, err := createLocalStatus(h.pm, h.chainID, bestBlock, nil)
	if err != nil {
		h.logger.Warn().Err(err).Msg("Failed to create localStatus message.")
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

func (h *V030Handshaker) handleGoAway(peerID types.PeerID, data p2pcommon.Message) (*types.Status, error) {
	goAway := &types.GoAwayNotice{}
	if err := p2putil.UnmarshalMessageBody(data.Payload(), goAway); err != nil {
		h.logger.Warn().Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Err(err).Msg("Remote peer sent goAway but failed to decode internal message")
		return nil, err
	}
	return nil, fmt.Errorf("remote peer refuse handshake: %s", goAway.GetMessage())
}

func (h *V030Handshaker) sendGoAway(msg string) {
	goMsg := createMessage(p2pcommon.GoAway, p2pcommon.NewMsgID(), &types.GoAwayNotice{Message: msg})
	if goMsg != nil {
		h.msgRW.WriteMsg(goMsg)
	}
}

func createLocalStatus(pm p2pcommon.PeerManager, chainID *types.ChainID, bestBlock *types.Block, genesis []byte) (*types.Status, error) {
	selfAddr := pm.SelfMeta().ToPeerAddress()
	// for backward compatibility
	selfAddr.Address = pm.SelfMeta().PrimaryAddress()
	selfAddr.Port = pm.SelfMeta().PrimaryPort()

	chainIDbytes, err := chainID.Bytes()
	if err != nil {
		return nil, err
	}
	// create message data
	statusMsg := &types.Status{
		Sender:        &selfAddr,
		ChainID:       chainIDbytes,
		BestBlockHash: bestBlock.BlockHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
		NoExpose:      pm.SelfMeta().Hidden,
		Version:       pm.SelfMeta().Version,
		Genesis:       genesis,
	}

	return statusMsg, nil
}

func createMessage(protocolID p2pcommon.SubProtocol, msgID p2pcommon.MsgID, msgBody p2pcommon.MessageBody) p2pcommon.Message {
	bytes, err := p2putil.MarshalMessageBody(msgBody)
	if err != nil {
		return nil
	}

	msg := p2pcommon.NewMessageValue(protocolID, msgID, p2pcommon.EmptyID, time.Now().UnixNano(), bytes)
	return msg
}
