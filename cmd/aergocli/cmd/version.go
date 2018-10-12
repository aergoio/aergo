/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Aergocli",
	Long:  `All software has versions. This is Aergo's`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("Aergocli v0.1 -- HEAD")
	},
}
