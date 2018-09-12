package chain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func TestGatherTXs(t *testing.T) {
	txOp := NewCompTxOp(
		TxOpFn(func(tx *types.Tx) (*types.BlockState, error) {
			fmt.Println("x")
			return nil, nil
		}),
		TxOpFn(func(tx *types.Tx) (*types.BlockState, error) {
			fmt.Println("y")
			return nil, nil
		}))
	_, err := txOp.Apply(nil)
	assert.New(t).Nil(err)
}

func TestGatherTXsWithError(t *testing.T) {
	txDo := NewCompTxOp(
		TxOpFn(func(tx *types.Tx) (*types.BlockState, error) {
			fmt.Println("haha")
			return nil, nil
		}),
		TxOpFn(func(tx *types.Tx) (*types.BlockState, error) {
			fmt.Println("blah")
			return nil, errors.New("blah blah error")
		}))
	_, err := txDo.Apply(nil)
	assert.New(t).NotNil(err)
}
