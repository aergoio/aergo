/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	aergorpc "github.com/aergoio/aergo/v2/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var rawPayload bool

var gettxCmd = &cobra.Command{
	Use:   "gettx",
	Short: "Get transaction information",
	Long:  "Get transaction information from aergosvr instance. \nIf transaction is in block, return transaction with index that represent where it's included",
	Args:  cobra.MinimumNArgs(1),
	Run:   execGetTX,
}

func init() {
	rootCmd.AddCommand(gettxCmd)
	gettxCmd.Flags().BoolVar(&rawPayload, "rawpayload", false, "show payload without encoding")
}

func execGetTX(cmd *cobra.Command, args []string) {
	txHash, err := base58.Decode(args[0])
	if err != nil {
		cmd.Printf("Failed decode: %s", err.Error())
		return
	}
	payloadEncodingType := util.Base58
	if rawPayload {
		payloadEncodingType = util.Raw
	}
	msg, err := client.GetTX(context.Background(), &aergorpc.SingleBytes{Value: txHash})
	if err == nil {
		cmd.Println(util.ConvTxEx(msg, payloadEncodingType))
	} else {
		msgblock, err := client.GetBlockTX(context.Background(), &aergorpc.SingleBytes{Value: txHash})
		if err != nil {
			cmd.Printf("Failed: %s", err.Error())
			return
		}
		cmd.Println(util.ConvTxInBlockEx(msgblock, payloadEncodingType))
	}

}
