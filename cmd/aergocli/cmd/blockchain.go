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

var printHex bool

func init() {
	rootCmd.AddCommand(blockchainCmd)
	blockchainCmd.Flags().BoolVar(&printHex, "hex", false, "Print bytes to hex format")
}

var blockchainCmd = &cobra.Command{
	Use:   "blockchain",
	Short: "Print current blockchain status",
	Run: func(cmd *cobra.Command, args []string) {
		serverAddr := GetServerAddress()
		opts := []grpc.DialOption{grpc.WithInsecure()}
		var client *util.ConnClient
		var ok bool
		if client, ok = util.GetClient(serverAddr, opts).(*util.ConnClient); !ok {
			panic("Internal error. wrong RPC client type")
		}
		defer client.Close()

		msg, err := client.Blockchain(context.Background(), &aergorpc.Empty{})
		if nil == err {
			if printHex {
				fmt.Println(util.ConvHexBlockchainStatus(msg))
			} else {
				fmt.Println(util.JSON(msg))
			}
		} else {
			fmt.Printf("Failed: %s\n", err.Error())
		}
	},
}
