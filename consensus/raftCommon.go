package consensus

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/etcd/raft"
	"github.com/aergoio/etcd/raft/raftpb"
)

type EntryType int8

const (
	EntryBlock EntryType = iota
	EntryEmpty           // it is generated when node becomes leader
	EntryConfChange
	InvalidMemberID = 0
)

type ConfChangePropose struct {
	Ctx context.Context
	Cc  *raftpb.ConfChange

	ReplyC chan *ConfChangeReply
}

type ConfChangeReply struct {
	Member *Member
	Err    error
}

var (
	WalEntryType_name = map[EntryType]string{
		0: "EntryBlock",
		1: "EntryEmpty",
		2: "EntryConfChange",
	}

	ErrInvalidMemberID        = errors.New("member id of conf change doesn't match")
	ErrEmptySnapData          = errors.New("failed to decode snapshot data. encoded data is empty")
	ErrInvalidMemberAttr      = errors.New("invalid member attribute")
	ErrorMembershipChangeSkip = errors.New("node is not raft leader, so skip membership change request")
)

type WalEntry struct {
	Type  EntryType
	Term  uint64
	Index uint64
	Data  []byte // hash is set if Type is EntryBlock
}

func (we *WalEntry) ToBytes() ([]byte, error) {
	var val bytes.Buffer
	encoder := gob.NewEncoder(&val)
	if err := encoder.Encode(we); err != nil {
		logger.Panic().Err(err).Msg("raft entry to bytes error")
	}

	return val.Bytes(), nil
}

func (we *WalEntry) ToString() string {
	if we == nil {
		return "wal entry is nil"
	}
	return fmt.Sprintf("wal entry[type:%s, index:%d, term:%d", WalEntryType_name[we.Type], we.Index, we.Term)
}

type RaftIdentity struct {
	ClusterID uint64
	ID        uint64
	Name      string
	PeerID    string // base58 encoded format
}

func (rid *RaftIdentity) ToString() string {
	if rid == nil {
		return "raft identity is nil"
	}
	return fmt.Sprintf("raft identity[name:%s, nodeid:%x, peerid:%s]", rid.Name, rid.ID, rid.PeerID)
}

type ChainWAL interface {
	ChainDB

	ClearWAL()
	ResetWAL(hardStateInfo *types.HardStateInfo) error
	WriteRaftEntry([]*WalEntry, []*types.Block, []*raftpb.ConfChange) error
	GetRaftEntry(idx uint64) (*WalEntry, error)
	HasWal(identity RaftIdentity) (bool, error)
	GetRaftEntryOfBlock(hash []byte) (*WalEntry, error)
	GetRaftEntryLastIdx() (uint64, error)
	GetRaftEntryIndexOfBlock(hash []byte) (uint64, error)
	GetHardState() (*raftpb.HardState, error)
	WriteHardState(hardstate *raftpb.HardState) error
	WriteSnapshot(snap *raftpb.Snapshot) error
	GetSnapshot() (*raftpb.Snapshot, error)
	WriteIdentity(id *RaftIdentity) error
	GetIdentity() (*RaftIdentity, error)
	WriteConfChangeProgress(id uint64, progress *types.ConfChangeProgress) error
	GetConfChangeProgress(id uint64) (*types.ConfChangeProgress, error)
}

type SnapshotData struct {
	Chain          ChainSnapshot `json:"chain"`
	Members        []*Member     `json:"members"`
	RemovedMembers []*Member
}

func NewSnapshotData(members []*Member, rmMembers []*Member, block *types.Block) *SnapshotData {
	if block == nil {
		return nil
	}

	return &SnapshotData{
		Chain:          *NewChainSnapshot(block),
		Members:        members,
		RemovedMembers: rmMembers,
	}
}

func (snapd *SnapshotData) Encode() ([]byte, error) {
	return json.Marshal(snapd)
}

func (snapd *SnapshotData) Decode(data []byte) error {
	if len(data) == 0 {
		return ErrEmptySnapData
	}
	return json.Unmarshal(data, snapd)
}

func (snapd *SnapshotData) Equal(t *SnapshotData) bool {
	if !snapd.Chain.Equal(&t.Chain) {
		return false
	}

	if len(t.Members) != len(snapd.Members) {
		return false
	}

	for i, m := range snapd.Members {
		tMbr := t.Members[i]

		if !m.Equal(tMbr) {
			return false
		}
	}

	return true
}

func (snapd *SnapshotData) ToString() string {
	var buf string

	buf += fmt.Sprintf("chain:%s, ", snapd.Chain.ToString())

	printMembers := func(mbrs []*Member, name string) {
		if len(mbrs) > 0 {
			buf += fmt.Sprintf("%s[", name)

			for i, m := range mbrs {
				buf += fmt.Sprintf("#%d{%s}", i, m.ToString())
			}

			buf += fmt.Sprintf("]")
		}
	}

	printMembers(snapd.Members, "members")
	printMembers(snapd.RemovedMembers, "removed members")

	return buf
}

type ChainSnapshot struct {
	No   types.BlockNo `json:"no"`
	Hash []byte        `json:"hash"`
}

func NewChainSnapshot(block *types.Block) *ChainSnapshot {
	if block == nil {
		return nil
	}

	return &ChainSnapshot{No: block.BlockNo(), Hash: block.BlockHash()}
}

func (csnap *ChainSnapshot) Equal(other *ChainSnapshot) bool {
	return csnap.No == other.No && bytes.Equal(csnap.Hash, other.Hash)
}

func (csnap *ChainSnapshot) ToString() string {
	if csnap == nil || csnap.Hash == nil {
		return "csnap: empty"
	}
	return fmt.Sprintf("chainsnap:(no=%d, hash=%s)", csnap.No, enc.ToString(csnap.Hash))
}

/*
func (csnap *ChainSnapshot) Encode() ([]byte, error) {
	var val bytes.Buffer

	encoder := gob.NewEncoder(&val)
	if err := encoder.Encode(csnap); err != nil {
		logger.Fatal().Err(err).Msg("failed to encode chainsnap")
		return nil, err
	}

	return val.Bytes(), nil
}

func DecodeChainSnapshot(data []byte) (*ChainSnapshot, error) {
	var snap ChainSnapshot
	var b bytes.Buffer
	b.Write(data)

	if data == nil {
		return nil, ErrEmptySnapData
	}

	decoder := gob.NewDecoder(&b)
	if err := decoder.Decode(&snap); err != nil {
		logger.Fatal().Err(err).Msg("failed to decode chainsnap")
		return nil, err
	}

	return &snap, nil
}*/

func ConfStateToString(conf *raftpb.ConfState) string {
	var buf string

	if len(conf.Nodes) > 0 {
		buf = fmt.Sprintf("node")
		for _, node := range conf.Nodes {
			buf = buf + fmt.Sprintf("[%x]", node)
		}
	}

	if len(conf.Learners) > 0 {
		buf = buf + fmt.Sprintf(".learner")
		for _, learner := range conf.Learners {
			buf = buf + fmt.Sprintf("[%x]", learner)
		}
	}
	return buf
}

func SnapToString(snap *raftpb.Snapshot, snapd *SnapshotData) string {
	var buf string
	buf = buf + fmt.Sprintf("snap=[index:%d term:%d conf:%s]", snap.Metadata.Index, snap.Metadata.Term, ConfStateToString(&snap.Metadata.ConfState))

	if snapd != nil {
		buf = buf + fmt.Sprintf(", %s", snapd.ToString())
	}

	return buf
}

type Member struct {
	types.MemberAttr
}

func NewMember(name string, address string, peerID types.PeerID, chainID []byte, when int64) *Member {
	//check unique
	m := &Member{MemberAttr: types.MemberAttr{Name: name, Address: address, PeerID: []byte(peerID)}}

	//make ID
	m.CalculateMemberID(chainID, when)

	return m
}

func (m *Member) Clone() *Member {
	newM := Member{MemberAttr: types.MemberAttr{ID: m.ID, Name: m.Name, Address: m.Address}}

	copy(newM.PeerID, m.PeerID)

	return &newM
}

func (m *Member) SetAttr(attr *types.MemberAttr) {
	m.MemberAttr = *attr
}

func (m *Member) SetMemberID(id uint64) {
	m.ID = id
}

func (m *Member) CalculateMemberID(chainID []byte, curTimestamp int64) {
	var buf []byte

	buf = append(buf, []byte(m.Name)...)
	buf = append(buf, []byte(chainID)...)
	buf = append(buf, []byte(fmt.Sprintf("%d", curTimestamp))...)

	hash := sha1.Sum(buf)
	m.ID = binary.LittleEndian.Uint64(hash[:8])
}

func (m *Member) IsValid() bool {
	if m.ID == InvalidMemberID || len(m.PeerID) == 0 || len(m.Name) == 0 || len(m.Address) == 0 {
		return false
	}

	if _, err := types.ParseMultiaddr(m.Address); err != nil {
		logger.Error().Err(err).Msg("parse address of member")
		return false
	}

	return true
}

func (m *Member) GetPeerID() types.PeerID {
	return types.PeerID(m.PeerID)
}

func (m *Member) Equal(other *Member) bool {
	return m.ID == other.ID &&
		bytes.Equal(m.PeerID, other.PeerID) &&
		m.Name == other.Name &&
		m.Address == other.Address &&
		bytes.Equal([]byte(m.PeerID), []byte(other.PeerID))
}

func (m *Member) ToString() string {
	data, err := json.Marshal(&m.MemberAttr)
	if err != nil {
		logger.Error().Err(err).Str("name", m.Name).Msg("can't unmarshal member")
		return ""
	}
	return string(data)
}

func (m *Member) HasDuplicatedAttr(x *Member) bool {
	if m.Name == x.Name || m.ID == x.ID || m.Address == x.Address || bytes.Equal(m.PeerID, x.PeerID) {
		return true
	}

	return false
}

// IsCompatible checks if name, url and peerid of this member are the same with other member
func (m *Member) IsCompatible(other *Member) bool {
	return m.Name == other.Name && m.Address == other.Address && bytes.Equal(m.PeerID, other.PeerID)
}

type MembersByName []*Member

func (mbrs MembersByName) Len() int {
	return len(mbrs)
}
func (mbrs MembersByName) Less(i, j int) bool {
	return mbrs[i].Name < mbrs[j].Name
}
func (mbrs MembersByName) Swap(i, j int) {
	mbrs[i], mbrs[j] = mbrs[j], mbrs[i]
}

// DummyRaftAccessor returns error if process request comes, or silently ignore raft message.
type DummyRaftAccessor struct {
}

var IllegalArgumentError = errors.New("illegal argument")

func (DummyRaftAccessor) Process(ctx context.Context, peerID types.PeerID, m raftpb.Message) error {
	return IllegalArgumentError
}

func (DummyRaftAccessor) IsIDRemoved(peerID types.PeerID) bool {
	return false
}

func (DummyRaftAccessor) ReportUnreachable(peerID types.PeerID) {
}

func (DummyRaftAccessor) ReportSnapshot(peerID types.PeerID, status raft.SnapshotStatus) {
}

func (DummyRaftAccessor) GetMemberByID(id uint64) *Member {
	return nil
}

func (DummyRaftAccessor) GetMemberByPeerID(peerID types.PeerID) *Member {
	return nil
}

func (DummyRaftAccessor) SaveFromRemote(r io.Reader, id uint64, msg raftpb.Message) (int64, error) {
	return 0, nil
}
