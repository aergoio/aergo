/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	aergorpc "github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func init() {
	receiptCmd := &cobra.Command{
		Use:   "receipt [flags] subcommand",
		Short: "receipt command",
	}
	rootCmd.AddCommand(receiptCmd)

	receiptCmd.AddCommand(
		&cobra.Command{
			Use:   "get [flags] tx_hash",
			Short: "get a receipt",
			Args:  cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				opts := []grpc.DialOption{grpc.WithInsecure()}
				client, ok := util.GetClient(GetServerAddress(), opts).(*util.ConnClient)
				if !ok {
					log.Fatal("Internal error. wrong RPC client type")
				}
				defer client.Close()

				txHash, err := util.DecodeB64(args[0])
				if err != nil {
					log.Fatal(err)
				}
				msg, err := client.GetReceipt(context.Background(), &aergorpc.SingleBytes{Value: txHash})
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(util.JSON(msg))
			},
		},
	)
}
