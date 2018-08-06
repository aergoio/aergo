package cmd

import (
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"unsafe"

	/*
		   #cgo CFLAGS: -I${SRCDIR}/../../../libtool/include/luajit-2.0
		   #cgo LDFLAGS: ${SRCDIR}/../../../libtool/lib/libluajit-5.1.a -lm -ldl
		   #include <string.h>
		   #include <stdlib.h>
		   #include <lualib.h>
		   #include <lauxlib.h>
		   #include <luajit.h>

		   static int loadjitmodule(lua_State *L)
		   {
		        lua_getglobal(L, "require");
		        lua_pushliteral(L, "jit.");
		        lua_pushvalue(L, -3);
		        lua_concat(L, 2);

		        if (lua_pcall(L, 1, 1, 0)) {
		            const char *msg = lua_tostring(L, -1);
		            if (msg && strncmp(msg, "module ", 7) != 0) {
		                return -1;
		            }
		        }

		        lua_getfield(L, -1, "start");

		        if (lua_isnil(L, -1))
		            return -1;
		        lua_remove(L, -2);

		        return 0;
		   }

		   int compile(char *code, char *byte)
		   {
			    int err;
		        lua_State *L = luaL_newstate();

		        luaL_openlibs(L);
		        lua_pushliteral(L, "bcsave");

		        if(loadjitmodule(L) != 0) {
		            printf("exit");
		            return -1;
		        }

		        lua_pushstring(L, code);
		        lua_pushstring(L, byte);

				err = lua_pcall(L, 2, 0, 0);
				if (err != 0) {
					printf ("Failed compile : %d\n", err);
					return -1;
				}

		        return 0;
		   }
	*/
	"C"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(compileCmd)
}

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "compile contract",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Printf("Fail:Not enough arguments\n")
			return
		}
		codeCstr := C.CString(args[0])
		byteCstr := C.CString(args[1])
		defer C.free(unsafe.Pointer(codeCstr))
		defer C.free(unsafe.Pointer(byteCstr))

		if C.compile(codeCstr, byteCstr) < 0 {
			fmt.Printf("Fail: compile\n")
			return
		}
		dat, err := ioutil.ReadFile(args[1])
		if err != nil {
			panic(err)
		}
		b := b64.StdEncoding.EncodeToString([]byte(dat))
		fmt.Println(b)
	},
}
