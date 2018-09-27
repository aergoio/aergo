/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"fmt"

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

var amount uint64

func init() {
	rootCmd.AddCommand(sendtxCmd)
	sendtxCmd.Flags().StringVar(&from, "from", "", "")
	sendtxCmd.Flags().StringVar(&to, "to", "", "")
	sendtxCmd.Flags().Uint64Var(&amount, "amount", 0, "")
}

func execSendTX(cmd *cobra.Command, args []string) {
	account, err := types.DecodeAddress(from)
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
	}
	recipient, err := types.DecodeAddress(to)
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
	}
	tx := &types.Tx{Body: &types.TxBody{Account: account, Recipient: recipient, Amount: amount}}
	msg, err := client.SendTX(context.Background(), tx)
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return
	}
	fmt.Println(base58.Encode(msg.Hash), msg.Error)
}
