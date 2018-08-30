/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/types"

	"github.com/libp2p/go-libp2p-peer"
)

// remote peer requests handler
func (p *PingHandler) handleAddressesRequest(msg *types.P2PMessage) {
	peerID := p.peer.ID()
	remotePeer := p.peer

	// get request dataㅕ
	data := &types.AddressesRequest{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		p.logger.Info().Err(err).Msg("fail to decode")
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, p.peer.ID(), nil)

	// generate response message
	resp := &types.AddressesResponse{MessageData: &types.MessageData{}}
	var addrList = make([]*types.PeerAddress, 0, len(p.pm.GetPeers()))
	for _, aPeer := range p.pm.GetPeers() {
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

func (p *PingHandler) checkAndAddPeerAddresses(peers []*types.PeerAddress) {
	selfPeerID := p.pm.ID()
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
		p.pm.NotifyPeerAddressReceived(peerMetas)
	}
}

// remote ping response handler
func (p *PingHandler) handleAddressesResponse(msg *types.P2PMessage) {
	peerID := p.peer.ID()
	remotePeer := p.peer

	data := &types.AddressesResponse{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, len(data.GetPeers()))
	valid := p.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.logger.Info().Msg("Failed to authenticate message")
		return
	}

	remotePeer.consumeRequest(data.MessageData.Id)
	if len(data.GetPeers()) > 0 {
		p.checkAndAddPeerAddresses(data.GetPeers())
	}
}
