/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"time"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"

	"github.com/mr-tron/base58/base58"

	"github.com/aergoio/aergo/v2/types"
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
		Ref:  blockHash,
		Size: uint32(cpSize),
	}

	msg, err := client.CurrentList(context.Background(), uparams)
	if err != nil {
		cmd.Printf("Failed: %s", err.Error())
		return
	}
	// TODO decorate other props also;e.g. uint64 timestamp to human readable time format!
	ppList := make([]JSONPolarisPeer, len(msg.Peers))
	for i, p := range msg.Peers {
		ppList[i] = NewJSONPolarisPeer(p)
	}
	cmd.Println(util.B58JSON(ppList))
}

type PolarisPeerAlias types.PolarisPeer

// JSONPolarisPeer is simple wrapper to print human readable json output.
type JSONPolarisPeer struct {
	*PolarisPeerAlias
	Connected time.Time `json:"connected"`
	LastCheck time.Time `json:"lastCheck"`
}

func NewJSONPolarisPeer(pp *types.PolarisPeer) JSONPolarisPeer {
	return JSONPolarisPeer{
		(*PolarisPeerAlias)(pp),
		time.Unix(0, pp.Connected),
		time.Unix(0, pp.LastCheck),
	}
}
