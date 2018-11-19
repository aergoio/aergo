/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"

	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var getstateCmd = &cobra.Command{
	Use:   "getstate",
	Short: "Get account state",
	Run:   execGetState,
}

func init() {
	getstateCmd.Flags().StringVar(&address, "address", "", "Get state from the address")
	getstateCmd.MarkFlagRequired("address")
	getstateCmd.Flags().StringVar(&stateroot, "root", "", "Get the state at a specified state root")
	getstateCmd.Flags().BoolVar(&proof, "proof", false, "Get the proof for the state")
	getstateCmd.Flags().BoolVar(&compressed, "compressed", false, "Get a compressed proof for the state")
	getstateCmd.Flags().BoolVar(&staking, "staking", false, "Get the staking info from the address")
	rootCmd.AddCommand(getstateCmd)
}

func execGetState(cmd *cobra.Command, args []string) {
	var root []byte
	var err error
	if len(stateroot) != 0 {
		root, err = base58.Decode(stateroot)
		if err != nil {
			cmd.Printf("decode error: %s", err.Error())
			return
		}
	}
	addr, err := types.DecodeAddress(address)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	if staking {
		msg, err := client.GetStaking(context.Background(),
			&types.SingleBytes{Value: addr})
		if err != nil {
			cmd.Printf("Failed: %s", err.Error())
			return
		}
		cmd.Printf("{account:%s, staked:%d, when:%d}\n",
			address, msg.GetAmount(), msg.GetWhen())

		return
	}

	if !proof {
		// NOTE GetState first queries the statedb buffer.
		// So the prefered way to get the state is with a proof
		msg, err := client.GetState(context.Background(),
			&types.SingleBytes{Value: addr})
		if err != nil {
			cmd.Printf("Failed: %s", err.Error())
			return
		}
		cmd.Printf("{account:%s, nonce:%d, balance:%d}\n",
			address, msg.GetNonce(), msg.GetBalance())
	} else {
		// Get the state and proof at a specific root.
		// If root is nil, the latest block is queried.
		msg, err := client.GetStateAndProof(context.Background(),
			&types.AccountAndRoot{Account: addr, Root: root, Compressed: compressed})
		if err != nil {
			cmd.Printf("Failed: %s", err.Error())
			return
		}
		cmd.Printf("{account:%s, nonce:%d, balance:%d, included:%t, merkle proof length:%d, height:%d}\n",
			address, msg.GetState().GetNonce(), msg.GetState().GetBalance(), msg.GetInclusion(), len(msg.GetAuditPath()), msg.GetHeight())
	}

}
