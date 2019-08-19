package contract

// helper functions
import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/aergoio/aergo/config"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aergoio/aergo-lib/db"
	luac "github.com/aergoio/aergo/cmd/aergoluac/util"
	"github.com/aergoio/aergo/contract/system"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/minio/sha256-simd"
)

type DummyChain struct {
	sdb           *state.ChainStateDB
	bestBlock     *types.Block
	cBlock        *types.Block
	bestBlockNo   types.BlockNo
	bestBlockId   types.BlockID
	blockIds      []types.BlockID
	blocks        []*types.Block
	testReceiptDB db.DB
	tmpDir        string
	timeout       int
}

var addressRegexp *regexp.Regexp
var traceState bool

func init() {
	addressRegexp, _ = regexp.Compile("^[a-zA-Z0-9]+$")
	//	traceState = true
}

func LoadDummyChain(opts ...func(d *DummyChain)) (*DummyChain, error) {
	dataPath, err := ioutil.TempDir("", "data")
	if err != nil {
		return nil, err
	}
	bc := &DummyChain{
		sdb:    state.NewChainStateDB(),
		tmpDir: dataPath,
	}
	defer func() {
		if err != nil {
			bc.Release()
		}
	}()

	err = bc.sdb.Init(string(db.BadgerImpl), dataPath, nil, false)
	if err != nil {
		return nil, err
	}
	genesis := types.GetTestGenesis()
	_ = bc.sdb.SetGenesis(genesis, nil)
	bc.bestBlockNo = genesis.Block().BlockNo()
	bc.bestBlockId = genesis.Block().BlockID()
	bc.blockIds = append(bc.blockIds, bc.bestBlockId)
	bc.blocks = append(bc.blocks, genesis.Block())
	bc.testReceiptDB = db.NewDB(db.BadgerImpl, path.Join(dataPath, "receiptDB"))
	_ = LoadTestDatabase(dataPath) // sql database
	StartLStateFactory()
	HardforkConfig = config.AllEnabledHardforkConfig

	// To pass the governance tests.
	types.InitGovernance("dpos", true)
	system.InitGovernance("dpos")

	for _, opt := range opts {
		opt(bc)
	}
	return bc, nil
}

func (bc *DummyChain) Release() {
	bc.testReceiptDB.Close()
	_ = os.RemoveAll(bc.tmpDir)
}

func (bc *DummyChain) BestBlockNo() uint64 {
	return bc.bestBlockNo
}

func (bc *DummyChain) newBState() *state.BlockState {
	bc.cBlock = &types.Block{
		Header: &types.BlockHeader{
			PrevBlockHash: bc.bestBlockId[:],
			BlockNo:       bc.bestBlockNo + 1,
			Timestamp:     time.Now().UnixNano(),
		},
	}
	return state.NewBlockState(bc.sdb.OpenNewStateDB(bc.sdb.GetRoot()))
}

func (bc *DummyChain) BeginReceiptTx() db.Transaction {
	return bc.testReceiptDB.NewTx()
}

func (bc *DummyChain) GetABI(contract string) (*types.ABI, error) {
	cState, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(strHash(contract)))
	if err != nil {
		return nil, err
	}
	return GetABI(cState)
}

func (bc *DummyChain) getReceipt(txHash []byte) *types.Receipt {
	r := new(types.Receipt)
	if err := r.UnmarshalBinary(bc.testReceiptDB.Get(txHash)); err != nil {
		return nil
	}
	return r
}

func (bc *DummyChain) GetAccountState(name string) (*types.State, error) {
	return bc.sdb.GetStateDB().GetAccountState(types.ToAccountID(strHash(name)))
}

func (bc *DummyChain) GetStaking(name string) (*types.Staking, error) {
	scs, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	if err != nil {
		return nil, err
	}
	return system.GetStaking(scs, strHash(name))
}

func (bc *DummyChain) GetBlockByNo(blockNo types.BlockNo) (*types.Block, error) {
	return bc.blocks[blockNo], nil
}

func (bc *DummyChain) GetBestBlock() (*types.Block, error) {
	return bc.bestBlock, nil
}

type luaTx interface {
	run(bs *state.BlockState, bc *DummyChain, bi *types.BlockHeaderInfo, receiptTx db.Transaction, timeout <-chan struct{}) error
}

type luaTxAccount struct {
	name    []byte
	balance *big.Int
}

var _ luaTx = (*luaTxAccount)(nil)

func NewLuaTxAccount(name string, balance uint64) *luaTxAccount {
	return &luaTxAccount{
		name:    strHash(name),
		balance: new(big.Int).SetUint64(balance),
	}
}

func NewLuaTxAccountBig(name string, balance *big.Int) *luaTxAccount {
	return &luaTxAccount{
		name:    strHash(name),
		balance: balance,
	}
}

func (l *luaTxAccount) run(bs *state.BlockState, bc *DummyChain, bi *types.BlockHeaderInfo, receiptTx db.Transaction, timeout <-chan struct{}) error {

	id := types.ToAccountID(l.name)
	accountState, err := bs.GetAccountState(id)
	if err != nil {
		return err
	}
	updatedAccountState := *accountState
	updatedAccountState.Balance = l.balance.Bytes()
	return bs.PutState(id, &updatedAccountState)
}

type luaTxSend struct {
	sender   []byte
	receiver []byte
	balance  *big.Int
}

var _ luaTx = (*luaTxSend)(nil)

func NewLuaTxSendBig(sender, receiver string, balance *big.Int) *luaTxSend {
	return &luaTxSend{
		sender:   strHash(sender),
		receiver: strHash(receiver),
		balance:  balance,
	}
}

func (l *luaTxSend) run(bs *state.BlockState, bc *DummyChain, bi *types.BlockHeaderInfo, receiptTx db.Transaction, timeout <-chan struct{}) error {

	senderID := types.ToAccountID(l.sender)
	receiverID := types.ToAccountID(l.receiver)

	if senderID == receiverID {
		return fmt.Errorf("sender and receiever cannot be same")
	}

	senderState, err := bs.GetAccountState(senderID)
	if err != nil {
		return err
	} else if senderState.GetBalanceBigInt().Cmp(l.balance) < 0 {
		return fmt.Errorf("insufficient balance to sender")
	}
	receiverState, err := bs.GetAccountState(receiverID)
	if err != nil {
		return err
	}

	updatedSenderState := *senderState
	updatedSenderState.Balance = new(big.Int).Sub(updatedSenderState.GetBalanceBigInt(), l.balance).Bytes()
	err = bs.PutState(senderID, &updatedSenderState)
	if err != nil {
		return err
	}

	updatedReceiverState := *receiverState
	updatedReceiverState.Balance = new(big.Int).Add(updatedReceiverState.GetBalanceBigInt(), l.balance).Bytes()
	return bs.PutState(receiverID, &updatedReceiverState)
}

type luaTxCommon struct {
	sender   []byte
	contract []byte
	amount   *big.Int
	code     []byte
	id       uint64
}

type luaTxDef struct {
	luaTxCommon
	cErr error
}

var _ luaTx = (*luaTxDef)(nil)

func NewLuaTxDef(sender, contract string, amount uint64, code string) *luaTxDef {
	L := luac.NewLState()
	if L == nil {
		return &luaTxDef{cErr: newVmStartError()}
	}
	defer luac.CloseLState(L)
	b, err := luac.Compile(L, code)
	if err != nil {
		return &luaTxDef{cErr: err}
	}
	codeWithInit := make([]byte, 4+len(b))
	binary.LittleEndian.PutUint32(codeWithInit, uint32(4+len(b)))
	copy(codeWithInit[4:], b)
	return &luaTxDef{
		luaTxCommon: luaTxCommon{
			sender:   strHash(sender),
			contract: strHash(contract),
			code:     codeWithInit,
			amount:   new(big.Int).SetUint64(amount),
			id:       newTxId(),
		},
		cErr: nil,
	}
}

func getCompiledABI(code string) ([]byte, error) {

	L := luac.NewLState()
	if L == nil {
		return nil, newVmStartError()
	}
	defer luac.CloseLState(L)
	b, err := luac.Compile(L, code)
	if err != nil {
		return nil, err
	}

	codeLen := binary.LittleEndian.Uint32(b[:4])

	return b[4+codeLen:], nil
}

func NewRawLuaTxDefBig(sender, contract string, amount *big.Int, code string) *luaTxDef {

	byteAbi, err := getCompiledABI(code)
	if err != nil {
		return &luaTxDef{cErr: err}
	}

	byteCode := []byte(code)
	payload := make([]byte, 8+len(byteCode)+len(byteAbi))
	binary.LittleEndian.PutUint32(payload[0:], uint32(len(byteCode)+len(byteAbi)+8))
	binary.LittleEndian.PutUint32(payload[4:], uint32(len(byteCode)))
	codeLen := copy(payload[8:], byteCode)
	copy(payload[8+codeLen:], byteAbi)

	return &luaTxDef{
		luaTxCommon: luaTxCommon{
			sender:   strHash(sender),
			contract: strHash(contract),
			code:     payload,
			amount:   amount,
			id:       newTxId(),
		},
		cErr: nil,
	}
}

func strHash(d string) []byte {
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

var luaTxId uint64 = 0

func newTxId() uint64 {
	luaTxId++
	return luaTxId
}

func (l *luaTxDef) hash() []byte {
	h := sha256.New()
	h.Write([]byte(strconv.FormatUint(l.id, 10)))
	b := h.Sum(nil)
	return b
}

func (l *luaTxDef) Constructor(args string) *luaTxDef {
	argsLen := len([]byte(args))
	if argsLen == 0 || l.cErr != nil {
		return l
	}

	code := make([]byte, len(l.code)+argsLen)
	codeLen := copy(code[0:], l.code)
	binary.LittleEndian.PutUint32(code[0:], uint32(codeLen))
	copy(code[codeLen:], args)

	l.code = code

	return l
}

func contractFrame(l *luaTxCommon, bs *state.BlockState,
	run func(s, c *state.V, id types.AccountID, cs *state.ContractState) error) error {

	creatorId := types.ToAccountID(l.sender)
	creatorState, err := bs.GetAccountStateV(l.sender)
	if err != nil {
		return err
	}

	contractId := types.ToAccountID(l.contract)
	contractState, err := bs.GetAccountStateV(l.contract)
	if err != nil {
		return err
	}

	eContractState, err := bs.OpenContractState(contractId, contractState.State())
	if err != nil {
		return err
	}

	creatorState.SubBalance(l.amount)
	contractState.AddBalance(l.amount)
	err = run(creatorState, contractState, contractId, eContractState)
	if err != nil {
		return err
	}

	err = bs.PutState(creatorId, creatorState.State())
	if err != nil {
		return err
	}
	return bs.PutState(contractId, contractState.State())

}

func (l *luaTxDef) run(bs *state.BlockState, bc *DummyChain, bi *types.BlockHeaderInfo, receiptTx db.Transaction, timeout <-chan struct{}) error {

	if l.cErr != nil {
		return l.cErr
	}

	return contractFrame(&l.luaTxCommon, bs,
		func(sender, contract *state.V, contractId types.AccountID, eContractState *state.ContractState) error {
			contract.State().SqlRecoveryPoint = 1

			stateSet := NewContext(bs, nil, sender, contract, eContractState, sender.ID(), l.hash(), bi, "", true,
				false, contract.State().SqlRecoveryPoint, ChainService, l.luaTxCommon.amount, timeout)

			if traceState {
				stateSet.traceFile, _ =
					os.OpenFile("test.trace", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
				defer func() {
					_ = stateSet.traceFile.Close()
				}()
			}

			_, _, _, err := Create(eContractState, l.code, l.contract, stateSet)
			if err != nil {
				return err
			}
			err = bs.StageContractState(eContractState)
			if err != nil {
				return err
			}
			return nil
		},
	)
}

type luaTxCall struct {
	luaTxCommon
	expectedErr string
}

var _ luaTx = (*luaTxCall)(nil)

func NewLuaTxCall(sender, contract string, amount uint64, code string) *luaTxCall {
	return &luaTxCall{
		luaTxCommon: luaTxCommon{
			sender:   strHash(sender),
			contract: strHash(contract),
			amount:   new(big.Int).SetUint64(amount),
			code:     []byte(code),
			id:       newTxId(),
		},
	}
}

func NewLuaTxCallBig(sender, contract string, amount *big.Int, code string) *luaTxCall {
	return &luaTxCall{
		luaTxCommon: luaTxCommon{
			sender:   strHash(sender),
			contract: strHash(contract),
			amount:   amount,
			code:     []byte(code),
			id:       newTxId(),
		},
	}
}

func (l *luaTxCall) hash() []byte {
	h := sha256.New()
	h.Write([]byte(strconv.FormatUint(l.id, 10)))
	b := h.Sum(nil)
	return b
}

func (l *luaTxCall) Fail(expectedErr string) *luaTxCall {
	l.expectedErr = expectedErr
	return l
}

func (l *luaTxCall) run(bs *state.BlockState, bc *DummyChain, bi *types.BlockHeaderInfo, receiptTx db.Transaction, timeout <-chan struct{}) error {
	err := contractFrame(&l.luaTxCommon, bs,
		func(sender, contract *state.V, contractId types.AccountID, eContractState *state.ContractState) error {
			stateSet := NewContext(bs, bc, sender, contract, eContractState, sender.ID(), l.hash(), bi, "", true,
				false, contract.State().SqlRecoveryPoint, ChainService, l.luaTxCommon.amount, timeout)
			if traceState {
				stateSet.traceFile, _ =
					os.OpenFile("test.trace", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
				defer func() {
					_ = stateSet.traceFile.Close()
				}()
			}
			rv, evs, _, err := Call(eContractState, l.code, l.contract, stateSet)
			if err != nil {
				r := types.NewReceipt(l.contract, err.Error(), "")
				r.TxHash = l.hash()
				b, _ := r.MarshalBinary()
				receiptTx.Set(l.hash(), b)
				return err
			}
			_ = bs.StageContractState(eContractState)
			r := types.NewReceipt(l.contract, "SUCCESS", rv)
			r.Events = evs
			r.TxHash = l.hash()
			blockHash := make([]byte, 32)
			for _, ev := range evs {
				ev.TxHash = r.TxHash
				ev.BlockHash = blockHash
			}
			b, _ := r.MarshalBinary()
			receiptTx.Set(l.hash(), b)
			return nil
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

func (bc *DummyChain) ConnectBlock(txs ...luaTx) error {
	blockState := bc.newBState()
	tx := bc.BeginReceiptTx()
	defer tx.Commit()
	defer CloseDatabase()

	timeout := make(chan struct{})
	go func() {
		<-time.Tick(time.Duration(bc.timeout) * time.Millisecond)
		timeout <- struct{}{}
	}()
	for _, x := range txs {
		if err := x.run(blockState, bc, types.NewBlockHeaderInfo(bc.cBlock), tx, timeout); err != nil {
			return err
		}
	}
	err := SaveRecoveryPoint(blockState)
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

func (bc *DummyChain) Query(contract, queryInfo, expectedErr string, expectedRvs ...string) error {
	cState, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(strHash(contract)))
	if err != nil {
		return err
	}
	rv, err := Query(strHash(contract), bc.newBState(), bc, cState, []byte(queryInfo))
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

func (bc *DummyChain) QueryOnly(contract, queryInfo string) (string, error) {
	cState, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(strHash(contract)))
	if err != nil {
		return "", err
	}
	rv, err := Query(strHash(contract), bc.newBState(), nil, cState, []byte(queryInfo))

	if err != nil {
		return "", err
	}

	return string(rv), nil
}

func StrToAddress(name string) string {
	return types.EncodeAddress(strHash(name))
}
