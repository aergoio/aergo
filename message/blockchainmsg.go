/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package message

import (
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
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
type GetMissing struct {
	Hashes   [][]byte
	StopHash []byte
}
type GetMissingRsp struct {
	TopMatched BlockHash
	TopNumber  types.BlockNo
	StopNumber types.BlockNo
	//Hashes   []BlockHash
	//Blocknos []types.BlockNo
}

type GetBlockByNo struct {
	BlockNo types.BlockNo
}
type GetBlockByNoRsp GetBlockRsp

type AddBlock struct {
	PeerID peer.ID
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
	State *types.State
	Err   error
}
type GetStateAndProof struct {
	Account    []byte
	Root       []byte
	Compressed bool
}
type GetStateAndProofRsp struct {
	StateProof *types.StateProof
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
	VarName         string
	VarIndex        string
	Root            []byte
	Compressed      bool
}
type GetStateQueryRsp struct {
	Result *types.StateQueryProof
	Err    error
}

// SyncBlockState is request to sync from remote peer. It returns sync result.
type SyncBlockState struct {
	PeerID    peer.ID
	BlockNo   types.BlockNo
	BlockHash []byte
}

// GetElected is request to get voting result about top N elect
type GetElected struct {
	N int
}

type GetVote struct {
	Addr []byte
}

// GetElectedRsp is return to get voting result
type GetVoteRsp struct {
	Top *types.VoteList
	Err error
}

type GetStaking struct {
	Addr []byte
}

type GetStakingRsp struct {
	Staking *types.Staking
	Err     error
}

type GetAnchors struct{}
type GetAnchorsRsp struct {
	Hashes [][]byte
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
