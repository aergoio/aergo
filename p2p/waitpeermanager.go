/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"errors"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/metric"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/p2p/subproto"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-core/network"
	"sort"
	"time"
)

func NewWaitingPeerManager(logger *log.Logger, pm *peerManager, actorService p2pcommon.ActorService, maxCap int, useDiscover, usePolaris bool) p2pcommon.WaitingPeerManager {
	var wpm p2pcommon.WaitingPeerManager
	if !useDiscover {
		sp := &staticWPManager{basePeerManager{pm: pm, logger: logger,workingJobs:make(map[types.PeerID]ConnWork)}}
		wpm = sp
	} else {
		dp := &dynamicWPManager{basePeerManager:basePeerManager{pm: pm, logger: logger, workingJobs:make(map[types.PeerID]ConnWork)}, maxPeers: maxCap}
		wpm = dp
	}

	return wpm
}

type basePeerManager struct {
	pm          *peerManager

	logger      *log.Logger
	workingJobs map[types.PeerID]ConnWork

}


func (dpm *basePeerManager) OnInboundConn(s network.Stream) {
	version := p2pcommon.P2PVersion031

	peerID := s.Conn().RemotePeer()
	tempMeta := p2pcommon.PeerMeta{ID: peerID}
	addr := s.Conn().RemoteMultiaddr()

	dpm.logger.Debug().Str(p2putil.LogFullID, peerID.Pretty()).Str("multiaddr", addr.String()).Msg("new inbound peer arrived")
	query := inboundConnEvent{meta: tempMeta, p2pVer: p2pcommon.P2PVersionUnknown, foundC: make(chan bool)}
	dpm.pm.inboundConnChan <- query
	if exist := <-query.foundC; exist {
		dpm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("same peer as inbound peer already exists.")
		s.Close()
		return
	}

	h := dpm.pm.hsFactory.CreateHSHandler(version, false, peerID)
	// check if remote peer is connected (already handshaked)
	completeMeta, added := dpm.tryAddPeer(false, tempMeta, s, h)
	if !added {
		s.Close()
	} else {
		if tempMeta.IPAddress != completeMeta.IPAddress {
			dpm.logger.Debug().Str("after", completeMeta.IPAddress).Msg("Update IP address of inbound remote peer")
		}
	}
}

func (dpm *basePeerManager) OnInboundConnLegacy(s network.Stream) {
	version := p2pcommon.P2PVersion030
	peerID := s.Conn().RemotePeer()
	tempMeta := p2pcommon.PeerMeta{ID: peerID}
	addr := s.Conn().RemoteMultiaddr()

	dpm.logger.Debug().Str(p2putil.LogFullID, peerID.Pretty()).Str("multiaddr", addr.String()).Msg("new legacy inbound peer arrived")
	query := inboundConnEvent{meta: tempMeta, p2pVer: version, foundC: make(chan bool)}
	dpm.pm.inboundConnChan <- query
	if exist := <-query.foundC; exist {
		dpm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("same peer as inbound peer already exists.")
		s.Close()
		return
	}

	h := dpm.pm.hsFactory.CreateHSHandler(version, false, peerID)
	// check if remote peer is connected (already handshaked)
	completeMeta, added := dpm.tryAddPeer(false, tempMeta, s, h)
	if !added {
		s.Close()
	} else {
		if tempMeta.IPAddress != completeMeta.IPAddress {
			dpm.logger.Debug().Str("after", completeMeta.IPAddress).Msg("Update IP address of inbound remote peer")
		}
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

func (dpm *basePeerManager) connectWaitingPeers(maxJob int) {
	// do try to connection at most maxJobs cnt,
	peers := make([]*p2pcommon.WaitingPeer, 0, len(dpm.pm.waitingPeers))
	for _, wp := range dpm.pm.waitingPeers {
		peers = append(peers,wp)
	}
	sort.Sort(byNextTrial(peers))

	added := 0
	now := time.Now()
	for _, wp := range peers {
		if added >= maxJob {
			break
		}
		if wp.NextTrial.Before(now) {
			// check if peer is currently working now
			if _, exist := dpm.workingJobs[wp.Meta.ID]; exist {
				continue
			}
			dpm.logger.Info().Int("trial", wp.TrialCnt).Str(p2putil.LogPeerID, p2putil.ShortForm(wp.Meta.ID)).Msg("Starting scheduled try to connect peer")

			dpm.workingJobs[wp.Meta.ID] = ConnWork{Meta: wp.Meta, PeerID:wp.Meta.ID, StartTime:time.Now()}
			go dpm.runTryOutboundConnect(wp)
			added++
		} else {
			break
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
	p2pversion, s, err := dpm.getStream(meta)
	if err != nil {
		dpm.logger.Info().Err(err).Str(p2putil.LogPeerID, p2putil.ShortForm(meta.ID)).Msg("Failed to get stream.")
		workResult.Result = err
		return
	}
	h := dpm.pm.hsFactory.CreateHSHandler(p2pversion, true, meta.ID)
	// handshake
	completeMeta, added := dpm.tryAddPeer(true, meta, s, h)
	if !added {
		s.Close()
		workResult.Result = errors.New("handshake failed")
		return
	} else {
		if meta.IPAddress != completeMeta.IPAddress {
			dpm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(completeMeta.ID)).Str("before", meta.IPAddress).Str("after", completeMeta.IPAddress).Msg("IP address of remote peer is changed to ")
		}
	}
}

func (dpm *basePeerManager) getStream(meta p2pcommon.PeerMeta) (p2pcommon.P2PVersion, network.Stream, error) {
	// try connect peer with possible versions
	s, err := dpm.pm.nt.GetOrCreateStream(meta, p2pcommon.P2PSubAddr, p2pcommon.LegacyP2PSubAddr)
	if err != nil {
		return p2pcommon.P2PVersionUnknown, nil, err
	}
	switch s.Protocol() {
	case p2pcommon.P2PSubAddr:
		return p2pcommon.P2PVersion031, s, nil
	case p2pcommon.LegacyP2PSubAddr:
		return p2pcommon.P2PVersion030, s, err
	default:
		return p2pcommon.P2PVersionUnknown, nil, fmt.Errorf("unknown p2p wire protocol %v",s.Protocol())
	}
}

// tryAddPeer will do check connecting peer and add. it will return peer meta information received from
// remote peer. stream s will be owned to remotePeer if succeed to add perr.
func (dpm *basePeerManager) tryAddPeer(outbound bool, meta p2pcommon.PeerMeta, s network.Stream, h p2pcommon.HSHandler) (p2pcommon.PeerMeta, bool) {
	var peerID = meta.ID
	rd := metric.NewReader(s)
	wt := metric.NewWriter(s)
	msgRW, remoteStatus, err := h.Handle(rd, wt, defaultHandshakeTTL)
	if err != nil {
		dpm.logger.Debug().Err(err).Bool("outbound",outbound).Str(p2putil.LogPeerID, p2putil.ShortForm(meta.ID)).Msg("Failed to handshake")
		if msgRW != nil {
			dpm.sendGoAway(msgRW, err.Error())
		}
		return meta, false
	}
	// update peer meta info using sent information from remote peer
	receivedMeta := p2pcommon.NewMetaFromStatus(remoteStatus, outbound)
	if receivedMeta.ID != peerID {
		dpm.logger.Debug().Str("received_peer_id", receivedMeta.ID.Pretty()).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("Inconsistent peerID")
		dpm.sendGoAway(msgRW, "Inconsistent peerID")
		return meta, false
	}
	// override options by configurations of nodd
	_, receivedMeta.Designated = dpm.pm.designatedPeers[peerID]
	// hidden is set by either remote peer's asking or local node's config
	if _, exist := dpm.pm.hiddenPeerSet[peerID]; exist {
		receivedMeta.Hidden = true
	}

	newPeer := newRemotePeer(receivedMeta, dpm.pm.GetNextManageNum(), dpm.pm, dpm.pm.actorService, dpm.logger, dpm.pm.mf, dpm.pm.signer, s, msgRW)
	newPeer.UpdateBlkCache(remoteStatus.GetBestBlockHash(), remoteStatus.GetBestHeight())

	// insert Handlers
	dpm.pm.handlerFactory.InsertHandlers(newPeer)

	dpm.pm.peerHandshaked <- newPeer
	return receivedMeta, true
}

func (dpm *basePeerManager) OnWorkDone(result p2pcommon.ConnWorkResult) {
	meta := result.Meta
	delete(dpm.workingJobs, meta.ID)
	wp, ok := dpm.pm.waitingPeers[meta.ID]
	if !ok {
		dpm.logger.Debug().Str(p2putil.LogPeerName, p2putil.ShortMetaForm(meta)).Err(result.Result).Msg("Connection job finished")
		return
	} else {
		dpm.logger.Debug().Str(p2putil.LogPeerName, p2putil.ShortMetaForm(meta)).Int("trial",wp.TrialCnt).Err(result.Result).Msg("Connection job finished")
	}
	wp.LastResult = result.Result
	// success to connect
	if result.Result == nil {
		dpm.logger.Debug().Str(p2putil.LogPeerName, p2putil.ShortMetaForm(meta)).Msg("Deleting unimportant failed peer.")
		delete(dpm.pm.waitingPeers,meta.ID)
	} else {
		// leave waitingpeer if needed to reconnect
		if  !setNextTrial(wp) {
			dpm.logger.Debug().Str(p2putil.LogPeerName, p2putil.ShortMetaForm(meta)).Time("next_time",wp.NextTrial).Msg("Failed Connection will be retried")
			delete(dpm.pm.waitingPeers,meta.ID)
		}
	}

}

func (dpm *basePeerManager) sendGoAway(rw p2pcommon.MsgReadWriter, msg string) {
	goMsg := &types.GoAwayNotice{Message: msg}
	// TODO code smell. non safe casting. too many member depth
	mo := dpm.pm.mf.NewMsgRequestOrder(false, subproto.GoAway, goMsg).(*pbRequestOrder)
	container := mo.message

	rw.WriteMsg(container)
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
		spm.pm.waitingPeers[peer.ID()] = &p2pcommon.WaitingPeer{Meta: peer.Meta(), NextTrial: time.Now().Add(firstReconnectColltime)}
	}
}


func (spm *staticWPManager) OnDiscoveredPeers(metas []p2pcommon.PeerMeta) int {
	// static manager don't need to discovered peer.
	return 0
}

type dynamicWPManager struct {
	basePeerManager

	maxPeers    int
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
		dpm.pm.waitingPeers[peer.ID()] = &p2pcommon.WaitingPeer{Meta: peer.Meta(), NextTrial: time.Now().Add(firstReconnectColltime)}
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
	if wp.Meta.Designated {
		wp.TrialCnt++
		wp.NextTrial = time.Now().Add(getNextInterval(wp.TrialCnt))
		return true
	} else {
		return false
	}
}
