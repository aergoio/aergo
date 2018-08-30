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
		func(tx *types.Tx) error {
			fmt.Println("x")
			return nil
		},
		func(tx *types.Tx) error {
			fmt.Println("y")
			return nil
		})
	err := txOp.Apply(nil)
	assert.New(t).Nil(err)
}

func TestGatherTXsWithError(t *testing.T) {
	txDo := NewCompTxOp(
		func(tx *types.Tx) error {
			fmt.Println("haha")
			return nil
		},
		func(tx *types.Tx) error {
			fmt.Println("blah")
			return errors.New("blah blah error")
		})
	err := txDo.Apply(nil)
	assert.New(t).NotNil(err)
}
