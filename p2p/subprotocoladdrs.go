/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"net"
	"strconv"
	"sync"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"

	"github.com/libp2p/go-libp2p-peer"
)

// AddressesProtocol type
type AddressesProtocol struct {
	log *log.Logger

	ps       PeerManager
	reqMutex sync.Mutex
}

// NewAddressesProtocol create address sub protocol handler
func NewAddressesProtocol(logger *log.Logger) *AddressesProtocol {
	p := &AddressesProtocol{log: logger,
		reqMutex: sync.Mutex{},
	}
	return p
}

func (p *AddressesProtocol) setPeerManager(pm PeerManager) {
	p.ps = pm
}

func (p *AddressesProtocol) startHandling() {
	// p.ps.SetStreamHandler(addressesRequest, p.onAddressesRequest)
	// p.ps.SetStreamHandler(addressesResponse, p.onAddressesResponse)
}

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

// AddressesToStringMap PeerAddress 객체를 맵으로 변환한 것을 반환한다.
// FIXME 개별 타입마다 일일이 이런거 만드는 것은 삽질이다. golang은 jackson같은게 없나보다.
func AddressesToStringMap(addrs []*types.PeerAddress) []map[string]string {
	arr := make([]map[string]string, len(addrs))
	for i, addr := range addrs {
		vMap := make(map[string]string)
		vMap["address"] = net.IP(addr.Address).String()
		vMap["port"] = strconv.Itoa(int(addr.Port))
		vMap["peerId"] = peer.ID(addr.PeerID).Pretty()
		arr[i] = vMap
	}
	return arr
}
