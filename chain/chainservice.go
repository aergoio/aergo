/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	cfg "github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/contract"
	"github.com/aergoio/aergo/v2/contract/enterprise"
	"github.com/aergoio/aergo/v2/contract/name"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/fee"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	lru "github.com/hashicorp/golang-lru"
)

var (
	logger = log.NewLogger("chain")

	dfltErrBlocks = 128

	ErrNotSupportedConsensus = errors.New("not supported by this consensus")
	ErrRecoNoBestStateRoot   = errors.New("state root of best block is not exist")
	ErrRecoInvalidSdbRoot    = errors.New("state root of sdb is invalid")

	TestDebugger *Debugger
)

// Core represents a storage layer of a blockchain (chain & state DB).
type Core struct {
	cdb *ChainDB
	sdb *state.ChainStateDB
}

// NewCore returns an instance of Core.
func NewCore(dbType string, dataDir string, testModeOn bool, forceResetHeight types.BlockNo) (*Core, error) {
	core := &Core{
		cdb: NewChainDB(),
		sdb: state.NewChainStateDB(),
	}

	err := core.init(dbType, dataDir, testModeOn, forceResetHeight)
	if err != nil {
		return nil, err
	}

	return core, nil
}

// Init prepares Core (chain & state DB).
func (core *Core) init(dbType string, dataDir string, testModeOn bool, forceResetHeight types.BlockNo) error {
	// init chaindb
	if err := core.cdb.Init(dbType, dataDir); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize chaindb")
		return err
	}

	if forceResetHeight > 0 {
		if err := core.cdb.ResetBest(forceResetHeight); err != nil {
			logger.Fatal().Err(err).Uint64("height", forceResetHeight).Msg("failed to reset chaindb")
			return err
		}
	}

	// init statedb
	bestBlock, err := core.cdb.GetBestBlock()
	if err != nil {
		return err
	}

	if err := core.sdb.Init(dbType, dataDir, bestBlock, testModeOn); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize statedb")
		return err
	}

	contract.LoadDatabase(dataDir)

	return nil
}

func (core *Core) initGenesis(genesis *types.Genesis, mainnet bool, testmode bool) (*types.Block, error) {

	gen := core.cdb.GetGenesisInfo()
	if gen == nil {
		logger.Info().Msg("generating genesis block..")
		if testmode {
			if !mainnet {
				logger.Warn().Msg("--testnet opt will ignored due to testmode")
			}
			genesis = types.GetTestGenesis()
		} else {
			if genesis == nil {
				if mainnet {
					//return nil, errors.New("to use mainnet, create genesis manually (visit http://docs.aergo.io)")
					genesis = types.GetMainNetGenesis()
				} else {
					genesis = types.GetTestNetGenesis()
				}
			}
		}

		err := core.sdb.SetGenesis(genesis, InitGenesisBPs)
		if err != nil {
			logger.Fatal().Err(err).Msg("cannot set statedb of genesisblock")
			return nil, err
		}

		err = core.cdb.addGenesisBlock(genesis)
		if err != nil {
			logger.Fatal().Err(err).Msg("cannot add genesisblock")
			return nil, err
		}
		gen = genesis
	} else {
		if !mainnet {
			logger.Warn().Msg("--testnet option will be ignored")
		}
		if testmode && !gen.HasDevChainID() {
			logger.Info().Str("chain id", gen.ID.ToJSON()).Msg("current genesis info")
			return nil, errors.New("do not run testmode on non dev-chain")
		}
	}

	initChainParams(gen)

	genesisBlock, _ := core.cdb.GetBlockByNo(0)

	logger.Info().Str("chain id", gen.ID.ToJSON()).
		Str("hash", enc.ToString(genesisBlock.GetHash())).Msg("chain initialized")
	return genesisBlock, nil
}

func (core *Core) GetGenesisInfo() *types.Genesis {
	return core.cdb.GetGenesisInfo()
}

// Close closes chain & state DB.
func (core *Core) Close() {
	if core.sdb != nil {
		core.sdb.Close()
	}
	if core.cdb != nil {
		core.cdb.Close()
	}
	contract.CloseDatabase()
}

// InitGenesisBlock initialize chain database and generate specified genesis block if necessary
func (core *Core) InitGenesisBlock(gb *types.Genesis, useTestnet bool) error {
	_, err := core.initGenesis(gb, useTestnet, false)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot initialize genesis block")
		return err
	}
	return nil
}

type IChainHandler interface {
	getBlock(blockHash []byte) (*types.Block, error)
	getBlockByNo(blockNo types.BlockNo) (*types.Block, error)
	getTx(txHash []byte) (*types.Tx, *types.TxIdx, error)
	getReceipt(txHash []byte) (*types.Receipt, error)
	getAccountVote(addr []byte) (*types.AccountVoteInfo, error)
	getVotes(id string, n uint32) (*types.VoteList, error)
	getStaking(addr []byte) (*types.Staking, error)
	getNameInfo(name string, blockNo types.BlockNo) (*types.NameInfo, error)
	getEnterpriseConf(key string) (*types.EnterpriseConfig, error)
	addBlock(newBlock *types.Block, usedBstate *state.BlockState, peerID types.PeerID) error
	getAnchorsNew() (ChainAnchor, types.BlockNo, error)
	findAncestor(Hashes [][]byte) (*types.BlockInfo, error)
	setSkipMempool(val bool)
	listEvents(filter *types.FilterInfo) ([]*types.Event, error)
	verifyBlock(block *types.Block) error
}

// ChainService manage connectivity of blocks
type ChainService struct {
	*component.BaseComponent
	consensus.ChainConsensus
	*Core

	cfg       *cfg.Config
	op        *OrphanPool
	errBlocks *lru.Cache

	validator *BlockValidator

	chainWorker   *ChainWorker
	chainManager  *ChainManager
	chainVerifier *ChainVerifier

	stat stats

	recovered  atomic.Value
	debuggable bool
}

var _ types.ChainAccessor = (*ChainService)(nil)

// NewChainService creates an instance of ChainService.
func NewChainService(cfg *cfg.Config) *ChainService {
	cs := &ChainService{
		cfg:  cfg,
		op:   NewOrphanPool(DfltOrphanPoolSize),
		stat: newStats(),
	}

	cs.setRecovered(false)

	var err error
	if cs.Core, err = NewCore(cfg.DbType, cfg.DataDir, cfg.EnableTestmode, types.BlockNo(cfg.Blockchain.ForceResetHeight)); err != nil {
		logger.Panic().Err(err).Msg("failed to initialize DB")
	}

	if err = Init(cfg.Blockchain.MaxBlockSize,
		cfg.Blockchain.CoinbaseAccount,
		cfg.Consensus.EnableBp,
		cfg.Blockchain.MaxAnchorCount,
		cfg.Blockchain.VerifierCount); err != nil {
		logger.Panic().Err(err).Msg("failed to init chainservice | invalid config: blockchain")
	}

	var verifyMode = cs.cfg.Blockchain.VerifyOnly || cs.cfg.Blockchain.VerifyBlock != 0

	cs.validator = NewBlockValidator(cs, cs.sdb, cs.cfg.Blockchain.VerifyBlock != 0)
	cs.BaseComponent = component.NewBaseComponent(message.ChainSvc, cs, logger)
	cs.chainManager = newChainManager(cs, cs.Core)
	cs.chainWorker = newChainWorker(cs, cs.cfg.Blockchain.NumWorkers, cs.Core)
	// TODO set VerifyOnly true if cs.cfg.Blockchain.VerifyBlock is not 0
	if verifyMode {
		if cs.cfg.Consensus.EnableBp {
			logger.Fatal().Err(err).Msg("can't be enableBp at verifyOnly mode")
		}
		cs.chainVerifier = newChainVerifier(cs, cs.Core, cs.cfg.Blockchain.VerifyBlock)
	}

	cs.errBlocks, err = lru.New(dfltErrBlocks)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init lru")
		return nil
	}

	// init genesis block
	if _, err := cs.initGenesis(nil, !cfg.UseTestnet, cfg.EnableTestmode); err != nil {
		logger.Panic().Err(err).Msg("failed to create a genesis block")
	}

	if err := cs.checkHardfork(); err != nil {
		logger.Panic().Err(err).Msg("check the hardfork compatibility")
	}

	if ConsensusName() == consensus.ConsensusName[consensus.ConsensusDPOS] {
		top, err := cs.getVotes(types.OpvoteBP.ID(), 1)
		if err != nil {
			logger.Debug().Err(err).Msg("failed to get elected BPs")
		} else {
			for _, res := range top.Votes {
				logger.Debug().Str("BP", enc.ToString(res.Candidate)).
					Str("votes", new(big.Int).SetBytes(res.Amount).String()).Msgf("BP vote stat")
			}
		}
	}

	// init related modules
	if !pubNet {
		fee.EnableZeroFee()
	}
	contract.PubNet = pubNet
	contract.TraceBlockNo = cfg.Blockchain.StateTrace
	contract.SetStateSQLMaxDBSize(cfg.SQL.MaxDbSize)
	contract.StartLStateFactory((cfg.Blockchain.NumWorkers+2)*(int(contract.MaxCallDepth(cfg.Hardfork.Version(math.MaxUint64)))+2), cfg.Blockchain.NumLStateClosers, cfg.Blockchain.CloseLimit)
	contract.InitContext(cfg.Blockchain.NumWorkers + 2)

	// For a strict governance transaction validation.
	types.InitGovernance(cs.ConsensusType(), cs.IsPublic())
	system.InitGovernance(cs.ConsensusType())

	//reset parameter of aergo.system
	systemState, err := cs.SDB().GetSystemAccountState()
	if err != nil {
		logger.Panic().Err(err).Msg("failed to read aergo.system state")
	}
	system.InitSystemParams(systemState, len(cs.GetGenesisInfo().BPs))

	// init Debugger
	cs.initDebugger()

	cs.startChilds()

	return cs
}

func (cs *ChainService) startChilds() {
	if !cs.cfg.Blockchain.VerifyOnly && cs.cfg.Blockchain.VerifyBlock == 0 {
		cs.chainManager.Start()
		cs.chainWorker.Start()
	} else {
		cs.chainVerifier.Start()
	}
}

func (cs *ChainService) initDebugger() {
	TestDebugger = newDebugger()
}

// SDB returns cs.sdb.
func (cs *ChainService) SDB() *state.ChainStateDB {
	return cs.sdb
}

// CDB returns cs.sdb as a consensus.ChainDbReader.
func (cs *ChainService) CDB() consensus.ChainDB {
	return cs.cdb
}

// WalDB returns cs.sdb as a consensus.ChainDbReader.
func (cs *ChainService) WalDB() consensus.ChainWAL {
	return cs.cdb
}

// GetConsensusInfo returns consensus-related information, which is different
// from consensus to consensus.
func (cs *ChainService) GetConsensusInfo() string {
	if cs.ChainConsensus == nil {
		return ""
	}

	return cs.Info()
}

func (cs *ChainService) GetChainStats() string {
	return cs.stat.JSON()
}

// GetEnterpriseConfig return EnterpiseConfig. if the given key does not exist, fill EnterpriseConfig with only the key and return
func (cs *ChainService) GetEnterpriseConfig(key string) (*types.EnterpriseConfig, error) {
	return cs.getEnterpriseConf(key)
}

func (cs *ChainService) GetSystemValue(key types.SystemValue) (*big.Int, error) {
	return cs.getSystemValue(key)
}

// SetChainConsensus sets cs.cc to cc.
func (cs *ChainService) SetChainConsensus(cc consensus.ChainConsensus) {
	cs.ChainConsensus = cc
	cs.cdb.cc = cc
}

// BeforeStart initialize chain database and generate empty genesis block if necessary
func (cs *ChainService) BeforeStart() {
}

// AfterStart ... do nothing
func (cs *ChainService) AfterStart() {
}

// BeforeStop close chain database and stop BlockValidator
func (cs *ChainService) BeforeStop() {
	cs.Close()

	cs.chainManager.Stop()
	cs.chainWorker.Stop()

	cs.validator.Stop()
}

func (cs *ChainService) notifyBlock(block *types.Block, isByBP bool) {
	cs.BaseComponent.RequestTo(message.P2PSvc,
		&message.NotifyNewBlock{
			Produced: isByBP,
			BlockNo:  block.Header.BlockNo,
			Block:    block,
		})
}

func (cs *ChainService) setRecovered(val bool) {
	cs.recovered.Store(val)
	return
}

func (cs *ChainService) isRecovered() bool {
	var val, ok bool
	aopv := cs.recovered.Load()
	if aopv == nil {
		logger.Panic().Msg("ChainService: recovered is nil")
	}
	if val, ok = aopv.(bool); !ok {
		logger.Panic().Msg("ChainService: recovered is not bool")
	}
	return val
}

// Receive actor message
func (cs *ChainService) Receive(context actor.Context) {
	if !cs.isRecovered() {
		err := cs.Recover()
		if err != nil {
			logger.Fatal().Err(err).Msg("CHAIN DATA IS CRASHED, BUT CAN'T BE RECOVERED")
		}

		cs.setRecovered(true)
	}

	switch msg := context.Message().(type) {
	case *message.AddBlock,
		*message.GetAnchors, //TODO move to ChainWorker (need chain lock)
		*message.GetAncestor:
		cs.chainManager.Request(msg, context.Sender())

		//pass to chainWorker
	case *message.GetBlock,
		*message.GetBlockByNo,
		*message.GetState,
		*message.GetStateAndProof,
		*message.GetTx,
		*message.GetReceipt,
		*message.GetABI,
		*message.GetQuery,
		*message.GetStateQuery,
		*message.GetElected,
		*message.GetVote,
		*message.GetStaking,
		*message.GetNameInfo,
		*message.GetEnterpriseConf,
		*message.GetParams,
		*message.ListEvents,
		*message.CheckFeeDelegation:
		cs.chainWorker.Request(msg, context.Sender())

		//handle directly
	case *message.GetBestBlockNo:
		context.Respond(message.GetBestBlockNoRsp{
			BlockNo: cs.getBestBlockNo(),
		})
	case *message.GetBestBlock:
		block, err := cs.GetBestBlock()
		if err != nil {
			logger.Error().Err(err).Msg("failed to get best block")
		}
		context.Respond(message.GetBestBlockRsp{
			Block: block,
			Err:   err,
		})
	case *message.MemPoolDelRsp:
		err := msg.Err
		if err != nil {
			logger.Error().Err(err).Msg("failed to remove txs from mempool")
		}
	case actor.SystemMessage,
		actor.AutoReceiveMessage,
		actor.NotInfluenceReceiveTimeout:
		//logger.Debugf("Received message. (%v) %s", reflect.TypeOf(msg), msg)

	default:
		//logger.Debugf("Missed message. (%v) %s", reflect.TypeOf(msg), msg)
	}
}

func (cs *ChainService) Statistics() *map[string]interface{} {
	if cs.chainVerifier != nil {
		return cs.chainVerifier.Statistics()
	}
	return &map[string]interface{}{
		"testmode": cs.cfg.EnableTestmode,
		"testnet":  cs.cfg.UseTestnet,
		"orphan":   cs.op.curCnt,
		"config":   cs.cfg.Blockchain,
	}
}

func (cs *ChainService) GetChainTree() ([]byte, error) {
	return cs.cdb.GetChainTree()
}

func (cs *ChainService) getVotes(id string, n uint32) (*types.VoteList, error) {
	switch ConsensusName() {
	case consensus.ConsensusName[consensus.ConsensusDPOS]:
		sdb := cs.sdb.OpenNewStateDB(cs.sdb.GetRoot())
		if n == 0 {
			return system.GetVoteResult(sdb, []byte(id), system.GetBpCount())
		}
		return system.GetVoteResult(sdb, []byte(id), int(n))
	case consensus.ConsensusName[consensus.ConsensusRAFT]:
		//return cs.GetBPs()
		return nil, ErrNotSupportedConsensus
	default:
		return nil, ErrNotSupportedConsensus
	}
}

func (cs *ChainService) getAccountVote(addr []byte) (*types.AccountVoteInfo, error) {
	if cs.GetType() != consensus.ConsensusDPOS {
		return nil, ErrNotSupportedConsensus
	}

	sdb := cs.sdb.OpenNewStateDB(cs.sdb.GetRoot())
	scs, err := sdb.GetSystemAccountState()
	if err != nil {
		return nil, err
	}
	namescs, err := sdb.GetNameAccountState()
	if err != nil {
		return nil, err
	}
	voteInfo, err := system.GetVotes(scs, name.GetAddress(namescs, addr))
	if err != nil {
		return nil, err
	}

	return &types.AccountVoteInfo{Voting: voteInfo}, nil
}

func (cs *ChainService) getStaking(addr []byte) (*types.Staking, error) {
	if cs.GetType() != consensus.ConsensusDPOS {
		return nil, ErrNotSupportedConsensus
	}

	sdb := cs.sdb.OpenNewStateDB(cs.sdb.GetRoot())
	scs, err := sdb.GetSystemAccountState()
	if err != nil {
		return nil, err
	}
	namescs, err := sdb.GetNameAccountState()
	if err != nil {
		return nil, err
	}
	staking, err := system.GetStaking(scs, name.GetAddress(namescs, addr))
	if err != nil {
		return nil, err
	}
	return staking, nil
}

func (cs *ChainService) getNameInfo(qname string, blockNo types.BlockNo) (*types.NameInfo, error) {
	var stateDB *state.StateDB
	if blockNo != 0 {
		block, err := cs.cdb.GetBlockByNo(blockNo)
		if err != nil {
			return nil, err
		}
		stateDB = cs.sdb.OpenNewStateDB(block.GetHeader().GetBlocksRootHash())
	} else {
		stateDB = cs.sdb.OpenNewStateDB(cs.sdb.GetRoot())
	}
	return name.GetNameInfo(stateDB, qname)
}

func (cs *ChainService) getEnterpriseConf(key string) (*types.EnterpriseConfig, error) {
	sdb := cs.sdb.OpenNewStateDB(cs.sdb.GetRoot())
	if strings.ToUpper(key) != enterprise.AdminsKey {
		return enterprise.GetConf(sdb, key)
	}
	return enterprise.GetAdmin(sdb)
}

func (cs *ChainService) getSystemValue(key types.SystemValue) (*big.Int, error) {
	stateDB := cs.sdb.GetStateDB()
	switch key {
	case types.StakingTotal:
		return system.GetStakingTotal(stateDB)
	case types.StakingMin:
		return system.GetStakingMinimum(), nil
	case types.GasPrice:
		return system.GetGasPrice(), nil
	case types.NamePrice:
		return system.GetNamePrice(), nil
	case types.TotalVotingPower:
		return system.GetTotalVotingPower(), nil
	case types.VotingReward:
		return system.GetVotingRewardAmount(), nil
	}
	return nil, fmt.Errorf("unsupported system value : %s", key)
}

type ChainManager struct {
	*SubComponent
	IChainHandler //to use chain APIs
	*Core         // TODO remove after moving GetQuery to ChainWorker
}

type ChainWorker struct {
	*SubComponent
	IChainHandler //to use chain APIs
	*Core
}

var (
	chainManagerName  = "Chain Manager"
	chainWorkerName   = "Chain Worker"
	chainVerifierName = "Chain Verifier"
)

func newChainManager(cs *ChainService, core *Core) *ChainManager {
	chainManager := &ChainManager{IChainHandler: cs, Core: core}
	chainManager.SubComponent = NewSubComponent(chainManager, cs.BaseComponent, chainManagerName, 1)

	return chainManager
}

func newChainWorker(cs *ChainService, cntWorker int, core *Core) *ChainWorker {
	chainWorker := &ChainWorker{IChainHandler: cs, Core: core}
	chainWorker.SubComponent = NewSubComponent(chainWorker, cs.BaseComponent, chainWorkerName, cntWorker)

	return chainWorker
}

func (cm *ChainManager) Receive(context actor.Context) {
	defer RecoverExit()

	switch msg := context.Message().(type) {

	case *message.AddBlock:
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		block := msg.Block
		logger.Debug().Str("hash", block.ID()).Str("prev", block.PrevID()).Uint64("bestno", cm.cdb.getBestBlockNo()).
			Uint64("no", block.GetHeader().GetBlockNo()).Str("peer", p2putil.ShortForm(msg.PeerID)).Bool("syncer", msg.IsSync).Msg("add block chainservice")

		var bstate *state.BlockState
		if msg.Bstate != nil {
			bstate = msg.Bstate.(*state.BlockState)
			if timeoutTx := bstate.TimeoutTx(); timeoutTx != nil {
				if logger.IsDebugEnabled() {
					logger.Debug().Str("hash", enc.ToString(timeoutTx.GetHash())).Msg("received timeout tx")
				}
				cm.TellTo(message.MemPoolSvc, &message.MemPoolDelTx{Tx: timeoutTx.GetTx()})
			}
		}
		err := cm.addBlock(block, bstate, msg.PeerID)
		if err != nil {
			logger.Error().Err(err).Uint64("no", block.GetHeader().BlockNo).Str("hash", block.ID()).Msg("failed to add block")
		}
		blkNo := block.GetHeader().GetBlockNo()
		blkHash := block.BlockHash()

		rsp := message.AddBlockRsp{
			BlockNo:   blkNo,
			BlockHash: blkHash,
			Err:       err,
		}

		context.Respond(&rsp)
	case *message.GetAnchors:
		anchor, lastNo, err := cm.getAnchorsNew()
		context.Respond(message.GetAnchorsRsp{
			Seq:    msg.Seq,
			Hashes: anchor,
			LastNo: lastNo,
			Err:    err,
		})
	case *message.GetAncestor:
		hashes := msg.Hashes
		ancestor, err := cm.findAncestor(hashes)
		context.Respond(message.GetAncestorRsp{
			Ancestor: ancestor,
			Err:      err,
		})
	case *actor.Started, *actor.Stopping, *actor.Stopped, *component.CompStatReq: // donothing
	default:
		debug := fmt.Sprintf("[%s] Missed message. (%v) %s", cm.name, reflect.TypeOf(msg), msg)
		logger.Debug().Msg(debug)
	}
}

func getAddressNameResolved(sdb *state.StateDB, account []byte) ([]byte, error) {
	if len(account) == types.NameLength {
		scs, err := sdb.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoName)))
		if err != nil {
			logger.Error().Str("hash", enc.ToString(account)).Err(err).Msg("failed to get state for account")
			return nil, err
		}
		return name.GetAddress(scs, account), nil
	}
	return account, nil
}

func (cw *ChainWorker) Receive(context actor.Context) {
	var sdb *state.StateDB

	getAccProof := func(sdb *state.StateDB, account, root []byte, compressed bool) (*types.AccountProof, error) {
		address, err := getAddressNameResolved(sdb, account)
		if err != nil {
			return nil, err
		}
		id := types.ToAccountID(address)
		proof, err := sdb.GetAccountAndProof(id[:], root, compressed)
		if err != nil {
			logger.Error().Str("hash", enc.ToString(address)).Err(err).Msg("failed to get state for account")
			return nil, err
		}
		proof.Key = address
		return proof, err
	}

	switch msg := context.Message().(type) {
	case *message.GetBlock:
		bid := types.ToBlockID(msg.BlockHash)
		block, err := cw.getBlock(bid[:])
		if err != nil {
			logger.Debug().Err(err).Str("hash", enc.ToString(msg.BlockHash)).Msg("block not found")
		}
		context.Respond(message.GetBlockRsp{
			Block: block,
			Err:   err,
		})
	case *message.GetBlockByNo:
		block, err := cw.getBlockByNo(msg.BlockNo)
		if err != nil {
			logger.Error().Err(err).Uint64("blockNo", msg.BlockNo).Msg("failed to get block by no")
		}
		context.Respond(message.GetBlockByNoRsp{
			Block: block,
			Err:   err,
		})
	case *message.GetState:
		sdb = cw.sdb.OpenNewStateDB(cw.sdb.GetRoot())
		address, err := getAddressNameResolved(sdb, msg.Account)
		if err != nil {
			context.Respond(message.GetStateRsp{
				Account: msg.Account,
				State:   nil,
				Err:     err,
			})
			return
		}
		id := types.ToAccountID(address)
		accState, err := sdb.GetAccountState(id)
		if err != nil {
			logger.Error().Str("hash", enc.ToString(address)).Err(err).Msg("failed to get state for account")
		}
		context.Respond(message.GetStateRsp{
			Account: address,
			State:   accState,
			Err:     err,
		})
	case *message.GetStateAndProof:
		sdb = cw.sdb.OpenNewStateDB(cw.sdb.GetRoot())
		stateProof, err := getAccProof(sdb, msg.Account, msg.Root, msg.Compressed)
		context.Respond(message.GetStateAndProofRsp{
			StateProof: stateProof,
			Err:        err,
		})
	case *message.GetTx:
		tx, txIdx, err := cw.getTx(msg.TxHash)
		context.Respond(message.GetTxRsp{
			Tx:    tx,
			TxIds: txIdx,
			Err:   err,
		})
	case *message.GetReceipt:
		receipt, err := cw.getReceipt(msg.TxHash)
		context.Respond(message.GetReceiptRsp{
			Receipt: receipt,
			Err:     err,
		})
	case *message.GetABI:
		sdb = cw.sdb.OpenNewStateDB(cw.sdb.GetRoot())
		address, err := getAddressNameResolved(sdb, msg.Contract)
		if err != nil {
			context.Respond(message.GetABIRsp{
				ABI: nil,
				Err: err,
			})
			break
		}
		contractState, err := sdb.OpenContractStateAccount(types.ToAccountID(address))
		if err == nil {
			abi, err := contract.GetABI(contractState, nil)
			context.Respond(message.GetABIRsp{
				ABI: abi,
				Err: err,
			})
		} else {
			context.Respond(message.GetABIRsp{
				ABI: nil,
				Err: err,
			})
		}
	case *message.GetQuery:
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		sdb = cw.sdb.OpenNewStateDB(cw.sdb.GetRoot())
		address, err := getAddressNameResolved(sdb, msg.Contract)
		if err != nil {
			context.Respond(message.GetQueryRsp{Result: nil, Err: err})
			break
		}
		ctrState, err := sdb.OpenContractStateAccount(types.ToAccountID(address))
		if err != nil {
			logger.Error().Str("hash", enc.ToString(address)).Err(err).Msg("failed to get state for contract")
			context.Respond(message.GetQueryRsp{Result: nil, Err: err})
		} else {
			bs := state.NewBlockState(sdb)
			ret, err := contract.Query(address, bs, cw.cdb, ctrState, msg.Queryinfo)
			context.Respond(message.GetQueryRsp{Result: ret, Err: err})
		}
	case *message.GetStateQuery:
		sdb = cw.sdb.OpenNewStateDB(cw.sdb.GetRoot())
		contractProof, err := getAccProof(sdb, msg.ContractAddress, msg.Root, msg.Compressed)
		if err != nil {
			context.Respond(message.GetStateQueryRsp{
				Result: nil,
				Err:    err,
			})
			return
		}

		var varProofs []*types.ContractVarProof
		if contractProof.Inclusion {
			contractTrieRoot := contractProof.State.StorageRoot
			for _, storageKey := range msg.StorageKeys {
				varProof, err := sdb.GetVarAndProof(storageKey, contractTrieRoot, msg.Compressed)
				varProof.Key = storageKey
				varProofs = append(varProofs, varProof)
				if err != nil {
					logger.Error().Str("hash", enc.ToString(contractProof.Key)).Err(err).Msg("failed to get state variable in contract")
				}
			}
		}

		stateQuery := &types.StateQueryProof{
			ContractProof: contractProof,
			VarProofs:     varProofs,
		}
		context.Respond(message.GetStateQueryRsp{
			Result: stateQuery,
			Err:    err,
		})
	case *message.GetElected:
		top, err := cw.getVotes(msg.Id, msg.N)
		context.Respond(&message.GetVoteRsp{
			Top: top,
			Err: err,
		})
	case *message.GetVote:
		info, err := cw.getAccountVote(msg.Addr)
		context.Respond(&message.GetAccountVoteRsp{
			Info: info,
			Err:  err,
		})
	case *message.GetStaking:
		staking, err := cw.getStaking(msg.Addr)
		context.Respond(&message.GetStakingRsp{
			Staking: staking,
			Err:     err,
		})
	case *message.GetNameInfo:
		owner, err := cw.getNameInfo(msg.Name, msg.BlockNo)
		context.Respond(&message.GetNameInfoRsp{
			Owner: owner,
			Err:   err,
		})
	case *message.GetEnterpriseConf:
		conf, err := cw.getEnterpriseConf(msg.Key)
		context.Respond(&message.GetEnterpriseConfRsp{
			Conf: conf,
			Err:  err,
		})
	case *message.ListEvents:
		events, err := cw.listEvents(msg.Filter)
		context.Respond(&message.ListEventsRsp{
			Events: events,
			Err:    err,
		})
	case *message.GetParams:
		context.Respond(&message.GetParamsRsp{
			BpCount:      system.GetBpCount(),
			MinStaking:   system.GetStakingMinimum(),
			MaxBlockSize: uint64(MaxBlockSize()),
		})
	case *message.CheckFeeDelegation:
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		sdb = cw.sdb.OpenNewStateDB(cw.sdb.GetRoot())
		ctrState, err := sdb.OpenContractStateAccount(types.ToAccountID(msg.Contract))
		if err != nil {
			logger.Error().Str("hash", enc.ToString(msg.Contract)).Err(err).Msg("failed to get state for contract")
			context.Respond(message.CheckFeeDelegationRsp{Err: err})
		} else {
			bs := state.NewBlockState(sdb)
			err := contract.CheckFeeDelegation(msg.Contract, bs, nil, cw.cdb, ctrState, msg.Payload, msg.TxHash, msg.Sender, msg.Amount)
			context.Respond(message.CheckFeeDelegationRsp{Err: err})
		}

	case *actor.Started, *actor.Stopping, *actor.Stopped, *component.CompStatReq: // donothing
	default:
		debug := fmt.Sprintf("[%s] Missed message. (%v) %s", cw.name, reflect.TypeOf(msg), msg)
		logger.Debug().Msg(debug)
	}
}

func (cs *ChainService) ConsensusType() string {
	return cs.GetGenesisInfo().ConsensusType()
}

func (cs *ChainService) IsPublic() bool {
	return cs.GetGenesisInfo().PublicNet()
}

func (cs *ChainService) checkHardfork() error {
	config := cs.cfg.Hardfork
	if Genesis.IsMainNet() {
		*config = *cfg.MainNetHardforkConfig
	} else if Genesis.IsTestNet() {
		*config = *cfg.TestNetHardforkConfig
	}
	dbConfig := cs.cdb.Hardfork(*config)
	if len(dbConfig) == 0 {
		return cs.cdb.WriteHardfork(config)
	}
	if err := config.CheckCompatibility(dbConfig, cs.cdb.getBestBlockNo()); err != nil {
		return err
	}
	return cs.cdb.WriteHardfork(config)
}

func (cs *ChainService) ChainID(bno types.BlockNo) *types.ChainID {
	b, err := cs.GetGenesisInfo().ID.Bytes()
	if err != nil {
		return nil
	}
	cid := new(types.ChainID)
	err = cid.Read(b)
	if err != nil {
		return nil
	}
	cid.Version = cs.cfg.Hardfork.Version(bno)
	return cid
}
