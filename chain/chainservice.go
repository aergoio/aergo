/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"sync/atomic"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/contract/name"
	"github.com/aergoio/aergo/contract/system"
	"github.com/aergoio/aergo/fee"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/hashicorp/golang-lru"
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
	getAccountVote(id []string, addr []byte) (*types.AccountVoteInfo, error)
	getVotes(id string, n uint32) (*types.VoteList, error)
	getStaking(addr []byte) (*types.Staking, error)
	getNameInfo(name string, blockNo types.BlockNo) (*types.NameInfo, error)
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
		logger.Fatal().Err(err).Msg("failed to initialize DB")
		panic(err)
	}

	if err = Init(cfg.Blockchain.MaxBlockSize,
		cfg.Blockchain.CoinbaseAccount,
		cfg.Consensus.EnableBp,
		cfg.Blockchain.MaxAnchorCount,
		cfg.Blockchain.VerifierCount); err != nil {
		logger.Error().Err(err).Msg("failed to init chainservice")
		panic("invalid config: blockchain")
	}

	cs.validator = NewBlockValidator(cs, cs.sdb)
	cs.BaseComponent = component.NewBaseComponent(message.ChainSvc, cs, logger)
	cs.chainManager = newChainManager(cs, cs.Core)
	cs.chainWorker = newChainWorker(cs, defaultChainWorkerCount, cs.Core)
	if cs.cfg.Blockchain.VerifyOnly {
		if cs.cfg.Consensus.EnableBp {
			logger.Fatal().Err(err).Msg("can't be enableBp at verifyOnly mode")
		}
		cs.chainVerifier = newChainVerifier(cs, cs.Core)
	}

	cs.errBlocks, err = lru.New(dfltErrBlocks)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init lru")
		return nil
	}

	// init genesis block
	if _, err := cs.initGenesis(nil, !cfg.UseTestnet, cfg.EnableTestmode); err != nil {
		logger.Fatal().Err(err).Msg("failed to create a genesis block")
		panic("failed to init genesis block")
	}

	if ConsensusName() == consensus.ConsensusName[consensus.ConsensusDPOS] {
		top, err := cs.getVotes(types.VoteBP[2:], 1)
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
	if !pubNet && cfg.Blockchain.ZeroFee {
		fee.EnableZeroFee()
	}
	logger.Info().Bool("enablezerofee", fee.IsZeroFee()).Msg("fee")
	contract.PubNet = pubNet
	contract.StartLStateFactory()

	// init Debugger
	cs.initDebugger()

	return cs
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

// CDB returns cs.sdb as a consensus.ChainDbReader.
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
	if !cs.cfg.Blockchain.VerifyOnly {
		cs.chainManager.Start()
		cs.chainWorker.Start()
	} else {
		cs.chainVerifier.Start()
	}
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
	var val bool
	aopv := cs.recovered.Load()
	if aopv != nil {
		val = aopv.(bool)
	} else {
		panic("ChainService: recovered is nil")
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

		if cs.cfg.Blockchain.VerifyOnly {
			cs.chainVerifier.Request(&message.VerifyStart{}, context.Sender())
		}
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
		*message.ListEvents:
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
		"orphan": cs.op.curCnt,
	}
}

func (cs *ChainService) GetChainTree() ([]byte, error) {
	return cs.cdb.GetChainTree()
}

func (cs *ChainService) getVotes(id string, n uint32) (*types.VoteList, error) {
	switch ConsensusName() {
	case consensus.ConsensusName[consensus.ConsensusDPOS]:
		return system.GetVoteResult(cs.sdb, []byte(id), int(n))
	case consensus.ConsensusName[consensus.ConsensusRAFT]:
		//return cs.GetBPs()
		return nil, ErrNotSupportedConsensus
	default:
		return nil, ErrNotSupportedConsensus
	}
}

func (cs *ChainService) getAccountVote(ids []string, addr []byte) (*types.AccountVoteInfo, error) {
	if cs.GetType() != consensus.ConsensusDPOS {
		return nil, ErrNotSupportedConsensus
	}

	scs, err := cs.sdb.GetSystemAccountState()
	if err != nil {
		return nil, err
	}

	var voteInfo types.AccountVoteInfo

	for _, id := range ids {
		vote, err := system.GetVote(scs, addr, []byte(id))
		if err != nil {
			return nil, err
		}
		var candidates []string
		to := vote.GetCandidate()
		if len(to) == 0 {
			continue
		}
		if id == types.VoteBP[2:] {
			for offset := 0; offset < len(to); offset += system.PeerIDLength {
				candidates = append(candidates, types.EncodeB58(to[offset:offset+system.PeerIDLength]))
			}
		} else {
			err := json.Unmarshal(to, &candidates)
			if err != nil {
				return nil, err
			}
		}
		voteInfo.Voting = append(voteInfo.Voting, &types.VoteInfo{Id: id, Candidates: candidates})
	}

	return &voteInfo, nil
}

func (cs *ChainService) getStaking(addr []byte) (*types.Staking, error) {
	if cs.GetType() != consensus.ConsensusDPOS {
		return nil, ErrNotSupportedConsensus
	}

	scs, err := cs.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	if err != nil {
		return nil, err
	}
	namescs, err := cs.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte(types.AergoName)))
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
		stateDB = cs.sdb.GetStateDB()
	}
	return name.GetNameInfo(stateDB, qname)
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
		logger.Info().Str("hash", block.ID()).Str("prev", block.PrevID()).Uint64("bestno", cm.cdb.getBestBlockNo()).
			Uint64("no", block.GetHeader().GetBlockNo()).Bool("syncer", msg.IsSync).Msg("add block chainservice")

		var bstate *state.BlockState
		if msg.Bstate != nil {
			bstate = msg.Bstate.(*state.BlockState)
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

func getAddressNameResolved(sdb *state.ChainStateDB, account []byte) ([]byte, error) {
	if len(account) <= types.NameLength {
		scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte(types.AergoName)))
		if err != nil {
			logger.Error().Str("hash", enc.ToString(account)).Err(err).Msg("failed to get state for account")
			return nil, err
		}
		return name.GetAddress(scs, account), nil
	}
	return account, nil
}

func (cw *ChainWorker) Receive(context actor.Context) {
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
		address, err := getAddressNameResolved(cw.sdb, msg.Account)
		if err != nil {
			context.Respond(message.GetStateRsp{
				Account: msg.Account,
				State:   nil,
				Err:     err,
			})
			return
		}
		id := types.ToAccountID(address)
		accState, err := cw.sdb.GetStateDB().GetAccountState(id)
		if err != nil {
			logger.Error().Str("hash", enc.ToString(address)).Err(err).Msg("failed to get state for account")
		}
		context.Respond(message.GetStateRsp{
			Account: address,
			State:   accState,
			Err:     err,
		})
	case *message.GetStateAndProof:
		address, err := getAddressNameResolved(cw.sdb, msg.Account)
		if err != nil {
			context.Respond(message.GetStateAndProofRsp{
				StateProof: nil,
				Err:        err,
			})
			break
		}
		id := types.ToAccountID(address)
		stateProof, err := cw.sdb.GetStateDB().GetAccountAndProof(id[:], msg.Root, msg.Compressed)
		if err != nil {
			logger.Error().Str("hash", enc.ToString(address)).Err(err).Msg("failed to get state for account")
		}
		stateProof.Key = address
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
		address, err := getAddressNameResolved(cw.sdb, msg.Contract)
		if err != nil {
			context.Respond(message.GetABIRsp{
				ABI: nil,
				Err: err,
			})
			break
		}
		contractState, err := cw.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(address))
		if err == nil {
			abi, err := contract.GetABI(contractState)
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
		address, err := getAddressNameResolved(cw.sdb, msg.Contract)
		if err != nil {
			context.Respond(message.GetQueryRsp{Result: nil, Err: err})
			break
		}
		ctrState, err := cw.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(address))
		if err != nil {
			logger.Error().Str("hash", enc.ToString(address)).Err(err).Msg("failed to get state for contract")
			context.Respond(message.GetQueryRsp{Result: nil, Err: err})
		} else {
			bs := state.NewBlockState(cw.sdb.OpenNewStateDB(cw.sdb.GetRoot()))
			ret, err := contract.Query(address, bs, cw.cdb, ctrState, msg.Queryinfo)
			context.Respond(message.GetQueryRsp{Result: ret, Err: err})
		}
	case *message.GetStateQuery:
		var varProofs []*types.ContractVarProof
		var contractProof *types.AccountProof
		var err error

		address, err := getAddressNameResolved(cw.sdb, msg.ContractAddress)
		if err != nil {
			context.Respond(message.GetStateQueryRsp{
				Result: nil,
				Err:    err,
			})
			break
		}
		id := types.ToAccountID(address)
		contractProof, err = cw.sdb.GetStateDB().GetAccountAndProof(id[:], msg.Root, msg.Compressed)
		if err != nil {
			logger.Error().Str("hash", enc.ToString(address)).Err(err).Msg("failed to get state for account")
		} else if contractProof.Inclusion {
			contractTrieRoot := contractProof.State.StorageRoot
			for _, storageKey := range msg.StorageKeys {
				trieKey := common.Hasher([]byte(storageKey))
				varProof, err := cw.sdb.GetStateDB().GetVarAndProof(trieKey, contractTrieRoot, msg.Compressed)
				varProof.Key = storageKey
				varProofs = append(varProofs, varProof)
				if err != nil {
					logger.Error().Str("hash", enc.ToString(address)).Err(err).Msg("failed to get state variable in contract")
				}
			}
		}
		contractProof.Key = address
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
		info, err := cw.getAccountVote(msg.Ids, msg.Addr)
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
	case *message.ListEvents:
		events, err := cw.listEvents(msg.Filter)
		context.Respond(&message.ListEventsRsp{
			Events: events,
			Err:    err,
		})
	case *actor.Started, *actor.Stopping, *actor.Stopped, *component.CompStatReq: // donothing
	default:
		debug := fmt.Sprintf("[%s] Missed message. (%v) %s", cw.name, reflect.TypeOf(msg), msg)
		logger.Debug().Msg(debug)
	}
}
