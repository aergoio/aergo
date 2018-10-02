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
func newAddressesReqHandler(pm PeerManager, peer *RemotePeer, logger *log.Logger, signer msgSigner) *addressesRequestHandler {
	ph := &addressesRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: AddressesRequest, pm: pm, peer: peer, actor: peer.actorServ, logger: logger, signer: signer}}
	return ph
}

func (ph *addressesRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.AddressesRequest{})
}

func (ph *addressesRequestHandler) handle(msgHeader *types.MsgHeader, msgBody proto.Message) {
	peerID := ph.peer.ID()
	remotePeer := ph.peer
	data := msgBody.(*types.AddressesRequest)
	debugLogReceiveMsg(ph.logger, ph.protocol, msgHeader.GetId(), peerID, nil)

	// check sender
	maxPeers := data.MaxSize

	// generate response message
	resp := &types.AddressesResponse{}
	var addrList = make([]*types.PeerAddress, 0, len(ph.pm.GetPeers()))
	addrCount := uint32(0)
	for _, aPeer := range ph.pm.GetPeers() {
		// exclude not running peer and requesting peer itself
		// TODO: apply peer status after fix status management bug
		if aPeer.meta.ID == peerID {
			continue
		}
		pAddr := aPeer.meta.ToPeerAddress()
		addrList = append(addrList, &pAddr)
		addrCount++
		if addrCount >= maxPeers {
			break
		}
	}
	resp.Peers = addrList
	// send response
	remotePeer.sendMessage(remotePeer.mf.newMsgResponseOrder(msgHeader.Id, AddressesResponse, resp))
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
func newAddressesRespHandler(pm PeerManager, peer *RemotePeer, logger *log.Logger, signer msgSigner) *addressesResponseHandler {
	ph := &addressesResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: AddressesResponse, pm: pm, peer: peer, actor: peer.actorServ, logger: logger, signer: signer}}
	return ph
}

func (ph *addressesResponseHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.AddressesResponse{})
}

func (ph *addressesResponseHandler) handle(msgHeader *types.MsgHeader, msgBody proto.Message) {
	peerID := ph.peer.ID()
	remotePeer := ph.peer
	data := msgBody.(*types.AddressesResponse)
	debugLogReceiveMsg(ph.logger, ph.protocol, msgHeader.GetId(), peerID, len(data.GetPeers()))

	remotePeer.consumeRequest(msgHeader.GetId())
	if len(data.GetPeers()) > 0 {
		ph.checkAndAddPeerAddresses(data.GetPeers())
	}
}
