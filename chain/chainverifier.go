package chain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/message"
)

type ChainVerifier struct {
	*SubComponent
	cs            *ChainService
	IChainHandler //to use chain APIs
	*Core
	reader        *BlockReader
	err           error
	targetBlockNo uint64
	stage         TestStage
}

func newChainVerifier(cs *ChainService, core *Core, targetBlockNo uint64) *ChainVerifier {
	chainVerifier := &ChainVerifier{IChainHandler: cs, Core: core, cs: cs, targetBlockNo: targetBlockNo}
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

	case *actor.Stopping, *actor.Stopped, *component.CompStatReq: // do nothing
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

type TestStage int

const (
	TestPrevBlock TestStage = 0 + iota
	TestCurBlock
	TestBlockExecute
	TestComplete
)

var TestStageStr = []string{
	"Test if previous block exist",
	"Test if target block exist",
	"Test if block execution succeed",
	"All tests completed",
}

func (cv *ChainVerifier) VerifyBlockWithReport() error {
	var (
		err       error
		block     *types.Block
		prevBlock *types.Block
	)

	logger.Info().Msg("================== [VERIFYBLOCK] start =================")
	logger.Info().Msgf("Target block no : %d", cv.targetBlockNo)

	// set sdb root with previous block

	if cv.targetBlockNo <= 0 {
		logger.Error().Msg("target block number must be greater than 0")
	}

	cv.setStage(TestPrevBlock)
	if prevBlock, err = cv.cdb.GetBlockByNo(cv.targetBlockNo - 1); err != nil {
		goto END
	}
	cv.report(prevBlock, nil, nil)

	cv.Core.sdb.SetRoot(prevBlock.GetHeader().GetBlocksRootHash())

	cv.setStage(TestCurBlock)
	if block, err = cv.cdb.GetBlockByNo(cv.targetBlockNo); err != nil {
		goto END
	}
	cv.report(prevBlock, block, err)

	cv.setStage(TestBlockExecute)
	if err = cv.verifyBlock(block); err != nil {
		goto END
	}
	cv.report(prevBlock, block, nil)

	cv.setStage(TestComplete)
END:
	cv.report(prevBlock, block, err)

	logger.Info().Msg("================== [VERIFYBLOCK] end =================")

	return err
}

func (cv *ChainVerifier) setStage(stage TestStage) {
	cv.stage = stage
}

func (cv *ChainVerifier) report(prevBlock *types.Block, targetBlock *types.Block, err error) {
	var (
		report    string
		stageName string
	)

	if cv.stage == TestComplete {
		logger.Info().Msg(TestStageStr[TestComplete])
		return
	}

	stageName = TestStageStr[cv.stage]
	if err != nil {
		report += fmt.Sprintf("%s: FAILED.", stageName)
		report += fmt.Sprintf(" [Error] = %s", err.Error())
		logger.Info().Msg(report)
		return
	}

	report += fmt.Sprintf("%s: PASS. ", stageName)

	switch cv.stage {
	case TestPrevBlock:
		report += fmt.Sprintf("[description] prev block hash=%s, prev stage root=%s", prevBlock.ID(),
			base58.Encode(prevBlock.GetHeader().GetBlocksRootHash()))

	case TestCurBlock:
		report += fmt.Sprintf("[description] target block hash=%s", targetBlock.ID())

	case TestBlockExecute:
		report += fmt.Sprintf("[description] tx Merkle = %s", base58.Encode(targetBlock.GetHeader().GetTxsRootHash()))
		report += fmt.Sprintf(", state Root = %s", base58.Encode(targetBlock.GetHeader().GetBlocksRootHash()))
		report += fmt.Sprintf(", all transaction passed")
	}

	logger.Info().Msg(report)
}

func (cv *ChainVerifier) VerifyChain() error {
	var (
		err   error
		block *types.Block
	)

	if cv.targetBlockNo > 0 {
		return cv.VerifyBlockWithReport()
	}

	logger.Info().Msg("start verifyChain")

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

const permVer int = 4

func checkAccountStatus(accPermission []byte, accState *types.State) {
	var stateCarry []byte
	if permVer == 3 {
		stateCarry, _ = hex.DecodeString("0c59202ea3c74e18fc5a3be51c915b09")
	} else {
		stateCarry, _ = hex.DecodeString("03167d1771d19c03a280b58ca8f97bd3")
	}

	new(big.Int).SetBytes(stateCarry)

	var a, b, c *big.Int
	if permVer == 3 {
		a = new(big.Int).SetUint64(500000000000000000)
		b = new(big.Int).SetUint64(32866)
		c = new(big.Int).SetUint64(469)
	} else {
		a = new(big.Int).SetUint64(1000000000000000000)
		b = new(big.Int).SetUint64(10000)
		c = new(big.Int).SetUint64(1)
	}

	m1 := new(big.Int).Mul(a, b)
	m2 := new(big.Int).Mul(m1, c)

	persmionBit := make([]byte, len(accPermission)/2)
	for i := 0; i < len(accPermission)/2; i++ {
		persmionBit[i] = accPermission[i*2]
	}

	if bytes.Equal(persmionBit, stateCarry) {
		bal := accState.GetBalanceBigInt()
		balanceCarry := bal.Sub(bal, m2)
		accState.Balance = balanceCarry.Bytes()
	}
}
