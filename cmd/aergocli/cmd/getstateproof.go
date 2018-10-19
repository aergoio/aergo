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

var getstateproofCmd = &cobra.Command{
	Use:   "getstateproof",
	Short: "Get account state with merkle proof",
	Run:   execGetStateProof,
}

var stateroot string
var proof bool

func init() {
	getstateproofCmd.Flags().StringVar(&address, "address", "", "Get state from the address")
	getstateproofCmd.MarkFlagRequired("address")
	getstateproofCmd.Flags().StringVar(&stateroot, "root", "", "Get state from the address")
	getstateproofCmd.Flags().BoolVar(&proof, "proof", true, "Get the proof for the state")
	rootCmd.AddCommand(getstateproofCmd)
}

func execGetStateProof(cmd *cobra.Command, args []string) {
	//TODO TEST after modifying state with TX, check test by setting state root
	var root []byte
	var err error
	if len(stateroot) != 0 {
		//if cmd.Flags().Changed("root") == false {
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
	//TODO make GetStateAndProofCompressed
	msg, err := client.GetStateAndProof(context.Background(),
		&types.AccountAndRoot{Account: addr, Root: root})
	if err != nil {
		cmd.Printf("Failed: %s", err.Error())
		return
	}
	cmd.Printf("{account:%s, nonce:%d, balance:%d, merkle proof length:%d}\n",
		address, msg.GetState().GetNonce(), msg.GetState().GetBalance(), len(msg.GetAuditPath()))
}
