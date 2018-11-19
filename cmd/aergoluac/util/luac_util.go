package util

/*
#cgo CFLAGS: -I${SRCDIR}/../../../libtool/include/luajit-2.0
#cgo LDFLAGS: ${SRCDIR}/../../../libtool/lib/libluajit-5.1.a -lm

#include <stdlib.h>
#include "compile.h"
*/
import "C"
import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/aergoio/aergo/cmd/aergocli/util"
	"io/ioutil"
	"log"
	"os"
	"unsafe"
)

var (
	b bytes.Buffer
)

func Compile(code string) ([]byte, error) {
	b.Reset()
	L := C.luac_vm_newstate()
	defer C.luac_vm_close(L)
	cstr := C.CString(code)
	defer C.free(unsafe.Pointer(cstr))
	if errMsg := C.vm_loadstring(L, cstr); errMsg != nil {
		return nil, errors.New(C.GoString(errMsg))
	}
	if errMsg := C.vm_stringdump(L); errMsg != nil {
		return nil, errors.New(C.GoString(errMsg))
	}
	return b.Bytes(), nil
}

func CompileFromFile(srcFileName, outFileName, abiFileName string) {
	cSrcFileName := C.CString(srcFileName)
	cOutFileName := C.CString(outFileName)
	cAbiFileName := C.CString(abiFileName)
	L := C.luac_vm_newstate()
	defer C.free(unsafe.Pointer(cSrcFileName))
	defer C.free(unsafe.Pointer(cOutFileName))
	defer C.free(unsafe.Pointer(cAbiFileName))
	defer C.luac_vm_close(L)

	if errMsg := C.vm_compile(L, cSrcFileName, cOutFileName, cAbiFileName); errMsg != nil {
		log.Fatal(C.GoString(errMsg))
	}
}

func DumpFromFile(srcFileName string) {
	cSrcFileName := C.CString(srcFileName)
	L := C.luac_vm_newstate()
	defer C.free(unsafe.Pointer(cSrcFileName))
	defer C.luac_vm_close(L)

	if errMsg := C.vm_loadfile(L, cSrcFileName); errMsg != nil {
		log.Fatal(C.GoString(errMsg))
	}
	if errMsg := C.vm_stringdump(L); errMsg != nil {
		log.Fatal(C.GoString(errMsg))
	}

	fmt.Print(util.EncodeCode(b.Bytes()))
}

func DumpFromStdin() {
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}
	var buf []byte
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		buf, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		var bBuf bytes.Buffer
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			bBuf.WriteString(scanner.Text() + "\n")
		}
		if err = scanner.Err(); err != nil {
			log.Fatal(err)
		}
		buf = bBuf.Bytes()
	}
	srcCode := C.CString(string(buf))
	L := C.luac_vm_newstate()
	defer C.free(unsafe.Pointer(srcCode))
	defer C.luac_vm_close(L)

	if errMsg := C.vm_loadstring(L, srcCode); errMsg != nil {
		log.Fatal(C.GoString(errMsg))
	}
	if errMsg := C.vm_stringdump(L); errMsg != nil {
		log.Fatal(C.GoString(errMsg))
	}
	fmt.Print(util.EncodeCode(b.Bytes()))
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
