/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package v200

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/internal/network"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	v030 "github.com/aergoio/aergo/v2/p2p/v030"
	"github.com/aergoio/aergo/v2/types"
)

var (
	ErrInvalidAgentStatus = errors.New("invalid agent status")
	ErrInvalidCertIssue   = errors.New("invalid issue request")
)

// V200Handshaker exchange status data over protocol version 1.0.0
type V200Handshaker struct {
	is p2pcommon.InternalService
	cm p2pcommon.CertificateManager
	vm p2pcommon.VersionedManager

	selfMeta p2pcommon.PeerMeta

	logger *log.Logger
	peerID types.PeerID

	msgRW p2pcommon.MsgReadWriter

	localGenesisHash []byte

	remoteMeta  p2pcommon.PeerMeta
	remoteCerts []*p2pcommon.AgentCertificateV1
	remoteHash  types.BlockID
	remoteNo    types.BlockNo
}

var _ p2pcommon.VersionedHandshaker = (*V200Handshaker)(nil)

func (h *V200Handshaker) GetMsgRW() p2pcommon.MsgReadWriter {
	return h.msgRW
}

func NewV200VersionedHS(is p2pcommon.InternalService, log *log.Logger, vm p2pcommon.VersionedManager, cm p2pcommon.CertificateManager, peerID types.PeerID, rwc io.ReadWriteCloser, genesis []byte) *V200Handshaker {
	h := &V200Handshaker{selfMeta: is.SelfMeta(), is: is, logger: log, peerID: peerID, localGenesisHash: genesis, vm: vm, cm: cm}
	// msg format is not changed
	h.msgRW = v030.NewV030MsgPipe(rwc)
	return h
}

// handshakeOutboundPeer start handshake with outbound peer
func (h *V200Handshaker) DoForOutbound(ctx context.Context) (*p2pcommon.HandshakeResult, error) {
	h.logger.Debug().Stringer(p2putil.LogPeerID, types.LogPeerShort(h.peerID)).Msg("Starting versioned handshake for outbound peer connection")

	// find my best block
	bestBlock, err := h.is.GetChainAccessor().GetBestBlock()
	if err != nil {
		return nil, err
	}
	localID := h.vm.GetChainID(bestBlock.Header.BlockNo)

	status, err := h.createLocalStatus(localID, bestBlock)
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
		hsResult := &p2pcommon.HandshakeResult{Meta: h.remoteMeta, BestBlockHash: h.remoteHash, BestBlockNo: h.remoteNo, MsgRW: h.msgRW, Certificates: h.remoteCerts, Hidden: remotePeerStatus.NoExpose}
		return hsResult, nil
	}
}

func (h *V200Handshaker) sendLocalStatus(ctx context.Context, hostStatus *types.Status) error {
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

func (h *V200Handshaker) receiveRemoteStatus(ctx context.Context) (*types.Status, error) {
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

	return remotePeerStatus, nil
}

func (h *V200Handshaker) checkRemoteStatus(remotePeerStatus *types.Status) error {
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
		return fmt.Errorf("different chainID : %s", remoteChainID.ToJSON())
	}

	h.remoteHash, err = types.ParseToBlockID(remotePeerStatus.BestBlockHash)
	if err != nil {
		h.sendGoAway("wrong block hash")
		return err
	}
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

	// do additional check for genesises are identical
	genHash := h.localGenesisHash
	if !bytes.Equal(genHash, remotePeerStatus.Genesis) {
		h.sendGoAway("different genesis block")
		return fmt.Errorf("different genesis block local: %v , remote %v", enc.ToString(genHash), enc.ToString(remotePeerStatus.Genesis))
	}

	h.remoteMeta = rMeta

	if err = h.checkByRole(remotePeerStatus); err != nil {
		h.sendGoAway("invalid certificate works")
		return fmt.Errorf("invalid certificate info: %v", err.Error())
	}

	return nil
}

// DoForInbound is handle handshake from inbound peer
func (h *V200Handshaker) DoForInbound(ctx context.Context) (*p2pcommon.HandshakeResult, error) {
	h.logger.Debug().Stringer(p2putil.LogPeerID, types.LogPeerShort(h.peerID)).Msg("Starting versioned handshake for inbound peer connection")

	// inbound: receive, check and send
	remotePeerStatus, err := h.receiveRemoteStatus(ctx)
	if err != nil {
		return nil, err
	}
	if err = h.checkRemoteStatus(remotePeerStatus); err != nil {
		return nil, err
	}

	bestBlock, err := h.is.GetChainAccessor().GetBestBlock()
	if err != nil {
		return nil, err
	}
	localID := h.vm.GetChainID(bestBlock.Header.BlockNo)

	// send my status message as response
	localStatus, err := h.createLocalStatus(localID, bestBlock)
	if err != nil {
		h.logger.Warn().Err(err).Msg("Failed to create localStatus message.")
		h.sendGoAway("internal error")
		return nil, err
	}
	err = h.sendLocalStatus(ctx, localStatus)
	if err != nil {
		return nil, err
	}
	hsResult := &p2pcommon.HandshakeResult{Meta: h.remoteMeta, BestBlockHash: h.remoteHash, BestBlockNo: h.remoteNo, MsgRW: h.msgRW, Certificates: h.remoteCerts, Hidden: remotePeerStatus.NoExpose}
	return hsResult, nil
}

func (h *V200Handshaker) handleGoAway(peerID types.PeerID, data p2pcommon.Message) (*types.Status, error) {
	goAway := &types.GoAwayNotice{}
	if err := p2putil.UnmarshalMessageBody(data.Payload(), goAway); err != nil {
		h.logger.Warn().Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Err(err).Msg("Remote peer sent goAway but failed to decode internal message")
		return nil, err
	}
	return nil, fmt.Errorf("remote peer refuse handshake: %s", goAway.GetMessage())
}

func (h *V200Handshaker) sendGoAway(msg string) {
	goMsg := createMessage(p2pcommon.GoAway, p2pcommon.NewMsgID(), &types.GoAwayNotice{Message: msg})
	if goMsg != nil {
		h.msgRW.WriteMsg(goMsg)
	}
}

func (h *V200Handshaker) checkByRole(status *types.Status) error {
	if h.remoteMeta.Role == types.PeerRole_Agent {
		err := h.checkAgent(status)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *V200Handshaker) checkAgent(status *types.Status) error {
	h.logger.Debug().Int("certCnt", len(status.Certificates)).Stringer(p2putil.LogPeerID, types.LogPeerShort(h.remoteMeta.ID)).Msg("checking peer as agent")

	// Agent must have at least one block producer
	if len(h.remoteMeta.ProducerIDs) == 0 {
		return ErrInvalidAgentStatus
	}
	producers := make(map[types.PeerID]bool)
	for _, id := range h.remoteMeta.ProducerIDs {
		producers[id] = true
	}
	certs := make([]*p2pcommon.AgentCertificateV1, len(status.Certificates))
	for i, pCert := range status.Certificates {
		cert, err := p2putil.CheckAndGetV1(pCert)
		if err != nil {
			h.logger.Info().Err(err).Stringer(p2putil.LogPeerID, types.LogPeerShort(h.remoteMeta.ID)).Msg("invalid agent certificate")
			return ErrInvalidAgentStatus
		}
		// check certificate
		if !types.IsSamePeerID(cert.AgentID, h.remoteMeta.ID) {
			h.logger.Info().Err(err).Stringer(p2putil.LogPeerID, types.LogPeerShort(h.remoteMeta.ID)).Msg("certificate is not for this agent")
			return ErrInvalidAgentStatus
		}
		if _, exist := producers[cert.BPID]; !exist {
			h.logger.Info().Err(err).Stringer(p2putil.LogPeerID, types.LogPeerShort(h.remoteMeta.ID)).Stringer("bpID", types.LogPeerShort(cert.BPID)).Msg("peer id of certificate not matched")
			return ErrInvalidAgentStatus
		}

		certs[i] = cert
	}
	h.remoteCerts = certs
	return nil
}

func (h *V200Handshaker) createLocalStatus(chainID *types.ChainID, bestBlock *types.Block) (*types.Status, error) {
	selfAddr := h.selfMeta.ToPeerAddress()
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
		NoExpose:      h.selfMeta.Hidden,
		Version:       p2pkey.NodeVersion(),
		Genesis:       h.localGenesisHash,
	}

	if h.selfMeta.Role == types.PeerRole_Agent {
		cs := h.cm.GetCertificates()
		h.logger.Debug().Int("certCnt", len(cs)).Msg("appending local certificates to status")
		pcs, err := p2putil.ConvertCertsToProto(cs)
		if err != nil {
			h.logger.Error().Err(err).Msg("failed to convert certificates")
			return nil, errors.New("internal error")
		}
		statusMsg.Certificates = pcs
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
