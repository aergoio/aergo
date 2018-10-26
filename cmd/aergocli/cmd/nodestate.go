/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/binary"

	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
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
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(number))
	msg, err := client.NodeState(context.Background(), &types.SingleBytes{Value: b})
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	cmd.Printf("%s\n", string(msg.Value))
}
