/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

/*
#cgo CFLAGS: -I${SRCDIR}/../../../contract/native -I${SRCDIR}/../../../libtool/include
#cgo LDFLAGS: ${SRCDIR}/../../../contract/native/lib/libascl.a
#cgo LDFLAGS: -L${SRCDIR}/../../../libtool/lib -lgmp -lbinaryen -lstdc++ -lm

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
	debug     bool
	version   bool
	printLex  bool
	printYacc bool
	printWat  bool
	optLvl    uint8
	stackSize uint32

	rootCmd = &cobra.Command{
		Use: "aergoscc [flags] file",
		Run: func(cmd *cobra.Command, args []string) {
			if version {
				fmt.Println(fmt.Sprintf("Aergo smart contract compiler %s", githash))
			} else if len(args) < 1 {
				fmt.Println("Error: no input files")
				os.Exit(1)
			} else {
				flags := C.flag_t{}

				if debug {
					flags.val |= C.FLAG_DEBUG
				}
				if printLex {
					flags.val |= C.FLAG_DUMP_LEX
				}
				if printYacc {
					flags.val |= C.FLAG_DUMP_YACC
				}
				if printWat {
					flags.val |= C.FLAG_DUMP_WAT
				}
				flags.opt_lvl = C.uint8_t(optLvl)
				flags.stack_size = C.uint32_t(stackSize)

				err := C.int(0)
				for _, arg := range args {
                    flags.path = C.CString(arg);
					err |= C.compile(flags)
				}
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

	rootCmd.Flags().BoolVarP(&debug, "debug", "g", false, "")
	rootCmd.Flags().Uint8VarP(&optLvl, "optimize", "O", 2, "")
	rootCmd.Flags().BoolVar(&printLex, "print-lex", false, "")
	rootCmd.Flags().BoolVar(&printYacc, "print-yacc", false, "")
	rootCmd.Flags().BoolVar(&printWat, "print-wat", false, "")
	rootCmd.Flags().Uint32Var(&stackSize, "stack", 0x100000, "")
	rootCmd.Flags().BoolVarP(&version, "version", "v", false, "")

	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Println("Usage:")
		fmt.Println("  aergoscc [options] file...")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  -h,--help                Display this information")
		fmt.Println("  -v,--version             Display the version of the compiler")
		fmt.Println("")
		fmt.Println("  -g                       Generate debug information")
		fmt.Println("  -O0,-O1,-O2              Set the optimization level (default 2)")
		fmt.Println("")
		fmt.Println("  --print-wat              Print WebAssembly text format to stdout")
		fmt.Println("  --stack <size>           Set the maximum stack size (default 1048576)")
		fmt.Println("")
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
