package chain

import (
	"fmt"
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"reflect"
	"time"
)

type ChainVerifier struct {
	*SubComponent
	cs            *ChainService
	IChainHandler //to use chain APIs
	*Core
	reader *BlockReader
	err    error
}

func newChainVerifier(cs *ChainService, core *Core) *ChainVerifier {
	chainVerifier := &ChainVerifier{IChainHandler: cs, Core: core, cs: cs}
	chainVerifier.SubComponent = NewSubComponent(chainVerifier, cs.BaseComponent, chainVerifierName, 1)

	var (
		bestBlock *types.Block
		err       error
	)

	if bestBlock, err = core.cdb.GetBestBlock(); err != nil {
		logger.Fatal().Msg("can't get best block in newChainVerifier()")
	}

	chainVerifier.reader = &BlockReader{cdb: core.cdb, curNo: 0, bestNo: bestBlock.BlockNo()}

	return chainVerifier
}

func (cv *ChainVerifier) Receive(context actor.Context) {
	defer RecoverExit()

	switch msg := context.Message().(type) {
	case *message.VerifyStart:

	case *actor.Started:
		logger.Info().Msg("verify chain service start")

		for !cv.cs.isRecovered() {
			logger.Debug().Msg("recovery of chain doesn't finished")
			time.Sleep(time.Second * 5)
		}

		if err := cv.VerifyChain(); err != nil {
			logger.Error().Err(err).Msg("failed to verify chain")
			cv.err = err
		}

		logger.Info().Msg("verify chain finished")

	case *actor.Stopping, *actor.Stopped, *component.CompStatReq: // donothing
	default:
		debug := fmt.Sprintf("[%s] Missed message. (%v) %s", cv.name, reflect.TypeOf(msg), msg)
		logger.Debug().Msg(debug)
	}
}

type BlockReader struct {
	cdb *ChainDB

	genesisBlock *types.Block
	curNo        uint64
	bestNo       uint64
}

func (br *BlockReader) getNext() (*types.Block, error) {
	if br.genesisBlock == nil {
		br.curNo = 0
	} else {
		if br.curNo+1 > br.bestNo {
			return nil, nil
		}
		br.curNo++
	}

	block, err := br.cdb.GetBlockByNo(br.curNo)
	if err != nil {
		logger.Error().Err(err).Uint64("no", br.curNo).Msg("failed to get next block")
		return nil, err
	}

	if br.curNo == 0 {
		br.genesisBlock = block
	}

	return block, nil
}

func (cv *ChainVerifier) VerifyChain() error {
	var (
		err   error
		block *types.Block
	)

	logger.Info().Msg("start verifychan")

	// get genesis block
	if block, err = cv.reader.getNext(); err != nil || block == nil {
		logger.Error().Err(err).Msg("failed to get genesis block")
		return err
	}

	if err = cv.Core.sdb.SetRoot(block.GetHeader().GetBlocksRootHash()); err != nil {
		logger.Error().Err(err).Msg("failed to set root of sdb to root hash of genesis block")
		return err
	}

	for {
		if block, err = cv.reader.getNext(); err != nil || block == nil {
			return err
		}

		if err := cv.verifyBlock(block); err != nil {
			logger.Error().Err(err).Uint64("no", block.BlockNo()).Str("block", block.ID()).Msg("failed to verify block")
			return err
		}
	}
}

func (cv *ChainVerifier) IsRunning() bool {
	return cv.reader.bestNo > cv.reader.curNo
}

func (cv *ChainVerifier) Statistics() *map[string]interface{} {
	var (
		state  string
		errStr string
	)

	if cv.err != nil {
		errStr = cv.err.Error()
	}

	if cv.IsRunning() {
		state = "running"
	} else {
		state = "finished"
	}

	return &map[string]interface{}{
		"verify":       state,
		"verifyBestNo": cv.reader.bestNo,
		"verifyCurNo":  cv.reader.curNo,
		"verifyErr":    errStr,
	}
}
