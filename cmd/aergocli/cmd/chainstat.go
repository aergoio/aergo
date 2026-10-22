/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package cmd

import (
	"context"

	aergorpc "github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chainstatCmd)
}

var chainstatCmd = &cobra.Command{
	Use:   "chainstat",
	Short: "Print chain statistics",
	Run: func(cmd *cobra.Command, args []string) {
		msg, err := client.ChainStat(context.Background(), &aergorpc.Empty{})
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		cmd.Println(msg.Report)
	},
}
