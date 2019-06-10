/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"strings"
	"sync"
)

type RaftRoleManager struct {
	p2ps      *P2P
	logger    *log.Logger
	raftBP    map[types.PeerID]bool
	raftMutex sync.Mutex
}

func (rm *RaftRoleManager) UpdateBP(toAdd []types.PeerID, toRemove []types.PeerID) {
	rm.raftMutex.Lock()
	defer rm.raftMutex.Unlock()
	changes := make([]p2pcommon.AttrModifier,0, len(toAdd)+len(toRemove))
	for _, pid := range toRemove {
		delete(rm.raftBP, pid)
		changes = append(changes, p2pcommon.AttrModifier{pid, p2pcommon.RaftWatcher})
		rm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(pid)).Msg("raftBP removed")
	}
	for _, pid := range toAdd {
		rm.raftBP[pid] = true
		changes = append(changes, p2pcommon.AttrModifier{pid, p2pcommon.RaftLeader})
		rm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(pid)).Msg("raftBP added")
	}
	rm.p2ps.pm.UpdatePeerRole(changes)
}

func (rm *RaftRoleManager) GetRole(pid types.PeerID) p2pcommon.PeerRole {
	rm.raftMutex.Lock()
	defer rm.raftMutex.Unlock()
	if _, found := rm.raftBP[pid]; found {
		// TODO check if leader or follower
		return p2pcommon.RaftLeader
	} else {
		return p2pcommon.RaftWatcher
	}
}

func (rm *RaftRoleManager) NotifyNewBlockMsg(mo p2pcommon.MsgOrder, peers []p2pcommon.RemotePeer) (skipped, sent int) {
	// TODO filter to only contain bp and trusted node.
	for _, neighbor := range peers {
		if neighbor != nil && neighbor.State() == types.RUNNING &&
			neighbor.Role() == p2pcommon.RaftWatcher {
			sent++
			neighbor.SendMessage(mo)
		} else {
			skipped++
		}
	}
	return
}

type DefaultRoleManager struct {
	p2ps *P2P
}

func (rm *DefaultRoleManager) UpdateBP(toAdd []types.PeerID, toRemove []types.PeerID) {
	for _, pid := range toRemove {
		if peer, found := rm.p2ps.pm.GetPeer(pid); found {
			peer.ChangeRole(p2pcommon.Watcher)
		}
	}
	for _, pid := range toAdd {
		if peer, found := rm.p2ps.pm.GetPeer(pid); found {
			peer.ChangeRole(p2pcommon.BlockProducer)
		}
	}
}

func (rm *DefaultRoleManager) GetRole(pid types.PeerID) p2pcommon.PeerRole {
	prettyID := pid.Pretty()
	bps := rm.p2ps.consacc.ConsensusInfo().Bps
	for _, bp := range bps {
		if strings.Contains(bp, prettyID) {
			return p2pcommon.BlockProducer
		}
	}
	return p2pcommon.Watcher
}

func (rm *DefaultRoleManager) NotifyNewBlockMsg(mo p2pcommon.MsgOrder, peers []p2pcommon.RemotePeer) (skipped, sent int) {
	// TODO filter to only contain bp and trusted node.
	for _, neighbor := range peers {
		if neighbor != nil && neighbor.State() == types.RUNNING {
			sent++
			neighbor.SendMessage(mo)
		} else {
			skipped++
		}
	}
	return
}

