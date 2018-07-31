/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"fmt"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	aergorpc "github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var gettxCmd = &cobra.Command{
	Use:   "gettx",
	Short: "Get transaction information",
	Args:  cobra.MinimumNArgs(1),
	Run:   execGetTX,
}

func init() {
	rootCmd.AddCommand(gettxCmd)
	// args := make([]string, 0, 10)
	// args = append(args, "subCommand")
	// blockCmd.SetArgs(args)
}

func execGetTX(cmd *cobra.Command, args []string) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var client *util.ConnClient
	var ok bool
	if client, ok = util.GetClient(GetServerAddress(), opts).(*util.ConnClient); !ok {
		panic("Internal error. wrong RPC client type")
	}
	defer client.Close()

	txHash, err := util.DecodeB64(args[0])
	if err != nil {
		fmt.Printf("decode error: %s", err.Error())
		return
	}
	msg, err := client.GetTX(context.Background(), &aergorpc.SingleBytes{Value: txHash})
	if nil == err {
		fmt.Println("Pending: ", util.JSON(msg))
	} else {
		msgblock, err := client.GetBlockTX(context.Background(), &aergorpc.SingleBytes{Value: txHash})
		if nil == err {
			fmt.Println("Confirm: ", util.JSON(msgblock))
		} else {
			fmt.Printf("Failed: %s\n", err.Error())
		}
	}

}
