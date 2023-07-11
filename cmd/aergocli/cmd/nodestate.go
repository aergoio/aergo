/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/binary"

	"github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
)

var (
	nodeCmd = &cobra.Command{
		Use:   "node",
		Short: "Show internal metric",
		Args:  cobra.MinimumNArgs(0),
		Run:   execNodeState,
	}

	component string
)

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.Flags().Uint64VarP(&timeout, "timeout", "t", 3, "Per module time out")
	nodeCmd.Flags().StringVarP(&component, "component", "c", "", "component name")
}

func execNodeState(cmd *cobra.Command, args []string) {
	var b []byte
	var nodeReq types.NodeReq

	if len(component) > 0 {
		b = []byte(component)
		nodeReq.Component = b
	}

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, timeout)
	nodeReq.Timeout = b

	msg, err := client.NodeState(context.Background(), &nodeReq)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	cmd.Printf("%s\n", string(msg.Value))
}
