/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package consensus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/etcd/raft"
	"github.com/aergoio/etcd/raft/raftpb"
)

// DefaultBlockIntervalSec  is the default block generation interval in seconds.
const DefaultBlockIntervalSec = int64(1)

var (
	// BlockIntervalSec is the block generation interval in seconds.
	BlockIntervalSec = DefaultBlockIntervalSec

	// BlockInterval is the maximum block generation time limit.
	BlockInterval = time.Second * time.Duration(DefaultBlockIntervalSec)

	logger = log.NewLogger("consensus")
)

var (
	ErrNotSupportedMethod = errors.New("not supported method in this consensus")
)

// InitBlockInterval initializes block interval parameters.
func InitBlockInterval(blockIntervalSec int64) {
	if blockIntervalSec > 0 {
		BlockIntervalSec = blockIntervalSec
		BlockInterval = time.Second * time.Duration(BlockIntervalSec)
	}
}

// ErrorConsensus is a basic error struct for consensus modules.
type ErrorConsensus struct {
	Msg string
	Err error
}

func (e ErrorConsensus) Error() string {
	errMsg := e.Msg
	if e.Err != nil {
		errMsg = fmt.Sprintf("%s: %s", errMsg, e.Err.Error())
	}
	return errMsg
}

// Constructor represents a function returning the Consensus interface for
// each implementation.
type Constructor func() (Consensus, error)

// Consensus is an interface for a consensus implementation.
type Consensus interface {
	ChainConsensus
	ConsensusAccessor
	Ticker() *time.Ticker
	QueueJob(now time.Time, jq chan<- interface{})
	BlockFactory() BlockFactory
	QuitChan() chan interface{}
}

type ConsensusAccessor interface {
	ConsensusInfo() *types.ConsensusInfo
	ClusterInfo([]byte) *types.GetClusterInfoResponse
	ConfChange(req *types.MembershipChange) (*Member, error)
	ConfChangeInfo(requestID uint64) (*types.ConfChangeProgress, error)
	// RaftAccessor returns AergoRaftAccessor. It is only valid if chain is raft consensus
	RaftAccessor() AergoRaftAccessor
}

// ChainDB is a reader interface for the ChainDB.
type ChainDB interface {
	GetBestBlock() (*types.Block, error)
	GetBlockByNo(blockNo types.BlockNo) (*types.Block, error)
	GetHashByNo(blockNo types.BlockNo) ([]byte, error)
	GetBlock(hash []byte) (*types.Block, error)
	GetGenesisInfo() *types.Genesis
	Get(key []byte) []byte
	NewTx() db.Transaction
}

// AergoRaftAccessor is interface to access raft messaging. It is wrapping raft message with aergo internal types
type AergoRaftAccessor interface {
	Process(ctx context.Context, peerID types.PeerID, m raftpb.Message) error
	IsIDRemoved(peerID types.PeerID) bool
	ReportUnreachable(peerID types.PeerID)
	ReportSnapshot(peerID types.PeerID, status raft.SnapshotStatus)

	SaveFromRemote(r io.Reader, id uint64, msg raftpb.Message) (int64, error)

	GetMemberByID(id uint64) *Member
	GetMemberByPeerID(peerID types.PeerID) *Member
}

type ConsensusType int

const (
	ConsensusDPOS ConsensusType = iota
	ConsensusRAFT
	ConsensusSBP
)

var ConsensusName = []string{"dpos", "raft", "sbp"}
var ConsensusTypes = map[string]ConsensusType{"dpos": ConsensusDPOS, "raft": ConsensusRAFT, "sbp": ConsensusSBP}

var CurConsensusType ConsensusType

func IsRaftName(consensus string) bool {
	return ConsensusName[ConsensusRAFT] == strings.ToLower(consensus)
}
func IsDposName(consensus string) bool {
	return ConsensusName[ConsensusDPOS] == strings.ToLower(consensus)
}

func SetCurConsensus(consensus string) {
	CurConsensusType = ConsensusTypes[consensus]
}

func UseRaft() bool {
	return CurConsensusType == ConsensusRAFT
}

func UseDpos() bool {
	return CurConsensusType == ConsensusDPOS
}

// ChainConsensus includes chain status and validation API.
type ChainConsensus interface {
	ChainConsensusCluster

	GetType() ConsensusType
	IsTransactionValid(tx *types.Tx) bool
	VerifyTimestamp(block *types.Block) bool
	VerifySign(block *types.Block) error
	IsBlockValid(block *types.Block, bestBlock *types.Block) error
	Update(block *types.Block)
	Save(tx TxWriter) error
	NeedReorganization(rootNo types.BlockNo) bool
	NeedNotify() bool
	HasWAL() bool // if consensus has WAL, block has already written in db
	IsConnectedBlock(block *types.Block) bool
	IsForkEnable() bool
	Info() string
}

type ChainConsensusCluster interface {
	MakeConfChangeProposal(req *types.MembershipChange) (*ConfChangePropose, error)
}

type TxWriter interface {
	Set(key, value []byte)
}

// Info represents an information for a consensus implementation.
type Info struct {
	Type   string
	Status *json.RawMessage `json:",omitempty"`
}

// NewInfo returns a new Info with name.
func NewInfo(name string) *Info {
	return &Info{Type: name}
}

// AsJSON returns i as a JSON string
func (i *Info) AsJSON() string {
	if m, err := json.Marshal(i); err == nil {
		return string(m)
	}
	return ""
}

// BlockFactory is an interface for a block factory implementation.
type BlockFactory interface {
	Start()
	JobQueue() chan<- interface{}
}

// Start run a selected consensus service.
func Start(c Consensus) {
	bf := c.BlockFactory()
	if c == nil || bf == nil {
		logger.Fatal().Msg("failed to start consensus service: no Consensus or BlockFactory")
	}

	go bf.Start()

	go func() {
		ticker := c.Ticker()
		for now := range ticker.C {
			c.QueueJob(now, bf.JobQueue())
			select {
			case <-c.QuitChan():
				logger.Info().Msg("shutdown initiated. stop the consensus service")
				return
			default:
			}
		}
	}()
}

// Stop shutdown consensus service.
func Stop(c Consensus) {
	close(c.QuitChan())
}
