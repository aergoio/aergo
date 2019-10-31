/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
)

type issueCertRequestHandler struct {
	BaseMsgHandler
	cm p2pcommon.CertificateManager
}

var _ p2pcommon.MessageHandler = (*issueCertRequestHandler)(nil)

type issueCertResponseHandler struct {
	BaseMsgHandler
	cm p2pcommon.CertificateManager
}

var _ p2pcommon.MessageHandler = (*issueCertResponseHandler)(nil)

// newAddressesReqHandler creates handler for PingRequest
func NewIssueCertReqHandler(pm p2pcommon.PeerManager, cm p2pcommon.CertificateManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *issueCertRequestHandler {
	ph := &issueCertRequestHandler{BaseMsgHandler{protocol: p2pcommon.IssueCertificateRequest, pm: pm, peer: peer, actor: actor, logger: logger}, cm } //TODO fill CertificateManager
	return ph
}

func (h *issueCertRequestHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.IssueCertificateRequest{})
}

func (h *issueCertRequestHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := h.peer
	// data := msgBody.(*types.IssueCertificateRequest)
	p2putil.DebugLogReceive(h.logger, h.protocol, msg.ID().String(), remotePeer, nil)

	resp := &types.IssueCertificateResponse{}
	cert, err := h.cm.CreateCertificate(remotePeer.Meta())
	if err != nil {
		if err == p2pcommon.ErrInvalidRole {
			resp.Status = types.ResultStatus_PERMISSION_DENIED
		} else {
			resp.Status = types.ResultStatus_UNAVAILABLE
		}
	} else {
		pcert, err := p2putil.ConvertCertToProto(cert)
		if err != nil {
			resp.Status = types.ResultStatus_INTERNAL
		} else {
			resp.Status = types.ResultStatus_OK
			resp.Certificate = pcert
		}
	}
	remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.IssueCertificateResponse, resp))
}

// newAddressesRespHandler creates handler for PingRequest
func NewIssueCertRespHandler(pm p2pcommon.PeerManager, cm p2pcommon.CertificateManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *issueCertResponseHandler {
	ph := &issueCertResponseHandler{BaseMsgHandler{protocol: p2pcommon.IssueCertificateResponse, pm: pm, peer: peer, actor: actor, logger: logger}, cm}
	return ph
}

func (h *issueCertResponseHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.IssueCertificateResponse{})
}

func (h *issueCertResponseHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := h.peer
	data := msgBody.(*types.IssueCertificateResponse)
	p2putil.DebugLogReceiveResponse(h.logger, h.protocol, msg.ID().String(), msg.OriginalID().String(), remotePeer, data)

	remotePeer.ConsumeRequest(msg.OriginalID())
	if data.Status == types.ResultStatus_OK {
		cert, err := p2putil.CheckAndGetV1(data.Certificate)
		if err == nil {
			h.cm.AddCertificate(cert)
		} else {
			h.logger.Info().Err(err).Msg("Failed to convert certificate")
		}
	}
}

type certRenewedNoticeHandler struct {
	BaseMsgHandler
	cm p2pcommon.CertificateManager
}


// newAddressesRespHandler creates handler for PingRequest
func NewCertRenewedNoticeHandler(pm p2pcommon.PeerManager, cm p2pcommon.CertificateManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *certRenewedNoticeHandler {
	ph := &certRenewedNoticeHandler{BaseMsgHandler{protocol: p2pcommon.CertificateRenewedNotice, pm: pm, peer: peer, actor: actor, logger: logger}, cm}
	return ph
}

func (h *certRenewedNoticeHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.CertificateRenewedNotice{})
}

func (h *certRenewedNoticeHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := h.peer
	data := msgBody.(*types.CertificateRenewedNotice)
	p2putil.DebugLogReceive(h.logger, h.protocol, msg.ID().String(), remotePeer, data)

	// TODO check authorization (i.e. if remotePeer is agent )
	if remotePeer.Role() != types.PeerRole_Agent {
		h.logger.Debug().Str("role", remotePeer.Role().String()).Str(p2putil.LogPeerName, remotePeer.Name()).Msg("invalid peer status. peer is not agent")
		return
	}
	// verify certificate
	cert, err := p2putil.CheckAndGetV1(data.Certificate)
	if err != nil {
		h.logger.Debug().Str(p2putil.LogPeerName, remotePeer.Name()).Err(err).Msg("cert verification failed")
		return
	}
	remotePeer.AddCertificate(cert)
}