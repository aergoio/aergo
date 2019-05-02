package main

import "github.com/spf13/cobra"

var githash = "No git hash provided"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Aergoluac",
	Long:  `All software has versions. This is Aergo's`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("Aergoluac %s\n", githash)
	},
}
