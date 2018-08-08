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
	"github.com/aergoio/aergo/pkg/log"
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
	logger = log.NewLogger(log.ChainSvc)
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
		logger.Fatal("failed to initialize chaindb: ", err)
	}

	// init statedb
	if err := cs.sdb.Init(cs.cfg.DataDir); err != nil {
		logger.Fatal("failed to initialize statedb: ", err)
	}

	// init genesis block
	if err := cs.initGenesis(cs.cfg.GenesisSeed); err != nil {
		logger.Fatal("failed to initialize genesis block: ", err)
	}

	return
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
	logger.Infof("chain initialized: seed=%v, genesis=%s",
		gb.Header.Timestamp, EncodeB64(gb.Hash))
	return nil
}

// Sync with peer
func (cs *ChainService) ChainSync(peerID peer.ID) {
	// handlt it like normal block (orphan)
	logger.Debugf("Best Block Request")
	anchors := cs.getAnchorsFromHash(nil)
	hashes := make([]message.BlockHash, 0)
	for _, a := range anchors {
		hashes = append(hashes, message.BlockHash(a))
		logger.Debugf("request blocks for sync: (%v)")
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
			logger.Errorf("failed to get best block. error %v", err)
		}
		res := block.Clone()
		context.Respond(message.GetBestBlockRsp{
			Block: res,
			Err:   err,
		})
	case *message.GetBlock:
		bkey := types.ToBlockKey(msg.BlockHash)
		block, err := cs.getBlock(bkey[:])
		if err != nil {
			logger.Errorf("failed to get block. block %v, error %v", bkey, err)
		}
		res := block.Clone()
		context.Respond(message.GetBlockRsp{
			Block: res,
			Err:   err,
		})
	case *message.GetBlockByNo:
		block, err := cs.getBlockByNo(msg.BlockNo)
		if err != nil {
			logger.Errorf("failed to get block by no. block no %v, error %v", msg.BlockNo, err)
		}
		res := block.Clone()
		context.Respond(message.GetBlockByNoRsp{
			Block: res,
			Err:   err,
		})
	case *message.AddBlock:
		bkey := types.ToBlockKey(msg.Block.GetHash())
		logger.Debugf("Add Block chainservice %v", bkey)
		_, err := cs.getBlock(bkey[:])
		if err == nil {
			logger.Debugf("already exist %v", bkey)
		} else {
			block := msg.Block.Clone()
			err := cs.addBlock(block, msg.PeerID)
			context.Respond(message.AddBlockRsp{
				BlockNo:   block.GetHeader().GetBlockNo(),
				BlockHash: block.GetHash(),
				Err:       err,
			})
		}
	case *message.MemPoolDelRsp:
		err := msg.Err
		if err != nil {
			logger.Debug("failed to remove txs from mempool: ", err)
		}
	case *message.GetState:
		akey := types.ToAccountKey(msg.Account)
		state, err := cs.sdb.GetAccountState(akey)
		if err != nil {
			logger.Debugf("failed to get state for account %v, error %v", akey, err)
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
