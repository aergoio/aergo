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

	cc chan consensus.ChainConsensus

	votes map[string]uint64 //candidate, sum of votes
}

var (
	logger = log.NewLogger("chain")
)

func NewChainService(cfg *cfg.Config) *ChainService {
	actor := &ChainService{
		cfg: cfg,
		cc:  make(chan consensus.ChainConsensus),
		cdb: NewChainDB(),
		sdb: state.NewStateDB(),
		op:  NewOrphanPool(),

		votes: map[string]uint64{},
	}
	actor.BaseComponent = component.NewBaseComponent(message.ChainSvc, actor, logger)

	return actor
}

func (cs *ChainService) receiveChainInfo() {
	// Get a Validation interface from the consensus service
	cs.ChainConsensus = <-cs.cc
	cs.cdb.ChainConsensus = cs.ChainConsensus
	// Disable the channel. warning: don't read from this channel!!!
	cs.cc = nil
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
	if err := cs.initGenesis(cs.cfg.GenesisSeed); err != nil {
		logger.Fatal().Err(err).Msg("failed to genesis block")
	}

	if err := cs.loadGovernace(); err != nil {
		logger.Fatal().Err(err).Msg("failed to load governance")
	}
}

func (cs *ChainService) AfterStart() {
	cs.receiveChainInfo()
}

func (cs *ChainService) InitGenesisBlock(gb *types.Genesis, dataDir string) error {

	if err := cs.initDB(dataDir); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize DB")
		return err
	}
	return cs.initGenesis(gb.Timestamp)
}
func (cs *ChainService) initGenesis(seed int64) error {
	gh, _ := cs.cdb.getHashByNo(0)
	if gh == nil || len(gh) == 0 {
		if cs.cdb.latest == 0 {
			genesisBlock, err := cs.cdb.generateGenesisBlock(seed)
			if err != nil {
				return err
			}
			err = cs.sdb.SetGenesis(genesisBlock)
			if err != nil {
				return err
			}
		}
	}
	gb, _ := cs.cdb.getBlockByNo(0)
	logger.Info().Int64("seed", gb.Header.Timestamp).Str("genesis", enc.ToString(gb.Hash)).Msg("chain initialized")

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

// SetValidationAPI send the Validation v of the chosen Consensus to ChainService cs.
func (cs *ChainService) SendChainInfo(ca consensus.ChainConsensus) {
	cs.cc <- ca
}

func (cs *ChainService) BeforeStop() {
	if cs.sdb != nil {
		cs.sdb.Close()
	}
	if cs.cdb != nil {
		cs.cdb.Close()
	}

	contract.DB.Close()
}

func (cs *ChainService) notifyBlock(block *types.Block) {
	cs.BaseComponent.RequestTo(message.P2PSvc,
		&message.NotifyNewBlock{
			BlockNo: block.Header.BlockNo,
			Block:   block.Clone(),
		})
	// if err != nil {
	// 	logger.Info("failed to notify block:", block.Header.BlockNo, ToJSON(block))
	// }
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
		res := block.Clone()
		context.Respond(message.GetBestBlockRsp{
			Block: res,
			Err:   err,
		})
	case *message.GetBlock:
		bid := types.ToBlockID(msg.BlockHash)
		block, err := cs.getBlock(bid[:])
		if err != nil {
			logger.Debug().Err(err).Str("hash", enc.ToString(msg.BlockHash)).Msg("block not found")
		}
		res := block.Clone()
		context.Respond(message.GetBlockRsp{
			Block: res,
			Err:   err,
		})
	case *message.GetBlockByNo:
		block, err := cs.getBlockByNo(msg.BlockNo)
		if err != nil {
			logger.Error().Err(err).Uint64("blockNo", msg.BlockNo).Msg("failed to get block by no")
		}
		res := block.Clone()
		context.Respond(message.GetBlockByNoRsp{
			Block: res,
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
			block := msg.Block.Clone()
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
		abi, err := contract.GetABI(msg.Contract)
		context.Respond(message.GetABIRsp{
			ABI: abi,
			Err: err,
		})
	case *message.SyncBlockState:
		cs.checkBlockHandshake(msg.PeerID, msg.BlockNo, msg.BlockHash)
	case *message.GetElected:
		top := cs.getVotes(msg.N)
		context.Respond(&message.GetElectedRsp{
			Top: top,
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
