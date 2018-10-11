package contract

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
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

	rv, err := bc.query("return_num", `{"Name":"return_num", "Args":[]}`, t)
	if err != nil {
		t.Error(err)
	}
	if rv != "[10]" {
		t.Errorf("expected: %s, bug got: %s", "[10]", rv)
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
	receipt := types.NewReceiptFromBytes(TempReceiptDb.Get(tx.hash()))
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
	receipt := types.NewReceiptFromBytes(TempReceiptDb.Get(tx.hash()))
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
	cState, err := bc.sdb.OpenContractStateAccount(types.ToAccountID(strHash("hello")))
	if err != nil {
		t.Error(err)
	}
	abi, err := GetABI(cState)
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

	rv, err := bc.query("query", `{"Name":"inc", "Args":[]}`, t)
	if err == nil {
		t.Error("error expected")

	}
	if !strings.Contains(err.Error(), "not permitted set in query") {
		t.Errorf("failed check error: %v", err)
	}

	rv, err = bc.query("query", `{"Name":"query", "Args":["key1"]}`, t)
	if err != nil {
		t.Errorf("contract query error: %v", err)
	}
	if rv != "[1]" {
		t.Errorf("expected: %s, bug got: %s", "[1]", rv)
	}
}

type blockChain struct {
	sdb         *state.ChainStateDB
	bestBlock   *types.Block
	cBlock      *types.Block
	bestBlockNo types.BlockNo
	bestBlockId types.BlockID
	rTx         db.Transaction
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

type luaTxDef struct {
	sender   []byte
	contract []byte
	amount   uint64
	code     []byte
	id       uint64
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
		sender:   strHash(sender),
		contract: strHash(contract),
		code:     codeWithInit,
		amount:   amount,
		id:       newTxId(),
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

func (l *luaTxDef) run(sdb *state.ChainStateDB, bs *state.BlockState, blockNo uint64, ts int64) error {
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
	eContractState, err := sdb.OpenContractState(&uContractState)
	if err != nil {
		return err
	}
	contractState.SqlRecoveryPoint = 1
	sqlTx, err := BeginTx(contractId, contractState.SqlRecoveryPoint)
	if err != nil {
		return err
	}
	err = sqlTx.Savepoint()
	if err != nil {
		return err
	}

	bcCtx := NewContext(sdb, bs, creatorState, eContractState,
		types.EncodeAddress(l.sender), hex.EncodeToString(l.hash()), blockNo, ts,
		"", 1, types.EncodeAddress(l.contract),
		0, nil, sqlTx.GetHandle())

	err = Create(eContractState, l.code, l.contract, l.hash(), bcCtx, bs.ReceiptTx())
	if err != nil {
		_ = sqlTx.RollbackToSavepoint()
		return err
	}
	err = sdb.CommitContractState(eContractState)
	if err != nil {
		_ = sqlTx.RollbackToSavepoint()
		return err
	}
	err = sqlTx.Release()
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

type luaTxCall struct {
	sender   []byte
	contract []byte
	amount   uint64
	code     []byte
	id       uint64
}

func newLuaTxCall(sender, contract string, amount uint64, code string) *luaTxCall {
	return &luaTxCall{
		sender:   strHash(sender),
		contract: strHash(contract),
		amount:   amount,
		code:     []byte(code),
		id:       newTxId(),
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
	creatorId := types.ToAccountID([]byte(l.sender))
	creatorState, err := bs.GetAccountState(creatorId)
	if err != nil {
		return err
	}

	contractId := types.ToAccountID([]byte(l.contract))
	contractState, err := bs.GetAccountState(contractId)
	if err != nil {
		return err
	}

	uContractState := types.State(*contractState)
	eContractState, err := sdb.OpenContractState(&uContractState)
	if err != nil {
		return err
	}
	sqlTx, err := BeginTx(contractId, contractState.SqlRecoveryPoint)
	if err != nil {
		return err
	}
	sqlTx.Savepoint()

	bcCtx := NewContext(sdb, bs, creatorState, eContractState,
		types.EncodeAddress(l.sender), hex.EncodeToString(l.hash()), blockNo, ts,
		"", 1, types.EncodeAddress(l.contract),
		0, nil, sqlTx.GetHandle())

	err = Call(eContractState, l.code, l.contract, l.hash(), bcCtx, bs.ReceiptTx())
	if err != nil {
		_ = sqlTx.RollbackToSavepoint()
		return err
	}
	err = sdb.CommitContractState(eContractState)
	if err != nil {
		_ = sqlTx.RollbackToSavepoint()
		return err
	}
	err = sqlTx.Release()
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
	return nil
}

func (bc *blockChain) query(contract, queryInfo string, t *testing.T) (string, error) {
	cState, err := bc.sdb.OpenContractStateAccount(types.ToAccountID(strHash(contract)))
	if err != nil {
		return "", err
	}
	ret, err := Query(strHash(contract), cState, []byte(queryInfo))
	if err != nil {
		return "", err
	}
	return string(ret), nil
}
