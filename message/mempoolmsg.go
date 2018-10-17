/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package message

import (
	"errors"

	"github.com/aergoio/aergo/types"
)

var (
	//ErrTxNotFound is returned by MemPool Service if transaction does not exists
	ErrTxNotFound = errors.New("tx not found in mempool")

	//ErrTxHasInvalidHash is returned by MemPool Service if transaction does have invalid hash
	ErrTxHasInvalidHash = errors.New("tx has invalid hash")

	//ErrTxAlreadyInMempool is returned by MemPool Service if transaction which already exists
	ErrTxAlreadyInMempool = errors.New("tx already in mempool")

	//ErrTxFormatInvalid is returned by MemPool Service if transaction does not exists
	ErrTxFormatInvalid = errors.New("tx invalid format")

	//ErrInsufficientBalance is returned by MemPool Service if account has not enough balance
	ErrInsufficientBalance = errors.New("not enough balance")

	//ErrTxNonceTooLow is returned by MemPool Service if transaction's nonce is already existed in block
	ErrTxNonceTooLow = errors.New("nonce is too low")

	//ErrTxNonceToohigh is for internal use only
	ErrTxNonceToohigh = errors.New("nonce is too high")
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

// MemPoolDel is interface of MemPool service for deleting transactions
// including given transactions
type MemPoolDel struct {
	Block *types.Block
}

// MemPoolDelRsp defines struct of result for MemPoolDel
type MemPoolDelRsp struct {
	Err error
}
