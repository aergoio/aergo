package vm_direct

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/contract"
	"github.com/aergoio/aergo/v2/contract/name"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/fee"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
)

type ChainType int

const (
	ChainTypeMainNet ChainType = iota
	ChainTypeTestNet
	ChainTypeUnitTest
)

var (
	logger *log.Logger
)

func init() {
	logger = log.NewLogger("vm_dummy")
}

type DummyChain struct {
	chaintype       ChainType
	HardforkConfig  *config.HardforkConfig
	sdb             *state.ChainStateDB
	cBlock          *types.Block
	bestBlock       *types.Block
	bestBlockNo     types.BlockNo
	bestBlockId     types.BlockID
	tmpDir          string
	gasPrice        *big.Int
	timestamp       int64
	coinbaseAccount []byte
}

func LoadDummyChainEx(chainType ChainType) (*DummyChain, error) {
	var err error
	var gasPrice *big.Int

	switch chainType {
	case ChainTypeMainNet:
		gasPrice = types.NewAmount(50, types.Gaer)
	case ChainTypeTestNet:
		gasPrice = types.NewAmount(50, types.Gaer)
	case ChainTypeUnitTest:
		gasPrice = types.NewAmount(1, types.Aer)
	}

	dataPath := "./data/state"

	bc := &DummyChain{
		sdb:       state.NewChainStateDB(),
		tmpDir:    dataPath,
		chaintype: chainType,
		gasPrice:  gasPrice,
	}
	defer func() {
		if err != nil {
			bc.Release()
		}
	}()

	if chainType == ChainTypeMainNet {
		bc.HardforkConfig = config.MainNetHardforkConfig
	} else if chainType == ChainTypeTestNet {
		bc.HardforkConfig = config.TestNetHardforkConfig
	} else {
		bc.HardforkConfig = config.AllEnabledHardforkConfig
	}

	// mainnet and testnet use badger db. the dummy tests use memory db.
	dbImpl := db.BadgerImpl
	if chainType == ChainTypeUnitTest {
		dbImpl = db.MemoryImpl
	}

	// clear folder if exists
	_ = os.RemoveAll(dataPath)
	// initialize the state database
	err = bc.sdb.Init(string(dbImpl), dataPath, nil, false)
	if err != nil {
		return nil, err
	}

	var genesis *types.Genesis

	switch chainType {
	case ChainTypeMainNet:
		genesis = types.GetMainNetGenesis()
	case ChainTypeTestNet:
		genesis = types.GetTestNetGenesis()
	case ChainTypeUnitTest:
		genesis = types.GetTestGenesis()
	}

	bc.sdb.SetGenesis(genesis, nil)
	bc.bestBlock = genesis.Block()
	bc.bestBlockNo = genesis.Block().BlockNo()
	bc.bestBlockId = genesis.Block().BlockID()

	// before starting the LState factory
	if chainType == ChainTypeUnitTest {
		fee.EnableZeroFee()
	} else {
		contract.PubNet = true
	}

	// state sql database
	contract.LoadTestDatabase(dataPath)
	contract.SetStateSQLMaxDBSize(1024)

	contract.StartVMPool(contract.MaxPossibleCallDepth())
	contract.InitContext(3)

	// To pass the governance tests.
	types.InitGovernance("dpos", true)

	// To pass dao parameters test
	scs, err := statedb.GetSystemAccountState(bc.sdb.GetStateDB())
	if err != nil {
		return nil, err
	}
	system.InitSystemParams(scs, 3)

	return bc, nil
}

func (bc *DummyChain) Release() {
	_ = os.RemoveAll(bc.tmpDir)
}

func (bc *DummyChain) SetBestBlockId(value []byte) {
	bc.bestBlockId = types.ToBlockID(value)
}

func (bc *DummyChain) SetBestBlockNo(value uint64) {
	bc.bestBlockNo = value
}

func (bc *DummyChain) BestBlockNo() uint64 {
	return bc.bestBlockNo
}

func (bc *DummyChain) GetBestBlock() (*types.Block, error) {
	return bc.bestBlock, nil
}

func (bc *DummyChain) GetBlockByNo(blockNo types.BlockNo) (*types.Block, error) {
	//return bc.blocks[blockNo], nil
	return bc.bestBlock, nil
}

func (bc *DummyChain) SetTimestamp(value int64) {
	bc.timestamp = value
}

func (bc *DummyChain) getTimestamp() int64 {

	if bc.timestamp != 0 {
		return bc.timestamp
	} else {
		return time.Now().UnixNano()
	}

}

func (bc *DummyChain) SetCoinbaseAccount(address []byte) {
	bc.coinbaseAccount = address
}

////////////////////////////////////////////////////////////////////////

func (bc *DummyChain) newBlockState() *state.BlockState {
	bc.cBlock = &types.Block{
		Header: &types.BlockHeader{
			PrevBlockHash: bc.bestBlockId[:],
			BlockNo:       bc.bestBlockNo + 1,
			Timestamp:     bc.getTimestamp(),
			ChainID:       types.MakeChainId(bc.bestBlock.GetHeader().ChainID, bc.HardforkConfig.Version(bc.bestBlockNo+1)),
		},
	}
	return state.NewBlockState(
		bc.sdb.OpenNewStateDB(bc.sdb.GetRoot()),
		state.SetPrevBlockHash(bc.cBlock.GetHeader().PrevBlockHash), // or .GetPrevBlockHash()
		state.SetGasPrice(bc.gasPrice),
	)
}

func (bc *DummyChain) ExecuteTxs(txs []*types.Tx) ([]*types.Receipt, error) {

	ex, err := newBlockExecutor(bc, txs)
	if err != nil {
		return nil, err
	}

	if err := ex.execute(); err != nil {
		return nil, err
	}

	bc.cBlock.SetBlocksRootHash(bc.sdb.GetRoot())
	bc.bestBlock = bc.cBlock
	bc.bestBlockNo = bc.bestBlockNo + 1
	bc.bestBlockId = types.ToBlockID(bc.cBlock.BlockHash())

	receipts := ex.BlockState.Receipts().Get()

	return receipts, nil
}

type TxExecFn func(blockState *state.BlockState, tx types.Transaction) error
type ValidatePostFn func() error

type blockExecutor struct {
	*state.BlockState
	sdb              *state.ChainStateDB
	execTx           TxExecFn
	txs              []*types.Tx
	validatePost     ValidatePostFn
	coinbaseAcccount []byte
	bi               *types.BlockHeaderInfo
}

func newBlockExecutor(bc *DummyChain, txs []*types.Tx) (*blockExecutor, error) {
	var exec TxExecFn
	var bi *types.BlockHeaderInfo

	blockState := bc.newBlockState()
	bi = types.NewBlockHeaderInfo(bc.cBlock)

	exec = NewTxExecutor(context.Background(), nil, bc, bi, contract.ChainService)

	scs, err := statedb.GetSystemAccountState(blockState.StateDB)
	if err != nil {
		return nil, err
	}
	blockState.SetGasPrice(system.GetGasPriceFromState(scs))

	blockState.Receipts().SetHardFork(bc.HardforkConfig, bc.bestBlockNo+1)

	return &blockExecutor{
		BlockState:       blockState,
		sdb:              bc.sdb,
		execTx:           exec,
		txs:              txs,
		coinbaseAcccount: bc.coinbaseAccount,
		//validatePost: func() error {
		//	return cs.validator.ValidatePost(blockState.GetRoot(), blockState.Receipts(), block)
		//},
		bi: bi,
	}, nil

}

func NewTxExecutor(execCtx context.Context, ccc consensus.ChainConsensusCluster, cdb contract.ChainAccessor, bi *types.BlockHeaderInfo, executionMode int) TxExecFn {

	return func(blockState *state.BlockState, tx types.Transaction) error {

		if blockState == nil {
			return errors.New("blockState is nil in txexec")
		}
		if bi.ForkVersion < 0 {
			return errors.New("ChainID.ForkVersion < 0")
		}

		blockSnap := blockState.Snapshot()

		err := executeTx(execCtx, ccc, cdb, blockState, tx, bi, executionMode)
		if err != nil {
			logger.Error().Err(err).Str("hash", base58.Encode(tx.GetHash())).Msg("tx failed")
			if err2 := blockState.Rollback(blockSnap); err2 != nil {
				logger.Panic().Err(err).Msg("failed to rollback block state")
			}
			return err
		}

		return nil
	}

}

// execute all transactions in the block
func (e *blockExecutor) execute() error {

	defer contract.CloseDatabase()

	for _, tx := range e.txs {
		// execute the transaction
		if err := e.execTx(e.BlockState, types.NewTransaction(tx)); err != nil {
			return err
		}
	}

	if err := SendBlockReward(e.BlockState, e.coinbaseAcccount); err != nil {
		return err
	}

	if err := contract.SaveRecoveryPoint(e.BlockState); err != nil {
		return err
	}

	if err := e.Update(); err != nil {
		return err
	}

	//if err := e.validatePost(); err != nil {
	//	return err
	//}

	if err := e.commit(); err != nil {
		return err
	}

	return nil
}

func (e *blockExecutor) commit() error {

	if err := e.BlockState.Commit(); err != nil {
		return err
	}

	if err := e.sdb.UpdateRoot(e.BlockState); err != nil {
		return err
	}

	return nil
}

const maxRetSize = 1024

func adjustReturnValue(ret string) string {
	if len(ret) > maxRetSize {
		modified, _ := json.Marshal(ret[:maxRetSize-4] + " ...")
		return string(modified)
	}
	return ret
}

func resetAccount(account *state.AccountState, fee *big.Int, nonce *uint64) error {
	account.Reset()
	if fee != nil {
		if account.Balance().Cmp(fee) < 0 {
			return &types.InternalError{Reason: "fee is greater than balance"}
		}
		account.SubBalance(fee)
	}
	if nonce != nil {
		account.SetNonce(*nonce)
	}
	return account.PutState()
}

func executeTx(
	execCtx context.Context,
	ccc consensus.ChainConsensusCluster,
	cdb contract.ChainAccessor,
	bs *state.BlockState,
	tx types.Transaction,
	bi *types.BlockHeaderInfo,
	executionMode int,
) error {

	var (
		txBody    = tx.GetBody()
		isQuirkTx = types.IsQuirkTx(tx.GetHash())
		account   []byte
		recipient []byte
		err       error
	)

	if account, err = name.Resolve(bs, txBody.GetAccount(), isQuirkTx); err != nil {
		return err
	}

	if tx.HasVerifedAccount() {
		txAcc := tx.GetVerifedAccount()
		tx.RemoveVerifedAccount()
		if !bytes.Equal(txAcc, account) {
			return types.ErrSignNotMatch
		}
	}

	err = tx.Validate(bi.ChainIdHash(), IsPublic())
	if err != nil {
		return err
	}

	sender, err := state.GetAccountState(account, bs.StateDB)
	if err != nil {
		return err
	}

	// check for sufficient balance
	senderState := sender.State()
	amount := tx.GetBody().GetAmountBigInt()
	balance := senderState.GetBalanceBigInt()

	switch tx.GetBody().GetType() {
	case types.TxType_NORMAL, types.TxType_REDEPLOY, types.TxType_TRANSFER, types.TxType_CALL, types.TxType_DEPLOY:
		if balance.Cmp(amount) <= 0 {
			// set the balance as amount + fee
			to_add := new(big.Int).SetUint64(1000000000000000000)
			balance = new(big.Int).Add(amount, to_add)
			senderState.Balance = balance.Bytes()
		}
	case types.TxType_GOVERNANCE:
		switch string(tx.GetBody().GetRecipient()) {
		case types.AergoSystem:
			if balance.Cmp(amount) <= 0 {
				// set the balance as amount + fee
				to_add := new(big.Int).SetUint64(1000000000000000000)
				balance = new(big.Int).Add(amount, to_add)
				senderState.Balance = balance.Bytes()
			}
		case types.AergoName:
			if balance.Cmp(amount) <= 0 {
				// set the balance as = amount
				senderState.Balance = amount.Bytes()
			}
		}
	case types.TxType_FEEDELEGATION:
		if balance.Cmp(amount) <= 0 {
			// set the balance as = amount
			senderState.Balance = amount.Bytes()
		}
	}

	err = tx.ValidateWithSenderState(senderState, bs.GasPrice, bi.ForkVersion)
	if err != nil {
		err = fmt.Errorf("%w: balance %s, amount %s, gasPrice %s, block %v, txhash: %s",
			err,
			sender.Balance().String(),
			tx.GetBody().GetAmountBigInt().String(),
			bs.GasPrice.String(),
			bi.No, base58.Encode(tx.GetHash()))
		return err
	}

	if recipient, err = name.Resolve(bs, txBody.Recipient, isQuirkTx); err != nil {
		return err
	}
	var receiver *state.AccountState
	status := "SUCCESS"
	if len(recipient) > 0 {
		receiver, err = state.GetAccountState(recipient, bs.StateDB)
		if receiver != nil && txBody.Type == types.TxType_REDEPLOY {
			status = "RECREATED"
			receiver.SetRedeploy()
		}
	} else {
		receiver, err = state.CreateAccountState(contract.CreateContractID(txBody.Account, txBody.Nonce), bs.StateDB)
		status = "CREATED"
	}
	if err != nil {
		return err
	}

	var txFee *big.Int
	var rv string
	var events []*types.Event
	switch txBody.Type {
	case types.TxType_NORMAL, types.TxType_REDEPLOY, types.TxType_TRANSFER, types.TxType_CALL, types.TxType_DEPLOY:
		rv, events, txFee, err = contract.Execute(execCtx, bs, cdb, tx.GetTx(), sender, receiver, bi, executionMode, false)
		sender.SubBalance(txFee)
	case types.TxType_GOVERNANCE:
		txFee = new(big.Int).SetUint64(0)
		events, err = executeGovernanceTx(ccc, bs, txBody, sender, receiver, bi)
		if err != nil {
			logger.Warn().Err(err).Str("txhash", base58.Encode(tx.GetHash())).Msg("governance tx Error")
		}
	case types.TxType_FEEDELEGATION:
		err = tx.ValidateMaxFee(receiver.Balance(), bs.GasPrice, bi.ForkVersion)
		if err != nil {
			return err
		}
		var contractState *statedb.ContractState
		contractState, err = statedb.OpenContractState(receiver.ID(), receiver.State(), bs.StateDB)
		if err != nil {
			return err
		}
		err = contract.CheckFeeDelegation(recipient, bs, bi, cdb, contractState, txBody.GetPayload(),
			tx.GetHash(), txBody.GetAccount(), txBody.GetAmount())
		if err != nil {
			if err != types.ErrNotAllowedFeeDelegation {
				logger.Warn().Err(err).Str("txhash", base58.Encode(tx.GetHash())).Msg("checkFeeDelegation Error")
				return err
			}
			return types.ErrNotAllowedFeeDelegation
		}
		rv, events, txFee, err = contract.Execute(execCtx, bs, cdb, tx.GetTx(), sender, receiver, bi, executionMode, true)
		receiver.SubBalance(txFee)
	}

	if err != nil {
		// Reset events on error
		if bi.ForkVersion >= 3 {
			events = nil
		}

		if !contract.IsRuntimeError(err) {
			return err
		}
		if txBody.Type != types.TxType_FEEDELEGATION || sender.AccountID() == receiver.AccountID() {
			sErr := resetAccount(sender, txFee, &txBody.Nonce)
			if sErr != nil {
				return sErr
			}
		} else {
			sErr := resetAccount(sender, nil, &txBody.Nonce)
			if sErr != nil {
				return sErr
			}
			sErr = resetAccount(receiver, txFee, nil)
			if sErr != nil {
				return sErr
			}
		}
		status = "ERROR"
		rv = err.Error()
	} else {
		if txBody.Type != types.TxType_FEEDELEGATION {
			if sender.Balance().Sign() < 0 {
				return &types.InternalError{Reason: "fee is greater than balance"}
			}
		} else {
			if receiver.Balance().Sign() < 0 {
				return &types.InternalError{Reason: "fee is greater than balance"}
			}
		}
		sender.SetNonce(txBody.Nonce)
		err = sender.PutState()
		if err != nil {
			return err
		}
		if sender.AccountID() != receiver.AccountID() {
			err = receiver.PutState()
			if err != nil {
				return err
			}
		}
		rv = adjustReturnValue(rv)
	}
	bs.BpReward.Add(&bs.BpReward, txFee)

	receipt := types.NewReceipt(receiver.ID(), status, rv)
	receipt.FeeUsed = txFee.Bytes()
	receipt.TxHash = tx.GetHash()
	receipt.Events = events
	receipt.FeeDelegation = txBody.Type == types.TxType_FEEDELEGATION
	isGovernance := txBody.Type == types.TxType_GOVERNANCE
	receipt.GasUsed = fee.ReceiptGasUsed(bi.ForkVersion, isGovernance, txFee, bs.GasPrice)

	return bs.AddReceipt(receipt)
}

func executeGovernanceTx(ccc consensus.ChainConsensusCluster, bs *state.BlockState, txBody *types.TxBody, sender, receiver *state.AccountState,
	blockInfo *types.BlockHeaderInfo) ([]*types.Event, error) {

	if len(txBody.Payload) <= 0 {
		return nil, types.ErrTxFormatInvalid
	}

	governance := string(txBody.Recipient)

	scs, err := statedb.OpenContractState(receiver.ID(), receiver.State(), bs.StateDB)
	if err != nil {
		return nil, err
	}

	var events []*types.Event

	switch governance {
	case types.AergoSystem:
		events, err = system.ExecuteSystemTx(scs, txBody, sender, receiver, blockInfo)
	case types.AergoName:
		events, err = name.ExecuteNameTx(bs, scs, txBody, sender, receiver, blockInfo)
	default:
		logger.Warn().Str("governance", governance).Msg("receive unknown recipient")
		err = types.ErrTxInvalidRecipient
	}

	if err == nil {
		err = statedb.StageContractState(scs, bs.StateDB)
	}

	return events, err
}

func SendBlockReward(bState *state.BlockState, coinbaseAccount []byte) error {
	bpReward := &bState.BpReward
	if bpReward.Cmp(new(big.Int).SetUint64(0)) <= 0 || coinbaseAccount == nil {
		logger.Debug().Str("reward", bpReward.String()).Msg("coinbase is skipped")
		return nil
	}

	receiverID := types.ToAccountID(coinbaseAccount)
	receiverState, err := bState.GetAccountState(receiverID)
	if err != nil {
		return err
	}

	receiverChange := receiverState.Clone()
	receiverChange.Balance = new(big.Int).Add(receiverChange.GetBalanceBigInt(), bpReward).Bytes()

	err = bState.PutState(receiverID, receiverChange)
	if err != nil {
		return err
	}

	logger.Debug().Str("reward", bpReward.String()).
		Str("newbalance", receiverChange.GetBalanceBigInt().String()).Msg("send reward to coinbase account")

	return nil
}

func IsPublic() bool {
	return true
}
