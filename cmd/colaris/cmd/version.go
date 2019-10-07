/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package cmd

import "github.com/spf13/cobra"

var githash = "No git hash provided"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Colaris",
	Long:  `The version of Colaris is following the one of Polaris`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("Colaris %s\n", githash)
	},
}
