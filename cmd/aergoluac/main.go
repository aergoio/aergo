/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package main

import (
	"log"
	"os"

	"github.com/aergoio/aergo/contract"
	"github.com/spf13/cobra"
)

var (
	rootCmd *cobra.Command
	abiFile string
)

func init() {
	log.SetOutput(os.Stderr)
	rootCmd = &cobra.Command{
		Use:   "aergoluac [flags] srcfile bcfile",
		Short: "compile a contract",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := contract.Compile(args[0], args[1], abiFile)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	rootCmd.PersistentFlags().StringVar(&abiFile, "abi", "", "abi filename")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
