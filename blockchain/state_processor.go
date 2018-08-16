/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

/*
#cgo CFLAGS: -I${SRCDIR}/../libtool/include/luajit-2.0
#cgo LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a -lm

#include <stdlib.h>
#include <string.h>
#include "lua_exec.h"
*/
import "C"
import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"unsafe"

	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
)

const contractDbName = "contracts.db"

var (
	ctrLog     *log.Logger
	contractDB db.DB
)

type Contract struct {
	code    []byte
	address []byte
}

type ContractExec struct {
	L        *C.lua_State
	contract *Contract
	err      error
}

func init() {
	ctrLog = log.NewLogger(log.Contract)
}

type LuaExecutor struct {
	instanceID []byte
	luaState   *LState
	definition []byte
	exContext  *LExContext
}

func NewBlockchainCtx(Sender []byte, blockHash []byte, txHash []byte, blockHeight uint64,
	timestamp int64, node string, confirmed bool, contractID []byte) *LBlockchainCtx {
	context := LBlockchainCtx{
		sender:      C.CString(base58.Encode(Sender)),
		blockHash:   C.CString(hex.EncodeToString(blockHash)),
		txHash:      C.CString(hex.EncodeToString(txHash)),
		blockHeight: C.ulonglong(blockHeight),
		timestamp:   C.longlong(timestamp),
		node:        C.CString(node),
		confirmed:   C.int(iconfirmed),
		contractId:  C.CString(base58.Encode(contractID)),
	}
}

func init() {
	ctrLog = log.NewLogger(log.Contract)
}

func newContractExec(contract *Contract, bcCtx *LBlockchainCtx) *ContractExec {
	ce := &ContractExec{
		contract: contract,
	}
	if cErrMsg := C.vm_loadbuff((*C.char)(unsafe.Pointer(&contract.code[0])),
		C.size_t(len(contract.code)),
		C.CString(base58.Encode(contract.address)), bcCtx,
		&ce.L,
	); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		C.free(unsafe.Pointer(cErrMsg))
		ctrLog.Error(errMsg)
		ce.err = errors.New(errMsg)
	}
	return ce
}

func (ce *ContractExec) call(abi *types.ABI) {
	if ce.err != nil {
		return
	}
	C.vm_getfield(ce.L, C.CString(abi.Name))
	for _, v := range abi.Args {
		switch arg := v.(type) {
		case string:
			C.lua_pushstring(ce.L, C.CString(arg))
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
	if cErrMsg := C.vm_pcall(ce.L, C.int(len(abi.Args))); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		C.free(unsafe.Pointer(cErrMsg))
		ctrLog.WithCtx("error", errMsg).Warnf("contract %s", base58.Encode(ce.contract.address))
		ce.err = errors.New(errMsg)
	}
}

func (ce *ContractExec) close() {
	if ce != nil && ce.L != nil {
		C.lua_close(ce.L)
	}
}

func ApplyCode(code, contractAddress, txHash []byte) error {
	var err error
	contract := getContract(contractAddress)
	if contract == nil {
		err = fmt.Errorf("cannot find contract %s", string(contractAddress))
		ctrLog.Warn(err.Error())
	}
	var abi types.ABI
	err = json.Unmarshal(code, &abi)
	if err != nil {
		ctrLog.WithCtx("error", err).Warn("contract %s", base58.Encode(contractAddress))
	}
	var ce *ContractExec
	defer ce.close()
	if err == nil {
		ctrLog.WithCtx("abi", abi).Debugf("contract %s", base58.Encode(contractAddress))
		ce = newContractExec(contract, bcCtx)
		ce.call(&abi)
		err = ce.err
	}
	receipt := types.NewReceipt(contractAddress, "SUCCESS")
	if err != nil {
		receipt.Status = err.Error()
	}
	contractDB.Set(txHash, receipt.Bytes())
	return err
}

func CreateContract(code, contractAddress, txHash []byte) error {
	ctrLog.WithCtx("contractAddress", base58.Encode(contractAddress)).Debug("new contract is deployed")
	contractDB.Set(contractAddress, code)
	receipt := types.NewReceipt(contractAddress, "CREATED")
	contractDB.Set(txHash, receipt.Bytes())
	return nil
}

func getContract(contractAddress []byte) *Contract {
	val := contractDB.Get(contractAddress)
	if len(val) > 0 {
		return &Contract{
			code:    val,
			address: contractAddress[:],
		}
	}
	return nil
}

func GetReceipt(txHash []byte) *types.Receipt {
	val := contractDB.Get(txHash)
	if len(val) == 0 {
		return nil
	}
	return types.NewReceiptFromBytes(val)
}
