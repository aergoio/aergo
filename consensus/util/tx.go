package util

import (
	"errors"

	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

var (
	// ErrQuit indicates that shutdown is initiated.
	ErrQuit = errors.New("shutdown initiated")

	errBlockSizeLimit = errors.New("the transactions included exceeded the block size limit")
)

// TxOp is an interface used by GatherTXs for apply some transaction related operation.
type TxOp interface {
	Apply(tx *types.Tx) error
}

// TxOpFn is the type of arguments for CompositeTxDo.
type TxOpFn func(tx *types.Tx) error

// Apply applies f to tx.
func (f TxOpFn) Apply(tx *types.Tx) error {
	return f(tx)
}

// NewCompTxOp returns a function which applies each function in fn.
func NewCompTxOp(fn ...TxOpFn) TxOp {
	return TxOpFn(func(tx *types.Tx) error {
		for _, f := range fn {
			if err := f.Apply(tx); err != nil {
				return err
			}
		}

		return nil
	})
}

// NewBlockLimitOp returns a TxOpFn which returns errBlockSizeLimit when the
// size of the collected transactions exceeds the maximum block size.
func NewBlockLimitOp(maxBlockBodySize int) TxOpFn {
	size := 0
	return TxOpFn(func(tx *types.Tx) error {
		if size += proto.Size(tx); size > maxBlockBodySize {
			return errBlockSizeLimit
		}
		return nil
	})
}

// GatherTXs returns transactions from txIn. The selection is done by applying
// txDo.
func GatherTXs(hs component.ICompSyncRequester, txOp TxOp) ([]*types.Tx, error) {
	txIn := FetchTXs(hs)
	if len(txIn) == 0 {
		return txIn, nil
	}

	end := 0
	for i, tx := range txIn {
		err := txOp.Apply(tx)
		if err == ErrQuit {
			return nil, err
		} else if err != nil {
			// TODO: Currently, there's only break condition. Later skip
			// conditions may be needed.
			break
		}
		end = i
	}

	return txIn[0 : end+1], nil
}
