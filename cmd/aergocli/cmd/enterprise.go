/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	aergorpc "github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var enterpriseKey string

func init() {
	rootCmd.AddCommand(enterpriseCmd)
	enterpriseCmd.AddCommand(enterpriseKeyCmd)
}

var enterpriseCmd = &cobra.Command{
	Use:   "enterprise subcommand",
	Short: "Enterprise command",
}

var enterpriseKeyCmd = &cobra.Command{
	Use:   "key <config key>",
	Short: "Print config values of enterprise",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		msg, err := client.GetEnterpriseConfig(context.Background(), &aergorpc.EnterpriseConfigKey{Key: args[0]})
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		cmd.Println(util.JSON(msg))
	},
}
