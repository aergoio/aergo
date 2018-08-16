/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package dpos

import (
	"errors"
	"time"

	"github.com/aergoio/aergo/consensus/util"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	slotQueueMax = 100
)

var errBpTimeout = errors.New("block production timeout")

// BlockFactory is the main data structure for DPoS block factory.
type BlockFactory struct {
	*component.ComponentHub
	jobQueue         chan interface{}
	workerQueue      chan *bpInfo
	bpTimeoutC       chan interface{}
	quit             <-chan interface{}
	maxBlockBodySize int
	sID              string
	privKey          crypto.PrivKey
	txOp             util.TxOp
}

// NewBlockFactory returns a new BlockFactory
func NewBlockFactory(hub *component.ComponentHub, id peer.ID, privKey crypto.PrivKey, quitC <-chan interface{}) *BlockFactory {
	bf := &BlockFactory{
		ComponentHub:     hub,
		jobQueue:         make(chan interface{}, slotQueueMax),
		workerQueue:      make(chan *bpInfo),
		bpTimeoutC:       make(chan interface{}, 1),
		maxBlockBodySize: util.MaxBlockBodySize(),
		sID:              id.Pretty(),
		privKey:          privKey,
		quit:             quitC,
	}

	bf.txOp = util.NewCompTxOp(
		// block size limit check
		util.NewBlockLimitOp(bf.maxBlockBodySize),
		// timeout check
		func(txIn *types.Tx) error {
			return bf.checkBpTimeout()
		},
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
	beginBlock := func(bpi *bpInfo) error {
		// This is only for draining an unconsumed message, which means
		// the previous block is generated within timeout. This code
		// is needed since an empty block will be generated without it.
		if err := bf.checkBpTimeout(); err == util.ErrQuit {
			return err
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
				log.DoLazyEval(func() string {
					return bpi.bestBlock.ID()
				}),
				log.DoLazyEval(func() string {
					return spew.Sdump(bpi.slot)
				}))

			if err := beginBlock(bpi); err == util.ErrQuit {
				return
			}

			notifyBpTimeout(bpi)

		case <-bf.quit:
			return
		}
	}
}

func (bf *BlockFactory) worker() {
	for {
		select {
		case bpi := <-bf.workerQueue:
			block, err := bf.generateBlock(bpi)
			if err == util.ErrQuit {
				return
			} else if err != nil {
				logger.Info().Err(err).Msg("failed to produce block")
				continue
			}

			util.ConnectBlock(bf, block)

		case <-bf.quit:
			return
		}
	}
}

func (bf *BlockFactory) generateBlock(bpi *bpInfo) (*types.Block, error) {
	block, err := util.GenerateBlock(bf, bpi.bestBlock, bf.txOp, bpi.slot.UnixNano())
	if err != nil {
		return nil, err
	}
	if err := block.Sign(bf.privKey); err != nil {
		return nil, err
	}

	logger.Info().Msgf("block %v(no=%v) produced by BP %v", block.ID(), block.GetHeader().GetBlockNo(), bf.sID)

	return block, nil
}

func (bf *BlockFactory) checkBpTimeout() error {
	select {
	case <-bf.bpTimeoutC:
		return errBpTimeout
	case <-bf.quit:
		return util.ErrQuit
	default:
		return nil
	}
}
