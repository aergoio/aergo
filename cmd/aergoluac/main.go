/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package main

/*
#cgo CFLAGS: -I${SRCDIR}/../../libtool/include/luajit-2.0
#cgo LDFLAGS: ${SRCDIR}/../../libtool/lib/libluajit-5.1.a -lm

#include <stdlib.h>
#include <string.h>
#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>

lua_State *vm_newstate()
{
	lua_State *L = luaL_newstate();
	luaL_openlibs(L);
	return L;
}

void vm_close(lua_State *L)
{
	if (L != NULL)
		lua_close(L);
}

static int kpt_lua_Writer(struct lua_State *L, const void *p, size_t sz, void *u)
{
	return (fwrite(p, sz, 1, (FILE *)u) != 1) && (sz != 0);
}

const char *vm_compile(lua_State *L, const char *code, const char *byte, const char *abi)
{
	const char *errMsg = NULL;
	FILE *f = NULL;

	if (luaL_loadfile(L, code) != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		lua_close(L);
		return errMsg;
	}
	f = fopen(byte, "wb");
	if (f == NULL) {
		return "cannot open a bytecode file";
	}
	if (lua_dump(L, kpt_lua_Writer, f) != 0) {
		errMsg = strdup(lua_tostring(L, -1));
	}
	fclose(f);

	if (abi != NULL && strlen(abi) > 0) {
		const char *r;
		if (lua_pcall(L, 0, 0, 0) != 0) {
		   errMsg = strdup(lua_tostring(L, -1));
		   return errMsg;
		}
		lua_getfield(L, LUA_GLOBALSINDEX, "abi");
		lua_getfield(L, -1, "generate");
		if (lua_pcall(L, 0, 1, 0) != 0) {
		    errMsg = strdup(lua_tostring(L, -1));
		    return errMsg;
		}
		if (!lua_isstring(L, -1)) {
		    return "cannot create a abi file";
		}
		r = lua_tostring(L, -1);
		f = fopen(abi, "wb");
		if (f == NULL) {
		    return "cannot open a abi file";
		}
		fwrite(r, 1, strlen(r), f);
		fclose(f);
	}

	return errMsg;
}
*/
import "C"
import (
	"log"
	"os"
	"unsafe"

	"github.com/spf13/cobra"
)

var (
	rootCmd *cobra.Command
	abiFile string
)

func init() {
	log.SetOutput(os.Stderr)
	rootCmd = &cobra.Command{
		Use:   "aergoluac [flags] srcfile bcfile",
		Short: "compile a contract",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			compile(args[0], args[1], abiFile)
		},
	}
	rootCmd.PersistentFlags().StringVar(&abiFile, "abi", "", "abi filename")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func compile(srcFileName, outFileName, abiFileName string) {
	cSrcFileName := C.CString(srcFileName)
	cOutFileName := C.CString(outFileName)
	cAbiFileName := C.CString(abiFileName)
	L := C.vm_newstate()
	defer C.free(unsafe.Pointer(cSrcFileName))
	defer C.free(unsafe.Pointer(cOutFileName))
	defer C.free(unsafe.Pointer(cAbiFileName))
	defer C.vm_close(L)

	if errMsg := C.vm_compile(L, cSrcFileName, cOutFileName, cAbiFileName); errMsg != nil {
		log.Fatal(C.GoString(errMsg))
	}
}
