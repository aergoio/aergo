/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

type RaftRoleManager struct {
	is        p2pcommon.InternalService
	actor     p2pcommon.ActorService
	logger    *log.Logger
	raftBP    map[types.PeerID]bool
	raftMutex sync.Mutex
}

func NewRaftRoleManager(is p2pcommon.InternalService, actor p2pcommon.ActorService, logger *log.Logger) *RaftRoleManager {
	return &RaftRoleManager{is: is, actor: actor, logger: logger, raftBP: make(map[types.PeerID]bool)}
}

func (rm *RaftRoleManager) Start() {
	// Do nothing in case of raft
}
func (rm *RaftRoleManager) Stop() {
	// Do nothing in case of raft
}

func (rm *RaftRoleManager) UpdateBP(toAdd []types.PeerID, toRemove []types.PeerID) {
	rm.raftMutex.Lock()
	defer rm.raftMutex.Unlock()
	changes := make([]p2pcommon.RoleModifier, 0, len(toAdd)+len(toRemove))
	for _, pid := range toRemove {
		delete(rm.raftBP, pid)
		changes = append(changes, p2pcommon.RoleModifier{ID: pid, Role: types.PeerRole_Watcher})
		rm.logger.Debug().Stringer(p2putil.LogPeerID, types.LogPeerShort(pid)).Msg("raftBP removed")
	}
	for _, pid := range toAdd {
		rm.raftBP[pid] = true
		changes = append(changes, p2pcommon.RoleModifier{ID: pid, Role: types.PeerRole_Producer})
		rm.logger.Debug().Stringer(p2putil.LogPeerID, types.LogPeerShort(pid)).Msg("raftBP added")
	}
	rm.is.PeerManager().UpdatePeerRole(changes)
}

func (rm *RaftRoleManager) SelfRole() types.PeerRole {
	return rm.is.SelfMeta().Role
}

func (rm *RaftRoleManager) GetRole(pid types.PeerID) types.PeerRole {
	rm.raftMutex.Lock()
	defer rm.raftMutex.Unlock()
	if _, found := rm.raftBP[pid]; found {
		return types.PeerRole_Producer
	} else {
		return types.PeerRole_Watcher
	}
}

func (rm *RaftRoleManager) CheckRole(remoteInfo p2pcommon.RemoteInfo, newRole types.PeerRole) bool {
	switch newRole {
	case types.PeerRole_Producer:
		return rm.GetRole(remoteInfo.Meta.ID) == types.PeerRole_Producer
	case types.PeerRole_Agent:
		// raft consensus does not allow agent
		return false
	default:
		return true
	}
}

func (rm *RaftRoleManager) FilterBPNoticeReceiver(block *types.Block, pm p2pcommon.PeerManager, targetZone p2pcommon.PeerZone) []p2pcommon.RemotePeer {
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
	return rm.FilterBPNoticeReceiver(block, pm, p2pcommon.InternalZone)
}

// time
const minimumRefreshInterval = time.Minute * 10
const getVotesMessageTimeout = time.Second * 2

type voteRank int

const (
	BP voteRank = iota
	Candidate
)

type DPOSRoleManager struct {
	is     p2pcommon.InternalService
	actor  p2pcommon.ActorService
	logger *log.Logger

	ticker *time.Ticker

	unionSet map[types.PeerID]voteRank
	bps      []types.PeerID
}

func NewDPOSRoleManager(is p2pcommon.InternalService, actor p2pcommon.ActorService, logger *log.Logger) *DPOSRoleManager {
	return &DPOSRoleManager{is: is, actor: actor, logger: logger, unionSet: make(map[types.PeerID]voteRank)}
}

func (rm *DPOSRoleManager) Start() {
	go func() {
		rm.logger.Info().Msg("Starting p2p dpos role manager")
		rm.ticker = time.NewTicker(minimumRefreshInterval)
		rm.refineProducers()
		for range rm.ticker.C {
			rm.refineProducers()
		}
	}()
}

func (rm *DPOSRoleManager) refineProducers() {
	if union, bps, err := rm.loadBPVotes(); err == nil {
		add, del := rm.collectAddDel(union)
		rm.unionSet = union
		rm.bps = bps
		if len(add)+len(del) > 0 {
			rm.logger.Info().Array("added", p2putil.NewLogPeerIdsMarshaller(add, 10)).Array("deleted", p2putil.NewLogPeerIdsMarshaller(del, 10)).Msg("found bp list changed, so telling peermanger to update peer role")
			rm.UpdateBP(add, del)
		}
	} else {
		rm.logger.Warn().Err(err).Msg("Failed to get bp vote result")
	}
}
func (rm *DPOSRoleManager) Stop() {
	rm.logger.Info().Msg("Finishing p2p dpos role manager")
	rm.ticker.Stop()
}

func (rm *DPOSRoleManager) UpdateBP(toAdd []types.PeerID, toRemove []types.PeerID) {
	changes := make([]p2pcommon.RoleModifier, 0, len(toAdd)+len(toRemove))
	for _, pid := range toRemove {
		changes = append(changes, p2pcommon.RoleModifier{ID: pid, Role: types.PeerRole_Watcher})
	}
	for _, pid := range toAdd {
		changes = append(changes, p2pcommon.RoleModifier{ID: pid, Role: types.PeerRole_Producer})
	}
	rm.is.PeerManager().UpdatePeerRole(changes)
}

func (rm *DPOSRoleManager) SelfRole() types.PeerRole {
	return rm.is.SelfMeta().Role
}

func (rm *DPOSRoleManager) GetRole(pid types.PeerID) types.PeerRole {
	if _, found := rm.unionSet[pid]; found {
		return types.PeerRole_Producer
	} else {
		return types.PeerRole_Watcher
	}
}

func (rm *DPOSRoleManager) CheckRole(remoteInfo p2pcommon.RemoteInfo, newRole types.PeerRole) bool {
	switch newRole {
	case types.PeerRole_Producer:
		return rm.GetRole(remoteInfo.Meta.ID) == types.PeerRole_Producer
	case types.PeerRole_Agent:
		for _, cert := range remoteInfo.Certificates {
			if rm.GetRole(cert.BPID) == types.PeerRole_Producer {
				return true
			}
		}
		// no certificate of accepted bp
		return false
	default:
		return true
	}
}

func (rm *DPOSRoleManager) loadBPVotes() (map[types.PeerID]voteRank, []types.PeerID, error) {
	bpCount := len(rm.is.ConsensusAccessor().ConsensusInfo().Bps)
	unionCap := bpCount * 2
	result, err := rm.actor.CallRequest(message.ChainSvc,
		&message.GetElected{Id: types.OpvoteBP.ID(), N: uint32(bpCount * 2)}, getVotesMessageTimeout)
	if err != nil {
		return nil, nil, err
	}
	rsp, ok := result.(*message.GetVoteRsp)
	if !ok {
		return nil, nil, fmt.Errorf("internal type error: actual type %s", reflect.TypeOf(result).String())
	}
	if rsp.Err != nil {
		return nil, nil, rsp.Err
	}
	bps := make([]types.PeerID, bpCount)
	union := make(map[types.PeerID]voteRank)
	for i, v := range rsp.Top.Votes {
		if i >= unionCap {
			// cut off low ranked candidates
			break
		}
		id, err := types.IDFromBytes(v.Candidate)
		if err != nil {
			return nil, nil, err
		}

		if i < bpCount {
			union[id] = BP
			bps[i] = id
		} else {
			union[id] = Candidate
		}
	}
	rm.logger.Debug().Array("bps", p2putil.NewLogPeerIdsMarshaller(bps, unionCap)).Int("unionSize", len(union)).Msg("reloaded bp list")

	return union, bps, nil
}

func (rm *DPOSRoleManager) collectAddDel(newRanks map[types.PeerID]voteRank) (add, del []types.PeerID) {
	for id := range newRanks {
		_, found := rm.unionSet[id]
		if !found {
			add = append(add, id)
		}
	}
	for id := range rm.unionSet {
		_, found := newRanks[id]
		if !found {
			del = append(del, id)
		}
	}
	return
}

func (rm *DPOSRoleManager) FilterBPNoticeReceiver(block *types.Block, pm p2pcommon.PeerManager, targetZone p2pcommon.PeerZone) []p2pcommon.RemotePeer {
	return pm.GetPeers()
}

func (rm *DPOSRoleManager) FilterNewBlockNoticeReceiver(block *types.Block, pm p2pcommon.PeerManager) []p2pcommon.RemotePeer {
	return pm.GetPeers()
}

type DPOSAgentRoleManager struct {
	DPOSRoleManager

	pdSet map[types.PeerID]bool
}

func NewDPOSAgentRoleManager(is p2pcommon.InternalService, actor p2pcommon.ActorService, logger *log.Logger, producers map[types.PeerID]bool) *DPOSAgentRoleManager {
	rm := &DPOSAgentRoleManager{DPOSRoleManager: *(NewDPOSRoleManager(is, actor, logger)), pdSet: producers}

	return rm
}

func (rm *DPOSAgentRoleManager) FilterBPNoticeReceiver(block *types.Block, pm p2pcommon.PeerManager, targetZone p2pcommon.PeerZone) []p2pcommon.RemotePeer {
	bpPeers := pm.GetProducerClassPeers()

	if targetZone == p2pcommon.ExternalZone {
		peers := make([]p2pcommon.RemotePeer, 0, len(bpPeers))
		for _, peer := range bpPeers {
			if peer.RemoteInfo().Zone == targetZone {
				peers = append(peers, peer)
			}
		}
		return peers
	} else {
		peers := make([]p2pcommon.RemotePeer, 0, len(rm.pdSet))
		for _, peer := range bpPeers {
			if _, found := rm.pdSet[peer.ID()]; found {
				peers = append(peers, peer)
			}
		}
		return peers
	}
}

func (rm *DPOSAgentRoleManager) FilterNewBlockNoticeReceiver(block *types.Block, pm p2pcommon.PeerManager) []p2pcommon.RemotePeer {
	return pm.GetPeers()
}
