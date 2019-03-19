/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

/*
#cgo CFLAGS: -I${SRCDIR}/../../../contract/native
#cgo LDFLAGS: ${SRCDIR}/../../../contract/native/lib/libascl.a
#cgo LDFLAGS: -L${SRCDIR}/../../../libtool/lib -lgmp -lbinaryen -lstdc++ -lm
#cgo LDFLAGS: -L${SRCDIR}/../../../libtool/lib -lASCLVM -lRuntime -lWASTParse -lWASM -lRegExp -lNFA -lLLVMJIT -lIR -lLogging -lPlatform -lWAVMUnwind
#cgo LDFLAGS: -lLLVMSupport -lLLVMCore -lLLVMPasses -lLLVMOrcJIT -lLLVMRuntimeDyld -lLLVMDebugInfoDWARF -lLLVMX86CodeGen -lLLVMX86AsmParser -lLLVMX86AsmPrinter -lLLVMX86Desc -lLLVMX86Disassembler -lLLVMX86Info -lLLVMX86Utils -lLLVMipo -lLLVMInstrumentation -lLLVMVectorize -lLLVMIRReader -lLLVMAsmParser -lLLVMLinker -lLLVMExecutionEngine -lLLVMRuntimeDyld -lLLVMAsmPrinter -lLLVMDebugInfoCodeView -lLLVMDebugInfoMSF -lLLVMGlobalISel -lLLVMSelectionDAG -lLLVMCodeGen -lLLVMScalarOpts -lLLVMInstCombine -lLLVMBitWriter -lLLVMTarget -lLLVMTransformUtils -lLLVMAnalysis -lLLVMProfileData -lLLVMX86AsmPrinter -lLLVMObject -lLLVMMCParser -lLLVMBitReader -lLLVMMCDisassembler -lLLVMMC -lLLVMX86Utils -lLLVMCore -lLLVMBinaryFormat -lLLVMSupport -lz -ltinfo -lpthread -lm -lLLVMDemangle -lrt -ldl

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
		Use:   "aergoscc [flags] file",
		Short: "Aergo smart contract compiler",
		Run: func(cmd *cobra.Command, args []string) {
			if version {
				fmt.Println(fmt.Sprintf("Aergo smart contract compiler %s", githash))
			} else if len(args) < 1 {
				fmt.Println("Error: no input files")
				cmd.Usage()
				os.Exit(1)
			} else if stackSize%65536 != 0 {
				fmt.Println("Error: stack size should be a multiple of 64KB")
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
					err |= C.compile(C.CString(arg), flags)
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
	rootCmd.Flags().Uint32Var(&stackSize, "stack-size", 0x100000, "")
	rootCmd.Flags().BoolVarP(&version, "version", "v", false, "")

	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Println("Usage:")
		fmt.Println("  aergoscc [options] file...")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --help                   Display this information")
		fmt.Println("  --version                Display the version of the compiler")
		fmt.Println("")
		fmt.Println("  -g                       Generate debug information")
		fmt.Println("  -O<level>                Set the optimization level (default 2)")
		fmt.Println("                           <level> should be 0, 1 or 2")
		fmt.Println("  --print-wat              Print WebAssembly text format to stdout")
		fmt.Println("  --stack-size <size>      Set the maximum stack size (default 1048576)")
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
