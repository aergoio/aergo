/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"fmt"

	"github.com/mr-tron/base58/base58"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var listblockheadersCmd = &cobra.Command{
	Use:   "listblocks",
	Short: "Get block headers list",
	Run:   execListBlockHeaders,
}
var gbhHash string
var gbhHeight int32
var gbhSize int
var gbhOffset int
var gbhAsc bool

func init() {
	rootCmd.AddCommand(listblockheadersCmd)

	listblockheadersCmd.Flags().StringVar(&gbhHash, "hash", "", "Block hash")
	listblockheadersCmd.Flags().Int32Var(&gbhHeight, "height", int32(-1), "Block height")
	listblockheadersCmd.Flags().IntVar(&gbhSize, "size", 20, "Max list size")
	listblockheadersCmd.Flags().IntVar(&gbhOffset, "offset", 0, "Offset")
	listblockheadersCmd.Flags().BoolVar(&gbhAsc, "asc", true, "Order by")

}

func execListBlockHeaders(cmd *cobra.Command, args []string) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var client *util.ConnClient
	var ok bool
	if client, ok = util.GetClient(GetServerAddress(), opts).(*util.ConnClient); !ok {
		panic("Internal error. wrong RPC client type")
	}
	defer client.Close()
	blockHash, err := base58.Decode(gbhHash)
	if err != nil {
		fmt.Printf("decode error: %s", err.Error())
		return
	}
	uparams := &types.ListParams{Hash: blockHash, Height: uint64(gbhHeight), Size: uint32(gbhSize)}

	msg2, err := client.ListBlockHeaders(context.Background(), uparams)
	if nil == err {
		fmt.Println(util.JSON(msg2))
	} else {
		fmt.Printf("Failed: %s\n", err.Error())
	}
}
