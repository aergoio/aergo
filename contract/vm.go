/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package contract

/*
#cgo CFLAGS: -I${SRCDIR}/../libtool/include/luajit-2.0
#cgo LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a -lm

#include <stdlib.h>
#include <string.h>
#include "vm.h"
*/
import "C"
import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
	"unsafe"

	"github.com/aergoio/aergo/state"
)

const DbName = "contracts.db"

var (
	ctrLog      *log.Logger
	DB          db.DB
	contractMap stateMap
)

type Contract struct {
	code    []byte
	address []byte
}

type stateMap struct {
	states map[string]*state.ContractState
	mu     sync.Mutex
}

type LState = C.struct_lua_State
type LBlockchainCtx = C.struct_blockchain_ctx

type Executor struct {
	L             *LState
	contract      *Contract
	err           error
	blockchainCtx *LBlockchainCtx
	jsonRet       string
}

func init() {
	ctrLog = log.NewLogger("contract")
	contractMap.init()
}

func NewContext(contractState *state.ContractState, Sender, txHash []byte, blockHeight uint64,
	timestamp int64, node string, confirmed bool, contractID []byte, query bool) *LBlockchainCtx {

	var iConfirmed, isQuery int
	if confirmed {
		iConfirmed = 1
	}
	if query {
		isQuery = 1
	}
	enContractId := types.EncodeAddress(contractID)
	enTxHash := hex.EncodeToString(txHash)

	stateKey := fmt.Sprintf("%s%s", enContractId, enTxHash)
	contractMap.register(stateKey, contractState)

	return &LBlockchainCtx{
		stateKey:    C.CString(stateKey),
		sender:      C.CString(types.EncodeAddress(Sender)),
		txHash:      C.CString(enTxHash),
		blockHeight: C.ulonglong(blockHeight),
		timestamp:   C.longlong(timestamp),
		node:        C.CString(node),
		confirmed:   C.int(iConfirmed),
		contractId:  C.CString(enContractId),
		isQuery:     C.int(isQuery),
	}
}

func newLState() *LState {
	return C.vm_newstate()
}

func (L *LState) Close() {
	if L != nil {
		C.lua_close(L)
	}
}

func newExecutor(contract *Contract, bcCtx *LBlockchainCtx) *Executor {
	address := C.CString(types.EncodeAddress(contract.address))
	defer C.free(unsafe.Pointer(address))

	ce := &Executor{
		contract: contract,
		L:        newLState(),
	}
	if ce.L == nil {
		ctrLog.Error().Str("error", "Failed: create lua state")
		ce.err = errors.New("Failed: create lua state")
		return ce
	}
	if cErrMsg := C.vm_loadbuff(
		ce.L,
		(*C.char)(unsafe.Pointer(&contract.code[0])),
		C.size_t(len(contract.code)),
		address,
		bcCtx,
	); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		C.free(unsafe.Pointer(cErrMsg))
		ctrLog.Error().Str("error", errMsg)
		ce.err = errors.New(errMsg)
	}
	return ce
}

func (ce *Executor) call(ci *types.CallInfo) {
	if ce.err != nil {
		return
	}
	abiStr := C.CString("abi")
	callStr := C.CString("call")
	abiName := C.CString(ci.Name)

	defer C.free(unsafe.Pointer(abiStr))
	defer C.free(unsafe.Pointer(callStr))
	defer C.free(unsafe.Pointer(abiName))

	C.vm_getfield(ce.L, abiStr)
	C.lua_getfield(ce.L, -1, callStr)
	C.lua_pushstring(ce.L, abiName)
	for _, v := range ci.Args {
		switch arg := v.(type) {
		case string:
			argC := C.CString(arg)
			C.lua_pushstring(ce.L, argC)
			C.free(unsafe.Pointer(argC))
		case int:
			C.lua_pushinteger(ce.L, C.long(arg))
		case bool:
			var b int
			if arg {
				b = 1
			}
			C.lua_pushboolean(ce.L, C.int(b))
		default:
			ce.err = errors.New("unsupported type")
			return
		}
	}
	nret := C.int(0)
	if cErrMsg := C.vm_pcall(ce.L, C.int(len(ci.Args)+1), &nret); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		C.free(unsafe.Pointer(cErrMsg))
		ctrLog.Warn().Str("error", errMsg).Msgf("contract %s", types.EncodeAddress(ce.contract.address))
		ce.err = errors.New(errMsg)
		return
	}
	ce.jsonRet = C.GoString(C.vm_get_json_ret(ce.L, nret))
}

func (ce *Executor) close() {
	if ce != nil {
		ce.L.Close()
		if ce.blockchainCtx != nil {
			context := ce.blockchainCtx
			contractMap.unregister(C.GoString(context.stateKey))

			C.free(unsafe.Pointer(context.sender))
			C.free(unsafe.Pointer(context.txHash))
			C.free(unsafe.Pointer(context.node))
			C.free(unsafe.Pointer(context.stateKey))
			C.free(unsafe.Pointer(context))
		}
	}
}

func Call(contractState *state.ContractState, code, contractAddress, txHash []byte, bcCtx *LBlockchainCtx, dbTx db.Transaction) error {
	var err error
	var ci types.CallInfo
	contract := getContract(contractState, contractAddress)
	if contract != nil {
		err = json.Unmarshal(code, &ci)
		if err != nil {
			ctrLog.Warn().AnErr("error", err).Msgf("contract %s", types.EncodeAddress(contractAddress))
		}
	} else {
		err = fmt.Errorf("cannot find contract %s", string(contractAddress))
		ctrLog.Warn().AnErr("err", err)
	}
	var ce *Executor
	if err == nil {
		ctrLog.Debug().Str("abi", string(code)).Msgf("contract %s", types.EncodeAddress(contractAddress))
		ce = newExecutor(contract, bcCtx)
		defer ce.close()
		ce.call(&ci)
		err = ce.err
	}
	var receipt types.Receipt
	if err == nil {
		receipt = types.NewReceipt(contractAddress, "SUCCESS", ce.jsonRet)
	} else {
		receipt = types.NewReceipt(contractAddress, err.Error(), "")
	}
	dbTx.Set(txHash, receipt.Bytes())
	return err
}

func Create(contractState *state.ContractState, code, contractAddress, txHash []byte, dbTx db.Transaction) error {
	ctrLog.Debug().Str("contractAddress", types.EncodeAddress(contractAddress)).Msg("new contract is deployed")
	err := contractState.SetCode(code)
	if err != nil {
		return err
	}
	receipt := types.NewReceipt(contractAddress, "CREATED", "{}")
	dbTx.Set(txHash, receipt.Bytes())

	return nil
}

func Query(contractAddress []byte, contractState *state.ContractState, queryInfo []byte) ([]byte, error) {
	var ci types.CallInfo
	var err error
	contract := getContract(contractState, contractAddress)
	if contract != nil {
		err = json.Unmarshal(queryInfo, &ci)
		if err != nil {
			ctrLog.Warn().AnErr("error", err).Msgf("contract %s", types.EncodeAddress(contractAddress))
		}
	} else {
		err = fmt.Errorf("cannot find contract %s", types.EncodeAddress(contractAddress))
		ctrLog.Warn().AnErr("err", err)
	}
	if err != nil {
		return nil, err
	}
	var ce *Executor

	bcCtx := NewContext(contractState, contractAddress, nil,
		0, 0, "", false, contractAddress, true)
	ctrLog.Debug().Str("abi", string(queryInfo)).Msgf("contract %s", types.EncodeAddress(contractAddress))
	ce = newExecutor(contract, bcCtx)
	defer ce.close()
	ce.call(&ci)
	err = ce.err

	return []byte(ce.jsonRet), err
}

func getContract(contractState *state.ContractState, contractAddress []byte) *Contract {
	val, err := contractState.GetCode()

	if err != nil {
		return nil
	}
	if len(val) > 0 {
		l := binary.LittleEndian.Uint32(val[0:])
		return &Contract{
			code:    val[4 : 4+l],
			address: contractAddress[:],
		}
	}
	return nil
}

func GetReceipt(txHash []byte) (*types.Receipt, error) {
	val := DB.Get(txHash)
	if len(val) == 0 {
		return nil, errors.New("cannot find a receipt")
	}
	return types.NewReceiptFromBytes(val), nil
}

func GetABI(contractState *state.ContractState, contractAddress []byte) (*types.ABI, error) {
	val, err := contractState.GetCode()
	if err != nil {
		return nil, err
	}
	if len(val) == 0 {
		return nil, errors.New("cannot find contract")
	}
	l := binary.LittleEndian.Uint32(val[0:])
	abi := new(types.ABI)
	if err := json.Unmarshal(val[4+l:], abi); err != nil {
		return nil, err
	}
	return abi, nil
}

func (sm *stateMap) init() {
	sm.states = make(map[string]*state.ContractState)
}

func (sm *stateMap) register(key string, state *state.ContractState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.states[key] != nil {
		err := fmt.Errorf("already exists contract state: %s", key)
		ctrLog.Warn().AnErr("err", err)
	}
	sm.states[key] = state
}

func (sm *stateMap) unregister(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.states[key] == nil {
		err := fmt.Errorf("cannot find contract state: %s", key)
		ctrLog.Warn().AnErr("err", err)
	}

	delete(sm.states, key)
}

func (sm *stateMap) lookup(key string) *state.ContractState {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.states[key]
}

//export LuaSetDB
func LuaSetDB(L *LState, stateKey *C.char, key *C.char, value *C.char) C.int {
	stateKeyString := C.GoString(stateKey)
	keyString := C.GoString(key)
	valueString := C.GoString(value)

	stateDb := contractMap.lookup(stateKeyString)
	if stateDb == nil {
		errMsg := C.CString("[System.LuaSetDB]not found contract state")
		C.lua_pushstring(L, errMsg)
		C.free(unsafe.Pointer(errMsg))
		return -1
	}

	err := stateDb.SetData([]byte(keyString), []byte(valueString))
	if err != nil {
		errMsg := C.CString(err.Error())
		C.lua_pushstring(L, errMsg)
		C.free(unsafe.Pointer(errMsg))
		return -1
	}
	return 0
}

//export LuaGetDB
func LuaGetDB(L *LState, stateKey *C.char, key *C.char) C.int {
	stateKeyString := C.GoString(stateKey)
	keyString := C.GoString(key)

	stateDb := contractMap.lookup(stateKeyString)
	if stateDb == nil {
		errMsg := C.CString("[System.LuaGetDB]not found contract state")
		C.lua_pushstring(L, errMsg)
		C.free(unsafe.Pointer(errMsg))
		return -1
	}

	data, err := stateDb.GetData([]byte(keyString))
	if err != nil {
		errMsg := C.CString(err.Error())
		C.lua_pushstring(L, errMsg)
		C.free(unsafe.Pointer(errMsg))
		return -1
	}

	if data == nil {
		return 0
	}
	dataString := C.CString(string(data))
	C.lua_pushstring(L, dataString)
	C.free(unsafe.Pointer(dataString))
	return 1
}
