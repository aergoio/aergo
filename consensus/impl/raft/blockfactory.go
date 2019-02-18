package raft

import (
	"errors"
	"fmt"
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

	"github.com/coreos/etcd/raft/raftpb"
)

const (
	slotQueueMax = 100
)

var (
	logger             *log.Logger
	RaftBpTick         = time.Second
	RaftSkipEmptyBlock = false
)

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
	prevBlock        *types.Block // best block of last job

	raftServer *raftServer
}

// GetConstructor build and returns consensus.Constructor from New function.
func GetConstructor(cfg *config.Config, hub *component.ComponentHub, cdb consensus.ChainDB,
	sdb *state.ChainStateDB) consensus.Constructor {
	return func() (consensus.Consensus, error) {
		return New(cfg, hub, cdb, sdb)
	}
}

var (
	ErrInvalidRaftID = errors.New("invalid raft id")
	ErrDupRaftUrl    = errors.New("duplicated raft bp urls")
)

func checkConfig(cfg *config.Config) error {
	//check Url
	// - each url is unique
	// - format is valid url
	//check ID
	// - 1 <= ID <= len(RaftBpUrls)
	consCfg := cfg.Consensus
	lenBpUrls := len(consCfg.RaftBpUrls)

	urlMap := make(map[string]bool)
	for _, url := range consCfg.RaftBpUrls {
		//TODO url is valid
		if _, ok := urlMap[url]; ok {
			return ErrDupRaftUrl
		} else {
			urlMap[url] = true
		}
	}

	raftID := cfg.Consensus.RaftID
	if raftID <= 0 || raftID > uint64(lenBpUrls) {
		return ErrInvalidRaftID
	}

	RaftSkipEmptyBlock = cfg.Consensus.RaftSkipEmpty
	return nil
}

// New returns a BlockFactory.
func New(cfg *config.Config, hub *component.ComponentHub, cdb consensus.ChainDB,
	sdb *state.ChainStateDB) (*BlockFactory, error) {

	bf := &BlockFactory{
		ComponentHub:     hub,
		ChainDB:          cdb,
		jobQueue:         make(chan interface{}, slotQueueMax),
		blockInterval:    RaftBpTick,
		maxBlockBodySize: chain.MaxBlockBodySize(),
		quit:             make(chan interface{}),
		sdb:              sdb,
	}

	if err := bf.initRaftServer(cfg); err != nil {
		logger.Error().Err(err).Msg("failed to init raft server")
		return bf, err
	}

	bf.txOp = chain.NewCompTxOp(
		chain.TxOpFn(func(bState *state.BlockState, txIn types.Transaction) error {
			select {
			case <-bf.quit:
				return chain.ErrQuit
			default:
				return nil
			}
		}),
	)

	return bf, nil
}

func (bf *BlockFactory) initRaftServer(cfg *config.Config) error {
	if err := checkConfig(cfg); err != nil {
		return err
	}

	proposeC := make(chan string, 1)
	confChangeC := make(chan raftpb.ConfChange, 1)

	peersList := cfg.Consensus.RaftBpUrls
	waldir := fmt.Sprintf("%s/raft/wal", cfg.DataDir)
	snapdir := fmt.Sprintf("%s/raft/snap", cfg.DataDir)

	bf.raftServer = newRaftServer(cfg.Consensus.RaftID, peersList, false, waldir, snapdir, nil, proposeC, confChangeC)

	bf.raftServer.WaitStartup()

	return nil
}

// Ticker returns a time.Ticker for the main consensus loop.
func (bf *BlockFactory) Ticker() *time.Ticker {
	return time.NewTicker(bf.blockInterval)
}

// QueueJob send a block triggering information to jq.
func (bf *BlockFactory) QueueJob(now time.Time, jq chan<- interface{}) {
	if !bf.raftServer.IsLeader() {
		logger.Debug().Msg("skip producing block because this bp is not leader")
		return
	}

	if b, _ := bf.GetBestBlock(); b != nil {
		if bf.prevBlock != nil && bf.prevBlock.BlockNo() == b.BlockNo() {
			logger.Debug().Msg("previous block not connected. skip to generate block")
			return
		}
		bf.prevBlock = b
		jq <- b
	}
}

// IsTransactionValid checks the onsensus level validity of a transaction
func (bf *BlockFactory) IsTransactionValid(tx *types.Tx) bool {
	// BlockFactory has no tx valid check.
	return true
}

// VerifyTimestamp checks the validity of the block timestamp.
func (bf *BlockFactory) VerifyTimestamp(*types.Block) bool {
	// BlockFactory don't need to check timestamp.
	return true
}

// VerifySign checks the consensus level validity of a block.
func (bf *BlockFactory) VerifySign(*types.Block) error {
	// BlockFactory has no block signature.
	return nil
}

// IsBlockValid checks the consensus level validity of a block.
func (bf *BlockFactory) IsBlockValid(*types.Block, *types.Block) error {
	// BlockFactory has no block valid check.
	return nil
}

// QuitChan returns the channel from which consensus-related goroutines check
// when shutdown is initiated.
func (bf *BlockFactory) QuitChan() chan interface{} {
	return bf.quit
}

// Update has nothging to do.
func (bf *BlockFactory) Update(block *types.Block) {
}

// Save has nothging to do.
func (bf *BlockFactory) Save(tx db.Transaction) error {
	return nil
}

// BlockFactory returns r itself.
func (bf *BlockFactory) BlockFactory() consensus.BlockFactory {
	return bf
}

// NeedReorganization has nothing to do.
func (bf *BlockFactory) NeedReorganization(rootNo types.BlockNo) bool {
	return true
}

// Start run a raft block factory service.
func (bf *BlockFactory) Start() {
	defer logger.Info().Msg("shutdown initiated. stop the service")

	runtime.LockOSThread()

	for {
		select {
		case e := <-bf.jobQueue:
			if prevBlock, ok := e.(*types.Block); ok {
				blockState := bf.sdb.NewBlockState(prevBlock.GetHeader().GetBlocksRootHash())

				ts := time.Now().UnixNano()

				txOp := chain.NewCompTxOp(
					bf.txOp,
					newTxExec(prevBlock.GetHeader().GetBlockNo()+1, ts, prevBlock.GetHash()),
				)

				block, err := chain.GenerateBlock(bf, prevBlock, blockState, txOp, ts, RaftSkipEmptyBlock)
				if err == chain.ErrQuit {
					return
				} else if err == chain.ErrBlockEmpty {
					continue
				} else if err != nil {
					logger.Info().Err(err).Msg("failed to produce block")
					continue
				}
				logger.Info().Uint64("no", block.GetHeader().GetBlockNo()).Str("hash", block.ID()).
					Str("TrieRoot", enc.ToString(block.GetHeader().GetBlocksRootHash())).
					Err(err).Msg("block produced")

				if !bf.raftServer.IsLeader() {
					logger.Info().Msg("skip producing block because this bp is not leader")
					continue
				}

				//if bestblock is changed, connecting block failed. new block is generated in next tick
				if err := chain.ConnectBlock(bf, block, blockState); err != nil {
					logger.Error().Msg(err.Error())
				}
			}
		case <-bf.quit:
			return
		}
	}
}

// JobQueue returns the queue for block production triggering.
func (bf *BlockFactory) JobQueue() chan<- interface{} {
	return bf.jobQueue
}
