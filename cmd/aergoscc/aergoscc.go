/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package main

/*
#cgo CFLAGS: -I${SRCDIR}/../../contract/native
#cgo LDFLAGS: ${SRCDIR}/../../contract/native/lib/libascl.a ${SRCDIR}/../../libtool/lib/libbinaryen.a -lstdc++ -lm

#include "compile.h"
#include "version.h"
*/
import "C"
import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "aergoscc [flags] file",
		Short: "Aergo smart contract compiler",
		Run: func(cmd *cobra.Command, args []string) {
			if version {
				fmt.Println(fmt.Sprintf("Aergo smart contract compiler %d.%d.%d", C.MAJOR_VER, C.MINOR_VER, C.PATCH_VER))
			} else if len(args) < 1 {
				cmd.Usage()
			} else {
				C.compile(C.CString(args[0]), C.FLAG_NONE)
			}
		},
	}
	output  string
	version bool
)

func init() {
	log.SetOutput(os.Stderr)
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "write the output into <string>")
	rootCmd.PersistentFlags().BoolVarP(&version, "version", "v", false, "display the compiler version")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
