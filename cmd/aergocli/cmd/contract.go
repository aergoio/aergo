package cmd

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

const char *compile(char *code, char *byte)
{
	int err;
	lua_State *L = luaL_newstate();

	luaL_openlibs(L);
	lua_pushliteral(L, "bcsave");

	if (loadjitmodule(L) != 0) {
		return "can't load jit module";
	}

	lua_pushstring(L, code);
	lua_pushstring(L, byte);

	err = lua_pcall(L, 2, 0, 0);
	if (err != 0) {
		return lua_tostring(L, -1);
	}
	return NULL;
}
*/
import "C"
import (
		"fmt"
	"io/ioutil"
	"log"
	"unsafe"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)
import (
	"context"

	"github.com/mr-tron/base58/base58"
)

func init() {
	rootCmd.AddCommand(compileCmd)
	rootCmd.AddCommand(deployCmd)
}

var compileCmd = &cobra.Command{
	Use:   "compile [flags] srcfile bcfile",
	Short: "compile contract",
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

var deployCmd = &cobra.Command{
	Use:   "deploy [flags] address bcfile",
	Short: "deploy contract",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serverAddr := GetServerAddress()
		opts := []grpc.DialOption{grpc.WithInsecure()}
		var client *util.ConnClient
		var ok bool
		if client, ok = util.GetClient(serverAddr, opts).(*util.ConnClient); !ok {
			log.Fatal("internal error. wrong RPC client type")
		}
		defer client.Close()
		var err error

		dat, err := ioutil.ReadFile(args[1])
		if err != nil {
			log.Fatalf(err.Error())
		}
		param, err := base58.Decode(args[0])
		if err != nil {
			log.Fatal(err.Error())
		}
		msg, err := client.GetState(context.Background(), &types.SingleBytes{Value: param})
		if err != nil {
			log.Fatal(err.Error())
		}
		tx := &types.Tx{Body: &types.TxBody{Nonce: msg.GetNonce() + 1,
			Account: []byte(param),
			Payload: []byte(dat)},
		}

		sign, err := client.SignTX(context.Background(), tx)
		if nil != err || sign == nil {
			log.Fatal(err.Error())
		}
		txs := []*types.Tx{sign}
		commit, err := client.CommitTX(context.Background(), &types.TxList{Txs: txs})

		for i, r := range commit.Results {
			fmt.Println(i+1, ":", util.EncodeB64(r.Hash), r.Error)
		}
	},
}
