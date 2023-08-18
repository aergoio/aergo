package exec

import (
	"fmt"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/types"
)

func init() {
	registerExec(&undoCommit{})
}

type undoCommit struct{}

func (c *undoCommit) Command() string {
	return "undo"
}

func (c *undoCommit) Syntax() string {
	return fmt.Sprintf("")
}

func (c *undoCommit) Usage() string {
	return "undo"
}

func (c *undoCommit) Describe() string {
	return "undo the previous transaction by disconnecting a block, which contains the tx"
}

func (c *undoCommit) Validate(args string) error {

	if context.Get().BestBlockNo() == 0 {
		return fmt.Errorf("There are no txs to undo")
	}
	return nil
}

func (c *undoCommit) Run(args string) (string, uint64, []*types.Event, error) {
	err := context.Get().DisConnectBlock()
	if err != nil {
		return "", 0, nil, err
	}
	return "Undo, Succesfully", 0, nil, nil
}
