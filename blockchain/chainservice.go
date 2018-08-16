/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (

	//"reflect"

	"github.com/aergoio/aergo-actor/actor"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

type ChainService struct {
	*component.BaseComponent
	consensus.ChainInfo

	cfg *cfg.Config
	cdb *ChainDB
	sdb *state.ChainStateDB
	op  *OrphanPool

	cac chan consensus.ChainInfo
}

var _ component.IComponent = (*ChainService)(nil)
var (
	logger = log.NewLogger("chain")
)

func NewChainService(cfg *cfg.Config) *ChainService {
	return &ChainService{
		BaseComponent: component.NewBaseComponent(message.ChainSvc, logger, cfg.EnableDebugMsg),
		cfg:           cfg,
		cac:           make(chan consensus.ChainInfo),
		cdb:           NewChainDB(),
		sdb:           state.NewStateDB(),
		op:            NewOrphanPool(),
	}
}

func (cs *ChainService) receiveChainInfo() {
	// Get a Validation interface from the consensus service
	cs.ChainInfo = <-cs.cac
	cs.cdb.ChainInfo = cs.ChainInfo
	// Disable the channel. warning: don't read from this channel!!!
	cs.cac = nil
}

func (cs *ChainService) Start() {
	cs.BaseComponent.Start(cs)

	cs.receiveChainInfo()

	// init chaindb
	if err := cs.cdb.Init(cs.cfg.DataDir); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize chaindb")
	}

	// init statedb
	if err := cs.sdb.Init(cs.cfg.DataDir); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize statedb")
	}

	// init genesis block
	if err := cs.initGenesis(cs.cfg.GenesisSeed); err != nil {
		logger.Fatal().Err(err).Msg("failed to genesis block")
	}
}

func (cs *ChainService) initGenesis(seed int64) error {
	gh, _ := cs.cdb.getHashByNo(0)
	if gh == nil || len(gh) == 0 {
		if cs.cdb.latest == 0 {
			genesisBlock := cs.cdb.generateGenesisBlock(seed)
			err := cs.sdb.SetGenesis(genesisBlock)
			if err != nil {
				return err
			}
		}
	}
	gb, _ := cs.cdb.getBlockByNo(0)
	logger.Info().Int64("seed", gb.Header.Timestamp).Str("genesis", EncodeB64(gb.Hash)).Msg("chain initialized")

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
		logger.Debug().Str("hash", EncodeB64(a)).Msg("request blocks for sync")
	}
	cs.Hub().Request(message.P2PSvc, &message.GetMissingBlocks{ToWhom: peerID, Hashes: hashes}, cs)
}

// SetValidationAPI send the Validation v of the chosen Consensus to ChainService cs.
func (cs *ChainService) SendChainInfo(ca consensus.ChainInfo) {
	cs.cac <- ca
}

func (cs *ChainService) Stop() {
	if cs.sdb != nil {
		cs.sdb.Close()
	}
	if cs.cdb != nil {
		cs.cdb.Close()
	}
	cs.BaseComponent.Stop()
}

func (cs *ChainService) notifyBlock(block *types.Block) {
	cs.BaseComponent.Hub().Request(message.P2PSvc,
		&message.NotifyNewBlock{
			BlockNo: block.Header.BlockNo,
			Block:   block.Clone(),
		}, cs)
	// if err != nil {
	// 	logger.Info("failed to notify block:", block.Header.BlockNo, ToJSON(block))
	// }
}

func (cs *ChainService) Receive(context actor.Context) {
	cs.BaseComponent.Receive(context)

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
			logger.Error().Err(err).Str("hash", EncodeB64(msg.BlockHash)).Msg("failed to get block")
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
		bid := types.ToBlockID(msg.Block.GetHash())
		logger.Debug().Str("hash", msg.Block.ID()).
			Uint64("blockNo", msg.Block.GetHeader().GetBlockNo()).Msg("add block chainservice")
		_, err := cs.getBlock(bid[:])
		if err == nil {
			logger.Debug().Str("hash", msg.Block.ID()).Msg("already exist")
		} else {
			block := msg.Block.Clone()
			err := cs.addBlock(block, msg.PeerID)
			if err != nil {
				logger.Error().Err(err).Str("hash", msg.Block.ID()).Msg("failed add block")
			}
			context.Respond(message.AddBlockRsp{
				BlockNo:   block.GetHeader().GetBlockNo(),
				BlockHash: block.GetHash(),
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
		state, err := cs.sdb.GetAccountState(id)
		if err != nil {
			logger.Error().Str("hash", EncodeB64(msg.Account)).Err(err).Msg("failed to get state for account")
		}
		res := state.Clone()
		context.Respond(message.GetStateRsp{
			State: res,
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
	case *message.SyncBlockState:
		cs.checkBlockHandshake(msg.PeerID, msg.BlockNo, msg.BlockHash)
	case actor.SystemMessage,
		actor.AutoReceiveMessage,
		actor.NotInfluenceReceiveTimeout:
		//logger.Debugf("Received message. (%v) %s", reflect.TypeOf(msg), msg)
	default:
		//logger.Debugf("Missed message. (%v) %s", reflect.TypeOf(msg), msg)
	}
}

func (cs *ChainService) GetChainTree() ([]byte, error) {
	return cs.cdb.GetChainTree()
}
