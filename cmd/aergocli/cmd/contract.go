package cmd

/*
#cgo CFLAGS: -I${SRCDIR}/../../../libtool/include/luajit-2.0
#cgo LDFLAGS: ${SRCDIR}/../../../libtool/lib/libluajit-5.1.a -lm -ldl
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
	lua_State *L = luaL_newstate();
	luaL_openlibs(L);

	if (luaL_loadfile(L, code) != 0) {
		return lua_tostring(L, -1);
	}
	FILE *D = fopen(byte, "wb");
	if (lua_dump(L, kpt_lua_Writer, D) != 0) {
		return lua_tostring(L, -1);
	}
	fclose(D);
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

		code, err := ioutil.ReadFile(args[1])
		if err != nil {
			log.Fatal(err.Error())
		}
		param, err := base58.Decode(args[0])
		if err != nil {
			log.Fatal(err.Error())
		}
		state, err := client.GetState(context.Background(), &types.SingleBytes{Value: param})
		if err != nil {
			log.Fatal(err.Error())
		}
		tx := &types.Tx{Body: &types.TxBody{Nonce: state.GetNonce() + 1,
			Account: []byte(param),
			Payload: []byte(code)},
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
