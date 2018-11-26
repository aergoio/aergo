/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"bytes"
	"errors"
	"runtime"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/contract/system"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

// ChainService manage connectivity of blocks
type ChainService struct {
	*component.BaseComponent
	consensus.ChainConsensus

	cfg *cfg.Config
	cdb *ChainDB
	sdb *state.ChainStateDB
	op  *OrphanPool

	validator *BlockValidator
}

var (
	logger           = log.NewLogger("chain")
	genesisBlock     *types.Block
	initialBestBlock *types.Block

	ErrBlockExist = errors.New("error! block already exist")
)

// NewChainService create instance of ChainService
func NewChainService(cfg *cfg.Config) *ChainService {
	cs := &ChainService{
		cfg: cfg,
		cdb: NewChainDB(),
		sdb: state.NewChainStateDB(),
		op:  NewOrphanPool(),
	}
	if err := Init(cfg.Blockchain.MaxBlockSize,
		cfg.Blockchain.CoinbaseAccount,
		types.DefaultCoinbaseFee,
		cfg.Consensus.EnableBp,
		cfg.Blockchain.MaxAnchorCount,
		cfg.Blockchain.UseFastSyncer); err != nil {
		logger.Error().Err(err).Msg("failed to init chainservice")
		panic("invalid config: blockchain")
	}

	cs.validator = NewBlockValidator(cs.sdb)
	cs.BaseComponent = component.NewBaseComponent(message.ChainSvc, cs, logger)

	if err := cs.initDB(cs.cfg.DbType, cs.cfg.DataDir); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize DB")
	}

	// init genesis block
	if _, err := cs.initGenesis(nil); err != nil {
		logger.Fatal().Err(err).Msg("failed to genesis block")
	}

	return cs
}

func (cs *ChainService) initDB(dbType string, dataDir string) error {
	// init chaindb
	if err := cs.cdb.Init(dbType, dataDir); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize chaindb")
		return err
	}

	// init statedb
	bestBlock, err := cs.GetBestBlock()
	if err != nil {
		return err
	}
	if err := cs.sdb.Init(dbType, dataDir, bestBlock, cs.cfg.EnableTestmode); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize statedb")
		return err
	}

	contract.LoadDatabase(dataDir)

	return nil
}

// SDB returns cs.sdb.
func (cs *ChainService) SDB() *state.ChainStateDB {
	return cs.sdb
}

// CDBReader returns cs.sdb as a consensus.ChainDbReader.
func (cs *ChainService) CDBReader() consensus.ChainDbReader {
	return cs.cdb
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

// InitGenesisBlock initialize chain database and generate specified genesis block if necessary
func (cs *ChainService) InitGenesisBlock(gb *types.Genesis, dbType string, dataDir string) error {

	if err := cs.initDB(dbType, dataDir); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize DB")
		return err
	}
	_, err := cs.initGenesis(gb)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot initialize genesis block")
		return err
	}
	return nil
}

func (cs *ChainService) initGenesis(genesis *types.Genesis) (*types.Block, error) {
	gh, _ := cs.cdb.getHashByNo(0)
	if gh == nil || len(gh) == 0 {
		logger.Info().Uint64("nom", cs.cdb.latest).Msg("current latest")
		if cs.cdb.latest == 0 {
			if genesis == nil {
				genesis = types.GetDefaultGenesis()
			}

			err := InitGenesisBPs(cs.sdb.GetStateDB(), genesis)
			if err != nil {
				logger.Fatal().Err(err).Msg("cannot set bp identifications")
				return nil, err
			}
			err = cs.sdb.SetGenesis(genesis)
			if err != nil {
				logger.Fatal().Err(err).Msg("cannot set statedb of genesisblock")
				return nil, err
			}

			err = cs.cdb.addGenesisBlock(genesis)
			if err != nil {
				logger.Fatal().Err(err).Msg("cannot add genesisblock")
				return nil, err
			}

			logger.Info().Msg("genesis block is generated")
		}
	}
	genesisBlock, _ := cs.cdb.GetBlockByNo(0)

	logger.Info().Str("genesis", enc.ToString(genesisBlock.Hash)).
		Str("stateroot", enc.ToString(genesisBlock.GetHeader().GetBlocksRootHash())).Msg("chain initialized")

	return genesisBlock, nil
}

// ChainSync synchronize with peer
func (cs *ChainService) ChainSync(peerID peer.ID, remoteBestHash []byte) {
	// handlt it like normal block (orphan)
	logger.Debug().Msg("Best Block Request")
	anchors := cs.getAnchorsFromHash(remoteBestHash)
	hashes := make([]message.BlockHash, 0)
	for _, a := range anchors {
		hashes = append(hashes, message.BlockHash(a))
		logger.Debug().Str("hash", enc.ToString(a)).Msg("request blocks for sync")
	}
	cs.RequestTo(message.P2PSvc, &message.GetMissingBlocks{ToWhom: peerID, Hashes: hashes})
}

// BeforeStop close chain database and stop BlockValidator
func (cs *ChainService) BeforeStop() {
	cs.CloseDB()

	cs.validator.Stop()
}

// CloseDB close chain database
func (cs *ChainService) CloseDB() {
	if cs.sdb != nil {
		cs.sdb.Close()
	}
	if cs.cdb != nil {
		cs.cdb.Close()
	}
	contract.CloseDatabase()
}

func (cs *ChainService) notifyBlock(block *types.Block) {
	cs.BaseComponent.RequestTo(message.P2PSvc,
		&message.NotifyNewBlock{
			BlockNo: block.Header.BlockNo,
			Block:   block,
		})
}

// Receive actor message
func (cs *ChainService) Receive(context actor.Context) {

	switch msg := context.Message().(type) {

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
	case *message.GetBlock:
		bid := types.ToBlockID(msg.BlockHash)
		block, err := cs.getBlock(bid[:])
		if err != nil {
			logger.Debug().Err(err).Str("hash", enc.ToString(msg.BlockHash)).Msg("block not found")
		}
		context.Respond(message.GetBlockRsp{
			Block: block,
			Err:   err,
		})
	case *message.GetBlockByNo:
		block, err := cs.getBlockByNo(msg.BlockNo)
		if err != nil {
			logger.Error().Err(err).Uint64("blockNo", msg.BlockNo).Msg("failed to get block by no")
		}
		context.Respond(message.GetBlockByNoRsp{
			Block: block,
			Err:   err,
		})
	case *message.AddBlock:
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		bid := msg.Block.BlockID()
		block := msg.Block
		logger.Debug().Str("hash", msg.Block.ID()).
			Uint64("blockNo", msg.Block.GetHeader().GetBlockNo()).Bool("syncer", msg.IsSync).Msg("add block chainservice")
		_, err := cs.getBlock(bid[:])
		if err == nil {
			logger.Debug().Str("hash", msg.Block.ID()).Msg("already exist")
			err = ErrBlockExist
		} else {
			var bstate *state.BlockState
			if msg.Bstate != nil {
				bstate = msg.Bstate.(*state.BlockState)
			}
			err = cs.addBlock(block, bstate, msg.PeerID)
			if err != nil && err != ErrBlockOrphan {
				logger.Error().Err(err).Str("hash", msg.Block.ID()).Msg("failed add block")
			}
		}

		rsp := message.AddBlockRsp{
			BlockNo:   block.GetHeader().GetBlockNo(),
			BlockHash: block.BlockHash(),
			Err:       err,
		}

		if msg.IsSync {
			cs.RequestTo(message.SyncerSvc, &rsp)
		} else {
			context.Respond(rsp)
		}

		cs.Hub().Tell(message.RPCSvc, block)
	case *message.MemPoolDelRsp:
		err := msg.Err
		if err != nil {
			logger.Error().Err(err).Msg("failed to remove txs from mempool")
		}
	case *message.GetState:
		id := types.ToAccountID(msg.Account)
		state, err := cs.sdb.GetStateDB().GetAccountState(id)
		if err != nil {
			logger.Error().Str("hash", enc.ToString(msg.Account)).Err(err).Msg("failed to get state for account")
		}
		context.Respond(message.GetStateRsp{
			State: state,
			Err:   err,
		})
	case *message.GetStateAndProof:
		id := types.ToAccountID(msg.Account)
		stateProof, err := cs.sdb.GetStateDB().GetStateAndProof(id[:], msg.Root, msg.Compressed)
		if err != nil {
			logger.Error().Str("hash", enc.ToString(msg.Account)).Err(err).Msg("failed to get state for account")
		}
		context.Respond(message.GetStateAndProofRsp{
			StateProof: stateProof,
			Err:        err,
		})
	case *message.GetMissing:
		stopHash := msg.StopHash
		hashes := msg.Hashes
		topHash, topNo, stopNo := cs.handleMissing(stopHash, hashes)
		context.Respond(message.GetMissingRsp{
			TopMatched: topHash,
			TopNumber:  topNo,
			StopNumber: stopNo,
		})
	case *message.GetAnchors:
		anchor, lastNo, err := cs.getAnchorsNew()
		context.Respond(message.GetAnchorsRsp{
			Hashes: anchor,
			LastNo: lastNo,
			Err:    err,
		})
	case *message.GetAncestor:
		hashes := msg.Hashes
		ancestor, err := cs.findAncestor(hashes)

		context.Respond(message.GetAncestorRsp{
			Ancestor: ancestor,
			Err:      err,
		})
	case *message.GetTx:
		tx, txIdx, err := cs.getTx(msg.TxHash)
		context.Respond(message.GetTxRsp{
			Tx:    tx,
			TxIds: txIdx,
			Err:   err,
		})
	case *message.GetReceipt:
		receipt, err := cs.getReceipt(msg.TxHash)
		context.Respond(message.GetReceiptRsp{
			Receipt: receipt,
			Err:     err,
		})
	case *message.GetABI:
		contractState, err := cs.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(msg.Contract))
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
		ctrState, err := cs.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(msg.Contract))
		if err != nil {
			logger.Error().Str("hash", enc.ToString(msg.Contract)).Err(err).Msg("failed to get state for contract")
			context.Respond(message.GetQueryRsp{Result: nil, Err: err})
		} else {
			bs := state.NewBlockState(cs.sdb.OpenNewStateDB(cs.sdb.GetRoot()))
			ret, err := contract.Query(msg.Contract, bs, ctrState, msg.Queryinfo)
			context.Respond(message.GetQueryRsp{Result: ret, Err: err})
		}
	case *message.GetStateQuery:
		var varProof *types.ContractVarProof
		var contractProof *types.StateProof
		var err error

		id := types.ToAccountID(msg.ContractAddress)
		contractProof, err = cs.sdb.GetStateDB().GetStateAndProof(id[:], msg.Root, msg.Compressed)
		if err != nil {
			logger.Error().Str("hash", enc.ToString(msg.ContractAddress)).Err(err).Msg("failed to get state for account")
		} else if contractProof.Inclusion {
			contractTrieRoot := contractProof.State.StorageRoot
			varId := bytes.NewBufferString("_")
			varId.WriteString(msg.VarName)
			varId.WriteString(msg.VarIndex)
			varTrieKey := common.Hasher(varId.Bytes())
			varProof, err = cs.sdb.GetStateDB().GetVarAndProof(varTrieKey, contractTrieRoot, msg.Compressed)
			if err != nil {
				logger.Error().Str("hash", enc.ToString(msg.ContractAddress)).Err(err).Msg("failed to get state variable in contract")
			}
		}
		stateQuery := &types.StateQueryProof{
			ContractProof: contractProof,
			VarProof:      varProof,
		}
		context.Respond(message.GetStateQueryRsp{
			Result: stateQuery,
			Err:    err,
		})
	case *message.SyncBlockState:
		cs.checkBlockHandshake(msg.PeerID, msg.BlockNo, msg.BlockHash)
	case *message.GetElected:
		top, err := cs.getVotes(msg.N)
		context.Respond(&message.GetVoteRsp{
			Top: top,
			Err: err,
		})
	case *message.GetVote:
		top, err := cs.getVote(msg.Addr)
		context.Respond(&message.GetVoteRsp{
			Top: top,
			Err: err,
		})
	case *message.GetStaking:
		staking, err := cs.getStaking(msg.Addr)
		context.Respond(&message.GetStakingRsp{
			Staking: staking,
			Err:     err,
		})

	case actor.SystemMessage,
		actor.AutoReceiveMessage,
		actor.NotInfluenceReceiveTimeout:
		//logger.Debugf("Received message. (%v) %s", reflect.TypeOf(msg), msg)

	default:
		//logger.Debugf("Missed message. (%v) %s", reflect.TypeOf(msg), msg)
	}
}

func (cs *ChainService) Statistics() *map[string]interface{} {
	return &map[string]interface{}{
		"orphan": cs.op.curCnt,
	}
}

func (cs *ChainService) GetChainTree() ([]byte, error) {
	return cs.cdb.GetChainTree()
}

func (cs *ChainService) getVotes(n int) (*types.VoteList, error) {
	scs, err := cs.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	if err != nil {
		return nil, err
	}
	return system.GetVoteResult(scs, n)
}

func (cs *ChainService) getVote(addr []byte) (*types.VoteList, error) {
	scs, err := cs.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	if err != nil {
		return nil, err
	}
	var voteList types.VoteList
	var tmp []*types.Vote
	voteList.Votes = tmp
	vote, err := system.GetVote(scs, addr)
	if err != nil {
		return nil, err
	}
	to := vote.GetCandidate()
	for offset := 0; offset < len(to); offset += system.PeerIDLength {
		vote := &types.Vote{
			Candidate: to[offset : offset+system.PeerIDLength],
			Amount:    vote.GetAmount(),
		}
		voteList.Votes = append(voteList.Votes, vote)
	}
	return &voteList, nil
}

func (cs *ChainService) getStaking(addr []byte) (*types.Staking, error) {
	scs, err := cs.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	if err != nil {
		return nil, err
	}
	staking, err := system.GetStaking(scs, addr)
	if err != nil {
		return nil, err
	}
	return staking, nil
}
