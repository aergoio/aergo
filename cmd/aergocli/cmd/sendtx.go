/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"errors"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var sendtxCmd = &cobra.Command{
	Use:   "sendtx",
	Short: "Send transaction",
	Args:  cobra.MinimumNArgs(0),
	RunE:  execSendTX,
}

func init() {
	rootCmd.AddCommand(sendtxCmd)
	sendtxCmd.Flags().StringVar(&from, "from", "", "Sender account address")
	sendtxCmd.MarkFlagRequired("from")
	sendtxCmd.Flags().StringVar(&to, "to", "", "Recipient account address")
	sendtxCmd.MarkFlagRequired("to")
	sendtxCmd.Flags().StringVar(&amount, "amount", "0", "How much in AER")
	sendtxCmd.MarkFlagRequired("amount")
}

func execSendTX(cmd *cobra.Command, args []string) error {
	account, err := types.DecodeAddress(from)
	if err != nil {
		return errors.New("Wrong address in --from flag\n" + err.Error())
	}
	recipient, err := types.DecodeAddress(to)
	if err != nil {
		return errors.New("Wrong address in --to flag\n" + err.Error())
	}
	amountBigInt, err := util.ParseUnit(amount)
	if err != nil {
		return errors.New("Wrong value in --amount flag\n" + err.Error())
	}
	tx := &types.Tx{Body: &types.TxBody{Account: account, Recipient: recipient, Amount: amountBigInt.Bytes()}}
	msg, err := client.SendTX(context.Background(), tx)
	if err != nil {
		cmd.Println(err.Error())
		return nil
	}
	cmd.Println(util.JSON(msg))
	return nil
}
