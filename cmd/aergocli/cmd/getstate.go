/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"

	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var getstateCmd = &cobra.Command{
	Use:   "getstate",
	Short: "Get account state",
	Run:   execGetState,
}

var address string

func init() {
	getstateCmd.Flags().StringVar(&address, "address", "", "Get state from the address")
	getstateCmd.MarkFlagRequired("address")
	rootCmd.AddCommand(getstateCmd)
}

func execGetState(cmd *cobra.Command, args []string) {
	param, err := types.DecodeAddress(address)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	msg, err := client.GetState(context.Background(),
		&types.SingleBytes{Value: param})
	if err != nil {
		cmd.Printf("Failed: %s", err.Error())
		return
	}
	cmd.Printf("{account:%s, nonce:%d, balance:%d}\n",
		address, msg.GetNonce(), msg.GetBalance())
}
