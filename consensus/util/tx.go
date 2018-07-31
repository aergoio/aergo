package util

import (
	"errors"

	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

var (
	// ErrQuit indicates that shutdown is initiated.
	ErrQuit = errors.New("shutdown initiated")
)

// TxDo is the type of arguments for CompositeTxDo.
type TxDo func(tx *types.Tx) error

// NewTxDo returns a function which applies each function in fn.x
func NewTxDo(fn ...TxDo) TxDo {
	return func(tx *types.Tx) error {
		for _, f := range fn {
			if err := f(tx); err != nil {
				return err
			}
		}

		return nil
	}
}

// GatherTXs returns transactions from txIn. The selection is done by applying
// txDo.
func GatherTXs(txIn []*types.Tx, txDo TxDo, maxBlockBodySize int) ([]*types.Tx, error) {
	if len(txIn) == 0 {
		return txIn, nil
	}

	end := 0
	size := 0
	for i, tx := range txIn {
		size += proto.Size(tx)
		if size > maxBlockBodySize {
			break
		}

		err := txDo(tx)
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
