package raftv2

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"strconv"
	"sync"
)

var (
	ErrNotExistRaftMember = errors.New("not exist member of raft cluster")
	ErrNoEnableSyncPeer   = errors.New("no peer to sync chain")
)

const (
	InvalidMemberID = MemberID(0)
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
	NodeID   MemberID

	Size uint16

	effeitveMembers *Members

	configMembers *Members
	members       *Members
}

type MemberID uint64

type Members struct {
	MapByID   map[MemberID]*Member // restore from DB or snapshot
	MapByName map[string]*Member

	Index map[peer.ID]MemberID // peer ID to raft ID mapping

	BPUrls []string //for raft server TODO remove
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

type Member struct {
	ID     MemberID `json:"id"`
	Name   string   `json:"name"`
	Url    string   `json:"url"`
	PeerID peer.ID  `json:"peerid"`
}

func newMembers(size uint16) *Members {
	return &Members{
		MapByID:   make(map[MemberID]*Member),
		MapByName: make(map[string]*Member),
		Index:     make(map[peer.ID]MemberID),
		BPUrls:    make([]string, 0),
	}
}

func newMember(name string, url string, peerID peer.ID, chainID []byte, when int64) *Member {
	//check unique
	bp := &Member{Name: name, Url: url, PeerID: peerID}

	//make ID
	bp.SetMemberID(chainID, when)

	return bp
}

func (bp *Member) SetMemberID(chainID []byte, curTimestamp int64) {
	var buf []byte

	buf = append(buf, []byte(bp.Name)...)
	buf = append(buf, []byte(chainID)...)
	buf = append(buf, []byte(fmt.Sprintf("%d", curTimestamp))...)

	hash := sha1.Sum(buf)
	bp.ID = MemberID(binary.LittleEndian.Uint64(hash[:8]))
}

func (bp *Member) isValid() bool {
	if bp.ID == InvalidMemberID || len(bp.PeerID) == 0 || len(bp.Name) == 0 || len(bp.Url) == 0 {
		return false
	}

	if _, err := parseToUrl(bp.Url); err != nil {
		logger.Error().Err(err).Msg("parse url of member")
		return false
	}

	return true
}

func (bp *Member) ToString() string {
	return fmt.Sprintf("member{Name:%s, ID:%x, Url:%s, PeerID:%s}", bp.Name, bp.ID, bp.Url, bp.PeerID)
}

func (bp *Member) Marshal() ([]byte, error) {
	return json.Marshal(bp)
}

func (bp *Member) Unmarshal(data []byte) error {
	return json.Unmarshal(data, bp)
}

func (bp *Member) hasDuplicatedAttr(x *Member) bool {
	if bp.Name == x.Name || bp.ID == x.ID || bp.Url == x.Url || bp.PeerID == x.PeerID {
		return true
	}

	return false
}

func NewCluster(chainID []byte, bf *BlockFactory, raftName string, size uint16, chainTimestamp int64) *Cluster {
	cl := &Cluster{
		chainID:            chainID,
		chainTimestamp:     chainTimestamp,
		ICompSyncRequester: bf,
		NodeName:           raftName,
		Size:               size,
		configMembers:      newMembers(size),
		members:            newMembers(0),
		cdb:                bf.ChainWAL,
	}

	cl.effeitveMembers = cl.configMembers

	return cl
}

// getEffectiveMembers returns configMembers if members doesn't loaded from DB or snapshot
func (cl *Cluster) getEffectiveMembers() *Members {
	return cl.effeitveMembers
}

func (cl *Cluster) Quorum() uint16 {
	return cl.Size/2 + 1
}

// getAnyPeerAddressToSync returns peer address that has block of no for sync
func (cl *Cluster) getAnyPeerAddressToSync() (peer.ID, error) {
	for _, member := range cl.getEffectiveMembers().MapByID {
		if member.Name != cl.NodeName {
			return member.PeerID, nil
		}
	}

	return "", ErrNoEnableSyncPeer
}

func (mbrs *Members) add(member *Member, nodeName string) error {
	for _, prevMember := range mbrs.MapByID {
		if prevMember.hasDuplicatedAttr(member) {
			logger.Error().Str("prev", prevMember.ToString()).Str("cur", member.ToString()).Msg("duplicated configuration for raft BP member")
			return ErrDupBP
		}
	}

	// check if mapping between raft id and PeerID is valid
	if nodeName == member.Name && member.PeerID != p2p.NodeID() {
		return ErrInvalidRaftPeerID
	}

	mbrs.MapByID[member.ID] = member
	mbrs.MapByName[member.Name] = member

	mbrs.Index[member.PeerID] = member.ID
	mbrs.BPUrls = append(mbrs.BPUrls, member.Url)

	logger.Debug().Str("member", member.ToString()).Msg("add raft member")

	return nil
}

func (mbrs *Members) getMemberByName(name string) *Member {
	member, ok := mbrs.MapByName[name]
	if !ok {
		return nil
	}

	return member
}

func (mbrs *Members) getMember(id MemberID) *Member {
	member, ok := mbrs.MapByID[id]
	if !ok {
		return nil
	}

	return member
}

func (mbrs *Members) getMemberPeerAddress(id MemberID) (peer.ID, error) {
	member := mbrs.getMember(id)
	if member == nil {
		return "", ErrNotExistRaftMember
	}

	//logger.Debug().Str("rid", MemberIDToString(id)).Str("peer", member.PeerID.Pretty()).Msg("raft member")
	return member.PeerID, nil
}

// hasDuplicatedMember returns true if any attributes of the given member is equal to the attributes of cluster members
func (mbrs *Members) hasDuplicatedMember(m *Member) error {
	for _, prevMember := range mbrs.MapByID {
		if prevMember.hasDuplicatedAttr(m) {
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
	var buf string

	myNode := cl.configMembers.getMemberByName(cl.NodeName)

	buf = fmt.Sprintf("raft cluster configure: total=%d, NodeName=%s, RaftID=%d", cl.Size, cl.NodeName, myNode.ID)
	buf += ", config members: " + cl.configMembers.toString()
	buf += ", runtime members: " + cl.members.toString()

	return buf
}

func (cl *Cluster) getRaftInfo(withStatus bool) *RaftInfo {
	var leader MemberID
	if cl.rs != nil {
		leader = cl.rs.GetLeader()
	}

	var leaderName string
	var m *Member

	if m = cl.getEffectiveMembers().getMember(leader); m != nil {
		leaderName = m.Name
	} else {
		leaderName = "id=" + strconv.FormatUint(uint64(leader), 10)
	}

	rinfo := &RaftInfo{Leader: leaderName, Total: strconv.FormatUint(uint64(cl.Size), 10), Name: cl.NodeName, RaftId: strconv.FormatUint(uint64(cl.NodeID), 10)}

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

	cons := emptyCons
	cons.Info = string(b)

	var i int = 0
	bps := make([]string, cl.Size)

	for id, m := range cl.getEffectiveMembers().MapByID {
		bp := &PeerInfo{Name: m.Name, RaftID: strconv.FormatUint(uint64(m.ID), 10), PeerID: m.PeerID.Pretty()}
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

func MemberIDToString(id MemberID) string {
	return fmt.Sprintf("%x", id)
}
