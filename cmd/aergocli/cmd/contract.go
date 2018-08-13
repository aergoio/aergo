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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"unsafe"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var client *util.ConnClient

func init() {
	contractCmd := &cobra.Command{
		Use:               "contract [flags] subcommand",
		Short:             "contract command",
		PersistentPreRun:  connectAergo,
		PersistentPostRun: disconnectAergo,
	}
	rootCmd.AddCommand(contractCmd)

	contractCmd.AddCommand(
		&cobra.Command{
			Use:               "compile [flags] srcfile bcfile",
			Short:             "compile a contract",
			Args:              cobra.MinimumNArgs(2),
			PersistentPreRun:  nil,
			PersistentPostRun: nil,
			Run: func(cmd *cobra.Command, args []string) {
				srcFileName := C.CString(args[0])
				outFileName := C.CString(args[1])
				defer C.free(unsafe.Pointer(srcFileName))
				defer C.free(unsafe.Pointer(outFileName))

				if err := C.compile(srcFileName, outFileName); err != nil {
					log.Fatal(C.GoString(err))
				}
			},
		},
	)
	contractCmd.AddCommand(
		&cobra.Command{
			Use:   "deploy [flags] address bcfile",
			Short: "deploy a contract",
			Args:  cobra.MinimumNArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				creator, err := base58.Decode(args[0])
				if err != nil {
					log.Fatal(err)
				}
				state, err := client.GetState(context.Background(), &types.SingleBytes{Value: creator})
				if err != nil {
					log.Fatal(err)
				}
				code, err := ioutil.ReadFile(args[1])
				if err != nil {
					log.Fatal(err)
				}
				tx := &types.Tx{
					Body: &types.TxBody{
						Nonce:   state.GetNonce() + 1,
						Account: []byte(creator),
						Payload: []byte(code),
					},
				}

				sign, err := client.SignTX(context.Background(), tx)
				if err != nil || sign == nil {
					log.Fatal(err)
				}
				txs := []*types.Tx{sign}
				commit, err := client.CommitTX(context.Background(), &types.TxList{Txs: txs})

				for i, r := range commit.Results {
					fmt.Println(i+1, ":", util.EncodeB64(r.Hash), r.Error)
				}
			},
		},
	)
	contractCmd.AddCommand(
		&cobra.Command{
			Use:   "call [flags] sender contract name args",
			Short: "deploy contract",
			Args:  cobra.MinimumNArgs(3),
			Run: func(cmd *cobra.Command, args []string) {
				caller, err := base58.Decode(args[0])
				if err != nil {
					log.Fatal(err)
				}
				state, err := client.GetState(context.Background(), &types.SingleBytes{Value: caller})
				if err != nil {
					log.Fatal(err)
				}
				contract, err := base58.Decode(args[1])
				if err != nil {
					log.Fatal(err)
				}
				var abi types.ABI
				abi.Name = args[2]
				if len(args) > 3 {
					err = json.Unmarshal([]byte(args[3]), &abi.Args)
					if err != nil {
						log.Fatal(err)
					}
				}
				payload, err := json.Marshal(abi)
				if err != nil {
					log.Fatal(err)
				}
				tx := &types.Tx{
					Body: &types.TxBody{
						Nonce:     state.GetNonce() + 1,
						Account:   []byte(caller),
						Recipient: []byte(contract),
						Payload:   payload,
					},
				}

				sign, err := client.SignTX(context.Background(), tx)
				if err != nil || sign == nil {
					log.Fatal(err)
				}
				txs := []*types.Tx{sign}
				commit, err := client.CommitTX(context.Background(), &types.TxList{Txs: txs})

				for i, r := range commit.Results {
					fmt.Println(i+1, ":", util.EncodeB64(r.Hash), r.Error)
				}
			},
		},
	)
}

func connectAergo(cmd *cobra.Command, args []string) {
	serverAddr := GetServerAddress()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var ok bool
	client, ok = util.GetClient(serverAddr, opts).(*util.ConnClient)
	if !ok {
		log.Fatal("internal error. wrong RPC client type")
	}
}

func disconnectAergo(cmd *cobra.Command, args []string) {
	if client != nil {
		client.Close()
	}
}
