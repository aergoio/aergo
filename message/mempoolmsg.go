/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package message

import (
	"github.com/aergoio/aergo/types"
)

// MemPoolSvc is exported name for MemPool service
const MemPoolSvc = "MemPoolSvc"

// MemPoolPut is interface of MemPool service for inserting transactions
type MemPoolPut struct {
	Tx *types.Tx
}

// MemPoolPutRsp defines struct of result for MemPoolPut
type MemPoolPutRsp struct {
	Err error
}

// MemPoolGet is interface of MemPool service for retrieving transactions
type MemPoolGet struct {
	MaxBlockBodySize uint32
}

// MemPoolGetRsp defines struct of result for MemPoolGet
type MemPoolGetRsp struct {
	Txs []*types.Tx
	Err error
}

// MemPoolExist is interface of MemPool service for retrieving transaction
// according to given hash
type MemPoolExist struct {
	Hash []byte
}

// MemPoolExistRsp defines struct of result for MemPoolExist
type MemPoolExistRsp struct {
	Tx *types.Tx
}

const MaxReqestHashes = 1000

type MemPoolExistEx struct {
	Hashes [][]byte
}
type MemPoolExistExRsp struct {
	Txs []*types.Tx
}

// MemPoolDel is interface of MemPool service for deleting transactions
// including given transactions
type MemPoolDel struct {
	Block *types.Block
}

// MemPoolDelRsp defines struct of result for MemPoolDel
type MemPoolDelRsp struct {
	Err error
}
