package cmd

import (
	"context"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	"github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chaininfoCmd)
}

var chaininfoCmd = &cobra.Command{
	Use:   "chaininfo",
	Short: "Print current blockchain information",
	Run: func(cmd *cobra.Command, args []string) {
		msg, err := client.GetChainInfo(context.Background(), &types.Empty{})
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		cmd.Println(util.ConvChainInfoMsg(msg))
	},
}
