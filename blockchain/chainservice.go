/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"os"
	"path"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/mempool"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

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
	logger = log.NewLogger("chain")
)

func NewChainService(cfg *cfg.Config, cc consensus.ChainConsensus, pool *mempool.MemPool) *ChainService {
	actor := &ChainService{
		ChainConsensus: cc,
		cfg:            cfg,
		cdb:            NewChainDB(),
		sdb:            state.NewStateDB(),
		op:             NewOrphanPool(),
	}
	Init(cfg.Blockchain.MaxBlockSize)
	if cc != nil {
		cc.SetStateDB(actor.sdb)
	}

	actor.validator = NewBlockValidator(actor.sdb)
	if pool != nil {
		pool.SetStateDb(actor.sdb)
	}
	actor.BaseComponent = component.NewBaseComponent(message.ChainSvc, actor, logger)

	return actor
}

func (cs *ChainService) initDB(dataDir string) error {
	// init chaindb
	if err := cs.cdb.Init(dataDir); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize chaindb")
		return err
	}

	// init statedb
	if err := cs.sdb.Init(dataDir); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize statedb")
		return err
	}
	return nil
}
func (cs *ChainService) BeforeStart() {

	if err := cs.initDB(cs.cfg.DataDir); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize DB")
	}

	// init genesis block
	if err := cs.initGenesis(nil); err != nil {
		logger.Fatal().Err(err).Msg("failed to genesis block")
	}

}

func (cs *ChainService) AfterStart() {
}

func (cs *ChainService) InitGenesisBlock(gb *types.Genesis, dataDir string) error {

	if err := cs.initDB(dataDir); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize DB")
		return err
	}
	err := cs.initGenesis(gb)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot initialize genesis block")
		return err
	}
	return nil
}
func (cs *ChainService) initGenesis(genesis *types.Genesis) error {
	gh, _ := cs.cdb.getHashByNo(0)
	if gh == nil || len(gh) == 0 {
		logger.Info().Uint64("nom", cs.cdb.latest).Msg("current latest")
		if cs.cdb.latest == 0 {
			if genesis == nil {
				genesis = types.GetDefaultGenesis()
			}
			err := cs.cdb.addGenesisBlock(types.GenesisToBlock(genesis))
			if err != nil {
				logger.Fatal().Err(err).Msg("cannot add genesisblock")
				return err
			}
			err = InitGenesisBPs(cs.sdb, genesis)
			if err != nil {
				logger.Fatal().Err(err).Msg("cannot set bp identifications")
				return err
			}
			err = cs.sdb.SetGenesis(genesis)
			if err != nil {
				logger.Fatal().Err(err).Msg("cannot set statedb of genesisblock")
				return err
			}
			logger.Info().Msg("genesis block is generated")
		}
	}
	gb, _ := cs.cdb.getBlockByNo(0)
	logger.Info().Str("genesis", enc.ToString(gb.Hash)).Msg("chain initialized")

	dbPath := path.Join(cs.cfg.DataDir, contract.DbName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	contract.DB = db.NewDB(db.BadgerImpl, dbPath)

	return nil
}

// Sync with peer
func (cs *ChainService) ChainSync(peerID peer.ID) {
	// handlt it like normal block (orphan)
	logger.Debug().Msg("Best Block Request")
	anchors := cs.getAnchorsFromHash(nil)
	hashes := make([]message.BlockHash, 0)
	for _, a := range anchors {
		hashes = append(hashes, message.BlockHash(a))
		logger.Debug().Str("hash", enc.ToString(a)).Msg("request blocks for sync")
	}
	cs.RequestTo(message.P2PSvc, &message.GetMissingBlocks{ToWhom: peerID, Hashes: hashes})
}

func (cs *ChainService) BeforeStop() {
	cs.CloseDB()

	cs.validator.Stop()
}

func (cs *ChainService) CloseDB() {
	if cs.sdb != nil {
		cs.sdb.Close()
	}
	if cs.cdb != nil {
		cs.cdb.Close()
	}
	if contract.DB != nil {
		contract.DB.Close()
	}
}

func (cs *ChainService) notifyBlock(block *types.Block) {
	cs.BaseComponent.RequestTo(message.P2PSvc,
		&message.NotifyNewBlock{
			BlockNo: block.Header.BlockNo,
			Block:   block,
		})
}

func (cs *ChainService) Receive(context actor.Context) {

	switch msg := context.Message().(type) {

	case *message.GetBestBlockNo:
		context.Respond(message.GetBestBlockNoRsp{
			BlockNo: cs.getBestBlockNo(),
		})
	case *message.GetBestBlock:
		block, err := cs.getBestBlock()
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
		bid := msg.Block.BlockID()
		logger.Debug().Str("hash", msg.Block.ID()).
			Uint64("blockNo", msg.Block.GetHeader().GetBlockNo()).Msg("add block chainservice")
		_, err := cs.getBlock(bid[:])
		if err == nil {
			logger.Debug().Str("hash", msg.Block.ID()).Msg("already exist")
		} else {
			block := msg.Block
			err := cs.addBlock(block, msg.Bstate, msg.PeerID)
			if err != nil {
				logger.Error().Err(err).Str("hash", msg.Block.ID()).Msg("failed add block")
			}
			context.Respond(message.AddBlockRsp{
				BlockNo:   block.GetHeader().GetBlockNo(),
				BlockHash: block.BlockHash(),
				Err:       err,
			})
		}
	case *message.MemPoolDelRsp:
		err := msg.Err
		if err != nil {
			logger.Error().Err(err).Msg("failed to remove txs from mempool")
		}
	case *message.GetState:
		id := types.ToAccountID(msg.Account)
		state, err := cs.sdb.GetAccountStateClone(id)
		if err != nil {
			logger.Error().Str("hash", enc.ToString(msg.Account)).Err(err).Msg("failed to get state for account")
		}
		context.Respond(message.GetStateRsp{
			State: state,
			Err:   err,
		})
	case *message.GetMissing:
		stopHash := msg.StopHash
		hashes := msg.Hashes
		mhashes, mnos := cs.handleMissing(stopHash, hashes)
		context.Respond(message.GetMissingRsp{
			Hashes:   mhashes,
			Blocknos: mnos,
		})
	case *message.GetTx:
		tx, txIdx, err := cs.getTx(msg.TxHash)
		context.Respond(message.GetTxRsp{
			Tx:    tx,
			TxIds: txIdx,
			Err:   err,
		})
	case *message.GetReceipt:
		receipt, err := contract.GetReceipt(msg.TxHash)
		context.Respond(message.GetReceiptRsp{
			Receipt: receipt,
			Err:     err,
		})
	case *message.GetABI:
		contractState, err := cs.sdb.OpenContractStateAccount(types.ToAccountID(msg.Contract))
		if err == nil {
			abi, err := contract.GetABI(contractState, msg.Contract)
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

		state, err := cs.sdb.OpenContractStateAccount(types.ToAccountID(msg.Contract))
		if err != nil {
			logger.Error().Str("hash", enc.ToString(msg.Contract)).Err(err).Msg("failed to get state for contract")
			context.Respond(message.GetQueryRsp{Result: nil, Err: err})
		} else {
			ret, err := contract.Query(msg.Contract, state, msg.Queryinfo)
			context.Respond(message.GetQueryRsp{Result: ret, Err: err})
		}
	case *message.SyncBlockState:
		cs.checkBlockHandshake(msg.PeerID, msg.BlockNo, msg.BlockHash)
	case *message.GetElected:
		top, err := cs.getVotes(msg.N)
		context.Respond(&message.GetElectedRsp{
			Top: top,
			Err: err,
		})

	case actor.SystemMessage,
		actor.AutoReceiveMessage,
		actor.NotInfluenceReceiveTimeout:
		//logger.Debugf("Received message. (%v) %s", reflect.TypeOf(msg), msg)

	default:
		//logger.Debugf("Missed message. (%v) %s", reflect.TypeOf(msg), msg)
	}
}

func (cs *ChainService) Statics() *map[string]interface{} {
	return &map[string]interface{}{
		"orphan": cs.op.curCnt,
	}
}

func (cs *ChainService) GetChainTree() ([]byte, error) {
	return cs.cdb.GetChainTree()
}
