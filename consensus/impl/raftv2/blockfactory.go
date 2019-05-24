package raftv2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pkey"
	"runtime"
	"sync"
	"time"

	"github.com/aergoio/aergo/internal/enc"
	"github.com/libp2p/go-libp2p-crypto"

	"github.com/aergoio/aergo-lib/log"
	bc "github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/chain"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const (
	slotQueueMax          = 100
	DefaultCommitQueueLen = 10
)

var (
	logger     *log.Logger
	httpLogger *log.Logger

	// blockIntervalMs is the block genration interval in milli-seconds.
	RaftTick           = DefaultTickMS
	RaftSkipEmptyBlock = false
	MaxCommitQueueLen  = DefaultCommitQueueLen

	BlockTimeout time.Duration
)

var (
	ErrClusterNotReady = errors.New("cluster is not ready")
	ErrNotRaftLeader   = errors.New("this node is not leader")
)

func init() {
	logger = log.NewLogger("raft")
	httpLogger = log.NewLogger("rafthttp")
}

type txExec struct {
	execTx bc.TxExecFn
}

func newTxExec(cdb consensus.ChainDB, blockNo types.BlockNo, ts int64, prevHash []byte, chainID []byte) chain.TxOp {
	// Block hash not determined yet
	return &txExec{
		execTx: bc.NewTxExecutor(contract.ChainAccessor(cdb), blockNo, ts, prevHash, contract.BlockFactory, chainID),
	}
}

func (te *txExec) Apply(bState *state.BlockState, tx types.Transaction) error {
	err := te.execTx(bState, tx)
	return err
}

type Work struct {
	*types.Block
}

func (work *Work) GetTimeout() time.Duration {
	return BlockTimeout
}

func (work *Work) ToString() string {
	return fmt.Sprintf("bestblock=%s", work.BlockID())
}

// BlockFactory implments a raft block factory which generate block each cfg.Consensus.BlockInterval if this node is leader of raft
//
// This can be used for testing purpose.
type BlockFactory struct {
	*component.ComponentHub
	consensus.ChainWAL

	bpc *Cluster

	workerQueue chan *Work
	jobQueue    chan interface{}
	bpTimeoutC  chan interface{}
	quit        chan interface{}

	maxBlockBodySize uint32
	ID               string
	privKey          crypto.PrivKey
	txOp             chain.TxOp
	sdb              *state.ChainStateDB
	prevBlock        *types.Block // best block of last job
	jobLock          sync.RWMutex

	raftOp     *RaftOperator
	raftServer *raftServer
}

// GetName returns the name of the consensus.
func GetName() string {
	return consensus.ConsensusName[consensus.ConsensusRAFT]
}

// GetConstructor build and returns consensus.Constructor from New function.
func GetConstructor(cfg *config.Config, hub *component.ComponentHub, cdb consensus.ChainWAL,
	sdb *state.ChainStateDB, pa p2pcommon.PeerAccessor) consensus.Constructor {
	return func() (consensus.Consensus, error) {
		return New(cfg, hub, cdb, sdb, pa)
	}
}

// New returns a BlockFactory.
func New(cfg *config.Config, hub *component.ComponentHub, cdb consensus.ChainWAL,
	sdb *state.ChainStateDB, pa p2pcommon.PeerAccessor) (*BlockFactory, error) {

	Init(consensus.BlockInterval)

	bf := &BlockFactory{
		ComponentHub:     hub,
		ChainWAL:         cdb,
		jobQueue:         make(chan interface{}, slotQueueMax),
		workerQueue:      make(chan *Work),
		bpTimeoutC:       make(chan interface{}, 1),
		quit:             make(chan interface{}),
		maxBlockBodySize: chain.MaxBlockBodySize(),
		ID:               p2pkey.NodeSID(),
		privKey:          p2pkey.NodePrivKey(),
		sdb:              sdb,
	}

	if cfg.Consensus.EnableBp {
		if err := bf.newRaftServer(cfg); err != nil {
			logger.Error().Err(err).Msg("failed to init raft server")
			return bf, err
		}

		bf.raftServer.SetPeerAccessor(pa)
	}

	bf.txOp = chain.NewCompTxOp(
		// timeout check
		chain.TxOpFn(func(bState *state.BlockState, txIn types.Transaction) error {
			return bf.checkBpTimeout()
		}),
	)

	return bf, nil
}

func Init(blockInterval time.Duration) {
	logger.Debug().Int64("timeout(ms)", BlockTimeout.Nanoseconds()/int64(time.Millisecond)).Msg("set block timeout")

	BlockTimeout = blockInterval
}

func (bf *BlockFactory) newRaftServer(cfg *config.Config) error {
	if err := bf.InitCluster(cfg); err != nil {
		return err
	}

	bf.raftOp = newRaftOperator(bf.raftServer)

	logger.Info().Str("name", bf.bpc.NodeName()).Msg("create raft server")

	bf.raftServer = newRaftServer(bf.ComponentHub, bf.bpc, cfg.Consensus.Raft.ListenUrl, !cfg.Consensus.Raft.NewCluster,
		cfg.Consensus.Raft.CertFile, cfg.Consensus.Raft.KeyFile, nil,
		RaftTick, bf.bpc.confChangeC, bf.raftOp.commitC, false, bf.ChainWAL)

	bf.bpc.rs = bf.raftServer
	bf.raftOp.rs = bf.raftServer

	return nil
}

// Ticker returns a time.Ticker for the main consensus loop.
func (bf *BlockFactory) Ticker() *time.Ticker {
	return time.NewTicker(consensus.BlockInterval)
}

// QueueJob send a block triggering information to jq.
func (bf *BlockFactory) QueueJob(now time.Time, jq chan<- interface{}) {
	bf.jobLock.Lock()
	defer bf.jobLock.Unlock()

	if !bf.raftServer.IsLeader() {
		logger.Debug().Msg("skip producing block because this bp is not leader")
		return
	}

	if b, _ := bf.GetBestBlock(); b != nil {
		if bf.prevBlock != nil && bf.prevBlock.BlockNo() == b.BlockNo() {
			logger.Debug().Uint64("bestno", b.BlockNo()).Msg("previous block not connected. skip to generate block")
			return
		}

		// If requested block remains in commit channel, block factory must wait until all requests are completed.
		// otherwise block of same height will be created and a fork will occur.
		if !bf.raftServer.commitProgress.IsReadyToPropose() {
			logger.Debug().Uint64("bestno", b.BlockNo()).Msg("pending request block not connected. skip to generate block")
			return
		}

		bf.prevBlock = b
		jq <- &Work{b}
	}
}

func (bf *BlockFactory) GetType() consensus.ConsensusType {
	return consensus.ConsensusRAFT
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
func (bf *BlockFactory) VerifySign(block *types.Block) error {
	valid, err := block.VerifySign()
	if !valid || err != nil {
		return &consensus.ErrorConsensus{Msg: "bad block signature", Err: err}
	}
	return nil
}

// IsBlockValid checks the consensus level validity of a block.
func (bf *BlockFactory) IsBlockValid(block *types.Block, bestBlock *types.Block) error {
	// BlockFactory has no block valid check.
	_, err := block.BPID()
	if err != nil {
		return &consensus.ErrorConsensus{Msg: "bad public key in block", Err: err}
	}
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
func (bf *BlockFactory) Save(tx consensus.TxWriter) error {
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
	bf.raftServer.Start()

	go bf.worker()
	go bf.controller()
}

func (bf *BlockFactory) controller() {
	defer shutdownMsg("block factory controller")

	beginBlock := func(work *Work) error {
		// This is only for draining an unconsumed message, which means
		// the previous block is generated within timeout. This code
		// is needed since an empty block will be generated without it.
		if err := bf.checkBpTimeout(); err == chain.ErrQuit {
			return err
		}

		select {
		case bf.workerQueue <- work:
		default:
			logger.Error().Msgf(
				"skip block production for %s due to a pending job", work.ToString())
		}
		return nil
	}

	notifyBpTimeout := func(work *Work) {
		timeout := work.GetTimeout()
		time.Sleep(timeout)
		bf.bpTimeoutC <- struct{}{}
		logger.Debug().Int64("timeout(ms)", timeout.Nanoseconds()/int64(time.Millisecond)).Msg("block production timeout signaled")
	}

	for {
		select {
		case info := <-bf.jobQueue:
			work := info.(*Work)

			logger.Debug().Msgf("received work: %s",
				log.DoLazyEval(func() string { return work.ToString() }))

			err := beginBlock(work)
			if err == chain.ErrQuit {
				return
			} else if err != nil {
				logger.Debug().Err(err).Msg("skip block production")
				continue
			}

			notifyBpTimeout(work)

		case <-bf.quit:
			return
		}
	}
}

// worker() is different for each consensus
func (bf *BlockFactory) worker() {
	defer logger.Info().Msg("shutdown initiated. stop the service")

	runtime.LockOSThread()

	for {
		select {
		case work := <-bf.workerQueue:
			if err := bf.generateBlock(work.Block); err != nil {
				if err == chain.ErrQuit {
					logger.Info().Msg("quit worker of block factory")
					return
				}

				logger.Error().Err(err).Msg("failed to produce block")
			}

		case cEntry, ok := <-bf.commitC():
			logger.Debug().Msg("received block to connect from raft")

			if !ok {
				logger.Fatal().Msg("commit channel for raft is closed")
				return
			}

			if cEntry.block == nil {
				bf.reset()
				continue
			}

			// add block that has produced by remote BP
			if err := bf.connect(cEntry.block); err != nil {
				logger.Error().Err(err).Msg("failed to connect block")
				return
			}

			bf.raftServer.commitProgress.UpdateConnect(cEntry)
		case <-bf.quit:
			return
		}
	}
}

func (bf *BlockFactory) generateBlock(bestBlock *types.Block) (err error) {
	defer func() {
		if panicMsg := recover(); panicMsg != nil {
			err = fmt.Errorf("panic ocurred during block generation - %v", panicMsg)
		}
	}()

	blockState := bf.sdb.NewBlockState(bestBlock.GetHeader().GetBlocksRootHash())

	ts := time.Now().UnixNano()

	txOp := chain.NewCompTxOp(
		bf.txOp,
		newTxExec(bf.ChainWAL, bestBlock.GetHeader().GetBlockNo()+1, ts, bestBlock.GetHash(), bestBlock.GetHeader().GetChainID()),
	)

	block, err := chain.GenerateBlock(bf, bestBlock, blockState, txOp, ts, RaftSkipEmptyBlock)
	if err == chain.ErrBlockEmpty {
		return nil
	} else if err != nil {
		logger.Info().Err(err).Msg("failed to generate block")
		return err
	}

	if err = block.Sign(bf.privKey); err != nil {
		logger.Error().Err(err).Msg("failed to sign in block")
		return nil
	}

	logger.Info().Str("blockProducer", bf.ID).Str("raftID", block.ID()).
		Str("sroot", enc.ToString(block.GetHeader().GetBlocksRootHash())).
		Uint64("no", block.GetHeader().GetBlockNo()).
		Str("hash", block.ID()).
		Msg("block produced")

	if !bf.raftServer.IsLeader() {
		logger.Info().Msg("dropped produced block because this bp became no longer leader")
		return nil
	}

	bf.raftOp.propose(block, blockState)

	return nil
}

func (bf *BlockFactory) commitC() chan *commitEntry {
	return bf.raftOp.commitC
}

func (bf *BlockFactory) reset() {
	bf.jobLock.Lock()
	defer bf.jobLock.Unlock()

	logger.Debug().Str("prev proposed", bf.raftOp.toString()).Msg("commit nil data, so reset block factory")

	bf.prevBlock = nil
}

// save block/block state to connect after commit
func (bf *BlockFactory) connect(block *types.Block) error {
	proposed := bf.raftOp.proposed
	var blockState *state.BlockState

	if proposed != nil {
		if !bytes.Equal(block.BlockHash(), proposed.block.BlockHash()) {
			logger.Warn().Uint64("prop-no", proposed.block.GetHeader().GetBlockNo()).Str("prop", proposed.block.ID()).Uint64("commit-no", block.GetHeader().GetBlockNo()).Str("commit", block.ID()).Msg("commited block is not proposed by me. this node is probably not leader")
			bf.raftOp.resetPropose()
		} else {
			blockState = proposed.blockState
		}
	}

	logger.Debug().Uint64("no", block.BlockNo()).
		Str("hash", block.ID()).
		Str("prev", block.PrevID()).
		Bool("proposed", blockState != nil).
		Msg("connect block")

	// if bestblock is changed, connecting block failed. new block is generated in next tick
	// On a slow server, chain service takes too long to add block in blockchain. In this case, raft server waits to send new block to commit channel.
	if err := chain.ConnectBlock(bf, block, blockState, time.Second*300); err != nil {
		logger.Error().Msg(err.Error())
		return err
	}

	return nil
}

/*
// waitUntilStartable wait until this chain synchronizes with more than half of all peers
func (bf *BlockFactory) waitSyncWithMajority() error {
	ticker := time.NewTicker(peerCheckInterval)

	for {
		select {
		case <-ticker.C:
			if synced, err := bf.bpc.hasSynced(); err != nil {
				logger.Error().Err(err).Msg("failed to check sync with a majority of peers")
				return err
			} else if synced {
				return nil
			}

		case <-bf.QuitChan():
			logger.Info().Msg("quit while wait sync")
			return ErrBFQuit
		default:
		}
	}
}
*/
// JobQueue returns the queue for block production triggering.
func (bf *BlockFactory) JobQueue() chan<- interface{} {
	return bf.jobQueue
}

// Info retuns an empty string.
func (bf *BlockFactory) Info() string {
	// TODO: Returns a appropriate information inx json format like current
	// leader, etc.
	info := consensus.NewInfo(GetName())
	if bf.raftServer == nil {
		return info.AsJSON()
	}

	b, err := json.Marshal(bf.bpc.getRaftInfo(false))
	if err != nil {
		logger.Error().Err(err).Msg("failed to marshalEntryData raft consensus")
	} else {
		m := json.RawMessage(b)
		info.Status = &m
	}

	return info.AsJSON()
}

func (bf *BlockFactory) ConsensusInfo() *types.ConsensusInfo {
	if bf.bpc == nil {
		return &types.ConsensusInfo{Type: GetName()}
	}
	return bf.bpc.toConsensusInfo()
}

func (bf *BlockFactory) NeedNotify() bool {
	return false
}

func (bf *BlockFactory) HasWAL() bool {
	return true
}

type ErrorMembershipChange struct {
	Err error
}

func (e ErrorMembershipChange) Error() string {
	return fmt.Sprintf("failed to change membership: %s", e.Err.Error())
}

// ConfChange change membership of raft cluster and returns new membership
func (bf *BlockFactory) ConfChange(req *types.MembershipChange) (*consensus.Member, error) {
	if bf.bpc == nil {
		return nil, ErrorMembershipChange{ErrClusterNotReady}
	}

	if !bf.raftServer.IsLeader() {
		return nil, ErrorMembershipChange{ErrNotRaftLeader}
	}

	var member *consensus.Member
	var err error
	if member, err = bf.bpc.ChangeMembership(req); err != nil {
		return nil, ErrorMembershipChange{err}
	}

	return member, nil
}

func (bf *BlockFactory) ClusterInfo() ([]*types.MemberAttr, []byte, error) {
	return bf.bpc.getMemberAttrs(), bf.bpc.chainID, nil
}

func (bf *BlockFactory) checkBpTimeout() error {
	select {
	case <-bf.bpTimeoutC:
		return chain.ErrTimeout{Kind: "block"}
	case <-bf.quit:
		return chain.ErrQuit
	default:
		return nil
	}
}

type Proposed struct {
	block      *types.Block
	blockState *state.BlockState
}

type RaftOperator struct {
	confChangeC chan *types.MembershipChange
	commitC     chan *commitEntry

	rs *raftServer

	proposed *Proposed
}

func newRaftOperator(rs *raftServer) *RaftOperator {
	confChangeC := make(chan *types.MembershipChange, 1)
	commitC := make(chan *commitEntry, MaxCommitQueueLen)

	return &RaftOperator{confChangeC: confChangeC, commitC: commitC, rs: rs}
}

func (rop *RaftOperator) propose(block *types.Block, blockState *state.BlockState) {
	rop.proposed = &Proposed{block: block, blockState: blockState}

	if err := rop.rs.Propose(block); err != nil {
		logger.Error().Err(err).Msg("propose error to raft")
		return
	}

	logger.Info().Msg("block proposed by blockfactory")
}

func (rop *RaftOperator) resetPropose() {
	rop.proposed = nil
	logger.Debug().Msg("reset proposed block")
}

func (rop *RaftOperator) toString() string {
	buf := "proposed:"
	if rop.proposed != nil && rop.proposed.block != nil {
		buf = buf + fmt.Sprintf("[no=%d, hash=%s]", rop.proposed.block.BlockNo(), rop.proposed.block.BlockID().String())
	} else {
		buf = buf + "empty"
	}
	return buf
}

func shutdownMsg(m string) {
	logger.Info().Msgf("shutdown initiated. stop the %s", m)
}
