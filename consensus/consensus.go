/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package consensus

import (
	"fmt"
	"time"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const (
	// DefaultBlockIntervalSec  is the default block generation interval in seconds.
	DefaultBlockIntervalSec = int64(1) // block production interval in sec

	// DefaultDposBpNumber is the default number of block producers.
	DefaultDposBpNumber = 23
)

var (
	// BlockIntervalSec is the block genration interval in seconds.
	BlockIntervalSec = DefaultBlockIntervalSec

	// BlockInterval is the maximum block generation time limit.
	BlockInterval = time.Second * time.Duration(DefaultBlockIntervalSec)

	logger = log.NewLogger("consensus")
)

// InitBlockInterval initializes block interval parameters.
func InitBlockInterval(blockIntervalSec int64) {
	if blockIntervalSec > 0 {
		BlockIntervalSec = blockIntervalSec
		BlockInterval = time.Second * time.Duration(BlockIntervalSec)
	}
}

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

// Constructor represents a function returning the Consensus interfactor for
// each implementation.
type Constructor func() (Consensus, error)

// Consensus is an interface for a consensus implementation.
type Consensus interface {
	ChainConsensus
	Ticker() *time.Ticker
	QueueJob(now time.Time, jq chan<- interface{})
	BlockFactory() BlockFactory
	QuitChan() chan interface{}
	SetChainAccessor(chainAccessor types.ChainAccessor)
}

// ChainDbReader is a reader interface for the ChainDB.
type ChainDbReader interface {
	GetBestBlock() (*types.Block, error)
	GetBlockByNo(blockNo types.BlockNo) (*types.Block, error)
	GetGenesisInfo() *types.Genesis
	Get(key []byte) []byte
}

// ChainConsensus includes chainstatus and validation API.
type ChainConsensus interface {
	SetStateDB(sdb *state.ChainStateDB)
	IsTransactionValid(tx *types.Tx) bool
	IsBlockValid(block *types.Block, bestBlock *types.Block) error
	Update(block *types.Block)
	Save(tx db.Transaction) error
	NeedReorganization(rootNo types.BlockNo) bool
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
			c.QueueJob(now, bf.JobQueue())
			select {
			case <-c.QuitChan():
				logger.Info().Msg("shutdown initiated. stop the consensus service")
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
