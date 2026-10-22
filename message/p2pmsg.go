/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package message

import (
	"fmt"
	"time"

	"github.com/aergoio/aergo/v2/types"
)

const P2PSvc = "p2pSvc"

// errors which async responses of p2p actor, such as GetBlockChunksRsp, can contains,
var (
	RemotePeerFailError  = fmt.Errorf("remote peer return error")
	PeerNotFoundError    = fmt.Errorf("remote peer was not found")
	MissingHashError     = fmt.Errorf("some block hash not found")
	UnexpectedBlockError = fmt.Errorf("unexpected blocks response")
	TooFewBlocksError    = fmt.Errorf("too few blocks received that expected")
	TooManyBlocksError   = fmt.Errorf("too many blocks received that expected")
	TooBigBlockError     = fmt.Errorf("block size limit exceeded")
	InvalidArgumentError = fmt.Errorf("invalid argument")
	WrongBlockHashError  = fmt.Errorf("wrong block hash")
)

// PingMsg send types.Ping to each peer.
// The actor returns true if sending is successful.
type PingMsg struct {
	ToWhom types.PeerID
}

// GetAddressesMsg send types.AddressesRequest to dest peer. the dest peer will send types.AddressesResponse.
// The actor returns true if sending is successful.
type GetAddressesMsg struct {
	ToWhom types.PeerID
	Size   uint32
	Offset uint32
}

// NotifyNewBlock send types.NewBlockNotice to other peers. The receiving peer will send GetBlockHeadersRequest or GetBlockRequest if needed.
// The actor returns true if sending is successful.
type NotifyNewBlock struct {
	Produced bool
	BlockNo  uint64
	Block    *types.Block
}

type BlockHash []byte
type TXHash []byte

// NotifyNewTransactions send types.NewTransactionsNotice to other peers.
// The actor returns true if sending is successful.
type NotifyNewTransactions struct {
	Txs []*types.Tx
}

// GetTransactions send types.GetTransactionsRequest to dest peer. The receiving peer will send types.GetTransactionsResponse
// The actor returns true if sending is successful.
type GetTransactions struct {
	ToWhom types.PeerID
	Hashes []TXHash
}

// TransactionsResponse is data from other peer, as a response of types.GetTransactionsRequest
// p2p module will send this to mempool actor.
type TransactionsResponse struct {
	txs []*types.Tx
}

// GetBlockHeaders send type.GetBlockRequest to dest peer
// The actor returns true if sending is successful.
type GetBlockHeaders struct {
	ToWhom types.PeerID
	// Hash is the first block to get. Height will be used when Hash mi empty
	Hash    BlockHash
	Height  uint64
	Asc     bool
	Offset  uint64
	MaxSize uint32
}

// BlockHeadersResponse is data from other peer, as a response of types.GetBlockRequest
// p2p module will send this to chainservice actor.
type BlockHeadersResponse struct {
	Hashes  []BlockHash
	Headers []*types.BlockHeader
}

// GetBlockInfos send types.GetBlockRequest to dest peer.
// The actor returns true if sending is successful.
type GetBlockInfos struct {
	ToWhom types.PeerID
	Hashes []BlockHash
}

type GetBlockChunks struct {
	Seq uint64
	GetBlockInfos
	TTL time.Duration
}

// BlockInfosResponse is data from other peer, as a response of types.GetBlockRequest
// p2p module will send this to chainservice actor.
type BlockInfosResponse struct {
	FromWhom types.PeerID
	Blocks   []*types.Block
}

type GetBlockChunksRsp struct {
	Seq    uint64
	ToWhom types.PeerID
	Blocks []*types.Block
	Err    error
}

// GetPeers requests p2p actor to get remote peers that is connected.
// The actor returns *GetPeersRsp
type GetPeers struct {
	NoHidden bool
	ShowSelf bool
}

type PeerInfo struct {
	Addr            *types.PeerAddress
	Certificates    []*types.AgentCertificate
	AcceptedRole    types.PeerRole
	Version         string
	Hidden          bool
	CheckTime       time.Time
	LastBlockHash   []byte
	LastBlockNumber uint64
	State           types.PeerState
	Self            bool
}

// GetPeersRsp contains peer meta information and current states.
type GetPeersRsp struct {
	Peers []*PeerInfo
}

type GetMetrics struct {
}

// GetSyncAncestor is sent from Syncer, send types.GetAncestorRequest to dest peer.
type GetSyncAncestor struct {
	Seq    uint64
	ToWhom types.PeerID
	Hashes [][]byte
}

// GetSyncAncestorRsp is data from other peer, as a response of types.GetAncestorRequest
type GetSyncAncestorRsp struct {
	Seq      uint64
	Ancestor *types.BlockInfo
}

type GetHashes struct {
	Seq      uint64
	ToWhom   types.PeerID
	PrevInfo *types.BlockInfo
	Count    uint64
}

type GetHashesRsp struct {
	Seq      uint64
	PrevInfo *types.BlockInfo
	Hashes   []BlockHash
	Count    uint64
	Err      error
}

type GetHashByNo struct {
	Seq     uint64
	ToWhom  types.PeerID
	BlockNo types.BlockNo
}

type GetHashByNoRsp struct {
	Seq       uint64
	BlockHash BlockHash
	Err       error
}

type GetSelf struct {
}

type GetCluster struct {
	BestBlockHash BlockHash
	ReplyC        chan *GetClusterRsp
}

type GetClusterRsp struct {
	ClusterID     uint64
	ChainID       BlockHash
	Members       []*types.MemberAttr
	Err           error
	HardStateInfo *types.HardStateInfo
}

type GetRaftTransport struct {
	Cluster interface{}
}

type RaftClusterEvent struct {
	BPAdded   []types.PeerID
	BPRemoved []types.PeerID
}

// ChangeDesignatedPeers will trigger connect or disconnect peers
type ChangeDesignatedPeers struct {
	Add    []types.PeerAddress
	Remove []types.PeerID
}

type SendRaft struct {
	ToWhom types.PeerID
	Body   interface{} // for avoiding dependency cycle, though it must be raftpb.Message.
}

type SendRaftRsp struct {
	Err error
}

type P2PWhiteListConfEnableEvent struct {
	Name string
	On   bool
}

type P2PWhiteListConfSetEvent struct {
	Name   string
	Values []string
}

type IssueAgentCertificate struct {
	ProducerID types.PeerID
}

type NotifyCertRenewed struct {
	Cert *types.AgentCertificate
}

type TossDirection bool

type TossBPNotice struct {
	Block *types.Block
	// toss notice to internal zone or not
	TossIn bool
	// OriginalMsg is actually p2pcommon.Message. it is declared by interface{} for ad-hoc way to avoid import cycle
	OriginalMsg interface{}
}
