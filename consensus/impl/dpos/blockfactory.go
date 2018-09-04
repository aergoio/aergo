/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package dpos

import (
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/consensus/chain"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	slotQueueMax = 100
)

type txExec struct {
	blockState *state.BlockState
	block      *types.Block
}

func newTxExec(blockNo types.BlockNo, prevHash types.BlockID) chain.TxOp {
	// Block hash not determined yet
	return &txExec{blockState: state.NewBlockState(blockNo, types.BlockID{}, prevHash)}
}

func (te *txExec) Apply(tx *types.Tx) (*state.BlockState, error) {
	return te.blockState, nil
}

// BlockFactory is the main data structure for DPoS block factory.
type BlockFactory struct {
	*component.ComponentHub
	jobQueue         chan interface{}
	workerQueue      chan *bpInfo
	bpTimeoutC       chan interface{}
	quit             <-chan interface{}
	maxBlockBodySize int
	ID               string
	privKey          crypto.PrivKey
	txOp             chain.TxOp
}

// NewBlockFactory returns a new BlockFactory
func NewBlockFactory(hub *component.ComponentHub, id peer.ID, privKey crypto.PrivKey, quitC <-chan interface{}) *BlockFactory {
	bf := &BlockFactory{
		ComponentHub:     hub,
		jobQueue:         make(chan interface{}, slotQueueMax),
		workerQueue:      make(chan *bpInfo),
		bpTimeoutC:       make(chan interface{}, 1),
		maxBlockBodySize: chain.MaxBlockBodySize(),
		ID:               enc.ToString([]byte(id)),
		privKey:          privKey,
		quit:             quitC,
	}

	bf.txOp = chain.NewCompTxOp(
		// block size limit check
		chain.NewBlockLimitOp(bf.maxBlockBodySize),
		// timeout check
		chain.TxOpFn(func(txIn *types.Tx) (*state.BlockState, error) {
			return nil, bf.checkBpTimeout()
		}),
	)

	return bf
}

// Start run a DPoS block factory service.
func (bf *BlockFactory) Start() {
	go func() {
		go bf.worker()
		go bf.controller()
	}()
}

// JobQueue returns the queue for block production triggering.
func (bf *BlockFactory) JobQueue() chan<- interface{} {
	return bf.jobQueue
}

func (bf *BlockFactory) controller() {
	defer shutdownMsg("block factory controller")

	beginBlock := func(bpi *bpInfo) error {
		// This is only for draining an unconsumed message, which means
		// the previous block is generated within timeout. This code
		// is needed since an empty block will be generated without it.
		if err := bf.checkBpTimeout(); err == chain.ErrQuit {
			return err
		}

		timeLeft := bpi.slot.RemainingTimeMS()
		if timeLeft <= 0 {
			return chain.ErrTimeout{Kind: "slot", Timeout: timeLeft}
		}

		select {
		case bf.workerQueue <- bpi:
		default:
			logger.Error().Msgf(
				"skip block production for the slot %v (best block: %v) due to a pending job",
				spew.Sdump(bpi.slot), bpi.bestBlock.ID())
		}
		return nil
	}

	notifyBpTimeout := func(bpi *bpInfo) {
		timeout := bpi.slot.GetBpTimeout()
		logger.Debug().Int64("timeout", timeout).Msg("block production timeout")
		time.Sleep(time.Duration(timeout) * time.Millisecond)
		// TODO: skip when the triggered block has already been genearted!
		bf.bpTimeoutC <- struct{}{}
	}

	for {
		select {
		case info := <-bf.jobQueue:
			bpi := info.(*bpInfo)
			logger.Debug().Msgf("received bpInfo: %v %v",
				log.DoLazyEval(func() string { return bpi.bestBlock.ID() }),
				log.DoLazyEval(func() string { return spew.Sdump(bpi.slot) }))

			err := beginBlock(bpi)
			if err == chain.ErrQuit {
				return
			} else if err != nil {
				logger.Debug().Err(err).Msg("skip block production")
				continue
			}

			notifyBpTimeout(bpi)

		case <-bf.quit:
			return
		}
	}
}

func (bf *BlockFactory) worker() {
	defer shutdownMsg("block factory worker")

	lpbNo := types.BlockNo(0)

	for {
		select {
		case bpi := <-bf.workerQueue:
			block, blockState, err := bf.generateBlock(bpi, lpbNo)
			if err == chain.ErrQuit {
				return
			} else if err != nil {
				logger.Info().Err(err).Msg("failed to produce block")
				continue
			}

			err = chain.ConnectBlock(bf, block, blockState)
			if err == nil {
				lpbNo = block.BlockNo()
			} else {
				logger.Error().Msg(err.Error())
			}

		case <-bf.quit:
			return
		}
	}
}

func (bf *BlockFactory) generateBlock(bpi *bpInfo, lpbNo types.BlockNo) (*types.Block, *state.BlockState, error) {
	/*
		txOp := chain.NewCompTxOp(
			bf.txOp,
			newTxExec(bpi.bestBlock.BlockNo()+1, bpi.bestBlock.BlockID()),
		)
	*/
	txOp := bf.txOp

	block, blockState, err := chain.GenerateBlock(bf, bpi.bestBlock, txOp, bpi.slot.UnixNano())
	if err != nil {
		return nil, nil, err
	}

	block.SetConfirms(block.BlockNo() - lpbNo)

	if err := block.Sign(bf.privKey); err != nil {
		return nil, nil, err
	}

	logger.Info().
		Str("BP", bf.ID).Str("id", block.ID()).
		Uint64("no", block.BlockNo()).Uint64("confirms", block.Confirms()).
		Uint64("lpb", lpbNo).
		Msg("block produced")

	return block, blockState, nil
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

func shutdownMsg(m string) {
	logger.Info().Msgf("shutdown initiated. stop the %s", m)
}
