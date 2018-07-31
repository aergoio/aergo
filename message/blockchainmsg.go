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
	Hashes   []BlockHash
	Blocknos []types.BlockNo
}

type GetBlockByNo struct {
	BlockNo types.BlockNo
}
type GetBlockByNoRsp GetBlockRsp

type AddBlock struct {
	PeerID peer.ID
	Block  *types.Block
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
type GetTx struct {
	TxHash []byte
}
type GetTxRsp struct {
	Tx  *types.Tx
	Err error
}
