/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package main

import "github.com/spf13/cobra"

var githash = "No git hash provided"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Polaris",
	Long:  `The version of Polaris is following the one of Arego`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("Polaris %s\n", githash)
	},
}
