/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

/*
#cgo CFLAGS: -I${SRCDIR}/../../../contract/native
#cgo LDFLAGS: ${SRCDIR}/../../../contract/native/lib/libascl.a ${SRCDIR}/../../../libtool/lib/libbinaryen.a -lstdc++ -lm

#include "compile.h"
*/
import "C"
import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var githash = "Unknown"

var (
	version  bool
	debug    bool
	printWat bool
	optLvl   uint8
	output   string

	rootCmd = &cobra.Command{
		Use:   "aergoscc [flags] file",
		Short: "Aergo smart contract compiler",
		Run: func(cmd *cobra.Command, args []string) {
			if version {
				fmt.Println(fmt.Sprintf("Aergo smart contract compiler %s", githash))
			} else if len(args) < 1 {
				cmd.Usage()
			} else {
				flags := C.flag_t{}

				if debug {
					flags.val |= C.FLAG_DEBUG
				}
				if printWat {
					flags.val |= C.FLAG_WAT_DUMP
				}
				flags.opt_lvl = C.int(optLvl)
				flags.outfile = C.CString(output)

				err := C.compile(C.CString(args[0]), flags)
				if err != 0 {
					os.Exit(1)
				}
			}
		},
	}
)

func init() {
	log.SetOutput(os.Stderr)
	rootCmd.SetOutput(os.Stdout)
	rootCmd.Flags().BoolVarP(&version, "version", "v", false, "display the compiler version")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "write the output into <string>")
	rootCmd.Flags().BoolVar(&printWat, "print-wat", false, "print WebAssembly text format to stdout")
	rootCmd.Flags().BoolVarP(&debug, "debug", "g", false, "generate debugging information")
	rootCmd.Flags().Uint8VarP(&optLvl, "optimize", "O", 2, "set the optimization level (-O0,-O1,-O2)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
