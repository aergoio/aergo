package sbp

import (
	"time"

	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/impl/dpos/param"
	"github.com/aergoio/aergo/consensus/util"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

const (
	blockMax     = 1 << 20
	slotQueueMax = 100
)

var logger *log.Logger

func init() {
	logger = log.NewLogger(log.SBP)
}

// SimpleBlockFactory implments a simple block factory which generate block each cfg.Consensus.BlockInterval.
//
// This can be used for testing purpose.
type SimpleBlockFactory struct {
	*component.ComponentHub
	jobQueue         chan interface{}
	blockInterval    int64
	maxBlockBodySize int
	onReorganizing   util.BcReorgStatus
}

// New returns a SimpleBlockFactory.
func New(cfg *config.Config, hub *component.ComponentHub) (*SimpleBlockFactory, error) {
	return &SimpleBlockFactory{
		ComponentHub:     hub,
		jobQueue:         make(chan interface{}, slotQueueMax),
		blockInterval:    cfg.Consensus.BlockInterval,
		maxBlockBodySize: blockMax,
		onReorganizing:   util.BcNoReorganizing,
	}, nil
}

// Ticker returns a time.Ticker for the main consensus loop.
func (s *SimpleBlockFactory) Ticker() *time.Ticker {
	return time.NewTicker(param.BlockInterval)
}

// QueueJob send a block triggering infomation to jq.
func (s *SimpleBlockFactory) QueueJob(now time.Time, jq chan<- interface{}) {
	if b := util.GetBestBlock(s); b != nil {
		jq <- b
	}
}

// IsTransactionValid checks the onsensus level validity of a transaction
func (s *SimpleBlockFactory) IsTransactionValid(tx *types.Tx) bool {
	// SimpleBlockFactory has no tx valid check.
	return true
}

// IsBlockValid checks the consensus level validity of a block.
func (s *SimpleBlockFactory) IsBlockValid(block *types.Block) error {
	// SimpleBlockFactory has no block valid check.
	return nil
}

// IsBlockReorganizing reports whether the blockchain is currently under
// reorganization.
func (s *SimpleBlockFactory) IsBlockReorganizing() bool {
	return util.OnReorganizing(&s.onReorganizing)
}

// SetReorganizing sets dpos.onReorganizing to 'OnReorganization.'
func (s *SimpleBlockFactory) SetReorganizing() {
	util.SetReorganizing(&s.onReorganizing)
}

// SetReorganizing sets dpos.onReorganizing to 'NoReorganization.'
func (s *SimpleBlockFactory) UnsetReorganizing() {
	util.UnsetReorganizing(&s.onReorganizing)
}

// BlockFactory returns s itself.
func (s *SimpleBlockFactory) BlockFactory() consensus.BlockFactory {
	return s
}

// Start run a simple block factory service.
func (s *SimpleBlockFactory) Start(quitC <-chan interface{}) {
	gatherTXs := func() []*types.Tx {
		txs := util.FetchTXs(s)
		if len(txs) == 0 {
			return txs
		}

		end := 0
		size := 0
		for i, tx := range txs {
			size += proto.Size(tx)
			if size > s.maxBlockBodySize {
				break
			}
			end = i
		}
		return txs[0 : end+1]
	}

	for {
		select {
		case e := <-s.jobQueue:
			if prevBlock, ok := e.(*types.Block); ok {
				block := types.NewBlock(prevBlock, gatherTXs())
				logger.Infof("block produced: no=%d, hash=%v",
					block.GetHeader().GetBlockNo(), block.ID())

				util.ConnectBlock(s, block)
			}
		case <-quitC:
			return
		}
	}
}

// JobQueue returns the queue for block production triggering.
func (s *SimpleBlockFactory) JobQueue() chan<- interface{} {
	return s.jobQueue
}
