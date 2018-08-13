/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package consensus

import (
	"fmt"
	"time"

	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
)

var (
	logger = log.NewLogger("consensus")
)

// ErrorConsensus is a basic error struct for consensus modules.
type ErrorConsensus struct {
	Msg string
	Err error
}

func (e ErrorConsensus) Error() string {
	errMsg := e.Msg
	if e.Err != nil {
		errMsg = fmt.Sprintf("%s: %s", errMsg, e.Err.Error())
	}
	return errMsg
}

// Consensus is an interface for a consensus implementation.
type Consensus interface {
	ChainInfo
	Ticker() *time.Ticker
	QueueJob(now time.Time, jq chan<- interface{})
	BlockFactory() BlockFactory
	QuitChan() chan interface{}
}

// ChainInfo includes chainstatus and validation API.
type ChainInfo interface {
	IsTransactionValid(tx *types.Tx) bool
	IsBlockValid(block *types.Block) error
	IsBlockReorganizing() bool
	SetReorganizing()
	UnsetReorganizing()
}

// BlockFactory is an interface for a block factory implementation.
type BlockFactory interface {
	Start()
	JobQueue() chan<- interface{}
}

// Start run a selected consesus service.
func Start(c Consensus) {
	bf := c.BlockFactory()
	if c == nil || bf == nil {
		logger.Fatal().Msg("failed to start consensus service: no Consensus or BlockFactory")
	}

	go bf.Start()

	go func() {
		ticker := c.Ticker()
		for now := range ticker.C {
			if c.IsBlockReorganizing() {
				continue
			}

			c.QueueJob(now, bf.JobQueue())
			select {
			case <-c.QuitChan():
				return
			default:
			}
		}
	}()
}

// Stop shutdown consensus service.
func Stop(c Consensus) {
	close(c.QuitChan())
}
