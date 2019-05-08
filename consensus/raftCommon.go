package consensus

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/libp2p/go-libp2p-peer"
	"net"
	"net/url"
)

type EntryType int8

const (
	EntryBlock EntryType = iota
	EntryEmpty           // it is generated when node becomes leader
	EntryConfChange
	InvalidMemberID = 0
)

var (
	WalEntryType_name = map[EntryType]string{
		0: "EntryBlock",
		1: "EntryEmpty",
		2: "EntryConfChange",
	}

	ErrURLInvalidScheme = errors.New("url has invalid scheme")
	ErrURLInvalidPort   = errors.New("url must have host:port style")
	ErrInvalidMemberID  = errors.New("member id of conf change doesn't match")
	ErrEmptySnapData    = errors.New("failed to decode snapshot data. encoded data is empty")
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
		panic("raft entry to bytes error")
		return nil, err
	}

	return val.Bytes(), nil
}

func (we *WalEntry) ToString() string {
	if we == nil {
		return "wal entry is nil"
	}
	return fmt.Sprintf("wal entry[type:%s, index:%d, term:%d", WalEntryType_name[we.Type], we.Index, we.Term)
}

type ChainWAL interface {
	ChainDB

	IsWALInited() bool
	GetBlock(blockHash []byte) (*types.Block, error)
	ReadAll() (state raftpb.HardState, ents []raftpb.Entry, err error)
	WriteRaftEntry([]*WalEntry, []*types.Block) error
	GetRaftEntry(idx uint64) (*WalEntry, error)
	HasWal() (bool, error)
	GetRaftEntryLastIdx() (uint64, error)
	GetHardState() (*raftpb.HardState, error)
	WriteHardState(hardstate *raftpb.HardState) error
	WriteSnapshot(snap *raftpb.Snapshot) error
	GetSnapshot() (*raftpb.Snapshot, error)
}

type SnapshotData struct {
	Chain   ChainSnapshot `json:"chain"`
	Members []*Member     `json:"members"`
}

func NewSnapshotData(members []*Member, block *types.Block) *SnapshotData {
	if block == nil {
		return nil
	}

	return &SnapshotData{
		Chain:   *NewChainSnapshot(block),
		Members: members,
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

	buf += fmt.Sprintf("chain:%s. members:[", snapd.Chain.ToString())

	for i, m := range snapd.Members {
		buf += fmt.Sprintf("#%d{%s}", i, m.ToString())
	}
	buf += fmt.Sprintf("]")

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

	buf = fmt.Sprintf("node")
	for _, node := range conf.Nodes {
		buf = buf + fmt.Sprintf("[%x]", node)
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
	/*
		ID     MemberID `json:"id"`
		Name   string   `json:"name"`
		Url    string   `json:"url"`
		PeerID peer.ID  `json:"peerid"`*/
	types.MemberAttr
}

func NewMember(name string, url string, peerID peer.ID, chainID []byte, when int64) *Member {
	//check unique
	m := &Member{types.MemberAttr{Name: name, Url: url, PeerID: []byte(peerID)}}

	//make ID
	m.CalculateMemberID(chainID, when)

	return m
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
	if m.ID == InvalidMemberID || len(m.PeerID) == 0 || len(m.Name) == 0 || len(m.Url) == 0 {
		return false
	}

	if _, err := ParseToUrl(m.Url); err != nil {
		logger.Error().Err(err).Msg("parse url of member")
		return false
	}

	return true
}

func (m *Member) GetPeerID() peer.ID {
	return peer.ID(m.PeerID)
}

func (m *Member) Equal(other *Member) bool {
	return m.ID == other.ID &&
		bytes.Equal(m.PeerID, other.PeerID) &&
		m.Name == other.Name &&
		m.Url == other.Url &&
		bytes.Equal([]byte(m.PeerID), []byte(other.PeerID))
}

func (m *Member) ToString() string {
	return fmt.Sprintf("{Name:%s, ID:%x, Url:%s, PeerID:%s}", m.Name, m.ID, m.Url, p2putil.ShortForm(peer.ID(m.PeerID)))
}

func (m *Member) HasDuplicatedAttr(x *Member) bool {
	if m.Name == x.Name || m.ID == x.ID || m.Url == x.Url || bytes.Equal(m.PeerID, x.PeerID) {
		return true
	}

	return false
}

/*
func (m *Member) MarshalJSON() ([]byte, error) {
	nj := NewJsonMember(m)
	return json.Marshal(nj)
}

func (m *Member) UnmarshalJSON(data []byte) error {
	var err error
	jm := JsonMember{}

	if err := json.Unmarshal(data, &jm); err != nil {
		return err
	}

	*m, err = jm.Member()
	if err != nil {
		return err
	}

	return nil
}
type JsonMember struct {
	ID     MemberID `json:"id"`
	Name   string   `json:"name"`
	Url    string   `json:"url"`
	PeerID string   `json:"peerid"`
}

func NewJsonMember(m *Member) JsonMember {
	return JsonMember{ID: m.ID, Name: m.Name, Url: m.Url, PeerID: peer.IDB58Encode(m.PeerID)}
}

func (jm *JsonMember) Member() (Member, error) {
	peerID, err := peer.IDB58Decode(jm.PeerID)
	if err != nil {
		return Member{}, err
	}

	return Member{
		ID:     jm.ID,
		Name:   jm.Name,
		Url:    jm.Url,
		PeerID: peerID,
	}, nil
}
*/

// IsCompatible checks if name, url and peerid of this member are the same with other member
func (m *Member) IsCompatible(other *Member) bool {
	return m.Name == other.Name && m.Url == other.Url && bytes.Equal(m.PeerID, other.PeerID)
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

func ParseToUrl(urlstr string) (*url.URL, error) {
	var urlObj *url.URL
	var err error

	if urlObj, err = url.Parse(urlstr); err != nil {
		return nil, err
	}

	if urlObj.Scheme != "http" && urlObj.Scheme != "https" {
		return nil, ErrURLInvalidScheme
	}

	if _, _, err := net.SplitHostPort(urlObj.Host); err != nil {
		return nil, ErrURLInvalidPort
	}

	return urlObj, nil
}
