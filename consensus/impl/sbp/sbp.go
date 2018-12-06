package sbp

import (
	"runtime"
	"time"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	bc "github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/chain"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const (
	slotQueueMax = 100
)

var logger *log.Logger

func init() {
	logger = log.NewLogger("sbp")
}

type txExec struct {
	execTx bc.TxExecFn
}

func newTxExec(blockNo types.BlockNo, ts int64) chain.TxOp {
	// Block hash not determined yet
	return &txExec{
		execTx: bc.NewTxExecutor(blockNo, ts, contract.BlockFactory),
	}
}

func (te *txExec) Apply(bState *state.BlockState, tx *types.Tx) error {
	err := te.execTx(bState, tx)
	return err
}

// SimpleBlockFactory implments a simple block factory which generate block each cfg.Consensus.BlockInterval.
//
// This can be used for testing purpose.
type SimpleBlockFactory struct {
	*component.ComponentHub
	jobQueue         chan interface{}
	blockInterval    time.Duration
	maxBlockBodySize uint32
	txOp             chain.TxOp
	quit             chan interface{}
	sdb              *state.ChainStateDB
	ca               types.ChainAccessor
	prevBlock        *types.Block
}

// GetConstructor build and returns consensus.Constructor from New function.
func GetConstructor(cfg *config.Config, hub *component.ComponentHub, cdb consensus.ChainDbReader) consensus.Constructor {
	return func() (consensus.Consensus, error) {
		return New(cfg, hub)
	}
}

// New returns a SimpleBlockFactory.
func New(cfg *config.Config, hub *component.ComponentHub) (*SimpleBlockFactory, error) {
	consensus.InitBlockInterval(cfg.Consensus.BlockInterval)

	s := &SimpleBlockFactory{
		ComponentHub:     hub,
		jobQueue:         make(chan interface{}, slotQueueMax),
		blockInterval:    consensus.BlockInterval,
		maxBlockBodySize: chain.MaxBlockBodySize(),
		quit:             make(chan interface{}),
	}

	s.txOp = chain.NewCompTxOp(
		chain.TxOpFn(func(bState *state.BlockState, txIn *types.Tx) error {
			select {
			case <-s.quit:
				return chain.ErrQuit
			default:
				return nil
			}
		}),
	)

	return s, nil
}

// Ticker returns a time.Ticker for the main consensus loop.
func (s *SimpleBlockFactory) Ticker() *time.Ticker {
	return time.NewTicker(s.blockInterval)
}

// QueueJob send a block triggering information to jq.
func (s *SimpleBlockFactory) QueueJob(now time.Time, jq chan<- interface{}) {
	if b, _ := s.ca.GetBestBlock(); b != nil {
		if s.prevBlock != nil && s.prevBlock.BlockNo() == b.BlockNo() {
			logger.Debug().Msg("previous block not connected. skip to generate block")
			return
		}
		s.prevBlock = b
		jq <- b
	}
}

// SetStateDB do nothing in the simple block factory, which do not execute
// transactions at all.
func (s *SimpleBlockFactory) SetStateDB(sdb *state.ChainStateDB) {
	s.sdb = sdb
}

// IsTransactionValid checks the onsensus level validity of a transaction
func (s *SimpleBlockFactory) IsTransactionValid(tx *types.Tx) bool {
	// SimpleBlockFactory has no tx valid check.
	return true
}

// IsBlockValid checks the consensus level validity of a block.
func (s *SimpleBlockFactory) IsBlockValid(*types.Block, *types.Block) error {
	// SimpleBlockFactory has no block valid check.
	return nil
}

// QuitChan returns the channel from which consensus-related goroutines check
// when shutdown is initiated.
func (s *SimpleBlockFactory) QuitChan() chan interface{} {
	return s.quit
}

// Update has nothging to do.
func (s *SimpleBlockFactory) Update(block *types.Block) {
}

// Save has nothging to do.
func (s *SimpleBlockFactory) Save(tx db.Transaction) error {
	return nil
}

// BlockFactory returns s itself.
func (s *SimpleBlockFactory) BlockFactory() consensus.BlockFactory {
	return s
}

func (s *SimpleBlockFactory) SetChainAccessor(chainAccessor types.ChainAccessor) {
	s.ca = chainAccessor
}

// NeedReorganization has nothing to do.
func (s *SimpleBlockFactory) NeedReorganization(rootNo types.BlockNo) bool {
	return true
}

// Start run a simple block factory service.
func (s *SimpleBlockFactory) Start() {
	defer logger.Info().Msg("shutdown initiated. stop the service")

	runtime.LockOSThread()

	for {
		select {
		case e := <-s.jobQueue:
			if prevBlock, ok := e.(*types.Block); ok {
				blockState := s.sdb.NewBlockState(prevBlock.GetHeader().GetBlocksRootHash())

				ts := time.Now().UnixNano()

				txOp := chain.NewCompTxOp(
					s.txOp,
					newTxExec(prevBlock.GetHeader().GetBlockNo()+1, ts),
				)

				block, err := chain.GenerateBlock(s, prevBlock, blockState, txOp, ts)
				if err == chain.ErrQuit {
					return
				} else if err != nil {
					logger.Info().Err(err).Msg("failed to produce block")
					continue
				}
				logger.Info().Uint64("no", block.GetHeader().GetBlockNo()).Str("hash", block.ID()).
					Str("TrieRoot", enc.ToString(block.GetHeader().GetBlocksRootHash())).
					Err(err).Msg("block produced")

				chain.ConnectBlock(s, block, blockState)
			}
		case <-s.quit:
			return
		}
	}
}

// JobQueue returns the queue for block production triggering.
func (s *SimpleBlockFactory) JobQueue() chan<- interface{} {
	return s.jobQueue
}
