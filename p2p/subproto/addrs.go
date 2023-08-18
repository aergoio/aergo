/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/network"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

type addressesRequestHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*addressesRequestHandler)(nil)

type addressesResponseHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*addressesResponseHandler)(nil)

// newAddressesReqHandler creates handler for PingRequest
func NewAddressesReqHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *addressesRequestHandler {
	ph := &addressesRequestHandler{BaseMsgHandler{protocol: p2pcommon.AddressesRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	return ph
}

func (ph *addressesRequestHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.AddressesRequest{})
}

func (ph *addressesRequestHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	peerID := ph.peer.ID()
	remotePeer := ph.peer
	data := msgBody.(*types.AddressesRequest)
	p2putil.DebugLogReceive(ph.logger, ph.protocol, msg.ID().String(), remotePeer, nil)

	// check sender
	maxPeers := data.MaxSize

	// generate response message
	resp := &types.AddressesResponse{}
	var addrList = make([]*types.PeerAddress, 0, len(ph.pm.GetPeers()))
	addrCount := uint32(0)
	for _, aPeer := range ph.pm.GetPeers() {
		// exclude not running peer and requesting peer itself
		// TODO: apply peer status after fix status management bug
		if aPeer.ID() == peerID {
			continue
		}
		if aPeer.RemoteInfo().Hidden {
			continue
		}

		pAddr := aPeer.Meta().ToPeerAddress()
		addrList = append(addrList, &pAddr)
		addrCount++
		if addrCount >= maxPeers {
			break
		}
	}
	resp.Peers = addrList
	// send response
	remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.AddressesResponse, resp))
}

// TODO need refactoring. This code is not bounded to a specific peer but rather whole peer pool, and cause code duplication in p2p.go
func (ph *addressesResponseHandler) checkAndAddPeerAddresses(peers []*types.PeerAddress) {
	selfPeerID := ph.pm.SelfNodeID()
	peerMetas := make([]p2pcommon.PeerMeta, 0, len(peers))
	for _, rPeerAddr := range peers {
		rPeerID := types.PeerID(rPeerAddr.PeerID)
		if selfPeerID == rPeerID {
			continue
		}
		if network.CheckAddressType(rPeerAddr.Address) == network.AddressTypeError {
			continue
		}
		meta := p2pcommon.FromPeerAddress(rPeerAddr)
		// for backward compatibility. old protocol return just single address in old field
		if len(meta.Addresses) == 0 {
			ma, err := types.ToMultiAddr(rPeerAddr.Address, rPeerAddr.Port)
			if err != nil {
				continue
			}
			meta.Addresses = []types.Multiaddr{ma}
		}

		peerMetas = append(peerMetas, meta)
	}
	if len(peerMetas) > 0 {
		ph.pm.NotifyPeerAddressReceived(peerMetas)
	}
}

// newAddressesRespHandler creates handler for PingRequest
func NewAddressesRespHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *addressesResponseHandler {
	ph := &addressesResponseHandler{BaseMsgHandler{protocol: p2pcommon.AddressesResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return ph
}

func (ph *addressesResponseHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.AddressesResponse{})
}

func (ph *addressesResponseHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := ph.peer
	data := msgBody.(*types.AddressesResponse)
	p2putil.DebugLogReceiveResponse(ph.logger, ph.protocol, msg.ID().String(), msg.OriginalID().String(), remotePeer, data)

	remotePeer.ConsumeRequest(msg.OriginalID())
	if len(data.GetPeers()) > 0 {
		ph.checkAndAddPeerAddresses(data.GetPeers())
	}
}
