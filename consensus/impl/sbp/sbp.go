package sbp

import (
	"runtime"
	"time"

	"github.com/aergoio/aergo-lib/log"
	bc "github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/consensus/chain"
	"github.com/aergoio/aergo/v2/contract"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
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

func newTxExec(cdb consensus.ChainDB, bi *types.BlockHeaderInfo) chain.TxOp {
	// Block hash not determined yet
	return &txExec{
		execTx: bc.NewTxExecutor(nil, contract.ChainAccessor(cdb), bi, contract.BlockFactory),
	}
}

func (te *txExec) Apply(bState *state.BlockState, tx types.Transaction) error {
	err := te.execTx(bState, tx)
	return err
}

// SimpleBlockFactory implments a simple block factory which generate block each cfg.Consensus.BlockInterval.
//
// This can be used for testing purpose.
type SimpleBlockFactory struct {
	*component.ComponentHub
	consensus.ChainDB
	jobQueue         chan interface{}
	blockInterval    time.Duration
	maxBlockBodySize uint32
	txOp             chain.TxOp
	quit             chan interface{}
	sdb              *state.ChainStateDB
	prevBlock        *types.Block
	bv               types.BlockVersionner
}

// GetName returns the name of the consensus.
func GetName() string {
	return consensus.ConsensusName[consensus.ConsensusSBP]
}

// GetConstructor build and returns consensus.Constructor from New function.
func GetConstructor(cfg *config.Config, hub *component.ComponentHub, cdb consensus.ChainDB,
	sdb *state.ChainStateDB) consensus.Constructor {
	return func() (consensus.Consensus, error) {
		return New(cfg.Hardfork, hub, cdb, sdb)
	}
}

// New returns a SimpleBlockFactory.
func New(
	bv types.BlockVersionner,
	hub *component.ComponentHub,
	cdb consensus.ChainDB,
	sdb *state.ChainStateDB,
) (*SimpleBlockFactory, error) {
	s := &SimpleBlockFactory{
		ComponentHub:     hub,
		ChainDB:          cdb,
		jobQueue:         make(chan interface{}, slotQueueMax),
		blockInterval:    consensus.BlockInterval,
		maxBlockBodySize: chain.MaxBlockBodySize(),
		quit:             make(chan interface{}),
		sdb:              sdb,
		bv:               bv,
	}
	s.txOp = chain.NewCompTxOp(
		chain.TxOpFn(func(bState *state.BlockState, txIn types.Transaction) error {
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
	if b, _ := s.GetBestBlock(); b != nil {
		if s.prevBlock != nil && s.prevBlock.BlockNo() == b.BlockNo() {
			logger.Debug().Msg("previous block not connected. skip to generate block")
			return
		}
		s.prevBlock = b
		jq <- b
	}
}

func (s *SimpleBlockFactory) GetType() consensus.ConsensusType {
	return consensus.ConsensusSBP
}

// IsTransactionValid checks the onsensus level validity of a transaction
func (s *SimpleBlockFactory) IsTransactionValid(tx *types.Tx) bool {
	// SimpleBlockFactory has no tx valid check.
	return true
}

// VerifyTimestamp checks the validity of the block timestamp.
func (s *SimpleBlockFactory) VerifyTimestamp(*types.Block) bool {
	// SimpleBlockFactory don't need to check timestamp.
	return true
}

// VerifySign checks the consensus level validity of a block.
func (s *SimpleBlockFactory) VerifySign(*types.Block) error {
	// SimpleBlockFactory has no block signature.
	return nil
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
func (s *SimpleBlockFactory) Save(tx consensus.TxWriter) error {
	return nil
}

// BlockFactory returns s itself.
func (s *SimpleBlockFactory) BlockFactory() consensus.BlockFactory {
	return s
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
				bi := types.NewBlockHeaderInfoFromPrevBlock(prevBlock, time.Now().UnixNano(), s.bv)
				blockState := s.sdb.NewBlockState(
					prevBlock.GetHeader().GetBlocksRootHash(),
					state.SetPrevBlockHash(prevBlock.BlockHash()),
				)
				blockState.SetGasPrice(system.GetGasPriceFromState(blockState))
				blockState.Receipts().SetHardFork(s.bv, bi.No)
				txOp := chain.NewCompTxOp(s.txOp, newTxExec(s.ChainDB, bi))

				block, err := chain.NewBlockGenerator(s, nil, bi, blockState, txOp, false).GenerateBlock()
				if err == chain.ErrQuit {
					return
				} else if err != nil {
					logger.Info().Err(err).Msg("failed to produce block")
					continue
				}
				logger.Info().Uint64("no", block.GetHeader().GetBlockNo()).Str("hash", block.ID()).
					Str("TrieRoot", enc.ToString(block.GetHeader().GetBlocksRootHash())).
					Err(err).Msg("block produced")

				chain.ConnectBlock(s, block, blockState, time.Second)
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

// Info retuns an empty string since SBP has no valuable consensus-related
// information.
func (s *SimpleBlockFactory) Info() string {
	return consensus.NewInfo(GetName()).AsJSON()
}

func (s *SimpleBlockFactory) ConsensusInfo() *types.ConsensusInfo {
	return &types.ConsensusInfo{Type: GetName()}
}

var dummyRaft consensus.DummyRaftAccessor

func (s *SimpleBlockFactory) RaftAccessor() consensus.AergoRaftAccessor {
	return &dummyRaft
}

func (s *SimpleBlockFactory) NeedNotify() bool {
	return true
}

func (s *SimpleBlockFactory) HasWAL() bool {
	return false
}

func (s *SimpleBlockFactory) IsForkEnable() bool {
	return true
}

func (s *SimpleBlockFactory) IsConnectedBlock(block *types.Block) bool {
	_, err := s.ChainDB.GetBlock(block.BlockHash())
	if err == nil {
		return true
	}

	return false
}

func (s *SimpleBlockFactory) ConfChange(req *types.MembershipChange) (*consensus.Member, error) {
	return nil, consensus.ErrNotSupportedMethod
}

func (s *SimpleBlockFactory) ConfChangeInfo(requestID uint64) (*types.ConfChangeProgress, error) {
	return nil, consensus.ErrNotSupportedMethod
}

func (s *SimpleBlockFactory) MakeConfChangeProposal(req *types.MembershipChange) (*consensus.ConfChangePropose, error) {
	return nil, consensus.ErrNotSupportedMethod
}

func (s *SimpleBlockFactory) ClusterInfo(bestBlockHash []byte) *types.GetClusterInfoResponse {
	return &types.GetClusterInfoResponse{ChainID: nil, Error: consensus.ErrNotSupportedMethod.Error(), MbrAttrs: nil, HardStateInfo: nil}
}

func ValidateGenesis(genesis *types.Genesis) error {
	return nil
}
