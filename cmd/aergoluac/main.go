/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package main

import (
	"github.com/aergoio/aergo/cmd/aergoluac/util"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	rootCmd *cobra.Command
	abiFile string
	payload bool
)

func init() {
	log.SetOutput(os.Stderr)
	rootCmd = &cobra.Command{
		Use:   "aergoluac --payload srcfile\n  aergoluac --abi abifile srcfile bcfile",
		Short: "Compile a lua contract",
		Long:  "Compile a lua contract. This command makes a bytecode file and a ABI file or prints a payload data.",
		Run: func(cmd *cobra.Command, args []string) {
			if payload {
				if len(args) == 0 {
					util.DumpFromStdin()
				} else {
					util.DumpFromFile(args[0])
				}
			} else {
				if len(args) < 2 {
					log.Fatal(cmd.UsageString())
				}
				util.CompileFromFile(args[0], args[1], abiFile)
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&abiFile, "abi", "a", "", "abi filename")
	rootCmd.PersistentFlags().BoolVar(&payload, "payload", false, "print the compilation result consisting of bytecode and abi")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
