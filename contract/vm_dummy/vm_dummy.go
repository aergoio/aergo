package vm_dummy

// helper functions
import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/cmd/aergoluac/util"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/contract"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/fee"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	sha256 "github.com/minio/sha256-simd"
)

var (
	logger *log.Logger
)

const (
	lStateMaxSize = 10 * 7
)

func init() {
	logger = log.NewLogger("vm_dummy")
}

type DummyChain struct {
	HardforkVersion int32
	sdb             *state.ChainStateDB
	bestBlock       *types.Block
	cBlock          *types.Block
	bestBlockNo     types.BlockNo
	bestBlockId     types.BlockID
	blockIds        []types.BlockID
	blocks          []*types.Block
	testReceiptDB   db.DB
	tmpDir          string
	timeout         int
	clearLState     func()
	gasPrice        *big.Int
	timestamp       int64
}

// overwrite config for dummychain
type DummyChainOptions func(d *DummyChain)

func SetHardForkVersion(forkVersion int32) DummyChainOptions {
	return func(dc *DummyChain) {
		dc.HardforkVersion = forkVersion
	}
}

func SetTimeout(timeout int) DummyChainOptions {
	return func(dc *DummyChain) {
		dc.timeout = timeout
	}
}

func SetPubNet() DummyChainOptions {
	return func(dc *DummyChain) {

		// public and private chains have different features.
		// private chains have the db module and public ones don't.
		// this is why we need to flush all Lua states and recreate
		// them when moving to and from public chain.

		contract.PubNet = true
		fee.DisableZeroFee()
		contract.FlushLStates()

		dc.clearLState = func() {
			contract.PubNet = false
			fee.EnableZeroFee()
			contract.FlushLStates()
		}
	}
}

func LoadDummyChain(opts ...DummyChainOptions) (*DummyChain, error) {
	dataPath, err := os.MkdirTemp("", "data")
	if err != nil {
		return nil, err
	}
	bc := &DummyChain{
		sdb:      state.NewChainStateDB(),
		tmpDir:   dataPath,
		gasPrice: types.NewAmount(1, types.Aer),
	}
	defer func() {
		if err != nil {
			bc.Release()
		}
	}()

	// reset the transaction id counter
	luaTxId = 0

	err = bc.sdb.Init(string(db.MemoryImpl), dataPath, nil, false)
	if err != nil {
		return nil, err
	}
	genesis := types.GetTestGenesis()
	bc.sdb.SetGenesis(genesis, nil)
	bc.bestBlock = genesis.Block()
	bc.bestBlockNo = genesis.Block().BlockNo()
	bc.bestBlockId = genesis.Block().BlockID()
	bc.blockIds = append(bc.blockIds, bc.bestBlockId)
	bc.blocks = append(bc.blocks, genesis.Block())
	bc.testReceiptDB = db.NewDB(db.MemoryImpl, path.Join(dataPath, "receiptDB"))
	contract.LoadTestDatabase(dataPath) // sql database
	contract.SetStateSQLMaxDBSize(1024)
	contract.StartLStateFactory(lStateMaxSize, config.GetDefaultNumLStateClosers(), 1)
	contract.InitContext(3)

	bc.HardforkVersion = 2

	// To pass the governance tests.
	types.InitGovernance("dpos", true)
	system.InitGovernance("dpos")

	// To pass dao parameters test
	scs, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	system.InitSystemParams(scs, 3)

	fee.EnableZeroFee()

	for _, opt := range opts {
		opt(bc)
	}
	return bc, nil
}

func (bc *DummyChain) Release() {
	bc.testReceiptDB.Close()
	if bc.clearLState != nil {
		bc.clearLState()
	}
	_ = os.RemoveAll(bc.tmpDir)
}

func (bc *DummyChain) BestBlockNo() uint64 {
	return bc.bestBlockNo
}

func (bc *DummyChain) SetTimestamp(is_increment bool, value int64) {
	if is_increment {
		if bc.timestamp == 0 {
			bc.timestamp = time.Now().UnixNano() / 1000000000
		}
		bc.timestamp += value
	} else {
		bc.timestamp = value
	}
}

func (bc *DummyChain) getTimestamp() int64 {

	if bc.timestamp > 0 {
		return bc.timestamp * 1000000000
	} else {
		return time.Now().UnixNano()
	}

}

func (bc *DummyChain) newBState() *state.BlockState {
	bc.cBlock = &types.Block{
		Header: &types.BlockHeader{
			PrevBlockHash: bc.bestBlockId[:],
			BlockNo:       bc.bestBlockNo + 1,
			Timestamp:     bc.getTimestamp(),
			ChainID:       types.MakeChainId(bc.bestBlock.GetHeader().ChainID, bc.HardforkVersion),
		},
	}
	return state.NewBlockState(
		bc.sdb.OpenNewStateDB(bc.sdb.GetRoot()),
		state.SetPrevBlockHash(bc.cBlock.GetHeader().PrevBlockHash),
		state.SetGasPrice(bc.gasPrice),
	)
}

func (bc *DummyChain) BeginReceiptTx() db.Transaction {
	return bc.testReceiptDB.NewTx()
}

func (bc *DummyChain) GetABI(code string) (*types.ABI, error) {
	cState, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(contract.StrHash(code)))
	if err != nil {
		return nil, err
	}
	return contract.GetABI(cState, nil)
}

func (bc *DummyChain) GetEvents(txhash []byte) []*types.Event {
	receipt := bc.GetReceipt(txhash)
	if receipt != nil {
		return receipt.Events
	}

	return nil
}

func (bc *DummyChain) GetReceipt(txHash []byte) *types.Receipt {
	r := new(types.Receipt)
	r.UnmarshalBinaryTest(bc.testReceiptDB.Get(txHash))
	return r
}

func (bc *DummyChain) GetAccountState(name string) (*types.State, error) {
	return bc.sdb.GetStateDB().GetAccountState(types.ToAccountID(contract.StrHash(name)))
}

func (bc *DummyChain) GetStaking(name string) (*types.Staking, error) {
	scs, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	if err != nil {
		return nil, err
	}
	return system.GetStaking(scs, contract.StrHash(name))
}

func (bc *DummyChain) GetBlockByNo(blockNo types.BlockNo) (*types.Block, error) {
	return bc.blocks[blockNo], nil
}

func (bc *DummyChain) GetBestBlock() (*types.Block, error) {
	return bc.bestBlock, nil
}

type LuaTxTester interface {
	run(bs *state.BlockState, bc *DummyChain, bi *types.BlockHeaderInfo, receiptTx db.Transaction) error
	Hash() []byte
	okMsg() string
}

type luaTxAccount struct {
	name    []byte
	balance *big.Int
	txId    uint64
}

var _ LuaTxTester = (*luaTxAccount)(nil)

func NewLuaTxAccount(name string, balance uint64, unit types.TokenUnit) *luaTxAccount {
	return NewLuaTxAccountBig(name, types.NewAmount(balance, unit))
}

func NewLuaTxAccountBig(name string, balance *big.Int) *luaTxAccount {
	return &luaTxAccount{
		name:    contract.StrHash(name),
		balance: balance,
		txId:    newTxId(),
	}
}

func (l *luaTxAccount) Hash() []byte {
	return hash(l.txId)
}

func (l *luaTxAccount) okMsg() string {
	return "SUCCESS"
}

func (l *luaTxAccount) run(bs *state.BlockState, bc *DummyChain, bi *types.BlockHeaderInfo, receiptTx db.Transaction) error {
	id := types.ToAccountID(l.name)
	accountState, err := bs.GetAccountState(id)
	if err != nil {
		return err
	}
	updatedAccountState := *accountState
	updatedAccountState.Balance = l.balance.Bytes()
	bs.PutState(id, &updatedAccountState)
	return nil
}

type luaTxSend struct {
	sender   []byte
	receiver []byte
	amount   *big.Int
	txId     uint64
}

var _ LuaTxTester = (*luaTxSend)(nil)

func NewLuaTxSendBig(sender, receiver string, amount *big.Int) *luaTxSend {
	return &luaTxSend{
		sender:   contract.StrHash(sender),
		receiver: contract.StrHash(receiver),
		amount:   amount,
		txId:     newTxId(),
	}
}

func (l *luaTxSend) Hash() []byte {
	return hash(l.txId)
}

func (l *luaTxSend) okMsg() string {
	return "SUCCESS"
}

func (l *luaTxSend) run(bs *state.BlockState, bc *DummyChain, bi *types.BlockHeaderInfo, receiptTx db.Transaction) error {
	senderID := types.ToAccountID(l.sender)
	receiverID := types.ToAccountID(l.receiver)

	if senderID == receiverID {
		return fmt.Errorf("sender and receiever cannot be same")
	}

	senderState, err := bs.GetAccountState(senderID)
	if err != nil {
		return err
	} else if senderState.GetBalanceBigInt().Cmp(l.amount) < 0 {
		return fmt.Errorf("insufficient balance to sender")
	}
	receiverState, err := bs.GetAccountState(receiverID)
	if err != nil {
		return err
	}

	updatedSenderState := types.State(*senderState)
	updatedSenderState.Balance = new(big.Int).Sub(updatedSenderState.GetBalanceBigInt(), l.amount).Bytes()
	bs.PutState(senderID, &updatedSenderState)

	updatedReceiverState := types.State(*receiverState)
	updatedReceiverState.Balance = new(big.Int).Add(updatedReceiverState.GetBalanceBigInt(), l.amount).Bytes()
	bs.PutState(receiverID, &updatedReceiverState)

	r := types.NewReceipt(l.receiver, l.okMsg(), "")
	r.TxHash = l.Hash()
	r.GasUsed = fee.TxGas(0)
	b, _ := r.MarshalBinaryTest()
	receiptTx.Set(l.Hash(), b)

	return nil
}

type luaTxContract interface {
	LuaTxTester
	sender() []byte
	recipient() []byte
	amount() *big.Int
	payload() []byte
	isFeeDelegate() bool
}

type luaTxContractCommon struct {
	_sender     []byte
	_recipient  []byte
	_amount     *big.Int
	_payload    []byte
	txId        uint64
	feeDelegate bool
}

func (l *luaTxContractCommon) Hash() []byte {
	return hash(l.txId)
}

func (l *luaTxContractCommon) sender() []byte {
	return l._sender
}

func (l *luaTxContractCommon) recipient() []byte {
	return l._recipient
}

func (l *luaTxContractCommon) amount() *big.Int {
	return l._amount
}

func (l *luaTxContractCommon) payload() []byte {
	return l._payload
}

func (l *luaTxContractCommon) isFeeDelegate() bool {
	return l.feeDelegate
}

func hash(id uint64) []byte {
	h := sha256.New()
	h.Write([]byte(strconv.FormatUint(id, 10)))
	b := h.Sum(nil)
	return b
}

type luaTxDeploy struct {
	luaTxContractCommon
	cErr error
}

var _ LuaTxTester = (*luaTxDeploy)(nil)

func NewLuaTxDeploy(sender, recipient string, amount uint64, code string) *luaTxDeploy {
	return NewLuaTxDeployBig(sender, recipient, types.NewAmount(amount, types.Aer), code)
}

/*
	func contract.StrHash(d string) []byte {
		// using real address
		if len(d) == types.EncodedAddressLength && addressRegexp.MatchString(d) {
			return types.ToAddress(d)
		} else {
			// using alias
			h := sha256.New()
			h.Write([]byte(d))
			b := h.Sum(nil)
			b = append([]byte{0x0C}, b...)
			return b
		}
	}
*/

var luaTxId uint64 = 0

func newTxId() uint64 {
	luaTxId++
	return luaTxId
}

func (l *luaTxDeploy) okMsg() string {
	return "CREATED"
}

func (l *luaTxDeploy) Constructor(args string) *luaTxDeploy {
	if len(args) == 0 || strings.Compare(args, "[]") == 0 || l.cErr != nil {
		return l
	}
	l._payload = util.NewLuaCodePayload(util.LuaCodePayload(l._payload).Code(), []byte(args))
	return l
}

func contractFrame(l luaTxContract, bs *state.BlockState, cdb contract.ChainAccessor, receiptTx db.Transaction,
	run func(s, c *state.V, id types.AccountID, cs *state.ContractState) (string, []*types.Event, *big.Int, error)) error {

	creatorId := types.ToAccountID(l.sender())
	creatorState, err := bs.GetAccountStateV(l.sender())
	if err != nil {
		return err
	}

	contractId := types.ToAccountID(l.recipient())
	contractState, err := bs.GetAccountStateV(l.recipient())
	if err != nil {
		return err
	}

	eContractState, err := bs.OpenContractState(contractId, contractState.State())
	if err != nil {
		return err
	}
	usedFee := contract.TxFee(len(l.payload()), types.NewAmount(1, types.Aer), 2)

	if l.isFeeDelegate() {
		balance := contractState.Balance()

		if usedFee.Cmp(balance) > 0 {
			return types.ErrInsufficientBalance
		}
		err = contract.CheckFeeDelegation(l.recipient(), bs, nil, cdb, eContractState, l.payload(),
			l.Hash(), l.sender(), l.amount().Bytes())
		if err != nil {
			if err != types.ErrNotAllowedFeeDelegation {
				logger.Debug().Err(err).Str("txhash", enc.ToString(l.Hash())).Msg("checkFeeDelegation Error")
				return err
			}
			return types.ErrNotAllowedFeeDelegation
		}
	}
	creatorState.SubBalance(l.amount())
	contractState.AddBalance(l.amount())
	rv, events, cFee, err := run(creatorState, contractState, contractId, eContractState)
	if cFee != nil {
		usedFee.Add(usedFee, cFee)
	}
	status := l.okMsg()
	if err != nil {
		status = "ERROR"
		rv = err.Error()
	}
	r := types.NewReceipt(l.recipient(), status, rv)
	r.TxHash = l.Hash()
	r.GasUsed = usedFee.Uint64()
	r.Events = events
	blockHash := make([]byte, 32)
	for _, event := range events {
		event.TxHash = r.TxHash
		event.BlockHash = blockHash
	}
	b, _ := r.MarshalBinaryTest()
	receiptTx.Set(l.Hash(), b)
	if err != nil {
		return err
	}

	if l.isFeeDelegate() {
		if contractState.Balance().Cmp(usedFee) < 0 {
			return types.ErrInsufficientBalance
		}
		contractState.SubBalance(usedFee)
	} else {
		if creatorState.Balance().Cmp(usedFee) < 0 {
			return types.ErrInsufficientBalance
		}
		creatorState.SubBalance(usedFee)
	}
	bs.PutState(creatorId, creatorState.State())
	bs.PutState(contractId, contractState.State())
	return nil

}

func (l *luaTxDeploy) run(bs *state.BlockState, bc *DummyChain, bi *types.BlockHeaderInfo, receiptTx db.Transaction) error {
	if l.cErr != nil {
		return l.cErr
	}
	return contractFrame(l, bs, bc, receiptTx,
		func(sender, contractV *state.V, contractId types.AccountID, eContractState *state.ContractState) (string, []*types.Event, *big.Int, error) {
			contractV.State().SqlRecoveryPoint = 1

			ctx := contract.NewVmContext(bs, nil, sender, contractV, eContractState, sender.ID(), l.Hash(), bi, "", true,
				false, contractV.State().SqlRecoveryPoint, contract.BlockFactory, l.amount(), math.MaxUint64, false)

			rv, events, ctrFee, err := contract.Create(eContractState, l.payload(), l.recipient(), ctx)
			if err != nil {
				return "", nil, ctrFee, err
			}
			err = bs.StageContractState(eContractState)
			if err != nil {
				return "", nil, ctrFee, err
			}
			return rv, events, ctrFee, nil
		},
	)
}

type luaTxCall struct {
	luaTxContractCommon
	expectedErr string
}

var _ LuaTxTester = (*luaTxCall)(nil)

func NewLuaTxCall(sender, recipient string, amount uint64, payload string) *luaTxCall {
	return NewLuaTxCallBig(sender, recipient, types.NewAmount(amount, types.Aer), payload)
}

func NewLuaTxCallBig(sender, recipient string, amount *big.Int, payload string) *luaTxCall {
	return &luaTxCall{
		luaTxContractCommon: luaTxContractCommon{
			_sender:    contract.StrHash(sender),
			_recipient: contract.StrHash(recipient),
			_amount:    amount,
			_payload:   []byte(payload),
			txId:       newTxId(),
		},
	}
}

func NewLuaTxCallFeeDelegate(sender, recipient string, amount uint64, payload string) *luaTxCall {
	return &luaTxCall{
		luaTxContractCommon: luaTxContractCommon{
			_sender:     contract.StrHash(sender),
			_recipient:  contract.StrHash(recipient),
			_amount:     types.NewAmount(amount, types.Aer),
			_payload:    []byte(payload),
			txId:        newTxId(),
			feeDelegate: true,
		},
	}
}

func (l *luaTxCall) Fail(expectedErr string) *luaTxCall {
	l.expectedErr = expectedErr
	return l
}

func (l *luaTxCall) run(bs *state.BlockState, bc *DummyChain, bi *types.BlockHeaderInfo, receiptTx db.Transaction) error {
	err := contractFrame(l, bs, bc, receiptTx,
		func(sender, contractV *state.V, contractId types.AccountID, eContractState *state.ContractState) (string, []*types.Event, *big.Int, error) {
			ctx := contract.NewVmContext(bs, bc, sender, contractV, eContractState, sender.ID(), l.Hash(), bi, "", true,
				false, contractV.State().SqlRecoveryPoint, contract.BlockFactory, l.amount(), math.MaxUint64, l.feeDelegate)

			rv, events, ctrFee, err := contract.Call(eContractState, l.payload(), l.recipient(), ctx)
			if err != nil {
				return "", nil, ctrFee, err
			}
			err = bs.StageContractState(eContractState)
			if err != nil {
				return "", nil, ctrFee, err
			}
			return rv, events, ctrFee, nil
		},
	)
	if l.expectedErr != "" {
		if err == nil {
			return fmt.Errorf("no error, expected: %s", l.expectedErr)
		}
		if !strings.Contains(err.Error(), l.expectedErr) {
			return err
		}
		return nil
	}
	return err
}

func (l *luaTxCall) okMsg() string {
	return "SUCCESS"
}

func (bc *DummyChain) ConnectBlock(txs ...LuaTxTester) error {
	blockState := bc.newBState()
	tx := bc.BeginReceiptTx()
	defer tx.Commit()
	defer contract.CloseDatabase()

	//timeout := make(chan struct{})
	//blockContext, _ := context.WithTimeout(context.Background(), time.Duration(bc.timeout)*time.Millisecond)
	//contract.SetBPTimeout(timeout)
	for _, x := range txs {
		if err := x.run(blockState, bc, types.NewBlockHeaderInfo(bc.cBlock), tx); err != nil {
			return err
		}
	}
	err := contract.SaveRecoveryPoint(blockState)
	if err != nil {
		return err
	}
	err = bc.sdb.Apply(blockState)
	if err != nil {
		return err
	}
	//FIXME newblock must be created after sdb.apply()
	bc.cBlock.SetBlocksRootHash(bc.sdb.GetRoot())
	bc.bestBlockNo = bc.bestBlockNo + 1
	bc.bestBlock = bc.cBlock
	bc.bestBlockId = types.ToBlockID(bc.cBlock.BlockHash())
	bc.blockIds = append(bc.blockIds, bc.bestBlockId)
	bc.blocks = append(bc.blocks, bc.cBlock)

	return nil
}

func (bc *DummyChain) DisConnectBlock() error {
	if len(bc.blockIds) == 1 {
		return errors.New("genesis block")
	}
	bc.bestBlockNo--
	bc.blockIds = bc.blockIds[0 : len(bc.blockIds)-1]
	bc.blocks = bc.blocks[0 : len(bc.blocks)-1]
	bc.bestBlockId = bc.blockIds[len(bc.blockIds)-1]

	bestBlock := bc.blocks[len(bc.blocks)-1]

	var sroot []byte
	if bestBlock != nil {
		sroot = bestBlock.GetHeader().GetBlocksRootHash()
	}
	return bc.sdb.SetRoot(sroot)
}

func (bc *DummyChain) Query(contract_name, queryInfo, expectedErr string, expectedRvs ...string) error {
	cState, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(contract.StrHash(contract_name)))
	if err != nil {
		return err
	}
	rv, err := contract.Query(contract.StrHash(contract_name), bc.newBState(), bc, cState, []byte(queryInfo))
	if expectedErr != "" {
		if err == nil {
			return fmt.Errorf("no error, expected: %s", expectedErr)
		}
		if !strings.Contains(err.Error(), expectedErr) {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}

	for _, ev := range expectedRvs {
		if ev != string(rv) {
			err = fmt.Errorf("expected: %s, but got: %s", ev, string(rv))
		} else {
			return nil
		}
	}
	return err
}

func (bc *DummyChain) QueryOnly(contract_name, queryInfo string, expectedErr string) (bool, string, error) {
	cState, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(contract.StrHash(contract_name)))
	if err != nil {
		return false, "", err
	}
	rv, err := contract.Query(contract.StrHash(contract_name), bc.newBState(), bc, cState, []byte(queryInfo))

	if expectedErr != "" {
		if err == nil {
			return false, "", fmt.Errorf("no error, expected: %s", expectedErr)
		}
		if !strings.Contains(err.Error(), expectedErr) {
			return false, "", err
		}
		return true, "", nil
	}

	if err != nil {
		return false, "", err
	}

	return false, string(rv), nil
}

func StrToAddress(name string) string {
	return types.EncodeAddress(contract.StrHash(name))
}
