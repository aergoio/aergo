/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"fmt"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Show internal metric",
	Args:  cobra.MinimumNArgs(0),
	Run:   execNodeState,
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}

func execNodeState(cmd *cobra.Command, args []string) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var client *util.ConnClient
	var ok bool
	if client, ok = util.GetClient(GetServerAddress(), opts).(*util.ConnClient); !ok {
		panic("Internal error. wrong RPC client type")
	}
	defer client.Close()

	msg, err := client.NodeState(context.Background(), &types.Empty{})

	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return
	}

	for _, s := range msg.Status {
		fmt.Printf("module: %s\n", s.Name)
		for _, is := range s.Stat {
			fmt.Printf("\t%s : %f\n", is.Name, is.Stat)
		}
	}
}
