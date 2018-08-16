/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package main

/*
#cgo CFLAGS: -I${SRCDIR}/../../libtool/include/luajit-2.0
#cgo LDFLAGS: ${SRCDIR}/../../libtool/lib/libluajit-5.1.a -lm -ldl
#include <string.h>
#include <stdlib.h>
#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>

static int kpt_lua_Writer(struct lua_State *L, const void *p, size_t sz, void *u)
{
	return (fwrite(p, sz, 1, (FILE *)u) != 1) && (sz != 0);
}

const char *compile(char *code, char *byte)
{
	const char *errMsg = NULL;
	lua_State *L = luaL_newstate();
	luaL_openlibs(L);

	if (luaL_loadfile(L, code) != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		lua_close(L);
		return errMsg;
	}
	FILE *D = fopen(byte, "wb");
	if (lua_dump(L, kpt_lua_Writer, D) != 0) {
		errMsg = strdup(lua_tostring(L, -1));
	}

	fclose(D);
	lua_close(L);

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

func init() {
	log.SetOutput(os.Stderr)
}

func main() {
	cmd := cobra.Command{
		Use:   "aergoluac [flags] srcfile bcfile",
		Short: "compile a contract",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			srcFileName := C.CString(args[0])
			outFileName := C.CString(args[1])
			defer C.free(unsafe.Pointer(srcFileName))
			defer C.free(unsafe.Pointer(outFileName))

			if err := C.compile(srcFileName, outFileName); err != nil {
				log.Fatal(C.GoString(err))
			}
		},
	}
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
