package chain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestGatherTXs(t *testing.T) {
	txOp := NewCompTxOp(
		TxOpFn(func(bState *state.BlockState, tx types.Transaction) error {
			fmt.Println("x")
			return nil
		}),
		TxOpFn(func(bState *state.BlockState, tx types.Transaction) error {
			fmt.Println("y")
			return nil
		}))
	err := txOp.Apply(nil, nil)
	assert.New(t).Nil(err)
}

func TestGatherTXsWithError(t *testing.T) {
	txDo := NewCompTxOp(
		TxOpFn(func(bState *state.BlockState, tx types.Transaction) error {
			fmt.Println("haha")
			return nil
		}),
		TxOpFn(func(bState *state.BlockState, tx types.Transaction) error {
			fmt.Println("blah")
			return errors.New("blah blah error")
		}))
	err := txDo.Apply(nil, nil)
	assert.New(t).NotNil(err)
}
