package raftv2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/p2p/p2pkey"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	raftlib "github.com/aergoio/etcd/raft"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/libp2p/go-libp2p-peer"
)

var (
	MaxConfChangeTimeOut = time.Second * 10

	ErrClusterHasNoMember     = errors.New("cluster has no member")
	ErrNotExistRaftMember     = errors.New("not exist member of raft cluster")
	ErrNoEnableSyncPeer       = errors.New("no peer to sync chain")
	ErrNotExistRuntimeMembers = errors.New("not exist runtime members of cluster")

	ErrInvalidMembershipReqType = errors.New("invalid type of membership change request")
	ErrPendingConfChange        = errors.New("pending membership change request is in progree. try again when it is finished")
	ErrConChangeTimeOut         = errors.New("timeouted membership change request")
	ErrConfChangeChannelBusy    = errors.New("channel of conf change propose is busy")
	ErrCCMemberIsNil            = errors.New("memeber is nil")
)

type RaftInfo struct {
	Leader string
	Total  string
	Name   string
	RaftId string
	Status *json.RawMessage
}

// raft cluster membership
// copy from dpos/bp
// TODO refactoring
// Cluster represents a cluster of block producers.
type Cluster struct {
	component.ICompSyncRequester
	sync.Mutex
	cdb consensus.ChainDB

	chainID        []byte
	chainTimestamp int64
	rs             *raftServer

	appliedIndex uint64
	appliedTerm  uint64

	NodeName string
	NodeID   consensus.MemberID

	Size uint32

	effectiveMembers *Members

	configMembers *Members
	members       *Members

	changeSeq   uint64
	confChangeC chan *consensus.ConfChangePropose

	savedChange *consensus.ConfChangePropose
}

type Members struct {
	MapByID   map[consensus.MemberID]*consensus.Member // restore from DB or snapshot
	MapByName map[string]*consensus.Member

	Index map[peer.ID]consensus.MemberID // peer ID to raft ID mapping

	BPUrls []string //for raft server TODO remove
}

func newMembers() *Members {
	return &Members{
		MapByID:   make(map[consensus.MemberID]*consensus.Member),
		MapByName: make(map[string]*consensus.Member),
		Index:     make(map[peer.ID]consensus.MemberID),
		BPUrls:    make([]string, 0),
	}
}

func (mbrs *Members) reset() {
	*mbrs = *newMembers()
}

func (mbrs *Members) ToArray() []*consensus.Member {
	count := len(mbrs.MapByID)

	var arrs = make([]*consensus.Member, count)

	i := 0
	for _, m := range mbrs.MapByID {
		arrs[i] = m
		i++
	}

	return arrs
}

func (mbrs *Members) toString() string {
	var buf string

	if mbrs == nil {
		return "[]"
	}

	buf += fmt.Sprintf("[")
	for _, bp := range mbrs.MapByID {
		buf += fmt.Sprintf("{%s}", bp.ToString())
	}
	buf += fmt.Sprintf("]")

	return buf
}

func NewCluster(chainID []byte, bf *BlockFactory, raftName string, chainTimestamp int64) *Cluster {
	cl := &Cluster{
		chainID:            chainID,
		chainTimestamp:     chainTimestamp,
		ICompSyncRequester: bf,
		NodeName:           raftName,
		configMembers:      newMembers(),
		members:            newMembers(),
		cdb:                bf.ChainWAL,
		confChangeC:        make(chan *consensus.ConfChangePropose),
	}

	cl.Lock()
	cl.changeSeq = 0
	cl.setEffectiveMembers(cl.configMembers)
	cl.Unlock()

	return cl
}

func (cl *Cluster) Recover(snapshot *raftpb.Snapshot) error {
	cl.Lock()
	defer cl.Unlock()

	var snapdata = &consensus.SnapshotData{}

	if err := snapdata.Decode(snapshot.Data); err != nil {
		return err
	}

	logger.Info().Str("snap", snapdata.ToString()).Msg("cluster recover from snapshot")
	cl.members.reset()

	cl.setEffectiveMembers(cl.members)

	// members restore
	for _, mbr := range snapdata.Members {
		cl.members.add(mbr)
	}

	return nil
}

func (cl *Cluster) isMatch(confstate *raftpb.ConfState) bool {
	var matched int
	for _, confID := range confstate.Nodes {
		if _, ok := cl.members.MapByID[consensus.MemberID(confID)]; !ok {
			return false
		}

		matched++
	}

	if matched != len(confstate.Nodes) {
		return false
	}

	return true
}

// getEffectiveMembers returns configMembers if members doesn't loaded from DB or snapshot
func (cl *Cluster) getEffectiveMembers() *Members {
	return cl.effectiveMembers
}

func (cl *Cluster) setEffectiveMembers(mbrs *Members) {
	cl.effectiveMembers = mbrs
	cl.Size = uint32(len(mbrs.MapByID))
}

func (cl *Cluster) Quorum() uint32 {
	return cl.Size/2 + 1
}

func (cl *Cluster) getStartPeers() ([]raftlib.Peer, error) {
	cl.Lock()
	defer cl.Unlock()

	if cl.Size == 0 {
		return nil, ErrClusterHasNoMember
	}

	rpeers := make([]raftlib.Peer, cl.Size)

	var i int
	for _, member := range cl.configMembers.MapByID {
		data, err := json.Marshal(member)
		if err != nil {
			return nil, err
		}
		rpeers[i] = raftlib.Peer{ID: uint64(member.ID), Context: data}
		i++
	}

	return rpeers, nil
}

// getAnyPeerAddressToSync returns peer address that has block of no for sync
func (cl *Cluster) getAnyPeerAddressToSync() (peer.ID, error) {
	cl.Lock()
	defer cl.Unlock()

	for _, member := range cl.getEffectiveMembers().MapByID {
		if member.Name != cl.NodeName {
			return member.PeerID, nil
		}
	}

	return "", ErrNoEnableSyncPeer
}

func (cl *Cluster) addMember(member *consensus.Member, fromConfig bool) error {
	cl.Lock()
	defer cl.Unlock()

	mbrs := cl.members

	logger.Debug().Bool("fromconfig", fromConfig).Str("member", member.ToString()).Msg("add member to members")

	if fromConfig {
		mbrs = cl.configMembers

		for _, prevMember := range mbrs.MapByID {
			if prevMember.HasDuplicatedAttr(member) {
				logger.Error().Str("prev", prevMember.ToString()).Str("cur", member.ToString()).Msg("duplicated configuration for raft BP member")
				return ErrDupBP
			}
		}

		// check if peerID of this node is valid
		if cl.NodeName == member.Name && member.PeerID != p2pkey.NodeID() {
			return ErrInvalidRaftPeerID
		}
	}

	mbrs.add(member)

	cl.setEffectiveMembers(mbrs)

	return nil
}

func (cl *Cluster) removeMember(member *consensus.Member) error {
	cl.Lock()
	defer cl.Unlock()

	mbrs := cl.members

	mbrs.remove(member)

	cl.setEffectiveMembers(mbrs)

	return nil
}

// CompatibleExistingCluster tests if members of existing cluster are matched with this cluster
func (cl *Cluster) CompatibleExistingCluster(existingCl *Cluster) bool {
	myMembers := cl.configMembers.ToArray()
	exMembers := existingCl.members.ToArray()

	if len(myMembers) != len(exMembers) {
		return false
	}

	// sort by name
	sort.Sort(consensus.MembersByName(myMembers))
	sort.Sort(consensus.MembersByName(exMembers))

	for i, myMember := range myMembers {
		exMember := exMembers[i]
		if !myMember.IsCompatible(exMember) {
			logger.Error().Str("my", myMember.ToString()).Str("existing", exMember.ToString()).Msg("not compatible with existing member configuration")
			return false
		}

		myMember.ID = exMember.ID

		cl.addMember(myMember, false)
	}

	return true
}

func (mbrs *Members) add(member *consensus.Member) {
	logger.Debug().Str("member", MemberIDToString(member.ID)).Msg("added raft member")

	mbrs.MapByID[member.ID] = member
	mbrs.MapByName[member.Name] = member
	mbrs.Index[member.PeerID] = member.ID
	mbrs.BPUrls = append(mbrs.BPUrls, member.Url)
}

func (mbrs *Members) remove(member *consensus.Member) {
	logger.Debug().Str("member", MemberIDToString(member.ID)).Msg("removed raft member")

	delete(mbrs.MapByID, member.ID)
	delete(mbrs.MapByName, member.Name)
	delete(mbrs.Index, member.PeerID)
}

func (mbrs *Members) getMemberByName(name string) *consensus.Member {
	member, ok := mbrs.MapByName[name]
	if !ok {
		return nil
	}

	return member
}

func (mbrs *Members) getMember(id consensus.MemberID) *consensus.Member {
	member, ok := mbrs.MapByID[id]
	if !ok {
		return nil
	}

	return member
}

func (mbrs *Members) getMemberPeerAddress(id consensus.MemberID) (peer.ID, error) {
	member := mbrs.getMember(id)
	if member == nil {
		return "", ErrNotExistRaftMember
	}

	return member.PeerID, nil
}

// hasDuplicatedMember returns true if any attributes of the given member is equal to the attributes of cluster members
func (mbrs *Members) hasDuplicatedMember(m *consensus.Member) error {
	for _, prevMember := range mbrs.MapByID {
		if prevMember.HasDuplicatedAttr(m) {
			logger.Error().Str("old", prevMember.ToString()).Str("new", m.ToString()).Msg("duplicated attribute for new member")
			return ErrDupBP
		}
	}
	return nil
}

func MaxUint64(x, y uint64) uint64 {
	if x < y {
		return y
	}
	return x
}

/*
// hasSynced get result of GetPeers request from P2P service and check if chain of this node is synchronized with majority of members
func (cc *Cluster) hasSynced() (bool, error) {
	var peers map[peer.ID]*message.PeerInfo
	var err error
	var peerBestNo uint64 = 0

	if cc.Size == 1 {
		return true, nil
	}

	// request GetPeers to p2p
	getBPPeers := func() (map[peer.ID]*message.PeerInfo, error) {
		peers := make(map[peer.ID]*message.PeerInfo)

		result, err := cc.RequestFuture(message.P2PSvc, &message.GetPeers{}, time.Second, "raft cluster sync test").Result()
		if err != nil {
			return nil, err
		}

		msg := result.(*message.GetPeersRsp)

		for _, peerElem := range msg.Peers {
			peerID := peer.ID(peerElem.Addr.PeerID)
			state := peerElem.State

			if peerElem.Self {
				continue
			}

			if state.Get() != types.RUNNING {
				logger.Debug().Str("peer", p2putil.ShortForm(peerID)).Msg("peer is not running")
				continue

			}

			// check if peer is not bp
			if _, ok := cc.Index[peerID]; !ok {
				continue
			}

			peers[peerID] = peerElem

			peerBestNo = MaxUint64(peerElem.LastBlockNumber, peerBestNo)
		}

		return peers, nil
	}

	if peers, err = getBPPeers(); err != nil {
		return false, err
	}

	if uint16(len(peers)) < (cc.Quorum() - 1) {
		logger.Debug().Msg("a majority of peers are not connected")
		return false, nil
	}

	var best *types.Block
	if best, err = cc.cdb.GetBestBlock(); err != nil {
		return false, err
	}

	if best.BlockNo()+DefaultMarginChainDiff < peerBestNo {
		logger.Debug().Uint64("best", best.BlockNo()).Uint64("peerbest", peerBestNo).Msg("chain was not synced with majority of peers")
		return false, nil
	}

	logger.Debug().Uint64("best", best.BlockNo()).Uint64("peerbest", peerBestNo).Int("margin", DefaultMarginChainDiff).Msg("chain has been synced with majority of peers")

	return true, nil
}
*/

func (cl *Cluster) toString() string {
	cl.Lock()
	defer cl.Unlock()

	var buf string

	myNode := cl.configMembers.getMemberByName(cl.NodeName)

	buf = fmt.Sprintf("raft cluster configure: total=%d, NodeName=%s, RaftID=%d", cl.Size, cl.NodeName, myNode.ID)
	buf += ", config members: " + cl.configMembers.toString()
	buf += ", runtime members: " + cl.members.toString()

	return buf
}

func (cl *Cluster) getRaftInfo(withStatus bool) *RaftInfo {
	cl.Lock()
	defer cl.Unlock()

	var leader consensus.MemberID
	if cl.rs != nil {
		leader = cl.rs.GetLeader()
	}

	var leaderName string
	var m *consensus.Member

	if m = cl.getEffectiveMembers().getMember(leader); m != nil {
		leaderName = m.Name
	} else {
		leaderName = "id=" + MemberIDToString(leader)
	}

	rinfo := &RaftInfo{Leader: leaderName, Total: strconv.FormatUint(uint64(cl.Size), 10), Name: cl.NodeName, RaftId: MemberIDToString(cl.NodeID)}

	if withStatus && cl.rs != nil {
		b, err := cl.rs.Status().MarshalJSON()
		if err != nil {
			logger.Error().Err(err).Msg("failed to marshalEntryData raft consensus")
		} else {
			m := json.RawMessage(b)
			rinfo.Status = &m
		}
	}
	return rinfo
}

func (cl *Cluster) toConsensusInfo() *types.ConsensusInfo {
	emptyCons := types.ConsensusInfo{
		Type: GetName(),
	}

	type PeerInfo struct {
		Name   string
		RaftID string
		PeerID string
	}

	b, err := json.Marshal(cl.getRaftInfo(true))
	if err != nil {
		logger.Error().Err(err).Msg("failed to marshalEntryData raft consensus")
		return &emptyCons
	}

	cl.Lock()
	defer cl.Unlock()

	cons := emptyCons
	cons.Info = string(b)

	var i int = 0
	bps := make([]string, cl.Size)

	for id, m := range cl.getEffectiveMembers().MapByID {
		bp := &PeerInfo{Name: m.Name, RaftID: MemberIDToString(m.ID), PeerID: m.PeerID.Pretty()}
		b, err = json.Marshal(bp)
		if err != nil {
			logger.Error().Err(err).Str("raftid", MemberIDToString(id)).Msg("failed to marshalEntryData raft consensus bp")
			return &emptyCons
		}
		bps[i] = string(b)

		i++
	}
	cons.Bps = bps

	return &cons
}

func (cl *Cluster) NewMemberFromAddReq(req *types.MembershipChange) (*consensus.Member, error) {
	peerID, err := peer.IDB58Decode(req.Attr.PeerID)
	if err != nil {
		return nil, err
	}
	return consensus.NewMember(req.Attr.Name, req.Attr.Url, peerID, cl.chainID, time.Now().UnixNano()), nil
}

func (cl *Cluster) ChangeMembership(req *types.MembershipChange) (*consensus.Member, error) {
	var (
		propose *consensus.ConfChangePropose
		err     error
	)

	if propose, err = cl.requestConfChange(req); err != nil {
		return nil, err
	}

	return cl.recvConfChangeReply(propose.ReplyC)
}

func (cl *Cluster) requestConfChange(req *types.MembershipChange) (*consensus.ConfChangePropose, error) {
	cl.Lock()
	defer cl.Unlock()

	if cl.savedChange != nil {
		return nil, ErrPendingConfChange
	}

	logger.Info().Str("request", util.JSON(req)).Msg("start change memgership of cluster")

	if req.Type == types.MembershipChangeType_REMOVE_MEMBER {
		panic("not implemented yet")
	}

	// make member
	member, err := cl.NewMemberFromAddReq(req)
	if err != nil {
		logger.Error().Err(err).Msg("failed to make new member")
		return nil, err
	}

	// TODO remove member

	// make raft confChange
	cc, err := cl.makeConfChange(req.Type, member)
	if err != nil {
		logger.Error().Err(err).Msg("failed to make confChange of raft")
		return nil, err
	}

	// validate member change
	if err = cl.validateChangeMembership(cc, member, false); err != nil {
		logger.Error().Err(err).Msg("failed to validate request of membership change")
		return nil, err
	}

	replyC := make(chan *consensus.ConfChangeReply)

	// TODO check cancel
	ctx, cancel := context.WithTimeout(context.Background(), MaxConfChangeTimeOut)
	defer cancel()

	// send proposeC (confChange, replyC)
	proposal := consensus.ConfChangePropose{Ctx: ctx, Cc: cc, ReplyC: replyC}

	cl.saveConfChangePropose(&proposal)

	select {
	case cl.confChangeC <- &proposal:
		logger.Info().Msg("proposal of confChange is sent to raft")
	default:
		logger.Error().Msg("proposal of confChange is dropped. confChange channel is busy")

		close(replyC)
		cl.resetSavedConfChangePropose()
		return nil, ErrConfChangeChannelBusy
	}

	return cl.savedChange, nil
}

func (cl *Cluster) recvConfChangeReply(replyC chan *consensus.ConfChangeReply) (*consensus.Member, error) {
	select {
	case reply, ok := <-replyC:
		if !ok {
			logger.Panic().Msg("reply channel of change request must not be closed")
		}

		if reply.Err != nil {
			logger.Error().Err(reply.Err).Msg("failed confChange")
			return nil, reply.Err
		}

		logger.Info().Str("cluster", cl.toString()).Msg("reply of confChange is succeed")

		return reply.Member, nil
	case <-time.After(MaxConfChangeTimeOut):
		// saved conf change must be reset in raft server after request completes
		logger.Warn().Msg("proposal of confChange is time-out")

		return nil, ErrConChangeTimeOut
	}
}

func (cl *Cluster) sendConfChangeReply(cc *raftpb.ConfChange, member *consensus.Member, err error) {
	cl.Lock()
	defer cl.Unlock()

	if cl.savedChange == nil || cl.savedChange.Cc.ID != cc.ID {
		return
	}

	propose := cl.savedChange
	cl.resetSavedConfChangePropose()

	logger.Debug().Str("req", util.JSON(propose.Cc)).Msg("send reply of conf change")

	propose.ReplyC <- &consensus.ConfChangeReply{Member: member, Err: err}
	close(propose.ReplyC)
}

func (cl *Cluster) saveConfChangePropose(ccPropose *consensus.ConfChangePropose) {
	logger.Debug().Uint64("ccid", ccPropose.Cc.ID).Msg("this confChange propose is saved in cluster")
	cl.savedChange = ccPropose
}

func (cl *Cluster) resetSavedConfChangePropose() {
	logger.Debug().Msg("reset saved confChange propose")

	cl.savedChange = nil
}

func (cl *Cluster) validateChangeMembership(cc *raftpb.ConfChange, member *consensus.Member, needlock bool) error {
	if member == nil {
		return ErrCCMemberIsNil
	}

	if needlock {
		cl.Lock()
		defer cl.Unlock()
	}

	if !member.IsValid() {
		logger.Error().Str("member", member.ToString()).Msg("member has invalid fields")
		return ErrInvalidMember
	}

	if cc.NodeID != uint64(member.ID) {
		return consensus.ErrInvalidMemberID
	}

	switch cc.Type {
	case raftpb.ConfChangeAddNode:
		if m := cl.members.getMember(member.ID); m != nil {
			return ErrCCAlreadyAdded
		}

		if err := cl.members.hasDuplicatedMember(member); err != nil {
			return err
		}

	case raftpb.ConfChangeRemoveNode:
		var m *consensus.Member
		if m = cl.members.getMember(member.ID); m == nil {
			return ErrCCNoMemberToRemove
		}
		*member = *m
	default:
		return ErrInvCCType
	}

	// - TODO UPDATE
	return nil
}

func (cl *Cluster) makeConfChange(reqType types.MembershipChangeType, member *consensus.Member) (*raftpb.ConfChange, error) {
	var changeType raftpb.ConfChangeType
	switch reqType {
	case types.MembershipChangeType_ADD_MEMBER:
		changeType = raftpb.ConfChangeAddNode
	case types.MembershipChangeType_REMOVE_MEMBER:
		changeType = raftpb.ConfChangeRemoveNode
	default:
		return nil, ErrInvalidMembershipReqType
	}

	cl.changeSeq++

	data, err := json.Marshal(member)
	if err != nil {
		return nil, err
	}

	cc := &raftpb.ConfChange{ID: cl.changeSeq, Type: changeType, NodeID: uint64(member.ID), Context: data}

	return cc, nil
}

func MemberIDToString(id consensus.MemberID) string {
	return fmt.Sprintf("%x", id)
}
