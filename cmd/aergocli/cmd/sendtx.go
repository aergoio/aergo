/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"

	"github.com/mr-tron/base58/base58"

	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var sendtxCmd = &cobra.Command{
	Use:   "sendtx",
	Short: "Send transaction",
	Args:  cobra.MinimumNArgs(0),
	Run:   execSendTX,
}

func init() {
	rootCmd.AddCommand(sendtxCmd)
	sendtxCmd.Flags().StringVar(&from, "from", "", "")
	sendtxCmd.MarkFlagRequired("from")
	sendtxCmd.Flags().StringVar(&to, "to", "", "")
	sendtxCmd.MarkFlagRequired("to")
	sendtxCmd.Flags().Uint64Var(&amount, "amount", 0, "")
	sendtxCmd.MarkFlagRequired("amount")
}

func execSendTX(cmd *cobra.Command, args []string) {
	account, err := types.DecodeAddress(from)
	if err != nil {
		cmd.Printf("Failed decode: %s\n", err.Error())
		return
	}
	recipient, err := types.DecodeAddress(to)
	if err != nil {
		cmd.Printf("Failed decode: %s\n", err.Error())
		return
	}
	tx := &types.Tx{Body: &types.TxBody{Account: account, Recipient: recipient, Amount: amount}}
	msg, err := client.SendTX(context.Background(), tx)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	cmd.Println(base58.Encode(msg.Hash), msg.Error)
}
