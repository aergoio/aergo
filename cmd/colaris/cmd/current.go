/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"

	"github.com/mr-tron/base58/base58"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var CurrentPeersCmd = &cobra.Command{
	Use:   "current",
	Short: "Get current peers list",
	Run:   execCurrentPeers,
}
var cpKey string
var cpSize int

func init() {
	rootCmd.AddCommand(CurrentPeersCmd)

	CurrentPeersCmd.Flags().StringVar(&cpKey, "ref", "", "Reference Key")
	CurrentPeersCmd.Flags().IntVar(&cpSize, "size", 20, "Max list size")

}

func execCurrentPeers(cmd *cobra.Command, args []string) {
	var blockHash []byte
	var err error

	if cmd.Flags().Changed("ref") == true {
		blockHash, err = base58.Decode(cpKey)
		if err != nil {
			cmd.Printf("Failed: %s", err.Error())
			return
		}
	}
	if cpSize <= 0 {
		cpSize = 20
	}

	uparams := &types.Paginations{
		Ref:   blockHash,
		Size:   uint32(cpSize),
	}

	msg, err := client.CurrentList(context.Background(), uparams)
	if err != nil {
		cmd.Printf("Failed: %s", err.Error())
		return
	}
	// TODO decorate other props also;e.g. uint64 timestamp to human readable time format!
	for _, p := range msg.Peers {
		if p.Verion == "" {
			p.Verion = "(old)"
		}
	}
	cmd.Println(util.JSON(msg))
}
