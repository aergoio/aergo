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
	debug     bool
	version   bool
	printLex  bool
	printYacc bool
	printWat  bool
	optLvl    uint8
	stackSize uint32
	output    string

	rootCmd = &cobra.Command{
		Use:   "aergoscc [flags] file",
		Short: "Aergo smart contract compiler",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if version {
				fmt.Println(fmt.Sprintf("Aergo smart contract compiler %s", githash))
			} else if len(args) < 1 {
				fmt.Println("Error: no input file")
				cmd.Usage()
			} else {
				if stackSize%65536 != 0 {
					fmt.Println("Error: stack size should be a multiple of 64KB")
					os.Exit(1)
				}

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

	rootCmd.Flags().BoolVarP(&debug, "debug", "g", false, "")
	rootCmd.Flags().Uint8VarP(&optLvl, "optimize", "O", 2, "")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "")
	rootCmd.Flags().BoolVar(&printLex, "print-lex", false, "")
	rootCmd.Flags().BoolVar(&printYacc, "print-yacc", false, "")
	rootCmd.Flags().BoolVar(&printWat, "print-wat", false, "")
	rootCmd.Flags().Uint32Var(&stackSize, "stack-size", 0x100000, "")
	rootCmd.Flags().BoolVarP(&version, "version", "v", false, "")

	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Println("Usage:")
		fmt.Println("  aergoscc [options] file")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --help                   Display this information")
		fmt.Println("  --version                Display the version of the compiler")
		fmt.Println("")
		fmt.Println("  -g                       Generates debug information")
		fmt.Println("  -O<level>                Sets the optimization level (default 2)")
		fmt.Println("                           <level> should be 0, 1 or 2")
		fmt.Println("  -o <file>                Writes the output into <file>")
		fmt.Println("  --print-wat              Prints WebAssembly text format to stdout")
		fmt.Println("  --stack-size <size>      Sets the maximum stack size (default 1048576)")
		fmt.Println("                           <size> should be a multiple of 64KB")
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
