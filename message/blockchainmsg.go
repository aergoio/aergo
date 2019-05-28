/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package message

import (
	"github.com/aergoio/aergo/types"
)

const ChainSvc = "ChainSvc"

type GetBestBlockNo struct{}
type GetBestBlockNoRsp struct {
	BlockNo types.BlockNo
}

type GetBestBlock struct{}
type GetBestBlockRsp GetBlockRsp

type GetBlock struct {
	BlockHash []byte
}
type GetBlockRsp struct {
	Block *types.Block
	Err   error
}

type GetBlockByNo struct {
	BlockNo types.BlockNo
}
type GetBlockByNoRsp GetBlockRsp

type AddBlock struct {
	PeerID types.PeerID
	Block  *types.Block
	Bstate interface{}
	IsSync bool
	// Bstate *types.BlockState
}
type AddBlockRsp struct {
	BlockNo   types.BlockNo
	BlockHash []byte
	Err       error
}
type GetState struct {
	Account []byte
}
type GetStateRsp struct {
	Account []byte
	State   *types.State
	Err     error
}
type GetStateAndProof struct {
	Account    []byte
	Root       []byte
	Compressed bool
}
type GetStateAndProofRsp struct {
	StateProof *types.AccountProof
	Err        error
}
type GetTx struct {
	TxHash []byte
}
type GetTxRsp struct {
	Tx    *types.Tx
	TxIds *types.TxIdx
	Err   error
}

type GetReceipt struct {
	TxHash []byte
}
type GetReceiptRsp struct {
	Receipt *types.Receipt
	Err     error
}

type GetABI struct {
	Contract []byte
}
type GetABIRsp struct {
	ABI *types.ABI
	Err error
}

type GetQuery struct {
	Contract  []byte
	Queryinfo []byte
}
type GetQueryRsp struct {
	Result []byte
	Err    error
}
type GetStateQuery struct {
	ContractAddress []byte
	StorageKeys     []string
	Root            []byte
	Compressed      bool
}
type GetStateQueryRsp struct {
	Result *types.StateQueryProof
	Err    error
}

// SyncBlockState is request to sync from remote peer. It returns sync result.
type SyncBlockState struct {
	PeerID    types.PeerID
	BlockNo   types.BlockNo
	BlockHash []byte
}

// GetElected is request to get voting result about top N elect
type GetElected struct {
	Id string
	N  uint32
}

type GetVote struct {
	Addr []byte
	Ids  []string
}

// GetElectedRsp is return to get voting result
type GetVoteRsp struct {
	Top *types.VoteList
	Err error
}
type GetAccountVoteRsp struct {
	Info *types.AccountVoteInfo
	Err  error
}

type GetStaking struct {
	Addr []byte
}

type GetStakingRsp struct {
	Staking *types.Staking
	Err     error
}

type GetNameInfo struct {
	Name    string
	BlockNo types.BlockNo
}

type GetNameInfoRsp struct {
	Owner *types.NameInfo
	Err   error
}

type GetAnchors struct {
	Seq uint64
}

type GetAnchorsRsp struct {
	Seq    uint64
	Hashes [][]byte
	LastNo types.BlockNo
	Err    error
}

// receive from p2p
type GetAncestor struct {
	Hashes   [][]byte
	StopHash []byte
}

// response to p2p for GetAncestor message
type GetAncestorRsp struct {
	Ancestor *types.BlockInfo
	Err      error
}

type ListEvents struct {
	Filter *types.FilterInfo
}

// response to p2p for GetAncestor message
type ListEventsRsp struct {
	Events []*types.Event
	Err    error
}
