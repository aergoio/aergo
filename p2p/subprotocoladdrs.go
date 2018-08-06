/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"encoding/json"
	"net"
	"strconv"
	"sync"

	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"

	"github.com/libp2p/go-libp2p-peer"

	inet "github.com/libp2p/go-libp2p-net"

	"github.com/multiformats/go-multicodec/protobuf"
)

// pattern: /protocol-name/request-or-response-message/version
const addressesRequest = "/peer/addressesreq/0.1"
const addressesResponse = "/peer/addressesresp/0.1"

// AddressesProtocol type
type AddressesProtocol struct {
	log log.ILogger

	ps       PeerManager
	reqMutex sync.Mutex
}

// NewAddressesProtocol create address sub protocol handler
func NewAddressesProtocol(logger log.ILogger) *AddressesProtocol {
	p := &AddressesProtocol{log: logger,
		reqMutex: sync.Mutex{},
	}
	return p
}

func (p *AddressesProtocol) initWith(p2pservice PeerManager) {
	p.ps = p2pservice
	p.ps.SetStreamHandler(addressesRequest, p.onAddressesRequest)
	p.ps.SetStreamHandler(addressesResponse, p.onAddressesResponse)
}

// GetAddresses send getAddress request to other peer
func (p *AddressesProtocol) GetAddresses(peerID peer.ID, size uint32) bool {
	remotePeer, ok := p.ps.GetPeer(peerID)
	if !ok {
		p.log.Warnf("Message %s to Unknown peer %s, check if a bug.", "addressRequest", peerID.Pretty())
		return false
	}
	senderAddr := p.ps.SelfMeta().ToPeerAddress()
	// create message data
	req := &types.AddressesRequest{MessageData: &types.MessageData{},
		Sender: &senderAddr, MaxSize: 50}
	remotePeer.sendMessage(newPbMsgRequestOrder(true, false, addressesRequest, req))
	return true
}

// remote peer requests handler
func (p *AddressesProtocol) onAddressesRequest(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	remotePeer, ok := p.ps.GetPeer(peerID)
	if !ok {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()

	// get request data
	data := &types.AddressesRequest{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Info(err)
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, nil)

	// generate response message
	resp := &types.AddressesResponse{MessageData: &types.MessageData{}}
	var addrList = make([]*types.PeerAddress, 0, len(p.ps.GetPeers()))
	for _, aPeer := range p.ps.GetPeers() {
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

func (p *AddressesProtocol) checkAndAddPeerAddresses(peers []*types.PeerAddress) {
	p.log.Debugf("Checking %d peers whether to added or not in peerstore; %s", len(peers), AddressesToStringMap(peers))
	selfPeerID := p.ps.ID()
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
		p.ps.NotifyPeerAddressReceived(peerMetas)
	}
}

// remote ping response handler
func (p *AddressesProtocol) onAddressesResponse(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	remotePeer, ok := p.ps.GetPeer(peerID)
	if !ok {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()

	data := &types.AddressesResponse{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID,
		log.DoLazyEval(func() string {
			str, _ := json.Marshal(AddressesToStringMap(data.GetPeers()))
			return string(str)
		}))
	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Info("Failed to authenticate message")
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
