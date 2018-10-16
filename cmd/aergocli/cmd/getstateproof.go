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

var getstateproofCmd = &cobra.Command{
	Use:   "getstateproof",
	Short: "Get account state with merkle proof",
	Run:   execGetStateProof,
}

func init() {
	getstateproofCmd.Flags().StringVar(&address, "address", "", "Get state and proof from the address")
	getstateproofCmd.MarkFlagRequired("address")
	rootCmd.AddCommand(getstateproofCmd)
}

func execGetStateProof(cmd *cobra.Command, args []string) {
	param, err := types.DecodeAddress(address)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	msg, err := client.GetStateAndProof(context.Background(),
		&types.SingleBytes{Value: param})
	if err != nil {
		cmd.Printf("Failed: %s", err.Error())
		return
	}
	cmd.Printf("{account:%s, nonce:%d, balance:%d, merkle proof length:%d}\n",
		address, msg.GetState().GetNonce(), msg.GetState().GetBalance(), len(msg.GetAuditPath()))
}
