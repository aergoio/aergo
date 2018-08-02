/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"context"
	"time"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-protocol"
)

// PeerStatus shows running status of peer
type PeerStatus int32

// indicating status of remote peer
const (
	DISCONNECTED = iota
	// STARTING means connection is estabished, but not fully handshaked yet.
	STARTING
	// RUNNING means normally workings now
	RUNNING
	STOPPED
)

const defaultPingInterval = time.Second * 60

// RemotePeer represent remote peer to which is connected
type RemotePeer struct {
	log          log.ILogger
	pingDuration time.Duration

	meta      PeerMeta
	status    PeerStatus
	actorServ ActorService
	ps        PeerManager
	stopChan  chan struct{}

	write      chan msgOrder
	closeWrite chan struct{}

	// used to access request data from response handlers
	requests    map[string]msgOrder
	consumeChan chan string
}

// msgOrder is abstraction information about the message that will be sent to peer
type msgOrder interface {
	GetRequestID() string
	IsRequest() bool
	IsGossip() bool
	IsNeedSign() bool
	// ResponseExpected means that remote peer is expected to send response to this request.
	ResponseExpected() bool
	GetProtocolID() protocol.ID
	SignWith(ps PeerManager) error
	SendOver(s inet.Stream) error
}

// newRemotePeer create an object which represent a remote peer.
func newRemotePeer(meta PeerMeta, p2ps PeerManager, iServ ActorService, log log.ILogger) RemotePeer {
	return RemotePeer{
		meta: meta, ps: p2ps, actorServ: iServ, log: log,
		pingDuration: defaultPingInterval,
		status:       STARTING,
		stopChan:     make(chan struct{}),
		write:        make(chan msgOrder),
		closeWrite:   make(chan struct{}),

		requests:    make(map[string]msgOrder),
		consumeChan: make(chan string, 10),
	}
}

// runPeer should be called by go routine
func (p *RemotePeer) runPeer() {
	p.log.Debugf("Starting peer %s ", p.meta.ID.Pretty())
	pingTicker := time.NewTicker(p.pingDuration)
	go p.runWrite()
RUNLOOP:
	for {
		select {
		case <-pingTicker.C:
			p.sendPing()
		// case hsMsg := <-p.hsChan:
		// 	p.startHandshake(hsMsg)
		case <-p.stopChan:
			p.status = STOPPED
			break RUNLOOP
		}
	}
	p.log.Infof("Finishing peer %s ", p.meta.ID.Pretty())
	pingTicker.Stop()
	p.closeWrite <- struct{}{}
	close(p.stopChan)
}

func (p *RemotePeer) runWrite() {
RUNLOOP:
	for {
		select {
		case m := <-p.write:
			p.writeToPeer(m)
		case rID := <-p.consumeChan:
			delete(p.requests, rID)
		case <-p.closeWrite:
			break RUNLOOP
		}
	}
	close(p.write)
	close(p.consumeChan)
}

// Stop stops aPeer works
func (p *RemotePeer) Stop() {
	p.stopChan <- struct{}{}
}

func (p *RemotePeer) sendMessage(msg msgOrder) {
	p.write <- msg
}

// consumeRequest remove request from request history.
func (p *RemotePeer) consumeRequest(requestID string) {
	p.consumeChan <- requestID
}

// startHandshake is run only in AergoPeer.RunPeer go routine
func (p *RemotePeer) handshakePeer(statusMsg *types.Status) {
	if p.status != STARTING {
		p.goAwayMsg("Invalid status msg")
		return
	}

	// check address. and apply current
	receivedMeta := FromPeerAddress(statusMsg.Sender)
	p.meta.IPAddress = receivedMeta.IPAddress
	p.meta.Port = receivedMeta.Port

	// TODO: check protocol version, blacklist, key authentication or etc.
	err := p.checkProtocolVersion()
	if err != nil {
		p.log.Infof("invalid protocol version of peer %v", p.meta.ID.Pretty())
		p.goAwayMsg("Handshake failed")
		return
	}

	// If all checked and validated. it's now handshaked. and then run sync.
	p.log.Infof("peer %s is handshaked and now running status", p.meta.ID.Pretty())
	p.status = RUNNING

	// notice to p2pmanager that handshaking is finished
	p.ps.NotifyPeerHandshake(p.meta.ID)

	p.actorServ.SendRequest(message.ChainSvc, &message.SyncBlockState{PeerID: p.meta.ID, BlockNo: statusMsg.BestHeight, BlockHash: statusMsg.BestBlockHash})
}

func (p *RemotePeer) writeToPeer(m msgOrder) {
	p.log.Debugf("Writing message %v:%v to peer %s", m.GetProtocolID(), m.GetRequestID(), p.meta.ID.Pretty())
	// sign the data
	if m.IsNeedSign() {
		err := m.SignWith(p.ps)
		if err != nil {
			p.log.Warn(err.Error())
			return
		}
	}
	s, err := p.ps.NewStream(context.Background(), p.meta.ID, m.GetProtocolID())
	if err != nil {
		p.log.Warn(err.Error())
		return
	}
	defer s.Close()

	err = m.SendOver(s)
	if err != nil {
		p.log.Warn(err.Error())
		return
	}
	if m.ResponseExpected() {
		p.requests[m.GetRequestID()] = m
	}
}

// this method MUST be called in same go routine as AergoPeer.RunPeer()
func (p *RemotePeer) sendPing() {
	p.log.Debugf("Sending ping to: %s....", p.meta.ID.Pretty())

	// find my best block
	bestBlock, err := extractBlockFromRequest(p.actorServ.CallRequest(message.ChainSvc, &message.GetBestBlock{}))
	if err != nil {
		p.log.Errorf("Failed to get best block %v", err.Error())
		return
	}
	// create message data
	pingMsg := &types.Ping{
		MessageData:   &types.MessageData{},
		BestBlockHash: bestBlock.GetHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	p.sendMessage(newPbMsgRequestOrder(true, false, pingRequest, pingMsg))
}

// sendStatus is called once when a peer is added.()
func (p *RemotePeer) sendStatus() {
	p.log.Infof("Sending status message to %s for handshaking.", p.meta.ID.Pretty())

	// find my best block
	bestBlock, err := extractBlockFromRequest(p.actorServ.CallRequest(message.ChainSvc, &message.GetBestBlock{}))
	if err != nil {
		p.log.Errorf("Failed to get best block %v", err.Error())
		return
	}
	meta := p.ps.SelfMeta().ToPeerAddress()
	// create message data
	statusMsg := &types.Status{
		MessageData:   &types.MessageData{},
		Sender:        &meta,
		BestBlockHash: bestBlock.GetHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	p.sendMessage(newPbMsgRequestOrder(false, true, statusRequest, statusMsg))
}

// send notice message and then disconnect. this routine should only run in RunPeer go routine
func (p *RemotePeer) goAwayMsg(msg string) {
	p.log.Infof("Peer is closing since by %s ", msg)
	p.sendMessage(newPbMsgRequestOrder(false, true, goAway, &types.GoAwayNotice{MessageData: &types.MessageData{}, Message: msg}))
	p.ps.RemovePeer(p.meta.ID)
}

func (p *RemotePeer) checkProtocolVersion() error {
	// TODO modify interface and put check code here
	return nil
}
