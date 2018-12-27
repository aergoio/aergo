/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/aergoio/aergo/cmd/aergoluac/util"
	"github.com/spf13/cobra"
)

var (
	rootCmd *cobra.Command
	abiFile string
	payload bool
)

func init() {
	rootCmd = &cobra.Command{
		Use:   "aergoluac --payload srcfile\n  aergoluac --abi abifile srcfile bcfile",
		Short: "Compile a lua contract",
		Long:  "Compile a lua contract. This command makes a bytecode file and a ABI file or prints a payload data.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if payload {
				if len(args) == 0 {
					err = util.DumpFromStdin()
				} else {
					err = util.DumpFromFile(args[0])
				}
			} else {
				if len(args) < 2 {
					return errors.New("2 arguments required: <srcfile> <bcfile>")
				}
				err = util.CompileFromFile(args[0], args[1], abiFile)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}
	rootCmd.PersistentFlags().StringVarP(&abiFile, "abi", "a", "", "abi filename")
	rootCmd.PersistentFlags().BoolVar(&payload, "payload", false, "print the compilation result consisting of bytecode and abi")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
