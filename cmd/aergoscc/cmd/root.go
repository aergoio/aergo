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
	compile  bool
	debug    bool
	version  bool
	printWat bool
	optLvl   uint8
	output   string

	rootCmd = &cobra.Command{
		Use:   "aergoscc [flags] file",
		Short: "Aergo smart contract compiler",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if version {
				fmt.Println(fmt.Sprintf("Aergo smart contract compiler %s", githash))
			} else if len(args) < 1 {
				cmd.Usage()
			} else {
				flags := C.flag_t{}

				if compile {
					flags.val |= C.FLAG_COMPILE
				}
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

	rootCmd.Flags().BoolVarP(&compile, "compile", "c", false, "compile only")
	rootCmd.Flags().BoolVarP(&debug, "debug", "g", false, "generate debugging information")
	rootCmd.Flags().Uint8VarP(&optLvl, "optimize", "O", 2, "set the optimization level (-O0,-O1,-O2)")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "write the output into <string>")
	rootCmd.Flags().BoolVar(&printWat, "print-wat", false, "print WebAssembly text format to stdout")
	rootCmd.Flags().BoolVarP(&version, "version", "v", false, "version for aergoscc")

	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Println("Usage:")
		fmt.Println("  aergoscc [options] file")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  -h, --help           Display this information")
		fmt.Println("  -v, --version        Display the version of the compiler")
		fmt.Println("  -c                   Compile only")
		fmt.Println("  -g                   Generate debugging information")
		fmt.Println("  -O<n>                Set the optimization level (n=0,1,2) (default 2)")
		fmt.Println("  -o <file>            Write the output into <file>")
		fmt.Println("  --print-wat          Print WebAssembly text format to stdout")
		return nil
	},
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
