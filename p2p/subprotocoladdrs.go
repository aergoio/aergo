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
func (ph *PingHandler) handleAddressesRequest(msg *types.P2PMessage) {
	peerID := ph.peer.ID()
	remotePeer := ph.peer

	// get request dataㅕ
	data := &types.AddressesRequest{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		ph.logger.Info().Err(err).Msg("fail to decode")
		return
	}
	debugLogReceiveMsg(ph.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, ph.peer.ID(), nil)

	// generate response message
	resp := &types.AddressesResponse{MessageData: &types.MessageData{}}
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

func (ph *PingHandler) checkAndAddPeerAddresses(peers []*types.PeerAddress) {
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

// remote ping response handler
func (ph *PingHandler) handleAddressesResponse(msg *types.P2PMessage) {
	peerID := ph.peer.ID()
	remotePeer := ph.peer

	data := &types.AddressesResponse{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(ph.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, len(data.GetPeers()))
	valid := ph.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		ph.logger.Info().Msg("Failed to authenticate message")
		return
	}

	remotePeer.consumeRequest(data.MessageData.Id)
	if len(data.GetPeers()) > 0 {
		ph.checkAndAddPeerAddresses(data.GetPeers())
	}
}
