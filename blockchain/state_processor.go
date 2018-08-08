/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

/*
#cgo CFLAGS: -I${SRCDIR}/../libtool/include/luajit-2.0
#cgo LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a -lm -ldl

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
	"unsafe"
	"github.com/aergoio/aergo/pkg/log"
)

var ctrLog *log.Logger

func init() {
	ctrLog = log.NewLogger(log.Contract)
}

func ApplyCode(code []byte, codeName []byte) error {
	if err := C.vm_run((*C.char)(unsafe.Pointer(&code[0])), C.size_t(len(code)),
		(*C.char)(unsafe.Pointer(&codeName))); err != nil {
		errMsg := C.GoString(err)
		C.free(unsafe.Pointer(err))
		ctrLog.Error(errMsg)
		return fmt.Errorf(errMsg)
	}
	return nil
}
