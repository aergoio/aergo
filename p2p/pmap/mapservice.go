/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package pmap

import (
	"bufio"
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"sync"
	"time"
)

const (
	MapQuery p2p.SubProtocol = 0x0100 + iota
	MapResponse
)

// PeerMapService is
type PeerMapService struct {
	*component.BaseComponent

	p2ps *p2p.P2P

	listen bool
	nt   p2p.NetworkTransport
	mutex        *sync.Mutex
	peerRegistry map[peer.ID]p2p.PeerMeta
}

func NewMapService(cfg *config.P2PConfig, p2ps *p2p.P2P) *PeerMapService {
	mapSvc := &PeerMapService{
		mutex: &sync.Mutex{} ,
		peerRegistry: make(map[peer.ID]p2p.PeerMeta),
	}
	mapSvc.BaseComponent = component.NewBaseComponent(message.MapSvc, mapSvc, log.NewLogger("map"))

	// init
	mapSvc.p2ps = p2ps

	return mapSvc
}

func (pms *PeerMapService) AfterStart() {
	if pms.listen {
		pms.nt = pms.p2ps.NetworkTransport()
		pms.nt.SetStreamHandler(p2p.AergoMapSub, pms.onConnect)
	}
}

func (pms *PeerMapService) BeforeStop() {
	if pms.listen {
		if pms.nt != nil {
			pms.nt.RemoveStreamHandler(p2p.AergoMapSub)
		}
	}
}

func (pms *PeerMapService) onConnect(s net.Stream) {
	peerID := s.Conn().RemotePeer()
	remotePeerMeta := p2p.PeerMeta{ID: peerID}
	pms.Logger.Debug().Str(p2p.LogPeerID, peerID.String()).Msg("Received map query")

	rw := p2p.NewV030ReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	defer s.Close()

	// receive input
	container, query, err := pms.readRequest(remotePeerMeta, rw)
	if err != nil {
		return
	}
	resp, err := pms.handleQuery(container, query)
	if err != nil {
		return
	}

	// response to peer
	if err = pms.writeResponse(container, remotePeerMeta, resp, rw); err != nil {
		return
	}

	// disconnect!
}

// tryAddPeer will do check connecting peer and add. it will return peer meta information received from
// remote peer setup some
func (pms *PeerMapService) readRequest(meta p2p.PeerMeta, rd p2p.MsgReader) (p2p.Message, *types.MapQuery, error) {
	data, err := rd.ReadMsg()
	if err != nil {
		return nil, nil, err
	}
	queryReq := &types.MapQuery{}
	err = p2p.UnmarshalMessage(data.Payload(), queryReq)
	if err != nil {
		return data, nil, err
	}

	return data, queryReq, nil
}

func (pms *PeerMapService) handleQuery(container p2p.Message, query *types.MapQuery) (*types.MapResponse, error) {
	receivedMeta := p2p.FromPeerAddress(query.Status.Sender)
	pms.Logger.Debug().Str(p2p.LogPeerID, receivedMeta.ID.String()).Msg("Handling query.")
	// TODO check more varification or request peer
	// must check peer is really capable to aergosvr

	// make response
	resp := &types.MapResponse{}

	if query.AddMe {
		pms.Logger.Debug().Str(p2p.LogPeerID, receivedMeta.ID.String()).Msg("AddMe is set, and register peer to peer registry")
		pms.registerPeer(receivedMeta)
	}

	return resp, nil
}

func (pms *PeerMapService) registerPeer(receivedMeta p2p.PeerMeta) error {
	peerID := receivedMeta.ID
	pms.mutex.Lock()
	defer  pms.mutex.Unlock()
	prev, ok := pms.peerRegistry[peerID]
	if !ok {
		pms.peerRegistry[peerID] = receivedMeta
	} else {
		pms.Logger.Info().Str("meta",prev.String()).Msg("Replacing previous peer info")
		pms.peerRegistry[peerID] = receivedMeta
	}
	return nil
}

func (pms *PeerMapService) writeResponse(reqContainer p2p.Message, meta p2p.PeerMeta, resp *types.MapResponse, wt p2p.MsgWriter) error {
	msgID := p2p.NewMsgID()
	respMsg, err := createV030Message(msgID, reqContainer.ID(), MapResponse, resp)
	if err != nil {
		return err
	}

	return wt.WriteMsg(respMsg)
}

func (pms *PeerMapService) Receive(context actor.Context) {
	rawMsg := context.Message()
	switch msg := rawMsg.(type) {
	default:
		pms.Logger.Debug().Interface("msg", msg) // TODO: temporal code for resolve compile error
	}

}

// TODO code duplication. it can result in a bug.
func createV030Message(msgID, orgID p2p.MsgID, subProtocol p2p.SubProtocol, innerMsg proto.Message) (*p2p.V030Message, error) {
	bytes, err := p2p.MarshalMessage(innerMsg)
	if err != nil {
		return nil, err
	}

	msg := p2p.NewV030Message(msgID, orgID, time.Now().Unix(), subProtocol, bytes)
	return msg, nil
}