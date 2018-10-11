package contract

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/minio/sha256-simd"
)

const (
	helloCode = `function hello(say) return "Hello " .. say end abi.register(hello)`

	systemCode = `function testState()
		string.format("creator: %s",system.getContractID())
		string.format("timestamp: %d",system.getTimestamp())
		string.format("blockheight: %d",system.getBlockheight())
		system.setItem("key1", 999)
		string.format("getitem : %s",system.getItem("key1"))
		return system.getSender(), system.getTxhash(),system.getContractID(), system.getTimestamp(), system.getBlockheight(), system.getItem("key1")
	  end 
abi.register(testState)`

	queryCode = `function inc()
		a = system.getItem("key1")
		if (a == nil) then
			system.setItem("key1", 1)
			return
		end
		system.setItem("key1", a + 1)
	end
	function query(a)
			return system.getItem(a)
	end
	abi.register(inc, query)`
)

func TestReturn(t *testing.T) {
	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "return_num", 10, "function return_num() return 10 end abi.register(return_num)"),
		newLuaTxCall("ktlee", "return_num", 10, `{"Name":"return_num", "Args":[]}`),
	)

	err := bc.query("return_num", `{"Name":"return_num", "Args":[]}`, "", "[10]")
	if err != nil {
		t.Error(err)
	}
}

func TestContractHello(t *testing.T) {
	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
	)
	bc.connectBlock(
		newLuaTxDef("ktlee", "hello", 1, helloCode),
	)
	tx := newLuaTxCall("ktlee", "hello", 1, `{"Name":"hello", "Args":["World"]}`)
	bc.connectBlock(tx)
	receipt := bc.getReceipt(tx.hash())
	if receipt.GetRet() != `["Hello World"]` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
}

func TestContractSystem(t *testing.T) {
	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
	)
	bc.connectBlock(
		newLuaTxDef("ktlee", "system", 1, systemCode),
	)
	tx := newLuaTxCall("ktlee", "system", 1, `{"Name":"testState", "Args":[]}`)
	bc.connectBlock(tx)
	receipt := bc.getReceipt(tx.hash())
	exRv := fmt.Sprintf(`["Amg6nZWXKB6YpNgBPv9atcjdm6hnFvs5wMdRgb2e9DmaF5g9muF2","0ce7f6c011776e8db7cd330b54174fd76f7d0216b612387a5ffcfb81e6f0919683","AmhNNBNY7XFk4p5ym4CJf8nTcRTEHjWzAeXJfhP71244CjBCAQU3",%d,3,999]`, bc.cBlock.Header.Timestamp)
	if receipt.GetRet() != exRv {
		t.Errorf("expected: %s, but got: %s", exRv, receipt.GetRet())
	}
}

func TestGetABI(t *testing.T) {
	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "hello", 1,
			`function hello(say) return "Hello " .. say end abi.register(hello)`),
	)
	abi, err := bc.getABI("hello")
	if err != nil {
		t.Error(err)
	}
	b, err := json.Marshal(abi)
	if err != nil {
		t.Error(err)
	}
	if string(b) != `{"version":"0.1","language":"lua","functions":[{"name":"hello","arguments":[{"name":"say"}]}]}` {
		t.Error(string(b))
	}
}

func TestContractQuery(t *testing.T) {
	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
	)
	bc.connectBlock(
		newLuaTxDef("ktlee", "query", 1, queryCode),
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
	)

	ktlee, err := bc.getAccountState("ktlee")
	if err != nil {
		t.Error(err)
	}
	if ktlee.Balance != uint64(98) {
		t.Error(ktlee.Balance)
	}
	t.Log(ktlee.Balance)
	query, err := bc.getAccountState("query")
	if err != nil {
		t.Error(err)
	}
	if query.Balance != uint64(2) {
		t.Error(query.Balance)
	}
	t.Log(query.Balance)

	err = bc.query("query", `{"Name":"inc", "Args":[]}`, "not permitted set in query", "")
	if err != nil {
		t.Error(err)
	}

	err = bc.query("query", `{"Name":"query", "Args":["key1"]}`, "", "[1]")
	if err != nil {
		t.Error(err)
	}
}

func TestRollback(t *testing.T) {
	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
	)
	bc.connectBlock(
		newLuaTxDef("ktlee", "query", 1, queryCode),
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
	)
	bc.connectBlock(
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
	)
	bc.connectBlock(
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
	)

	err := bc.query("query", `{"Name":"query", "Args":["key1"]}`, "", "[5]")
	if err != nil {
		t.Error(err)
	}

	err = bc.disconnectBlock()
	if err != nil {
		t.Error(err)
	}
	err = bc.query("query", `{"Name":"query", "Args":["key1"]}`, "", "[3]")
	if err != nil {
		t.Error(err)
	}

	err = bc.disconnectBlock()
	if err != nil {
		t.Error(err)
	}

	err = bc.query("query", `{"Name":"query", "Args":["key1"]}`, "", "[1]")
	if err != nil {
		t.Error(err)
	}

	bc.connectBlock(
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
	)

	err = bc.query("query", `{"Name":"query", "Args":["key1"]}`, "", "[2]")
	if err != nil {
		t.Error(err)
	}
}

type blockChain struct {
	sdb         *state.ChainStateDB
	bestBlock   *types.Block
	cBlock      *types.Block
	bestBlockNo types.BlockNo
	bestBlockId types.BlockID
	rTx         db.Transaction
	blockIds    []types.BlockID
}

func loadBlockChain(t *testing.T) *blockChain {
	bc := &blockChain{sdb: state.NewChainStateDB()}
	dataPath, err := ioutil.TempDir("", "data")
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	err = bc.sdb.Init(dataPath)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	genesis := types.GetTestGenesis()
	bc.sdb.SetGenesis(genesis)
	bc.bestBlockNo = genesis.Block.BlockNo()
	bc.bestBlockId = genesis.Block.BlockID()
	bc.blockIds = append(bc.blockIds, bc.bestBlockId)

	TempReceiptDb = db.NewDB(db.BadgerImpl, path.Join(dataPath, "receiptDB"))
	LoadDatabase(dataPath)
	return bc
}

func (bc *blockChain) newBState() *state.BlockState {
	b := types.Block{
		Header: &types.BlockHeader{
			PrevBlockHash: []byte(bc.bestBlockId.String()),
			BlockNo:       bc.bestBlockNo + 1,
			Timestamp:     time.Now().Unix(),
		},
	}
	bc.cBlock = &b
	// blockInfo := types.NewBlockInfo(b.BlockNo(), b.BlockID(), bc.bestBlockId)
	return state.NewBlockState(b.BlockID(), bc.sdb.OpenNewStateDB(bc.sdb.GetRoot()), TempReceiptDb.NewTx())
}

func (bc *blockChain) getABI(contract string) (*types.ABI, error) {
	cState, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(strHash(contract)))
	if err != nil {
		return nil, err
	}
	return GetABI(cState)
}

func (bc *blockChain) getReceipt(txHash []byte) *types.Receipt {
	return types.NewReceiptFromBytes(TempReceiptDb.Get(txHash))
}

func (bc *blockChain) getAccountState(name string) (*types.State, error) {
	return bc.sdb.GetStateDB().GetAccountState(types.ToAccountID(strHash(name)))
}

type luaTx interface {
	run(sdb *state.ChainStateDB, bs *state.BlockState, blockNo uint64, ts int64) error
}

type luaTxAccount struct {
	name    []byte
	balance uint64
}

func newLuaTxAccount(name string, balance uint64) *luaTxAccount {
	return &luaTxAccount{
		name:    strHash(name),
		balance: balance,
	}
}

func (l *luaTxAccount) run(sdb *state.ChainStateDB, bs *state.BlockState, blockNo uint64, ts int64) error {
	id := types.ToAccountID(l.name)
	accountState, err := bs.GetAccountState(id)
	if err != nil {
		return err
	}
	updatedAccountState := types.State(*accountState)
	updatedAccountState.Balance = l.balance
	bs.PutState(id, &updatedAccountState)
	return nil
}

type luaTxCommon struct {
	sender   []byte
	contract []byte
	amount   uint64
	code     []byte
	id       uint64
}

type luaTxDef struct {
	luaTxCommon
}

func newLuaTxDef(sender, contract string, amount uint64, code string) *luaTxDef {
	luac := exec.Command("../bin/aergoluac", "--payload")
	stdin, err := luac.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, code)
	}()
	out, err := luac.Output()
	if err != nil {
		log.Fatal(err)
	}
	b, err := util.DecodeCode(string(out))
	if err != nil {
		log.Fatal(err)
	}
	codeWithInit := make([]byte, 4+len(b))
	binary.LittleEndian.PutUint32(codeWithInit, uint32(4+len(b)))
	copy(codeWithInit[4:], b)
	return &luaTxDef{
		luaTxCommon{
			sender:   strHash(sender),
			contract: strHash(contract),
			code:     codeWithInit,
			amount:   amount,
			id:       newTxId(),
		},
	}
}

func strHash(d string) []byte {
	h := sha256.New()
	h.Write([]byte(d))
	b := h.Sum(nil)
	b = append([]byte{0x0C}, b...)
	return b
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
	b = append([]byte{0x0C}, b...)
	return b
}

func contractFrame(l *luaTxCommon, bs *state.BlockState,
	run func(s, c *types.State, id types.AccountID, cs *state.ContractState) error) error {

	creatorId := types.ToAccountID(l.sender)
	creatorState, err := bs.GetAccountState(creatorId)
	if err != nil {
		return err
	}

	contractId := types.ToAccountID(l.contract)
	contractState, err := bs.GetAccountState(contractId)
	if err != nil {
		return err
	}

	uContractState := types.State(*contractState)
	eContractState, err := bs.StateDB.OpenContractState(&uContractState)
	if err != nil {
		return err
	}

	err = run(creatorState, &uContractState, contractId, eContractState)
	if err != nil {
		return err
	}

	uCallerState := types.State(*creatorState)
	uCallerState.Balance -= l.amount
	uContractState.Balance += l.amount

	bs.PutState(creatorId, &uCallerState)
	bs.PutState(contractId, &uContractState)
	return nil

}
func (l *luaTxDef) run(sdb *state.ChainStateDB, bs *state.BlockState, blockNo uint64, ts int64) error {
	return contractFrame(&l.luaTxCommon, bs,
		func(senderState, uContractState *types.State, contractId types.AccountID, eContractState *state.ContractState) error {
			uContractState.SqlRecoveryPoint = 1
			sqlTx, err := BeginTx(contractId, uContractState.SqlRecoveryPoint)
			if err != nil {
				return err
			}
			err = sqlTx.Savepoint()
			if err != nil {
				return err
			}

			bcCtx := NewContext(sdb, bs, senderState, eContractState,
				types.EncodeAddress(l.sender), hex.EncodeToString(l.hash()), blockNo, ts,
				"", 1, types.EncodeAddress(l.contract),
				0, nil, sqlTx.GetHandle())

			err = Create(eContractState, l.code, l.contract, l.hash(), bcCtx, bs.ReceiptTx())
			if err != nil {
				_ = sqlTx.RollbackToSavepoint()
				return err
			}
			err = bs.StateDB.CommitContractState(eContractState)
			if err != nil {
				_ = sqlTx.RollbackToSavepoint()
				return err
			}
			err = sqlTx.Release()
			if err != nil {
				return err
			}
			return nil
		},
	)
}

type luaTxCall struct {
	luaTxCommon
}

func newLuaTxCall(sender, contract string, amount uint64, code string) *luaTxCall {
	return &luaTxCall{
		luaTxCommon{
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
	b = append([]byte{0x0C}, b...)
	return b
}

func (l *luaTxCall) run(sdb *state.ChainStateDB, bs *state.BlockState, blockNo uint64, ts int64) error {
	return contractFrame(&l.luaTxCommon, bs,
		func(senderState, uContractState *types.State, contractId types.AccountID, eContractState *state.ContractState) error {
			sqlTx, err := BeginTx(contractId, uContractState.SqlRecoveryPoint)
			if err != nil {
				return err
			}
			sqlTx.Savepoint()

			bcCtx := NewContext(sdb, bs, senderState, eContractState,
				types.EncodeAddress(l.sender), hex.EncodeToString(l.hash()), blockNo, ts,
				"", 1, types.EncodeAddress(l.contract),
				0, nil, sqlTx.GetHandle())

			err = Call(eContractState, l.code, l.contract, l.hash(), bcCtx, bs.ReceiptTx())
			if err != nil {
				_ = sqlTx.RollbackToSavepoint()
				return err
			}
			err = bs.StateDB.CommitContractState(eContractState)
			if err != nil {
				_ = sqlTx.RollbackToSavepoint()
				return err
			}
			err = sqlTx.Release()
			if err != nil {
				return err
			}
			return nil
		},
	)
}

func (bc *blockChain) connectBlock(txs ...luaTx) error {
	blockState := bc.newBState()
	defer blockState.CommitReceipt()

	for _, x := range txs {
		x.run(bc.sdb, blockState, bc.cBlock.Header.BlockNo, bc.cBlock.Header.Timestamp)
	}
	err := SaveRecoveryPoint(bc.sdb, blockState)
	if err != nil {
		return err
	}
	err = bc.sdb.Apply(blockState)
	if err != nil {
		return err
	}
	bc.bestBlockNo = bc.bestBlockNo + 1
	bc.bestBlockId = blockState.GetBlockHash()
	bc.blockIds = append(bc.blockIds, bc.bestBlockId)
	return nil
}

func (bc *blockChain) disconnectBlock() error {
	if len(bc.blockIds) == 1 {
		return errors.New("genesis block")
	}
	bc.bestBlockNo--
	bc.blockIds = bc.blockIds[0 : len(bc.blockIds)-1]
	bc.bestBlockId = bc.blockIds[len(bc.blockIds)-1]
	return bc.sdb.Rollback(bc.bestBlockId)
}

func (bc *blockChain) query(contract, queryInfo string, expectedErr, expectedRv string) error {
	cState, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(strHash(contract)))
	if err != nil {
		return err
	}
	rv, err := Query(strHash(contract), cState, []byte(queryInfo))
	if expectedErr != "" {
		if err == nil || !strings.Contains(err.Error(), expectedErr) {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}

	if expectedRv != string(rv) {
		return fmt.Errorf("expected: %s, but got: %s", expectedRv, string(rv))
	}
	return nil
}
