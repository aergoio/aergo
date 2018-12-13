/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"

	"github.com/libp2p/go-libp2p-peer"
)

type addressesRequestHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*addressesRequestHandler)(nil)

type addressesResponseHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*addressesResponseHandler)(nil)

// newAddressesReqHandler creates handler for PingRequest
func newAddressesReqHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *addressesRequestHandler {
	ph := &addressesRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: AddressesRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	return ph
}

func (ph *addressesRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.AddressesRequest{})
}

func (ph *addressesRequestHandler) handle(msg Message, msgBody proto.Message) {
	peerID := ph.peer.ID()
	remotePeer := ph.peer
	data := msgBody.(*types.AddressesRequest)
	debugLogReceiveMsg(ph.logger, ph.protocol, msg.ID().String(), peerID, nil)

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
		pAddr := aPeer.Meta().ToPeerAddress()
		addrList = append(addrList, &pAddr)
		addrCount++
		if addrCount >= maxPeers {
			break
		}
	}
	resp.Peers = addrList
	// send response
	remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), AddressesResponse, resp))
}

func (ph *addressesResponseHandler) checkAndAddPeerAddresses(peers []*types.PeerAddress) {
	selfPeerID := ph.pm.SelfNodeID()
	peerMetas := make([]PeerMeta, 0, len(peers))
	for _, rPeerAddr := range peers {
		rPeerID := peer.ID(rPeerAddr.PeerID)
		if selfPeerID == rPeerID {
			continue
		}
		meta := FromPeerAddress(rPeerAddr)
		peerMetas = append(peerMetas, meta)
	}
	if len(peerMetas) > 0 {
		ph.pm.NotifyPeerAddressReceived(peerMetas)
	}
}

// newAddressesRespHandler creates handler for PingRequest
func newAddressesRespHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *addressesResponseHandler {
	ph := &addressesResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: AddressesResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return ph
}

func (ph *addressesResponseHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.AddressesResponse{})
}

func (ph *addressesResponseHandler) handle(msg Message, msgBody proto.Message) {
	peerID := ph.peer.ID()
	remotePeer := ph.peer
	data := msgBody.(*types.AddressesResponse)
	debugLogReceiveResponseMsg(ph.logger, ph.protocol, msg.ID().String(), msg.OriginalID().String(), peerID, len(data.GetPeers()))

	remotePeer.consumeRequest(msg.OriginalID())
	if len(data.GetPeers()) > 0 {
		ph.checkAndAddPeerAddresses(data.GetPeers())
	}
}
