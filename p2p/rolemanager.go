/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"strings"
	"sync"
)

type RaftRoleManager struct {
	p2ps      *P2P
	raftBP    map[types.PeerID]bool
	raftMutex sync.Mutex
}

func (rr *RaftRoleManager) UpdateBP(toAdd []types.PeerID, toRemove []types.PeerID) {
	rr.raftMutex.Lock()
	defer rr.raftMutex.Unlock()

	for _, pid := range toRemove {
		delete(rr.raftBP, pid)
	}
	for _, pid := range toAdd {
		rr.raftBP[pid] = true
	}
}

func (rr *RaftRoleManager) GetRole(pid types.PeerID) p2pcommon.PeerRole {
	rr.raftMutex.Lock()
	defer rr.raftMutex.Unlock()
	if _, found := rr.raftBP[pid]; found {
		// TODO check if leader or follower
		return p2pcommon.RaftLeader
	} else {
		return p2pcommon.RaftWatcher
	}
}

func (rr *RaftRoleManager) NotifyNewBlockMsg(mo p2pcommon.MsgOrder, peers []p2pcommon.RemotePeer) (skipped, sent int) {
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

func (dr *DefaultRoleManager) UpdateBP(toAdd []types.PeerID, toRemove []types.PeerID) {
	// do nothing for now
}

func (dr *DefaultRoleManager) GetRole(pid types.PeerID) p2pcommon.PeerRole {
	prettyID := pid.Pretty()
	bps := dr.p2ps.consacc.ConsensusInfo().Bps
	for _, bp := range bps {
		if strings.Contains(bp, prettyID) {
			return p2pcommon.BlockProducer
		}
	}
	return p2pcommon.Watcher
}

func (dr *DefaultRoleManager) NotifyNewBlockMsg(mo p2pcommon.MsgOrder, peers []p2pcommon.RemotePeer) (skipped, sent int) {
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