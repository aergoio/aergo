/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	"github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
)

var metricCmd = &cobra.Command{
	Use:   "metric",
	Short: "Show metric informations",
	Run:   execMetric,
}

var ()

func init() {
	rootCmd.AddCommand(metricCmd)
}

func execMetric(cmd *cobra.Command, args []string) {
	req := &types.MetricsRequest{}

	msg, err := client.Metric(context.Background(), req)
	if err != nil {
		cmd.Printf("Failed to get metric from server: %s\n", err.Error())
		return
	}
	// address and peerid should be encoded, respectively
	cmd.Println(util.JSON(msg))
}
