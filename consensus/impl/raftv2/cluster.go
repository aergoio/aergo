package raftv2

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/message"
	raftlib "github.com/aergoio/etcd/raft"
	"github.com/aergoio/etcd/raft/raftpb"
)

var (
	MaxConfChangeTimeOut = time.Second * 100

	ErrClusterHasNoMember   = errors.New("cluster has no member")
	ErrNotExistRaftMember   = errors.New("not exist member of raft cluster")
	ErrNoEnableSyncPeer     = errors.New("no peer to sync chain")
	ErrMemberAlreadyApplied = errors.New("member is already added")

	ErrInvalidMembershipReqType = errors.New("invalid type of membership change request")
	ErrPendingConfChange        = errors.New("pending membership change request is in progree. try again when it is finished")
	ErrConChangeTimeOut         = errors.New("timeouted membership change request")
	ErrConfChangeChannelBusy    = errors.New("channel of conf change propose is busy")
	ErrCCMemberIsNil            = errors.New("memeber is nil")
	ErrNotMatchedRaftName       = errors.New("mismatched name of raft identity")
	ErrNotMatchedRaftPeerID     = errors.New("mismatched peerid of raft identity")
	ErrNotExitRaftProgress      = errors.New("progress of this node doesn't exist")
	ErrUnhealtyNodeExist        = errors.New("can't add some node if unhealthy nodes exist")
	ErrRemoveHealthyNode        = errors.New("remove of a healthy node may cause the cluster to hang")
)

const (
	MembersNameInit    = "init"
	MembersNameApplied = "applied"
	MembersNameRemoved = "removed"
	InvalidClusterID   = 0
)

type RaftInfo struct {
	Leader string
	Total  uint32
	Name   string
	RaftId string
	Status *json.RawMessage
}

type NotifyFn func(event *message.RaftClusterEvent)

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

	identity consensus.RaftIdentity

	Size uint32

	// @ MatchClusterAndConfState
	// cluster members must match nodes of confstate. otherwise confchange may fail and be skipped by comparing with cluster members.
	// Mismatch of cluster and confstate occures when node joins a exising cluster. Joined node starts from latest members, but confstate is empty.
	// If snapshot is written before all confchange logs  be applied, mismatched state is written to disk.
	// After recovery from snapshot, problems will happen.
	members *Members // using for 1. booting
	//           2. send cluster info to remote
	appliedMembers *Members // using for 1. verifying runtime confchange.
	// 			 2. creating snapshot
	//           3. recover from snapshot

	// raft http reject message from removed member
	// TODO for p2p
	removedMembers *Members

	changeSeq   uint64
	confChangeC chan *consensus.ConfChangePropose

	savedChange *consensus.ConfChangePropose

	notifyFn NotifyFn
}

type Members struct {
	name      string
	MapByID   map[uint64]*consensus.Member // restore from DB or snapshot
	MapByName map[string]*consensus.Member

	Index map[types.PeerID]uint64 // peer ID to raft ID mapping

	Addresses []string //for raft server TODO remove
}

func newMembers(name string) *Members {
	return &Members{
		name:      name,
		MapByID:   make(map[uint64]*consensus.Member),
		MapByName: make(map[string]*consensus.Member),
		Index:     make(map[types.PeerID]uint64),
		Addresses: make([]string, 0),
	}
}

func (mbrs *Members) len() int {
	return len(mbrs.MapByID)
}

func (mbrs *Members) ToArray() []*consensus.Member {
	count := len(mbrs.MapByID)

	var arrs = make([]*consensus.Member, count)

	i := 0
	for _, m := range mbrs.MapByID {
		arrs[i] = m
		i++
	}

	sort.Sort(consensus.MembersByName(arrs))

	return arrs
}

func (mbrs *Members) ToMemberAttrArray() []*types.MemberAttr {
	count := len(mbrs.MapByID)

	var arrs = make([]*types.MemberAttr, count)

	mbrArray := mbrs.ToArray()

	i := 0
	for _, m := range mbrArray {
		arrs[i] = &m.MemberAttr
		i++
	}

	return arrs
}

func (mbrs *Members) toString() string {
	var buf string

	buf += fmt.Sprintf("%s", mbrs.name)

	if mbrs == nil {
		return "[]"
	}

	mbrsArr := mbrs.ToArray()
	sort.Sort(consensus.MembersByName(mbrsArr))

	buf += fmt.Sprintf("[")
	for _, bp := range mbrsArr {
		buf += fmt.Sprintf("%s", bp.ToString())
	}
	buf += fmt.Sprintf("]")

	return buf
}

func NewCluster(chainID []byte, bf *BlockFactory, raftName string, p2pPeerID types.PeerID, chainTimestamp int64, notifyFn NotifyFn) *Cluster {
	cl := &Cluster{
		chainID:            chainID,
		chainTimestamp:     chainTimestamp,
		ICompSyncRequester: bf,
		identity:           consensus.RaftIdentity{Name: raftName},
		members:            newMembers(MembersNameInit),
		appliedMembers:     newMembers(MembersNameApplied),
		removedMembers:     newMembers(MembersNameRemoved),
		confChangeC:        make(chan *consensus.ConfChangePropose),
	}
	if bf != nil {
		cl.cdb = bf.ChainWAL
	}

	if len(p2pPeerID) > 0 {
		cl.identity.PeerID = types.IDB58Encode(p2pPeerID)
	}
	cl.notifyFn = notifyFn

	return cl
}

func NewClusterFromMemberAttrs(clusterID uint64, chainID []byte, memberAttrs []*types.MemberAttr) (*Cluster, error) {
	cl := NewCluster(chainID, nil, "", "", 0, nil)

	for _, mbrAttr := range memberAttrs {
		var mbr consensus.Member

		mbr.SetAttr(mbrAttr)

		if err := cl.isValidMember(&mbr); err != nil {
			logger.Error().Err(err).Str("mbr", mbr.ToString()).Msg("fail to add member")
			return nil, err
		}

		if err := cl.addMember(&mbr, false); err != nil {
			logger.Error().Err(err).Str("mbr", mbr.ToString()).Msg("fail to add member")
			return nil, err
		}
	}

	if clusterID == InvalidClusterID {
		return nil, ErrClusterNotReady
	}
	cl.identity.ClusterID = clusterID

	return cl, nil
}

func (cl *Cluster) ClusterID() uint64 {
	return cl.identity.ClusterID
}

func (cl *Cluster) NodeName() string {
	return cl.identity.Name
}

func (cl *Cluster) NodeID() uint64 {
	return cl.identity.ID
}

func (cl *Cluster) NodePeerID() string {
	return cl.identity.PeerID
}

func (cl *Cluster) SetNodeID(nodeid uint64) {
	cl.identity.ID = nodeid
}

func (cl *Cluster) SetClusterID(clusterid uint64) {
	logger.Debug().Str("id", EtcdIDToString(clusterid)).Msg("set cluster ID")

	cl.identity.ClusterID = clusterid
}

// RecoverIdentity reset node id and name of cluster.
// raft identity is saved in WAL and reset when server is restarted
func (cl *Cluster) RecoverIdentity(id *consensus.RaftIdentity) error {
	cl.Lock()
	defer cl.Unlock()

	// check name
	if cl.identity.Name != id.Name {
		return ErrNotMatchedRaftName
	}

	if cl.identity.PeerID != id.PeerID {
		return ErrNotMatchedRaftPeerID
	}

	if id.ClusterID == 0 {
		return ErrInvalidRaftIdentity
	}

	cl.identity = *id

	logger.Info().Str("identity", id.ToString()).Msg("recover raft identity of this node")

	return nil
}

func (cl *Cluster) Recover(snapshot *raftpb.Snapshot) (bool, error) {
	var snapdata = &consensus.SnapshotData{}

	if err := snapdata.Decode(snapshot.Data); err != nil {
		return false, err
	}

	logger.Info().Str("snap", snapdata.ToString()).Msg("cluster recover from snapshot")

	if cl.isAllMembersEqual(snapdata.Members, snapdata.RemovedMembers) {
		logger.Info().Msg("cluster recover skipped since all members are equal to previous configure")
		return true, nil
	}

	cl.ResetMembers()

	// members restore
	for _, mbr := range snapdata.Members {
		if err := cl.addMember(mbr, true); err != nil {
			return false, err
		}
	}

	for _, mbr := range snapdata.RemovedMembers {
		cl.RemovedMembers().add(mbr)
	}

	logger.Info().Str("info", cl.toStringWithLock()).Msg("cluster recovered")

	return false, nil
}

func (cl *Cluster) ResetMembers() {
	cl.Lock()
	defer cl.Unlock()

	cl.members = newMembers(MembersNameInit)
	cl.appliedMembers = newMembers(MembersNameApplied)
	cl.removedMembers = newMembers(MembersNameRemoved)

	cl.Size = 0
}

func (cl *Cluster) isMatch(confstate *raftpb.ConfState) bool {
	var matched int

	if len(cl.AppliedMembers().MapByID) != len(confstate.Nodes) {
		return false
	}

	for _, confID := range confstate.Nodes {
		if _, ok := cl.AppliedMembers().MapByID[confID]; !ok {
			return false
		}

		matched++
	}

	return true
}

func (cl *Cluster) Members() *Members {
	return cl.members
}

func (cl *Cluster) AppliedMembers() *Members {
	return cl.appliedMembers
}

func (cl *Cluster) RemovedMembers() *Members {
	return cl.removedMembers
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
	for _, member := range cl.members.MapByID {
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
func (cl *Cluster) getAnyPeerAddressToSync() (types.PeerID, error) {
	cl.Lock()
	defer cl.Unlock()

	for _, member := range cl.Members().MapByID {
		if member.Name != cl.NodeName() {
			return member.GetPeerID(), nil
		}
	}

	return "", ErrNoEnableSyncPeer
}

func (cl *Cluster) isValidMember(member *consensus.Member) error {
	cl.Lock()
	defer cl.Unlock()

	mbrs := cl.members

	for _, prevMember := range mbrs.MapByID {
		if prevMember.HasDuplicatedAttr(member) {
			logger.Error().Str("prev", prevMember.ToString()).Str("cur", member.ToString()).Msg("duplicated configuration for raft BP member")
			return ErrDupBP
		}
	}

	// check if peerID of this node is valid
	if cl.NodeName() == member.Name && base58.Encode([]byte(member.GetPeerID())) != cl.NodePeerID() {
		logger.Error().Str("config", member.GetPeerID().String()).Str("cluster peerid", cl.NodePeerID()).Msg("peerID value is not matched with P2P")
		return ErrInvalidRaftPeerID
	}

	return nil
}

func (cl *Cluster) addMember(member *consensus.Member, applied bool) error {
	logger.Info().Str("member", member.ToString()).Bool("applied", applied).Msg("member add")

	cl.Lock()
	defer cl.Unlock()

	if applied {
		if cl.AppliedMembers().isExist(member.ID) {
			return ErrMemberAlreadyApplied
		}
		logger.Debug().Str("member", member.ToString()).Msg("add to applied members")
		cl.AppliedMembers().add(member)

		// notify to p2p TODO temporary code
		peerID, err := types.IDFromBytes(member.PeerID)
		if err != nil {
			logger.Panic().Err(err).Str("peerid", base58.Encode(member.PeerID)).Msg("invalid member peerid")
		}

		if cl.notifyFn != nil {
			cl.notifyFn(&message.RaftClusterEvent{BPAdded: []types.PeerID{peerID}})
		}
	}

	if cl.members.isExist(member.ID) {
		logger.Debug().Str("member", member.ToString()).Msg("omit adding to init members")
		return nil
	}

	cl.members.add(member)
	cl.Size++

	return nil
}

func (cl *Cluster) removeMember(member *consensus.Member) error {
	logger.Info().Str("member", member.ToString()).Msg("member remove")

	cl.Lock()
	defer cl.Unlock()

	cl.AppliedMembers().remove(member)
	cl.members.remove(member)
	cl.removedMembers.add(member)

	cl.Size--
	// notify to p2p TODO temporary code
	peerID, err := types.IDFromBytes(member.PeerID)
	if err != nil {
		logger.Panic().Err(err).Str("peerid", base58.Encode(member.PeerID)).Msg("invalid member peerid")
	}

	if cl.notifyFn != nil {
		cl.notifyFn(&message.RaftClusterEvent{BPRemoved: []types.PeerID{peerID}})
	}

	return nil
}

// ValidateAndMergeExistingCluster tests if members of existing cluster are matched with this cluster
func (cl *Cluster) ValidateAndMergeExistingCluster(existingCl *Cluster) bool {
	cl.Lock()
	defer cl.Unlock()

	if !bytes.Equal(existingCl.chainID, cl.chainID) {
		logger.Error().Msg("My chainID is different from the existing cluster")
		return false
	}

	// check if this node is already added in existing cluster
	remoteMember := existingCl.Members().getMemberByName(cl.NodeName())
	if remoteMember == nil {
		logger.Error().Msg("This node doesn't exist in the existing cluster")
		return false
	}

	// TODO check my network config is equal to member of remote
	if base58.Encode(remoteMember.PeerID) != cl.NodePeerID() {
		logger.Error().Msg("peerid is different with peerid of member of existing cluster")
	}

	cl.members = existingCl.Members()
	cl.Size = existingCl.Size

	myNodeID := existingCl.getNodeID(cl.NodeName())

	// reset self nodeID of cluster
	cl.SetNodeID(myNodeID)
	cl.SetClusterID(existingCl.ClusterID())

	logger.Debug().Str("my", cl.toStringWithLock()).Msg("cluster merged with existing cluster")
	return true
}

func (cl *Cluster) getMemberAttrs() ([]*types.MemberAttr, error) {
	cl.Lock()
	defer cl.Unlock()

	attrs := make([]*types.MemberAttr, cl.members.len())

	if cl.members.len() == 0 {
		return nil, ErrClusterHasNoMember
	}

	var i = 0
	for _, mbr := range cl.members.MapByID {
		// copy attr since it can be modified
		attr := mbr.MemberAttr
		attrs[i] = &attr
		i++
	}

	return attrs, nil
}

// IsIDRemoved return true if given raft id is not exist in cluster
func (cl *Cluster) IsIDRemoved(id uint64) bool {
	return cl.RemovedMembers().isExist(id)
}

// GenerateID generate cluster ID by hashing IDs of all initial members
func (cl *Cluster) GenerateID(useBackup bool) {
	var buf []byte

	if useBackup {
		blk, err := cl.cdb.GetBestBlock()
		if err != nil || blk == nil {
			logger.Fatal().Msg("failed to get best block from backup datafiles")
		}

		buf = append(buf, blk.BlockHash()...)
	}

	mbrs := cl.Members().ToArray()
	sort.Sort(consensus.MembersByName(mbrs))

	for _, mbr := range mbrs {
		logger.Debug().Str("id", EtcdIDToString(mbr.GetID())).Msg("member ID")

		buf = append(buf, types.Uint64ToBytes(mbr.GetID())...)
	}

	hash := sha1.Sum(buf)
	cl.identity.ClusterID = binary.LittleEndian.Uint64(hash[:8])

	logger.Info().Str("id", EtcdIDToString(cl.ClusterID())).Msg("generate cluster ID")
}

func (cl *Cluster) isAllMembersEqual(members []*consensus.Member, RemovedMembers []*consensus.Member) bool {
	membersEqual := func(x []*consensus.Member, y []*consensus.Member) bool {
		if len(x) != len(y) {
			return false
		}

		for i, mX := range x {
			mY := y[i]
			if !mX.Equal(mY) {
				return false
			}
		}

		return true
	}

	clMembers := cl.AppliedMembers().ToArray()
	clRemovedMembers := cl.RemovedMembers().ToArray()

	sort.Sort(consensus.MembersByName(members))
	sort.Sort(consensus.MembersByName(RemovedMembers))

	if !membersEqual(clMembers, members) {
		return false
	}

	if !membersEqual(clRemovedMembers, RemovedMembers) {
		return false
	}

	return true
}

func (mbrs *Members) add(member *consensus.Member) {
	mbrs.MapByID[member.ID] = member
	mbrs.MapByName[member.Name] = member
	mbrs.Index[member.GetPeerID()] = member.ID
	mbrs.Addresses = append(mbrs.Addresses, member.Address)
}

func (mbrs *Members) remove(member *consensus.Member) {
	delete(mbrs.MapByID, member.ID)
	delete(mbrs.MapByName, member.Name)
	delete(mbrs.Index, member.GetPeerID())
}

func (mbrs *Members) getMemberByName(name string) *consensus.Member {
	member, ok := mbrs.MapByName[name]
	if !ok {
		return nil
	}

	return member
}

func (mbrs *Members) isExist(id uint64) bool {
	return mbrs.getMember(id) != nil
}

func (mbrs *Members) getMember(id uint64) *consensus.Member {
	member, ok := mbrs.MapByID[id]
	if !ok {
		return nil
	}

	return member
}

func (mbrs *Members) getMemberByPeerID(pid types.PeerID) *consensus.Member {
	return mbrs.getMember(mbrs.Index[pid])
}

func (mbrs *Members) getMemberPeerAddress(id uint64) (types.PeerID, error) {
	member := mbrs.getMember(id)
	if member == nil {
		return "", ErrNotExistRaftMember
	}

	return member.GetPeerID(), nil
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
		var peers map[types.PeerID]*message.PeerInfo
		var err error
		var peerBestNo uint64 = 0

		if cc.Size == 1 {
			return true, nil
		}

		// request GetPeers to p2p
		getBPPeers := func() (map[types.PeerID]*message.PeerInfo, error) {
			peers := make(map[types.PeerID]*message.PeerInfo)

			result, err := cc.RequestFuture(message.P2PSvc, &message.GetPeers{}, time.Second, "raft cluster sync test").Result()
			if err != nil {
				return nil, err
			}

			msg := result.(*message.GetPeersRsp)

			for _, peerElem := range msg.Peers {
				peerID := types.PeerID(peerElem.Addr.PeerID)
				state := peerElem.State

				if peerElem.Self {
					continue
				}

				if state.Get() != types.RUNNING {
					logger.Debug().Stringer("peer", types.LogPeerShort(peerID)).Msg("peer is not running")
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
func (cl *Cluster) toStringWithLock() string {
	var buf string

	buf = fmt.Sprintf("total=%d, cluserID=%x, NodeName=%s, RaftID=%x, ", cl.Size, cl.ClusterID(), cl.NodeName(), cl.NodeID())
	buf += "members: " + cl.members.toString()
	buf += ", appliedMembers: " + cl.appliedMembers.toString()

	return buf
}

func (cl *Cluster) toString() string {
	cl.Lock()
	defer cl.Unlock()

	return cl.toStringWithLock()
}

func (cl *Cluster) getNodeID(name string) uint64 {
	m, ok := cl.Members().MapByName[name]
	if !ok {
		return consensus.InvalidMemberID
	}

	return m.ID
}

func (cl *Cluster) getRaftInfo(withStatus bool) *RaftInfo {
	cl.Lock()
	defer cl.Unlock()

	var leader uint64
	if cl.rs != nil {
		leader = cl.rs.GetLeader()
	}

	var leaderName string
	var m *consensus.Member

	if m = cl.Members().getMember(leader); m != nil {
		leaderName = m.Name
	} else {
		leaderName = "id=" + EtcdIDToString(leader)
	}

	rinfo := &RaftInfo{Leader: leaderName, Total: cl.Size, Name: cl.NodeName(), RaftId: EtcdIDToString(cl.NodeID())}

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
		Addr   string
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
	if cl.Size != 0 {
		bps := make([]string, cl.Size)

		for id, m := range cl.Members().MapByID {
			bp := &PeerInfo{Name: m.Name, RaftID: EtcdIDToString(m.ID), PeerID: m.GetPeerID().String(), Addr: m.Address}
			b, err = json.Marshal(bp)
			if err != nil {
				logger.Error().Err(err).Str("raftid", EtcdIDToString(id)).Msg("failed to marshalEntryData raft consensus bp")
				return &emptyCons
			}
			bps[i] = string(b)

			i++
		}
		cons.Bps = bps
	}

	return &cons
}

func (cl *Cluster) NewMemberFromAddReq(req *types.MembershipChange) (*consensus.Member, error) {
	if len(req.Attr.Name) == 0 || len(req.Attr.Address) == 0 || len(req.Attr.PeerID) == 0 {
		return nil, consensus.ErrInvalidMemberAttr
	}

	return consensus.NewMember(req.Attr.Name, req.Attr.Address, types.PeerID(req.Attr.PeerID), cl.chainID, time.Now().UnixNano()), nil
}

func (cl *Cluster) NewMemberFromRemoveReq(req *types.MembershipChange) (*consensus.Member, error) {
	if req.Attr.ID == consensus.InvalidMemberID {
		return nil, consensus.ErrInvalidMemberID
	}

	member := consensus.NewMember("", "", types.PeerID(""), cl.chainID, 0)
	member.SetMemberID(req.Attr.ID)

	return member, nil
}

func (cl *Cluster) ChangeMembership(req *types.MembershipChange, nowait bool) (*consensus.Member, error) {
	var (
		proposal *consensus.ConfChangePropose
		err      error
	)

	submit := func() error {
		cl.Lock()
		defer cl.Unlock()

		if proposal, err = cl.makeProposal(req, nowait); err != nil {
			logger.Error().Uint64("requestID", req.GetRequestID()).Msg("failed to make proposal for membership change")
			return err
		}

		if err = cl.isEnableChangeMembership(proposal.Cc); err != nil {
			logger.Error().Err(err).Msg("failed cluster availability check to change membership")
			return err
		}

		if err = cl.submitProposal(proposal, nowait); err != nil {
			return err
		}

		return nil
	}

	if err = submit(); err != nil {
		return nil, err
	}

	if nowait {
		return nil, nil
	}

	return cl.recvConfChangeReply(proposal.ReplyC)
}

func (cl *Cluster) makeProposal(req *types.MembershipChange, nowait bool) (*consensus.ConfChangePropose, error) {
	if cl.savedChange != nil {
		logger.Error().Str("cc", types.RaftConfChangeToString(cl.savedChange.Cc)).Msg("already exist pending conf change")
		return nil, ErrPendingConfChange
	}

	var (
		replyC chan *consensus.ConfChangeReply
		member *consensus.Member
		err    error
	)

	switch req.Type {
	case types.MembershipChangeType_ADD_MEMBER:
		member, err = cl.NewMemberFromAddReq(req)

	case types.MembershipChangeType_REMOVE_MEMBER:
		member, err = cl.NewMemberFromRemoveReq(req)

	default:
		return nil, ErrInvalidMembershipReqType
	}

	if err != nil {
		logger.Error().Err(err).Uint64("requestID", req.GetRequestID()).Msg("failed to make new member")
		return nil, err
	}

	// make raft confChange
	cc, err := cl.makeConfChange(req.GetRequestID(), req.Type, member)
	if err != nil {
		logger.Error().Err(err).Uint64("requestID", req.GetRequestID()).Msg("failed to make conf change of raft")
		return nil, err
	}

	// validate member change
	if err = cl.validateChangeMembership(cc, member, false); err != nil {
		logger.Error().Err(err).Uint64("requestID", req.GetRequestID()).Msg("failed to validate request of membership change")
		return nil, err
	}

	if !nowait {
		replyC = make(chan *consensus.ConfChangeReply, 1)
	}

	// TODO check cancel
	ctx, cancel := context.WithTimeout(context.Background(), MaxConfChangeTimeOut)
	defer cancel()

	// send proposeC (confChange, replyC)
	proposal := consensus.ConfChangePropose{Ctx: ctx, Cc: cc, ReplyC: replyC}

	return &proposal, nil
}

func (cl *Cluster) submitProposal(proposal *consensus.ConfChangePropose, nowait bool) error {
	if cl.savedChange != nil {
		return ErrPendingConfChange
	}

	cl.saveConfChangePropose(proposal)

	select {
	case cl.confChangeC <- proposal:
		logger.Info().Uint64("requestID", proposal.Cc.ID).Msg("proposal of conf change is sent to raft")
	default:
		logger.Error().Uint64("requestID", proposal.Cc.ID).Msg("proposal of conf change is dropped. confChange channel is busy")

		if !nowait {
			close(proposal.ReplyC)
		}
		cl.resetSavedConfChangePropose()
		return ErrConfChangeChannelBusy
	}

	return nil
}

func (cl *Cluster) recvConfChangeReply(replyC chan *consensus.ConfChangeReply) (*consensus.Member, error) {
	select {
	case reply, ok := <-replyC:
		if !ok {
			logger.Panic().Msg("reply channel of change request must not be closed")
		}

		if reply.Err != nil {
			logger.Error().Err(reply.Err).Msg("failed conf change")
			return nil, reply.Err
		}

		logger.Info().Str("cluster", cl.toString()).Str("target", reply.Member.ToString()).Msg("reply of conf change is succeed")

		return reply.Member, nil
	case <-time.After(MaxConfChangeTimeOut):
		// saved conf change must be reset in raft server after request completes
		logger.Warn().Msg("proposal of conf change is time-out")

		return nil, ErrConChangeTimeOut
	}
}

func (cl *Cluster) AfterConfChange(cc *raftpb.ConfChange, member *consensus.Member, err error) {
	cl.Lock()
	defer cl.Unlock()

	// TODO XXX if leader is rebooted, savedChange will be nil, so need to handle this situation
	if cl.savedChange == nil || cl.savedChange.Cc.ID != cc.ID {
		return
	}

	propose := cl.savedChange

	logger.Info().Str("req", util.JSON(propose.Cc)).Msg("conf change succeed")

	cl.resetSavedConfChangePropose()

	if propose.ReplyC != nil {
		propose.ReplyC <- &consensus.ConfChangeReply{Member: member, Err: err}
		close(propose.ReplyC)
	}
}

func (cl *Cluster) saveConfChangePropose(ccPropose *consensus.ConfChangePropose) {
	logger.Debug().Uint64("ccid", ccPropose.Cc.ID).Msg("this conf change propose is saved in cluster")
	cl.savedChange = ccPropose
}

func (cl *Cluster) resetSavedConfChangePropose() {
	var ccid uint64

	if cl.savedChange == nil {
		return
	}

	ccid = cl.savedChange.Cc.ID

	logger.Debug().Uint64("requestID", ccid).Msg("reset saved conf change propose")

	cl.savedChange = nil
}

var (
	ErrRaftStatusEmpty = errors.New("raft status is empty")
)

// isEnableChangeMembership check if membership change request can stop cluster.
// case add : current avaliable node < (n + 1)/ 2 + 1
// case remove : avaliable node except node to remove < (n - 1) / 2 - 1
//
// Default :
//   - Add : 1 node라도 장애 node or slow node가 존재하면 add는 불가
//     slow node기준 - block 높이가 100이상 차이 나는 경우
//   - Remove :
//     현재 cluster가 available 해야함
//     node를 뺄때는 (정상node - 1) >= (n - 1) / 2 + 1 이어야함
//     slow node는 항상 뺄수 있다. slow node를 뺌으로써 cluster를 정상으로 만들기 위함
//   - force 모드: 무조건 실행 한다.
func (cl *Cluster) isEnableChangeMembership(cc *raftpb.ConfChange) error {
	status := cl.rs.Status()
	if status.ID == 0 {
		logger.Debug().Msg("raft node is not initialized")
		return ErrRaftStatusEmpty
	}

	cp, err := cl.rs.GetClusterProgress()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get cluster progress")
		return err
	}

	logger.Info().Str("info", cp.ToString()).Msg("cluster progress")

	getHealthyMembers := func(cp *ClusterProgress) int {
		var healthy int

		for _, mp := range cp.MemberProgresses {
			if mp.Status == MemberProgressStateHealthy {
				healthy++
			}
		}

		return healthy
	}

	isClusterAvilable := func(total int, healthy int) bool {
		quorum := total/2 + 1

		logger.Info().Int("quorum", quorum).Int("total", total).Int("healthy", healthy).Msg("cluster quorum")

		return healthy >= quorum
	}

	healthy := getHealthyMembers(cp)

	if !isClusterAvilable(cp.N, healthy) {
		logger.Warn().Msg("curretn cluster status doesn't satisfy quorum")
	}

	switch {
	case cc.Type == raftpb.ConfChangeAddNode:
		for _, mp := range cp.MemberProgresses {
			if mp.Status != MemberProgressStateHealthy {
				logger.Error().Uint64("slowgap", MaxSlowNodeGap).Str("unhealthy member", mp.ToString()).Msg("exist unhealthy member in cluster. If you want add some node, fix the unhealthy node and try again")
				return ErrUnhealtyNodeExist
			}
		}

		return nil
	case cc.Type == raftpb.ConfChangeRemoveNode:
		mp, ok := cp.MemberProgresses[cc.NodeID]
		if !ok {
			logger.Error().Uint64("id", cc.NodeID).Msg("not exist progress of member")
			return ErrNotExitRaftProgress
		}

		if mp.Status != MemberProgressStateHealthy {
			logger.Warn().Uint64("memberid", mp.MemberID).Msg("try to remove slow node")
			return nil
		}

		if !isClusterAvilable(cp.N-1, healthy-1) {
			logger.Error().Msg("can't remove healthy node. If you remove this node, cluster can be stop.")
			return ErrRemoveHealthyNode
		}

		return nil
	default:
		logger.Error().Msg("type of conf change is invalid")
		return ErrInvalidMembershipReqType
	}
}

func (cl *Cluster) validateChangeMembership(cc *raftpb.ConfChange, member *consensus.Member, needlock bool) error {
	if member == nil {
		return ErrCCMemberIsNil
	}

	if needlock {
		cl.Lock()
		defer cl.Unlock()
	}

	appliedMembers := cl.AppliedMembers()

	if member.ID == consensus.InvalidMemberID {
		return consensus.ErrInvalidMemberID
	}
	if cl.RemovedMembers().isExist(member.ID) {
		return ErrCCAlreadyRemoved
	}

	switch cc.Type {
	case raftpb.ConfChangeAddNode:
		if !member.IsValid() {
			logger.Error().Str("member", member.ToString()).Msg("member has invalid fields")
			return ErrInvalidMember
		}

		if m := appliedMembers.getMember(member.ID); m != nil {
			return ErrCCAlreadyAdded
		}

		if err := appliedMembers.hasDuplicatedMember(member); err != nil {
			return err
		}

	case raftpb.ConfChangeRemoveNode:
		var m *consensus.Member

		if m = appliedMembers.getMember(member.ID); m == nil {
			return ErrCCNoMemberToRemove
		}

		*member = *m
	default:
		return ErrInvCCType
	}

	// - TODO UPDATE
	return nil
}

func (cl *Cluster) makeConfChange(reqID uint64, reqType types.MembershipChangeType, member *consensus.Member) (*raftpb.ConfChange, error) {
	var changeType raftpb.ConfChangeType
	switch reqType {
	case types.MembershipChangeType_ADD_MEMBER:
		changeType = raftpb.ConfChangeAddNode
	case types.MembershipChangeType_REMOVE_MEMBER:
		changeType = raftpb.ConfChangeRemoveNode
	default:
		return nil, ErrInvalidMembershipReqType
	}

	logger.Debug().Uint64("requestID", reqID).Str("member", member.ToString()).Msg("conf change target member")

	cl.changeSeq++

	data, err := json.Marshal(member)
	if err != nil {
		return nil, err
	}

	// generateConfChangeID
	cc := &raftpb.ConfChange{ID: reqID, Type: changeType, NodeID: uint64(member.ID), Context: data}

	return cc, nil
}

func EtcdIDToString(id uint64) string {
	return fmt.Sprintf("%x", id)
}
