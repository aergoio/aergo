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
	changes := make([]p2pcommon.AttrModifier, 0, len(toAdd)+len(toRemove))
	for _, pid := range toRemove {
		delete(rm.raftBP, pid)
		changes = append(changes, p2pcommon.AttrModifier{pid, types.PeerRole_Watcher})
		rm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(pid)).Msg("raftBP removed")
	}
	for _, pid := range toAdd {
		rm.raftBP[pid] = true
		changes = append(changes, p2pcommon.AttrModifier{pid, types.PeerRole_Producer})
		rm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(pid)).Msg("raftBP added")
	}
	rm.p2ps.pm.UpdatePeerRole(changes)
}

func (rm *RaftRoleManager) SelfRole() types.PeerRole {
	return rm.p2ps.selfMeta.Role
}

func (rm *RaftRoleManager) GetRole(pid types.PeerID) types.PeerRole {
	rm.raftMutex.Lock()
	defer rm.raftMutex.Unlock()
	if _, found := rm.raftBP[pid]; found {
		// TODO check if leader or follower
		return types.PeerRole_Producer
	} else {
		return types.PeerRole_Watcher
	}
}

func (rm *RaftRoleManager) FilterBPNoticeReceiver(block *types.Block, pm p2pcommon.PeerManager) []p2pcommon.RemotePeer {
	peers := pm.GetPeers()
	filtered := make([]p2pcommon.RemotePeer, 0, len(peers))
	for _, neighbor := range peers {
		if neighbor.AcceptedRole() != types.PeerRole_Producer {
			filtered = append(filtered, neighbor)
		}
	}
	return filtered
}
func (rm *RaftRoleManager) FilterNewBlockNoticeReceiver(block *types.Block, pm p2pcommon.PeerManager) []p2pcommon.RemotePeer {
	return rm.FilterBPNoticeReceiver(block, pm)
}

type DefaultRoleManager struct {
	p2ps *P2P

	agentsSet          map[types.PeerID]bool
	bpSet              map[types.PeerID]bool
	blockManagePeerSet map[types.PeerID]map[types.PeerID]bool
}

func (rm *DefaultRoleManager) UpdateBP(toAdd []types.PeerID, toRemove []types.PeerID) {
	changes := make([]p2pcommon.AttrModifier, 0, len(toAdd)+len(toRemove))
	for _, pid := range toRemove {
		changes = append(changes, p2pcommon.AttrModifier{pid, types.PeerRole_Watcher})
	}
	for _, pid := range toAdd {
		changes = append(changes, p2pcommon.AttrModifier{pid, types.PeerRole_Producer})
	}
	rm.p2ps.pm.UpdatePeerRole(changes)
}

func (rm *DefaultRoleManager) SelfRole() types.PeerRole {
	return rm.p2ps.selfMeta.Role
}

func (rm *DefaultRoleManager) GetRole(pid types.PeerID) types.PeerRole {
	prettyID := pid.Pretty()
	bps := rm.p2ps.consacc.ConsensusInfo().Bps
	for _, bp := range bps {
		if strings.Contains(bp, prettyID) {
			return types.PeerRole_Producer
		}
	}
	return types.PeerRole_Watcher
}

func (rm *DefaultRoleManager) FilterBPNoticeReceiver(block *types.Block, pm p2pcommon.PeerManager) []p2pcommon.RemotePeer {
	return pm.GetPeers()
}

func (rm *DefaultRoleManager) FilterNewBlockNoticeReceiver(block *types.Block, pm p2pcommon.PeerManager) []p2pcommon.RemotePeer {
	return pm.GetPeers()
}

type DposAgentRoleManager struct {
	DefaultRoleManager
	cm p2pcommon.CertificateManager
	inChargeSet map[types.PeerID]bool
}

func (rm *DposAgentRoleManager) FilterBPNoticeReceiver(block *types.Block, pm p2pcommon.PeerManager) []p2pcommon.RemotePeer {
	// agent always receive this message when he is in charged
	bpID, err := block.BPID()
	if err != nil {
		rm.p2ps.Debug().Err(err).Str("blockID",block.BlockID().String()).Msg("invalid block public key")
		return nil
	}
	pmimpl := pm.(*peerManager)
	peers := make([]p2pcommon.RemotePeer, 0, len(pmimpl.bpClassPeers))
	for _, peer := range pmimpl.bpClassPeers {
		if !p2putil.ContainsID(peer.Meta().ProducerIDs, bpID) {
			peers = append(peers, peer)
		}
	}
	return peers
}
func (rm *DposAgentRoleManager) FilterNewBlockNoticeReceiver(block *types.Block, pm p2pcommon.PeerManager) []p2pcommon.RemotePeer {
	return pm.GetPeers()
}