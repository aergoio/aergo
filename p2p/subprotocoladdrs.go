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
func newAddressesReqHandler(pm PeerManager, peer *RemotePeer, logger *log.Logger) *addressesRequestHandler {
	ph := &addressesRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: addressesRequest, pm: pm, peer: peer, actor: peer.actorServ, logger: logger}}
	return ph
}

func (ph *addressesRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.AddressesRequest{})
}

func (ph *addressesRequestHandler) handle(msgHeader *types.MessageData, msgBody proto.Message) {
	peerID := ph.peer.ID()
	remotePeer := ph.peer
	data := msgBody.(*types.AddressesRequest)
	debugLogReceiveMsg(ph.logger, ph.protocol, msgHeader.GetId(), peerID, nil)

	// generate response message
	resp := &types.AddressesResponse{}
	var addrList = make([]*types.PeerAddress, 0, len(ph.pm.GetPeers()))
	for _, aPeer := range ph.pm.GetPeers() {
		// exclude not running peer and requesting peer itself
		// TODO: apply peer status after fix status management bug
		if aPeer.meta.ID == peerID {
			continue
		}
		pAddr := aPeer.meta.ToPeerAddress()
		addrList = append(addrList, &pAddr)
	}
	resp.Peers = addrList
	// send response
	remotePeer.sendMessage(newPbMsgResponseOrder(data.MessageData.Id, true, addressesResponse, resp))
}

func (ph *addressesResponseHandler) checkAndAddPeerAddresses(peers []*types.PeerAddress) {
	selfPeerID := ph.pm.ID()
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
func newAddressesRespHandler(pm PeerManager, peer *RemotePeer, logger *log.Logger) *addressesResponseHandler {
	ph := &addressesResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: addressesResponse, pm: pm, peer: peer, actor: peer.actorServ, logger: logger}}
	return ph
}

func (ph *addressesResponseHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.AddressesResponse{})
}

func (ph *addressesResponseHandler) handle(msgHeader *types.MessageData, msgBody proto.Message) {
	peerID := ph.peer.ID()
	remotePeer := ph.peer
	data := msgBody.(*types.AddressesResponse)
	debugLogReceiveMsg(ph.logger, ph.protocol, msgHeader.GetId(), peerID, len(data.GetPeers()))

	remotePeer.consumeRequest(msgHeader.GetId())
	if len(data.GetPeers()) > 0 {
		ph.checkAndAddPeerAddresses(data.GetPeers())
	}
}
