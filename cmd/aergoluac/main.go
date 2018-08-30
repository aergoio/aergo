/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package main

/*
#cgo CFLAGS: -I${SRCDIR}/../../libtool/include/luajit-2.0
#cgo LDFLAGS: ${SRCDIR}/../../libtool/lib/libluajit-5.1.a -lm

#include <stdlib.h>
#include "compile.h"
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"unsafe"

	"github.com/btcsuite/btcutil/base58"
	"github.com/spf13/cobra"
)

var (
	rootCmd *cobra.Command
	abiFile string
	payload bool
	b       bytes.Buffer
)

func init() {
	log.SetOutput(os.Stderr)
	rootCmd = &cobra.Command{
		Use:   "aergoluac [flags] srcfile bcfile",
		Short: "compile a contract",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if payload {
				dump(args[0])
			} else {
				if len(args) < 2 {
					log.Fatal(cmd.UsageString())
				}
				compile(args[0], args[1], abiFile)
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&abiFile, "abi", "a", "", "abi filename")
	rootCmd.PersistentFlags().BoolVar(&payload, "payload", false, "print a base58 encoded payload")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func compile(srcFileName, outFileName, abiFileName string) {
	cSrcFileName := C.CString(srcFileName)
	cOutFileName := C.CString(outFileName)
	cAbiFileName := C.CString(abiFileName)
	L := C.vm_newstate()
	defer C.free(unsafe.Pointer(cSrcFileName))
	defer C.free(unsafe.Pointer(cOutFileName))
	defer C.free(unsafe.Pointer(cAbiFileName))
	defer C.vm_close(L)

	if errMsg := C.vm_compile(L, cSrcFileName, cOutFileName, cAbiFileName); errMsg != nil {
		log.Fatal(C.GoString(errMsg))
	}
}

func dump(srcFileName string) {
	cSrcFileName := C.CString(srcFileName)
	L := C.vm_newstate()
	defer C.free(unsafe.Pointer(cSrcFileName))
	defer C.vm_close(L)

	if errMsg := C.vm_stringdump(L, cSrcFileName); errMsg != nil {
		log.Fatal(C.GoString(errMsg))
	}
	fmt.Print(base58.Encode(b.Bytes()))
}

//export addLen
func addLen(length C.int) {
	var l [4]byte
	binary.LittleEndian.PutUint32(l[:], uint32(length))
	b.Write(l[:])
}

//export addByteN
func addByteN(p *C.char, length C.int) {
	s := C.GoStringN(p, length)
	b.WriteString(s)
}
