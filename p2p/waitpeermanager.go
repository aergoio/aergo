/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/libp2p/go-libp2p/core/network"
)

// connPeerResult is result of connection work of the waitingPeerManager.
type connPeerResult struct {
	msgRW p2pcommon.MsgReadWriter

	remote       p2pcommon.RemoteInfo
	bestHash     types.BlockID
	bestNo       types.BlockNo
	Certificates []*p2pcommon.AgentCertificateV1
}

func NewWaitingPeerManager(logger *log.Logger, is p2pcommon.InternalService, pm *peerManager, lm p2pcommon.ListManager, maxCap int, useDiscover bool) p2pcommon.WaitingPeerManager {
	var wpm p2pcommon.WaitingPeerManager
	if !useDiscover {
		sp := &staticWPManager{basePeerManager: basePeerManager{is: is, pm: pm, lm: lm, logger: logger, workingJobs: make(map[types.PeerID]ConnWork)}}
		wpm = sp
	} else {
		dp := &dynamicWPManager{basePeerManager: basePeerManager{is: is, pm: pm, lm: lm, logger: logger, workingJobs: make(map[types.PeerID]ConnWork)}, maxPeers: maxCap}
		wpm = dp
	}

	return wpm
}

type basePeerManager struct {
	is p2pcommon.InternalService
	pm *peerManager
	lm p2pcommon.ListManager

	logger      *log.Logger
	workingJobs map[types.PeerID]ConnWork
}

func (dpm *basePeerManager) OnInboundConn(s network.Stream) {
	peerID := s.Conn().RemotePeer()
	addr := s.Conn().RemoteMultiaddr()
	ip, port, err := types.GetIPPortFromMultiaddr(addr)
	if err != nil {
		dpm.logger.Warn().Err(err).Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Msg("Can't get ip address and port from inbound peer")
		s.Close()
	}
	conn := p2pcommon.RemoteConn{Outbound: false, IP: ip, Port: port}
	tempMeta := p2pcommon.PeerMeta{ID: peerID, Addresses: []types.Multiaddr{addr}}

	dpm.logger.Info().Str(p2putil.LogFullID, peerID.String()).Str("multiaddr", addr.String()).Msg("new inbound peer arrived")
	if banned, _ := dpm.lm.IsBanned(ip.String(), peerID); banned {
		dpm.logger.Info().Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Str("multiaddr", addr.String()).Msg("inbound peer is banned by list manager")
		s.Close()
		return
	}

	query := inboundConnEvent{conn: conn, meta: tempMeta, p2pVer: p2pcommon.P2PVersionUnknown, foundC: make(chan bool)}
	dpm.pm.inboundConnChan <- query
	if exist := <-query.foundC; exist {
		dpm.logger.Debug().Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Msg("same peer as inbound peer already exists.")
		s.Close()
		return
	}

	h := dpm.pm.hsFactory.CreateHSHandler(false, peerID)
	// check if remote peer is connected (already handshaked)
	_, added := dpm.tryAddPeer(false, tempMeta, s, h)
	if !added {
		s.Close()
	}
}

func (dpm *basePeerManager) CheckAndConnect() {
	dpm.logger.Debug().Msg("checking space to connect more peers")
	maxJobs := dpm.getRemainingSpaces()
	if maxJobs == 0 {
		return
	}
	dpm.connectWaitingPeers(maxJobs)
}

func (dpm *basePeerManager) InstantConnect(meta p2pcommon.PeerMeta) {
	if _, ok := dpm.pm.remotePeers[meta.ID]; ok {
		// skip	if peer is already connected
		return
	} else if wp, ok := dpm.pm.waitingPeers[meta.ID]; ok {
		// reset next trial to try connect
		wp.NextTrial = time.Now().Add(-time.Hour)
		wp.TrialCnt = 0
	} else {
		// add to waiting peer
		_, designated := dpm.pm.designatedPeers[meta.ID]
		dpm.pm.waitingPeers[meta.ID] = &p2pcommon.WaitingPeer{Meta: meta, Designated: designated, NextTrial: time.Now().Add(-time.Hour)}
	}
	dpm.connectWaitingPeers(1)
}

func (dpm *basePeerManager) connectWaitingPeers(maxJob int) {
	// do try to connection at most maxJobs cnt,
	peers := make([]*p2pcommon.WaitingPeer, 0, len(dpm.pm.waitingPeers))
	for _, wp := range dpm.pm.waitingPeers {
		peers = append(peers, wp)
	}
	sort.Sort(byNextTrial(peers))

	added := 0
	now := time.Now()
	for _, wp := range peers {
		if added >= maxJob {
			break
		}
		if !wp.NextTrial.After(now) {
			// check if peer is currently working now
			if _, exist := dpm.workingJobs[wp.Meta.ID]; exist {
				continue
			}
			// 2019.09.02 connecting to outbound peer is not affected by whitelist. inbound peer will block
			//if banned, _ := dpm.lm.IsBanned(wp.Meta.IPAddress, wp.Meta.ID); banned {
			//	dpm.logger.Info().Str(p2putil.LogPeerName, p2putil.ShortMetaForm(wp.Meta)).Msg("Skipping banned peer")
			//	continue
			//}
			dpm.logger.Info().Int("trial", wp.TrialCnt).Stringer(p2putil.LogPeerID, types.LogPeerShort(wp.Meta.ID)).Msg("Starting scheduled try to connect peer")

			dpm.workingJobs[wp.Meta.ID] = ConnWork{Meta: wp.Meta, PeerID: wp.Meta.ID, StartTime: time.Now()}
			go dpm.runTryOutboundConnect(wp)
			added++
		} else {
			continue
		}
	}
}

// getRemainingSpaces check and return the number that can do connection work.
// the number depends on the number of current works and the number of waiting peers
func (dpm *basePeerManager) getRemainingSpaces() int {
	// simpler version. just check total count
	// has space to add more connection
	if len(dpm.pm.waitingPeers) <= 0 {
		return 0
	}
	affordWorker := p2pcommon.MaxConcurrentHandshake - len(dpm.workingJobs)
	if affordWorker <= 0 {
		return 0
	}
	return affordWorker
}

func (dpm *basePeerManager) runTryOutboundConnect(wp *p2pcommon.WaitingPeer) {
	workResult := p2pcommon.ConnWorkResult{Meta: wp.Meta, TargetPeer: wp}
	defer func() {
		dpm.pm.workDoneChannel <- workResult
	}()

	meta := wp.Meta
	s, err := dpm.getStream(meta)
	if err != nil {
		dpm.logger.Info().Err(err).Stringer(p2putil.LogPeerID, types.LogPeerShort(meta.ID)).Msg("Failed to get stream.")
		workResult.Result = err
		return
	}
	h := dpm.pm.hsFactory.CreateHSHandler(true, meta.ID)
	// handshake
	_, added := dpm.tryAddPeer(true, meta, s, h)
	if !added {
		s.Close()
		workResult.Result = errors.New("handshake failed")
		return
		//} else {
		//	if meta.IPAddress != completeMeta.IPAddress {
		//		dpm.logger.Debug().Stringer(p2putil.LogPeerID, types.LogPeerShort(completeMeta.ID)).Str("before", meta.IPAddress).Str("after", completeMeta.IPAddress).Msg("IP address of remote peer is changed to ")
		//	}
	}
}

// getStream returns is wire handshake is legacy or newer
func (dpm *basePeerManager) getStream(meta p2pcommon.PeerMeta) (network.Stream, error) {
	// try connect peer with possible versions
	s, err := dpm.pm.nt.GetOrCreateStream(meta, p2pcommon.P2PSubAddr)
	if err != nil {
		return nil, err
	}
	switch s.Protocol() {
	case p2pcommon.P2PSubAddr:
		return s, nil
	default:
		return nil, fmt.Errorf("unknown p2p wire protocol %v", s.Protocol())
	}
}

// tryAddPeer will do check connecting peer and add. it will return peer meta information received from
// remote peer. stream s will be owned to remotePeer if succeed to add peer.
func (dpm *basePeerManager) tryAddPeer(outbound bool, meta p2pcommon.PeerMeta, s network.Stream, h p2pcommon.HSHandler) (p2pcommon.PeerMeta, bool) {
	hResult, err := h.Handle(s, defaultHandshakeTTL)
	if err != nil {
		dpm.logger.Debug().Err(err).Bool("outbound", outbound).Stringer(p2putil.LogPeerID, types.LogPeerShort(meta.ID)).Msg("Failed to handshake")
		return meta, false
	}

	// update peer meta info using sent information from remote peer
	remoteInfo := dpm.createRemoteInfo(s.Conn(), *hResult, outbound)

	dpm.pm.peerConnected <- connPeerResult{remote: remoteInfo, msgRW: hResult.MsgRW, bestHash: hResult.BestBlockHash, bestNo: hResult.BestBlockNo, Certificates: hResult.Certificates}
	return remoteInfo.Meta, true
}

// createRemoteInfo create incomplete struct, field acceptedRole is not set yet
func (dpm *basePeerManager) createRemoteInfo(conn network.Conn, r p2pcommon.HandshakeResult, outbound bool) p2pcommon.RemoteInfo {
	rma := conn.RemoteMultiaddr()
	ip, port, err := types.GetIPPortFromMultiaddr(rma)
	if err != nil {
		panic("conn information is wrong : " + err.Error())
	}

	connection := p2pcommon.RemoteConn{IP: ip, Port: port, Outbound: outbound}
	zone := p2pcommon.PeerZone(p2putil.IsContainedIP(ip, dpm.is.LocalSettings().InternalZones))
	ri := p2pcommon.RemoteInfo{Meta: r.Meta, Connection: connection, Hidden: r.Hidden, Certificates: r.Certificates, AcceptedRole: types.PeerRole_Watcher, Zone: zone}

	// TODO Is it OK to this function has logic for policy?
	// check role
	switch r.Meta.Role {
	case types.PeerRole_Producer:
		// TODO check consensus and peer id is in to top vote list or bp list
		ri.AcceptedRole = types.PeerRole_Producer
	case types.PeerRole_Agent:
		// check if agent has at least one certificate
		if len(r.Certificates) > 0 {
			ri.AcceptedRole = types.PeerRole_Agent
		} else {
			dpm.logger.Debug().Stringer(p2putil.LogPeerID, types.LogPeerShort(r.Meta.ID)).Msg("treat peer which claims agent but with no certificates, as Watcher")
		}
	default:
		ri.AcceptedRole = r.Meta.Role
	}

	return ri
}

func (dpm *basePeerManager) OnWorkDone(result p2pcommon.ConnWorkResult) {
	meta := result.Meta
	delete(dpm.workingJobs, meta.ID)
	wp, ok := dpm.pm.waitingPeers[meta.ID]
	if !ok {
		dpm.logger.Debug().Str(p2putil.LogPeerName, p2putil.ShortMetaForm(meta)).Err(result.Result).Msg("Connection job finished")
		return
	} else {
		dpm.logger.Debug().Str(p2putil.LogPeerName, p2putil.ShortMetaForm(meta)).Int("trial", wp.TrialCnt).Err(result.Result).Msg("Connection job finished")
	}
	wp.LastResult = result.Result
	// success to connect
	if result.Result == nil {
		dpm.logger.Debug().Str(p2putil.LogPeerName, p2putil.ShortMetaForm(meta)).Msg("Connected job succeeded, so delete it from waiting peers")
		delete(dpm.pm.waitingPeers, meta.ID)
	} else {
		// leave waiting peer if needed to reconnect
		if !setNextTrial(wp) {
			dpm.logger.Debug().Str(p2putil.LogPeerName, p2putil.ShortMetaForm(meta)).Msg("Connected job failed, but will not retry unimportant peer.")
			delete(dpm.pm.waitingPeers, meta.ID)
		} else {
			dpm.logger.Debug().Str(p2putil.LogPeerName, p2putil.ShortMetaForm(meta)).Time("next_time", wp.NextTrial).Msg("Connected job failed, and will retry important peer")
		}
	}
}

type staticWPManager struct {
	basePeerManager
}

func (spm *staticWPManager) OnPeerConnect(pid types.PeerID) {
	delete(spm.pm.waitingPeers, pid)
}

func (spm *staticWPManager) OnPeerDisconnect(peer p2pcommon.RemotePeer) {
	// if peer is designated peer , try reconnect by add peermeta to waiting peer
	if _, ok := spm.pm.designatedPeers[peer.ID()]; ok {
		spm.logger.Debug().Str(p2putil.LogPeerID, peer.Name()).Msg("server will try to reconnect designated peer after cooltime")
		// These peers must have cool time.
		spm.pm.waitingPeers[peer.ID()] = &p2pcommon.WaitingPeer{Meta: peer.Meta(), Designated: true, NextTrial: time.Now().Add(firstReconnectCoolTime)}
	}
}

func (spm *staticWPManager) OnDiscoveredPeers(metas []p2pcommon.PeerMeta) int {
	// static manager don't need to discovered peer.
	return 0
}

type dynamicWPManager struct {
	basePeerManager

	maxPeers int
}

func (dpm *dynamicWPManager) OnPeerConnect(pid types.PeerID) {
	// remove peer from wait pool
	delete(dpm.pm.waitingPeers, pid)
}

func (dpm *dynamicWPManager) OnPeerDisconnect(peer p2pcommon.RemotePeer) {
	// if peer is designated peer or trusted enough , try reconnect by add peermeta to waiting peer
	// TODO check by trust level is not implemented yet.
	if _, ok := dpm.pm.designatedPeers[peer.ID()]; ok {
		dpm.logger.Debug().Str(p2putil.LogPeerID, peer.Name()).Msg("server will try to reconnect designated peer after cooltime")
		// These peers must have cool time.
		dpm.pm.waitingPeers[peer.ID()] = &p2pcommon.WaitingPeer{Meta: peer.Meta(), Designated: true, NextTrial: time.Now().Add(firstReconnectCoolTime)}
		//dpm.pm.addAwait(peer.Meta())
	}
}

func (dpm *dynamicWPManager) OnDiscoveredPeers(metas []p2pcommon.PeerMeta) int {
	addedWP := 0
	for _, meta := range metas {
		if _, ok := dpm.pm.remotePeers[meta.ID]; ok {
			// skip connected peer
			continue
		} else if _, ok := dpm.pm.waitingPeers[meta.ID]; ok {
			// skip already waiting peer
			continue
		}

		// TODO check blacklist later.
		dpm.pm.waitingPeers[meta.ID] = &p2pcommon.WaitingPeer{Meta: meta, NextTrial: time.Now()}
		addedWP++
	}
	return addedWP
}

func (dpm *dynamicWPManager) CheckAndConnect() {
	dpm.logger.Debug().Msg("checking space to connect more peers")
	maxJobs := dpm.getRemainingSpaces()
	if maxJobs == 0 {
		return
	}
	dpm.connectWaitingPeers(maxJobs)
}

func (dpm *dynamicWPManager) getRemainingSpaces() int {
	// simpler version. just check total count
	// has space to add more connection
	affordCnt := dpm.maxPeers - len(dpm.pm.remotePeers) - len(dpm.workingJobs)
	if affordCnt <= 0 {
		return 0
	}
	affordWorker := dpm.basePeerManager.getRemainingSpaces()
	if affordCnt < affordWorker {
		return affordCnt
	} else {
		return affordWorker
	}
}

type inboundConnEvent struct {
	conn   p2pcommon.RemoteConn
	meta   p2pcommon.PeerMeta
	p2pVer p2pcommon.P2PVersion
	foundC chan bool
}

type byNextTrial []*p2pcommon.WaitingPeer

func (a byNextTrial) Len() int           { return len(a) }
func (a byNextTrial) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byNextTrial) Less(i, j int) bool { return a[i].NextTrial.Before(a[j].NextTrial) }

type ConnWork struct {
	PeerID    types.PeerID
	Meta      p2pcommon.PeerMeta
	StartTime time.Time
}

// setNextTrial check if peer is worthy to connect, and set time when the server try to connect next time.
// It will true if this node is worth to try connect again, or return false if not.
func setNextTrial(wp *p2pcommon.WaitingPeer) bool {
	if wp.Designated {
		wp.TrialCnt++
		wp.NextTrial = time.Now().Add(getNextInterval(wp.TrialCnt))
		return true
	} else {
		return false
	}
}
