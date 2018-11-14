/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	aergorpc "github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var gettxCmd = &cobra.Command{
	Use:   "gettx",
	Short: "Get transaction information",
	Long:  "Get transaction information from aergosvr instance. \nIf transaction is in block, return transaction with index that represent where it's included",
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
	txHash, err := base58.Decode(args[0])
	if err != nil {
		cmd.Printf("Failed decode: %s", err.Error())
		return
	}
	msg, err := client.GetTX(context.Background(), &aergorpc.SingleBytes{Value: txHash})
	if err == nil {
		cmd.Println(util.TxConvBase58Addr(msg))
	} else {
		msgblock, err := client.GetBlockTX(context.Background(), &aergorpc.SingleBytes{Value: txHash})
		if err != nil {
			cmd.Printf("Failed: %s", err.Error())
			return
		}
		cmd.Println(util.TxInBlockConvBase58Addr(msgblock))
	}

}
