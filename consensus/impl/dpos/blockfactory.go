/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package dpos

import (
	"runtime"
	"time"

	"github.com/aergoio/aergo-lib/log"
	bc "github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/consensus/chain"
	"github.com/aergoio/aergo/consensus/impl/dpos/slot"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p-crypto"
)

const (
	slotQueueMax = 100
)

type txExec struct {
	execTx bc.TxExecFn
}

func newTxExec(blockNo types.BlockNo, ts int64) chain.TxOp {
	// Block hash not determined yet
	return &txExec{
		execTx: bc.NewTxExecutor(blockNo, ts, contract.BlockFactory),
	}
}

func (te *txExec) Apply(bState *state.BlockState, tx *types.Tx) error {
	err := te.execTx(bState, tx)
	return err
}

// BlockFactory is the main data structure for DPoS block factory.
type BlockFactory struct {
	*component.ComponentHub
	jobQueue         chan interface{}
	workerQueue      chan *bpInfo
	bpTimeoutC       chan interface{}
	quit             <-chan interface{}
	maxBlockBodySize uint32
	ID               string
	privKey          crypto.PrivKey
	txOp             chain.TxOp
	sdb              *state.ChainStateDB
}

// NewBlockFactory returns a new BlockFactory
func NewBlockFactory(hub *component.ComponentHub, quitC <-chan interface{}) *BlockFactory {
	bf := &BlockFactory{
		ComponentHub:     hub,
		jobQueue:         make(chan interface{}, slotQueueMax),
		workerQueue:      make(chan *bpInfo),
		bpTimeoutC:       make(chan interface{}, 1),
		maxBlockBodySize: chain.MaxBlockBodySize(),
		quit:             quitC,
		ID:               p2p.NodeSID(),
		privKey:          p2p.NodePrivKey(),
	}

	bf.txOp = chain.NewCompTxOp(
		// timeout check
		chain.TxOpFn(func(bState *state.BlockState, txIn *types.Tx) error {
			return bf.checkBpTimeout()
		}),
	)

	return bf
}

func (bf *BlockFactory) setStateDB(sdb *state.ChainStateDB) {
	bf.sdb = sdb.Clone()
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
		time.Sleep(time.Duration(timeout) * time.Millisecond)
		// TODO: skip when the triggered block has already been genearted!
		bf.bpTimeoutC <- struct{}{}
		logger.Debug().Int64("timeout", timeout).Msg("block production timeout signaled")
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
	defer shutdownMsg("the block factory worker")

	runtime.LockOSThread()

	lpbNo := bsLoader.lpbNo()
	logger.Info().Uint64("lastly produced block", lpbNo).
		Msg("start the block factory worker")

	for {
		select {
		case bpi := <-bf.workerQueue:
			block, blockState, err := bf.generateBlock(bpi, lpbNo)
			if err == chain.ErrQuit {
				return
			} else if err != nil {
				if err == chain.ErrBestBlock {
					// This means the best block is beging changed by the chain
					// service. If the chain service quickly executes the
					// block, there may be still some remaining time to produce
					// block in the current slot, though. Thus retry block
					// production by reseting lastJob to nil.
					lastJob.setIf(nil, func(s *slot.Slot) bool { return slot.Equal(s, bpi.slot) })

					logger.Info().Err(err).Msg("retry block production")
				} else {
					logger.Info().Err(err).Msg("failed to produce block")
				}
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
	ts := bpi.slot.UnixNano()

	blockState := bf.sdb.NewBlockState(bpi.bestBlock.GetHeader().GetBlocksRootHash())

	txOp := chain.NewCompTxOp(
		bf.txOp,
		newTxExec(bpi.bestBlock.GetHeader().GetBlockNo()+1, ts),
	)

	block, err := chain.GenerateBlock(bf, bpi.bestBlock, blockState, txOp, ts)
	if err != nil {
		return nil, nil, err
	}

	block.SetConfirms(block.BlockNo() - lpbNo)

	if err := block.Sign(bf.privKey); err != nil {
		return nil, nil, err
	}

	logger.Info().
		Str("BP", bf.ID).Str("id", block.ID()).
		Str("sroot", enc.ToString(block.GetHeader().GetBlocksRootHash())).
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
