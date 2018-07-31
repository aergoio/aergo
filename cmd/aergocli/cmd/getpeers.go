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

	"github.com/aergoio/aergo/cmd/aergocli/util"
	aergorpc "github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
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
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var client *util.ConnClient
	var ok bool
	if client, ok = util.GetClient(GetServerAddress(), opts).(*util.ConnClient); !ok {
		panic("Internal error. wrong RPC client type")
	}
	defer client.Close()

	msg2, err := client.GetPeers(context.Background(), &aergorpc.Empty{})
	if err != nil {
		fmt.Printf("Failed to get peer from server: %s\n", err.Error())
		return
	}
	// address and peerid should be encoded, respectively
	resultView := make([]map[string]string, 0, len(msg2.Peers))
	for _, peer := range msg2.Peers {
		peerData := make(map[string]string)
		peerData["Address"] = net.IP(peer.Address).String()
		peerData["Port"] = strconv.Itoa(int(peer.Port))
		peerData["PeerID"] = base58.Encode(peer.PeerID)
		resultView = append(resultView, peerData)
	}
	encoder := json.NewEncoder(os.Stdout)
	if nil == err {
		encoder.Encode(resultView)
	} else {
		fmt.Printf("Failed: %s\n", err.Error())
	}
}

func Must(a0 string, a1 error) string {
	return a0
}
