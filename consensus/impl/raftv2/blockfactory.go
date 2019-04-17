package raftv2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"runtime"
	"sync"
	"time"

	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p"
	"github.com/libp2p/go-libp2p-crypto"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	bc "github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/chain"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"

	"github.com/aergoio/etcd/raft/raftpb"
)

const (
	slotQueueMax = 100
)

var (
	logger             *log.Logger
	httpLogger         *log.Logger
	RaftTick           = DefaultTickMS
	RaftSkipEmptyBlock = false
	peerCheckInterval  = time.Second * 3
)

var (
	ErrBFQuit = errors.New("block factory quit")

	/*
		ErrEmptyProposed = errors.New("proposed block is not exist")
		ErrInvalidPropsedBlock = errors.New("proposed block is not equal to commited block")
	*/
)

func init() {
	logger = log.NewLogger("raftv2")
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

// BlockFactory implments a raft block factory which generate block each cfg.Consensus.BlockInterval if this node is leader of raft
//
// This can be used for testing purpose.
type BlockFactory struct {
	*component.ComponentHub
	consensus.ChainWAL

	bpc              *Cluster
	jobQueue         chan interface{}
	quit             chan interface{}
	blockInterval    time.Duration
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
	return consensus.ConsensusName[consensus.ConsensusRAFTV2]
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

	bf := &BlockFactory{
		ComponentHub:     hub,
		ChainWAL:         cdb,
		jobQueue:         make(chan interface{}, slotQueueMax),
		blockInterval:    time.Second * time.Duration(cfg.Consensus.BlockInterval),
		maxBlockBodySize: chain.MaxBlockBodySize(),
		quit:             make(chan interface{}),
		ID:               p2p.NodeSID(),
		privKey:          p2p.NodePrivKey(),
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

type Proposed struct {
	block      *types.Block
	blockState *state.BlockState
}

type RaftOperator struct {
	confChangeC chan raftpb.ConfChange
	commitC     chan *types.Block

	rs *raftServer

	proposed *Proposed
}

func newRaftOperator(rs *raftServer) *RaftOperator {
	confChangeC := make(chan raftpb.ConfChange, 1)
	commitC := make(chan *types.Block)

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

func (bf *BlockFactory) newRaftServer(cfg *config.Config) error {
	if err := bf.InitCluster(cfg); err != nil {
		return err
	}

	bf.raftOp = newRaftOperator(bf.raftServer)

	snapdir := fmt.Sprintf("%s/raft/snap", cfg.DataDir)

	logger.Info().Uint64("RaftID", bf.bpc.ID).Msg("raft server start")

	bf.raftServer = newRaftServer(bf.ComponentHub, bf.bpc, cfg.Consensus.Raft.RaftListenUrl, false, snapdir,
		cfg.Consensus.Raft.RaftCertFile, cfg.Consensus.Raft.RaftKeyFile,
		nil, RaftTick, bf.raftOp.confChangeC, bf.raftOp.commitC, false, bf.ChainWAL)

	bf.bpc.rs = bf.raftServer
	bf.raftOp.rs = bf.raftServer

	return nil
}

// Ticker returns a time.Ticker for the main consensus loop.
func (bf *BlockFactory) Ticker() *time.Ticker {
	return time.NewTicker(bf.blockInterval)
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
		//TODO is it ok if last job was failed?
		if bf.prevBlock != nil && bf.prevBlock.BlockNo() == b.BlockNo() {
			logger.Debug().Msg("previous block not connected. skip to generate block")
			return
		}
		bf.prevBlock = b
		jq <- b
	}
}

func (bf *BlockFactory) GetType() consensus.ConsensusType {
	return consensus.ConsensusRAFTV2
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

	bf.raftServer.Start()

	runtime.LockOSThread()

	// wait to commit all uncommited log in WAL, and start
	bf.raftServer.WaitStartup()

	for {
		select {
		case e := <-bf.jobQueue:
			if prevBlock, ok := e.(*types.Block); ok {
				if err := bf.build(prevBlock); err != nil {
					return
				}
			}
		case block, ok := <-bf.commitC():
			if !ok {
				logger.Fatal().Msg("commit channel for raft is closed")
				return
			}

			if block == nil {
				bf.reset()
				continue
			}

			// add block that has produced by remote BP
			if err := bf.connect(block); err != nil {
				logger.Fatal().Err(err).Msg("failed to connect block")
				return
			}
		case <-bf.quit:
			return
		}
	}
}

func (bf *BlockFactory) build(prevBlock *types.Block) error {
	blockState := bf.sdb.NewBlockState(prevBlock.GetHeader().GetBlocksRootHash())

	ts := time.Now().UnixNano()

	txOp := chain.NewCompTxOp(
		bf.txOp,
		newTxExec(bf.ChainWAL, prevBlock.GetHeader().GetBlockNo()+1, ts, prevBlock.GetHash(), prevBlock.GetHeader().GetChainID()),
	)

	block, err := chain.GenerateBlock(bf, prevBlock, blockState, txOp, ts, RaftSkipEmptyBlock)
	if err == chain.ErrBlockEmpty {
		return nil
	} else if err != nil {
		logger.Info().Err(err).Msg("failed to produce block")
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
		logger.Info().Msg("skip producing block because this bp is not leader")
		return nil
	}

	bf.raftOp.propose(block, blockState)

	return nil
}

func (bf *BlockFactory) commitC() chan *types.Block {
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
			logger.Warn().Uint64("prop-no", proposed.block.GetHeader().GetBlockNo()).Str("prop", proposed.block.ID()).Uint64("commit-no", block.GetHeader().GetBlockNo()).Str("commit", block.ID()).Msg("commited block is not commited. this node is probably not leader")
			bf.raftOp.resetPropose()
		} else {
			blockState = proposed.blockState
		}
	}

	logger.Debug().Uint64("no", block.BlockNo()).
		Str("hash", block.ID()).
		Str("prev", block.PrevID()).
		Msg("connect block")

	//if bestblock is changed, connecting block failed. new block is generated in next tick
	if err := chain.ConnectBlock(bf, block, blockState); err != nil {
		logger.Error().Msg(err.Error())
		return err
	}

	return nil
}

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
