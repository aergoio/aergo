/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package message

import (
	"fmt"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"time"
)

const P2PSvc = "p2pSvc"

// errors which async responses of p2p actor, such as GetBlockChunksRsp, can contains,
var (
	RemotePeerFailError  = fmt.Errorf("remote peer return err")
	MissingHashError     = fmt.Errorf("some block hash not found")
	UnexpectedBlockError = fmt.Errorf("unexpected blocks response")
)

// PingMsg send types.Ping to each peer.
// The actor returns true if sending is successful.
type PingMsg struct {
	ToWhom peer.ID
}

// GetAddressesMsg send types.AddressesRequest to dest peer. the dest peer will send types.AddressesResponse.
// The actor returns true if sending is successful.
type GetAddressesMsg struct {
	ToWhom peer.ID
	Size   uint32
	Offset uint32
}

// NotifyNewBlock send types.NewBlockNotice to other peers. The receiving peer will send GetBlockHeadersRequest or GetBlockRequest if needed.
// The actor returns true if sending is successful.
type NotifyNewBlock struct {
	BlockNo uint64
	Block   *types.Block
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
	ToWhom peer.ID
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
	ToWhom peer.ID
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
	ToWhom peer.ID
	Hashes []BlockHash
}

type GetBlockChunks struct {
	GetBlockInfos
	TTL time.Duration
}

// GetMissingBlocks send types.GetMissingRequest to dest peer.
// The actor returns true if sending is successful.
// Not used (need to be async operation)
type GetMissingBlocks struct {
	ToWhom peer.ID
	Hashes []BlockHash
}

// BlockInfosResponse is data from other peer, as a response of types.GetBlockRequest
// p2p module will send this to chainservice actor.
type BlockInfosResponse struct {
	FromWhom peer.ID
	Blocks   []*types.Block
}

type GetBlockChunksRsp struct {
	ToWhom peer.ID
	Blocks []*types.Block
	Err    error
}

// GetPeers requests p2p actor to get remote peers that is connected.
// The actor returns *GetPeersRsp
type GetPeers struct {
}

// GetPeersRsp contains peer meta information and current states.
type GetPeersRsp struct {
	Peers []*types.PeerAddress
	// last received block notice
	LastBlks []*types.NewBlockNotice
	States   []types.PeerState
}

// GetSyncAncestor is sent from Syncer, send types.GetAncestorRequest to dest peer.
type GetSyncAncestor struct {
	ToWhom peer.ID
	Hashes [][]byte
}

// GetSyncAncestorRsp is data from other peer, as a response of types.GetAncestorRequest
type GetSyncAncestorRsp struct {
	Ancestor *types.BlockInfo
}

type GetHashes struct {
	ToWhom   peer.ID
	PrevInfo *types.BlockInfo
	Count    uint64
}

type GetHashesRsp struct {
	PrevInfo *types.BlockInfo
	Hashes   []BlockHash
	Count    uint64
	Err      error
}
