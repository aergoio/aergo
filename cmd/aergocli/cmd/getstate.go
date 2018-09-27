/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"fmt"

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
		fmt.Printf("Failed: %s\n", err.Error())
	}
	msg, err := client.GetState(context.Background(),
		&types.SingleBytes{Value: param})
	if nil == err {
		fmt.Printf("{account:%s, nonce:%d, balance:%d}\n",
			address, msg.GetNonce(), msg.GetBalance())
	} else {
		fmt.Printf("Failed: %s\n", err.Error())
	}
}
