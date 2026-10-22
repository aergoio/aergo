/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package message

import (
	"github.com/aergoio/aergo/v2/types"
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
	Txs []types.Transaction
	Err error
}

// MemPoolList is  interface of MemPool service for retrieving hashes of transactions
type MemPoolList struct {
	Limit int
}

type MemPoolListRsp struct {
	Hashes  []types.TxID
	HasMore bool
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

// MemPoolExistEx is for getting retrieving multiple transactions.
type MemPoolExistEx struct {
	Hashes [][]byte
}

// MemPoolExistExRsp can contains nil element if requested tx is missing in mempool.
type MemPoolExistExRsp struct {
	Txs []*types.Tx
}

type MemPoolSetWhitelist struct {
	Accounts []string
}

type MemPoolEnableWhitelist struct {
	On bool
}

// MemPoolDel is interface of MemPool service for deleting transactions
// including given transactions
type MemPoolDel struct {
	Block *types.Block
}

type MemPoolTxStat struct {
}

type MemPoolTxStatRsp struct {
	Data []byte
}

type MemPoolTx struct {
	Accounts []types.Address
}

type MemPoolTxRsp MemPoolTxStatRsp

// MemPoolDelRsp defines struct of result for MemPoolDel
type MemPoolDelRsp struct {
	Err error
}

// MemPoolDelTx is interface of MemPool service for deleting a transaction
type MemPoolDelTx struct {
	Tx *types.Tx
}

// MemPoolDelTxRsp defines struct of result for MemPoolDelTx
type MemPoolDelTxRsp struct {
	Err error
}
