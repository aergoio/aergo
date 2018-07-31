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
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	slotQueueMax = 100
)

var (
	errBpTimeout = errors.New("block production timeout")
	errQuit      = errors.New("shutdown initiated")
)

// BlockFactory is the main data structure for DPoS block factory.
type BlockFactory struct {
	*component.ComponentHub
	jobQueue         chan interface{}
	beginBlock       chan *bpInfo
	bpTimeoutC       chan interface{}
	quit             <-chan interface{}
	maxBlockBodySize int
	sID              string
	privKey          crypto.PrivKey
}

// NewBlockFactory returns a new BlockFactory
func NewBlockFactory(hub *component.ComponentHub, id peer.ID, privKey crypto.PrivKey) *BlockFactory {
	return &BlockFactory{
		ComponentHub:     hub,
		jobQueue:         make(chan interface{}, slotQueueMax),
		beginBlock:       make(chan *bpInfo),
		bpTimeoutC:       make(chan interface{}, 1),
		maxBlockBodySize: util.MaxBlockBodySize(),
		sID:              id.Pretty(),
		privKey:          privKey,
	}
}

// Start run a DPoS block factory service.
func (bf *BlockFactory) Start(quitC <-chan interface{}) {
	bf.quit = quitC

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
		if err := bf.checkBpTimeout(); err == errQuit {
			return err
		}

		select {
		case bf.beginBlock <- bpi:
		default:
			logger.Errorf(
				"skip block production for the slot %v (best block: %v) due to a pending job",
				spew.Sdump(bpi.slot), bpi.bestBlock.ID())
		}
		return nil
	}

	notifyBpTimeout := func(bpi *bpInfo) {
		timeout := bpi.slot.GetBpTimeout()
		logger.Debugf("block production timeout: %vms", timeout)
		time.Sleep(time.Duration(timeout) * time.Millisecond)
		// TODO: skip when the triggered block has already been genearted!
		bf.bpTimeoutC <- struct{}{}
	}

	for {
		select {
		case info := <-bf.jobQueue:
			bpi := info.(*bpInfo)
			logger.Debugf("received bpInfo: %v %v",
				log.DoLazyEval(func() string {
					return bpi.bestBlock.ID()
				}),
				log.DoLazyEval(func() string {
					return spew.Sdump(bpi.slot)
				}))

			if err := beginBlock(bpi); err == errQuit {
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
		case bpi := <-bf.beginBlock:
			if err := bf.checkBpTimeout(); err != nil {
				continue
			}

			block, err := bf.generateBlock(bpi)
			if err == errQuit {
				return
			}

			// TODO: Send the generated block to ChainManager.
			util.ConnectBlock(bf, block)

		case <-bf.quit:
			return
		}
	}
}

func (bf *BlockFactory) generateBlock(bpi *bpInfo) (*types.Block, error) {
	txs, err := bf.gatherTXs(util.FetchTXs(bf))
	if err != nil {
		return nil, err
	}

	block := types.NewBlock(bpi.bestBlock, txs)
	if err := block.Sign(bf.privKey); err != nil {
		return nil, err
	}

	logger.Infof("block %v produced by BP %v", block.ID(), bf.sID)

	return block, nil
}

func (bf *BlockFactory) checkBpTimeout() error {
	select {
	case <-bf.bpTimeoutC:
		return errBpTimeout
	case <-bf.quit:
		return errQuit
	default:
		return nil
	}
}

func (bf *BlockFactory) gatherTXs(txIn []*types.Tx) ([]*types.Tx, error) {
	checkTimeout := func(txIn *types.Tx) error {
		return bf.checkBpTimeout()
	}

	return util.GatherTXs(txIn, util.NewTxDo(checkTimeout), bf.maxBlockBodySize)
}
