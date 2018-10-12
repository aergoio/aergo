/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/mr-tron/base58/base58"

	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var getpeersCmd = &cobra.Command{
	Use:   "getpeers",
	Short: "Get Peer list",
	Run:   execGetPeers,
}

func init() {
	rootCmd.AddCommand(getpeersCmd)
}

func execGetPeers(cmd *cobra.Command, args []string) {
	msg, err := client.GetPeers(context.Background(), &types.Empty{})
	if err != nil {
		fmt.Printf("Failed to get peer from server: %s\n", err.Error())
		return
	}
	// address and peerid should be encoded, respectively
	resultView := make([]map[string]string, 0, len(msg.Peers))
	for i, peer := range msg.Peers {
		peerData := make(map[string]string)
		peerState := types.PeerState(msg.States[i]).String()
		peerData["Address"] = net.IP(peer.Address).String()
		peerData["Port"] = strconv.Itoa(int(peer.Port))
		peerData["PeerID"] = base58.Encode(peer.PeerID)
		peerData["State"] = peerState
		resultView = append(resultView, peerData)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "\t")
	encoder.Encode(resultView)
}

func Must(a0 string, a1 error) string {
	return a0
}
