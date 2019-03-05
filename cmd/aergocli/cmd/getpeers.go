/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var getpeersCmd = &cobra.Command{
	Use:   "getpeers",
	Short: "Get Peer list",
	Run:   execGetPeers,
}


var nohidden bool
var showself bool

func init() {
	rootCmd.AddCommand(getpeersCmd)
	getpeersCmd.Flags().BoolVar(&nohidden, "nohidden",false,"exclude hidden peers")
	getpeersCmd.Flags().BoolVar(&showself, "self",false,"show self peer info")
}

func execGetPeers(cmd *cobra.Command, args []string) {
	msg, err := client.GetPeers(context.Background(), &types.PeersParams{NoHidden:nohidden, ShowSelf:showself})
	if err != nil {
		cmd.Printf("Failed to get peer from server: %s\n", err.Error())
		return
	}
	// address and peerid should be encoded, respectively
	cmd.Println(util.PeerListToString(msg))
}

func Must(a0 string, a1 error) string {
	return a0
}
