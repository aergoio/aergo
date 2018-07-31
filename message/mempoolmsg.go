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
	ErrTxNotFound          = errors.New("tx not found in mempool")
	ErrTxHasInvalidHash    = errors.New("tx has invalid hash")
	ErrTxAlreadyInMempool  = errors.New("tx already in mempool")
	ErrTxFormatInvalid     = errors.New("tx invalid format")
	ErrInsufficientBalance = errors.New("not enough balance")
	ErrTxNonceTooLow       = errors.New("nonce is too low")
	ErrTxNonceToohigh      = errors.New("nonce is too high")
)

const MemPoolSvc = "MemPoolSvc"

type MemPoolGenerateSampleTxs struct {
	MaxCount int
}
type MemPoolPut struct {
	Txs []*types.Tx
}
type MemPoolPutRsp struct {
	Err []error
}
type MemPoolGet struct {
}
type MemPoolGetRsp struct {
	Txs []*types.Tx
	Err error
}
type MemPoolExist struct {
	Hash []byte
}
type MemPoolExistRsp struct {
	Tx *types.Tx
}
type MemPoolDel struct {
	BlockNo uint64
	Txs     []*types.Tx
}
type MemPoolDelRsp struct {
	Err error
}
