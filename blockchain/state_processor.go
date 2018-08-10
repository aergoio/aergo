/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

/*
#cgo CFLAGS: -I${SRCDIR}/../libtool/include/luajit-2.0
#cgo LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a

#include <string.h>
#include <stdlib.h>
#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>

static const char* vm_run(const char *code, size_t sz, const char *name)
{
	int err;
	lua_State *L = luaL_newstate();
	const char *errMsg = NULL;

	luaL_openlibs(L);
	err = luaL_loadbuffer(L, code, sz, name);
	if (err != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		lua_close(L);
		return errMsg;
	}

	err = lua_pcall(L, 0, 0, 0);
	if (err != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		lua_close(L);
		return errMsg;
	}
	lua_close(L);
	return NULL;
}
*/
import "C"
import (
	"fmt"
		"github.com/aergoio/aergo/pkg/db"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
		"github.com/mr-tron/base58/base58"
	"encoding/json"
)

const dbName = "contracts.db"

var (
	ctrLog     *log.Logger
	contractDB db.DB
)

type Contract struct {
	code []byte
}

func init() {
	ctrLog = log.NewLogger(log.Contract)
	contractDB = db.NewDB(db.BadgerImpl, dbName)
}

func ApplyCode(code, contractAddress, txHash []byte) error {
	var err error
	contract := getContract(contractAddress)
	if contract == nil {
		err = fmt.Errorf("cannot find contract %s", string(contractAddress))
		ctrLog.Warn(err.Error())
	}
	var abi types.ABI
	json.Unmarshal(code, &abi)
	ctrLog.Debugf("contract call: %#v", abi)
	/*
	vm := NewLuaVM(code, call)
	vm.Run()
	if cErrMsg := C.vm_run((*C.char)(unsafe.Pointer(&contract.code[0])),
		C.size_t(len(contract.code)),
		(*C.char)(unsafe.Pointer(&contractAddress)),
	); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		C.free(unsafe.Pointer(cErrMsg))
		ctrLog.Error(errMsg)
		err = errors.New(errMsg)
	}
	*/
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
			code: val,
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
