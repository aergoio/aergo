/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"fmt"
	"unsafe"
	
	/*
		#cgo CFLAGS: -I${SRCDIR}/../libtool/include/luajit-2.0
		#cgo LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a -lm -ldl

		#include <string.h>
		#include <stdlib.h>
		#include <lualib.h>
		#include <lauxlib.h>
		#include <luajit.h>

		static int vm_run (const char *code, size_t sz, const char *name)
		{
			int err;
			lua_State *L = luaL_newstate();

			luaL_openlibs(L);
			err = luaL_loadbuffer(L, code, sz, name);
			if (err != 0) {
				printf("Failed lua load:%d\n", err);
				return -1;
			}

			err = lua_pcall(L, 0, 0, 0);
			if (err != 0) {
				printf("Failed lua execute:%d\n", err);
				return -1;
			}
			return 0;
		}
	*/
	"C"
)

func ApplyCode(code []byte, codeName []byte) {
	/*codeCstr := C.CString(code.String())
	codeNameCstr := C.CString(codeName.String())
	defer C.free(unsafe.Pointer(codeCstr))
	defer C.free(unsafe.Pointer(codeNameCstr))
    */
	if C.vm_run((*C.char)(unsafe.Pointer(&code[0])), C.size_t(len(code)), 
		(*C.char)(unsafe.Pointer(&codeName))) < 0 {
		fmt.Printf("Fail: execute\n")
		return
	}
}
