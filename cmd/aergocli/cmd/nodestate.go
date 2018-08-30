/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/binary"
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
	nodeCmd.Flags().Uint64VarP(&number, "timeout", "t", 3, "Per module time out")
}

func execNodeState(cmd *cobra.Command, args []string) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var client *util.ConnClient
	var ok bool
	if client, ok = util.GetClient(GetServerAddress(), opts).(*util.ConnClient); !ok {
		panic("Internal error. wrong RPC client type")
	}
	defer client.Close()

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(number))
	msg, err := client.NodeState(context.Background(), &types.SingleBytes{Value: b})
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return
	}
	fmt.Printf("%s\n", string(msg.Value))
}
