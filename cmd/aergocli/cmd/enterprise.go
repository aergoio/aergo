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
	enterpriseCmd.Flags().StringVar(&enterpriseKey, "key", "all", "query the state of enterprise config by key")
}

var enterpriseCmd = &cobra.Command{
	Use:   "enterprise",
	Short: "Print current enterprise status",
	Run: func(cmd *cobra.Command, args []string) {
		msg, err := client.GetEnterpriseConfig(context.Background(), &aergorpc.EnterpriseConfigKey{Key: enterpriseKey})
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		cmd.Println(util.JSON(msg))
	},
}
