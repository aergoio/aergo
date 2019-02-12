package raft

import (
	"runtime"
	"strings"
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

	"github.com/coreos/etcd/raft/raftpb"
)

const (
	slotQueueMax = 100
)

var logger *log.Logger

func init() {
	logger = log.NewLogger("raft")
}

type txExec struct {
	execTx bc.TxExecFn
}

func newTxExec(blockNo types.BlockNo, ts int64, prevHash []byte) chain.TxOp {
	// Block hash not determined yet
	return &txExec{
		execTx: bc.NewTxExecutor(blockNo, ts, prevHash, contract.BlockFactory),
	}
}

func (te *txExec) Apply(bState *state.BlockState, tx types.Transaction) error {
	err := te.execTx(bState, tx)
	return err
}

// BlockFactory implments a simple block factory which generate block each cfg.Consensus.BlockInterval.
//
// This can be used for testing purpose.
type BlockFactory struct {
	*component.ComponentHub
	consensus.ChainDB
	jobQueue         chan interface{}
	blockInterval    time.Duration
	maxBlockBodySize uint32
	txOp             chain.TxOp
	quit             chan interface{}
	sdb              *state.ChainStateDB
	prevBlock        *types.Block

	//TODO refactoring
	raftServer *raftServer
}

// GetConstructor build and returns consensus.Constructor from New function.
func GetConstructor(cfg *config.Config, hub *component.ComponentHub, cdb consensus.ChainDB,
	sdb *state.ChainStateDB) consensus.Constructor {
	return func() (consensus.Consensus, error) {
		return New(cfg.Consensus, hub, cdb, sdb)
	}
}

// New returns a BlockFactory.
func New(cfg *config.ConsensusConfig, hub *component.ComponentHub, cdb consensus.ChainDB,
	sdb *state.ChainStateDB) (*BlockFactory, error) {
	r := &BlockFactory{
		ComponentHub:     hub,
		ChainDB:          cdb,
		jobQueue:         make(chan interface{}, slotQueueMax),
		blockInterval:    consensus.BlockInterval,
		maxBlockBodySize: chain.MaxBlockBodySize(),
		quit:             make(chan interface{}),
		sdb:              sdb,
	}

	proposeC := make(chan string, 1)
	confChangeC := make(chan raftpb.ConfChange, 1)

	//TODO peers, id from config
	peers := ""
	peersList := strings.Split(peers, ",")
	id := 1

	r.raftServer = newRaftServer(id, peersList, false, nil, proposeC, confChangeC)

	r.txOp = chain.NewCompTxOp(
		chain.TxOpFn(func(bState *state.BlockState, txIn types.Transaction) error {
			select {
			case <-r.quit:
				return chain.ErrQuit
			default:
				return nil
			}
		}),
	)

	return r, nil
}

// Ticker returns a time.Ticker for the main consensus loop.
func (r *BlockFactory) Ticker() *time.Ticker {
	return time.NewTicker(r.blockInterval)
}

// QueueJob send a block triggering information to jq.
func (r *BlockFactory) QueueJob(now time.Time, jq chan<- interface{}) {

	if b, _ := r.GetBestBlock(); b != nil {
		if r.prevBlock != nil && r.prevBlock.BlockNo() == b.BlockNo() {
			logger.Debug().Msg("previous block not connected. skip to generate block")
			return
		}
		r.prevBlock = b
		jq <- b
	}
}

// IsTransactionValid checks the onsensus level validity of a transaction
func (r *BlockFactory) IsTransactionValid(tx *types.Tx) bool {
	// BlockFactory has no tx valid check.
	return true
}

// VerifyTimestamp checks the validity of the block timestamp.
func (r *BlockFactory) VerifyTimestamp(*types.Block) bool {
	// BlockFactory don't need to check timestamp.
	return true
}

// VerifySign checks the consensus level validity of a block.
func (r *BlockFactory) VerifySign(*types.Block) error {
	// BlockFactory has no block signature.
	return nil
}

// IsBlockValid checks the consensus level validity of a block.
func (r *BlockFactory) IsBlockValid(*types.Block, *types.Block) error {
	// BlockFactory has no block valid check.
	return nil
}

// QuitChan returns the channel from which consensus-related goroutines check
// when shutdown is initiated.
func (r *BlockFactory) QuitChan() chan interface{} {
	return r.quit
}

// Update has nothging to do.
func (r *BlockFactory) Update(block *types.Block) {
}

// Save has nothging to do.
func (r *BlockFactory) Save(tx db.Transaction) error {
	return nil
}

// BlockFactory returns r itself.
func (r *BlockFactory) BlockFactory() consensus.BlockFactory {
	return r
}

// NeedReorganization has nothing to do.
func (r *BlockFactory) NeedReorganization(rootNo types.BlockNo) bool {
	return true
}

// Start run a simple block factory service.
func (r *BlockFactory) Start() {
	defer logger.Info().Msg("shutdown initiated. stop the service")

	runtime.LockOSThread()

	for {
		select {
		case e := <-r.jobQueue:
			if prevBlock, ok := e.(*types.Block); ok {
				blockState := r.sdb.NewBlockState(prevBlock.GetHeader().GetBlocksRootHash())

				ts := time.Now().UnixNano()

				txOp := chain.NewCompTxOp(
					r.txOp,
					newTxExec(prevBlock.GetHeader().GetBlockNo()+1, ts, prevBlock.GetHash()),
				)

				block, err := chain.GenerateBlock(r, prevBlock, blockState, txOp, ts)
				if err == chain.ErrQuit {
					return
				} else if err != nil {
					logger.Info().Err(err).Msg("failed to produce block")
					continue
				}
				logger.Info().Uint64("no", block.GetHeader().GetBlockNo()).Str("hash", block.ID()).
					Str("TrieRoot", enc.ToString(block.GetHeader().GetBlocksRootHash())).
					Err(err).Msg("block produced")

				chain.ConnectBlock(r, block, blockState)
			}
		case <-r.quit:
			return
		}
	}
}

// JobQueue returns the queue for block production triggering.
func (r *BlockFactory) JobQueue() chan<- interface{} {
	return r.jobQueue
}
