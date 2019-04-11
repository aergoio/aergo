/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"context"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
	"time"
)

const (
	macConcurrentQueryCount = 4
	// TODO manage cooltime and reconnect interval together in same file.
	firstReconnectColltime = time.Minute
)

func NewPeerFinder(logger *log.Logger, pm *peerManager, actorService p2pcommon.ActorService, maxCap int, useDiscover, usePolaris bool) p2pcommon.PeerFinder {
	if !useDiscover {
		return &staticPeerFinder{pm:pm, logger:logger}
	} else {
		dp := &dynamicPeerFinder{logger: logger, pm: pm, actorService: actorService, maxCap: maxCap, usePolaris:usePolaris}
		dp.qStats = make(map[peer.ID]*queryStat)
		return dp
	}

}

// staticPeerFinder is for BP or backup node. it will not query to polaris or other peers
type staticPeerFinder struct {
	pm     *peerManager
	logger *log.Logger
}

func (dp *staticPeerFinder) OnPeerDisconnect(peer p2pcommon.RemotePeer) {
	if _, ok := dp.pm.designatedPeers[peer.ID()]; ok {
		// These peers must have cool time.
		dp.pm.waitingPeers[peer.ID()] = &p2pcommon.WaitingPeer{Meta: peer.Meta(), NextTrial: time.Now().Add(firstReconnectColltime)}
		dp.pm.addAwait(peer.Meta())
	}
}

func (dp *staticPeerFinder) OnPeerConnect(pid peer.ID) {
	if _, ok := dp.pm.designatedPeers[pid]; ok {
		delete(dp.pm.waitingPeers, pid)
		dp.pm.cancelAwait(pid)
	}
}

func (dp *staticPeerFinder) OnDiscoveredPeers(metas []p2pcommon.PeerMeta) {
}

func (dp *staticPeerFinder) CheckAndFill() {
	// find if there are not connected designated peers. designated peer
	for _, meta := range dp.pm.designatedPeers {
		if _, found := dp.pm.remotePeers[meta.ID]; !found {
			if _, foundInWait := dp.pm.waitingPeers[meta.ID]; !foundInWait {
				dp.pm.waitingPeers[meta.ID] = &p2pcommon.WaitingPeer{Meta: meta, NextTrial: time.Now()}
			}
		}
	}
}

func (dp *staticPeerFinder) AddWaitings(metas []p2pcommon.PeerMeta) {
}

func (dp *staticPeerFinder) PickPeer(ctx context.Context) (p2pcommon.PeerMeta, error) {
	return p2pcommon.PeerMeta{}, errors.New("no peers in pool")
}

var _ p2pcommon.PeerFinder = (*staticPeerFinder)(nil)

// dynamicPeerFinder is triggering map query to Polaris or address query to other connected peer
// to discover peer
// It is not thread-safe. Thread safety is responsible to the caller.
type dynamicPeerFinder struct {
	logger       *log.Logger
	pm           *peerManager
	actorService p2pcommon.ActorService
	usePolaris   bool

	// qStats are logs of query. all connected peers must exist queryStat.
	qStats map[peer.ID]*queryStat
	maxCap int

	polarisTurn time.Time
}

var _ p2pcommon.PeerFinder = (*dynamicPeerFinder)(nil)

func (dp *dynamicPeerFinder) OnPeerDisconnect(peer p2pcommon.RemotePeer) {
	// And check if to connect more peers
	delete(dp.qStats, peer.ID())
	if _, ok := dp.pm.designatedPeers[peer.ID()]; ok {
		dp.logger.Debug().Str(p2putil.LogPeerID, peer.Name()).Msg("server will try to reconnect designated peer after cooltime")
		// These peers must have cool time.
		dp.pm.waitingPeers[peer.ID()] = &p2pcommon.WaitingPeer{Meta: peer.Meta(), NextTrial: time.Now().Add(firstReconnectColltime)}
		dp.pm.addAwait(peer.Meta())
	}
}

func (dp *dynamicPeerFinder) OnPeerConnect(pid peer.ID) {
	dp.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(pid)).Msg("check and remove peerid in pool")
	if stat := dp.qStats[pid]; stat == nil {
		// first query will be sent quickly
		dp.qStats[pid] = &queryStat{pid: pid, nextTurn: time.Now().Add(p2pcommon.PeerFirstInterval)}
	}
	// remove peer from wait pool
	delete(dp.pm.waitingPeers, pid)
	dp.pm.cancelAwait(pid)
}

func (dp *dynamicPeerFinder) OnDiscoveredPeers(metas []p2pcommon.PeerMeta) {
	for _, meta := range metas {
		if _, ok := dp.qStats[meta.ID]; ok {
			// skip connected peer
			continue
		} else if _, ok := dp.pm.waitingPeers[meta.ID]; ok {
			continue
		}
		dp.pm.waitingPeers[meta.ID] = &p2pcommon.WaitingPeer{Meta: meta, NextTrial: time.Now()}
	}
}

func (dp *dynamicPeerFinder) CheckAndFill() {
	toConnCount := dp.maxCap - len(dp.pm.waitingPeers)
	if toConnCount <= 0 {
		// if enough peer is collected already, skip collect
		return
	}
	now := time.Now()
	// query to polaris
	if dp.usePolaris && now.After(dp.polarisTurn) {
		dp.polarisTurn = now.Add(p2pcommon.PolarisQueryInterval)
		dp.logger.Debug().Time("next_turn", dp.polarisTurn).Msg("quering to polaris")
		dp.actorService.SendRequest(message.P2PSvc, &message.MapQueryMsg{Count: MaxAddrListSizePolaris})
	}
	// query to peers
	queried := 0
	for _, stat := range dp.qStats {
		if stat.nextTurn.Before(now) {
			// slowly collect
			stat.lastCheck = now
			stat.nextTurn = now.Add(p2pcommon.PeerQueryInterval)
			dp.actorService.SendRequest(message.P2PSvc, &message.GetAddressesMsg{ToWhom: stat.pid, Size: MaxAddrListSizePeer, Offset: 0})
			queried++
			if queried >= macConcurrentQueryCount {
				break
			}
		}
	}
}

type queryStat struct {
	pid       peer.ID
	lastCheck time.Time
	nextTurn  time.Time
}
