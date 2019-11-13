package util

/*
#cgo CFLAGS: -I${SRCDIR}/../../../libtool/include/luajit-2.1
#cgo LDFLAGS: ${SRCDIR}/../../../libtool/lib/libluajit-5.1.a -lm

#include <stdlib.h>
#include <lualib.h>
#include "compile.h"
*/
import "C"
import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"unsafe"

	"github.com/aergoio/aergo/cmd/aergoluac/encoding"
)

func NewLState() *C.lua_State {
	L := C.luac_vm_newstate()
	if L == nil {
		runtime.GC()
		L = C.luac_vm_newstate()
	}
	return L
}

func CloseLState(L *C.lua_State) {
	if L != nil {
		C.luac_vm_close(L)
	}
}

func Compile(L *C.lua_State, code string) (ByteCodeABI, error) {
	cStr := C.CString(code)
	defer C.free(unsafe.Pointer(cStr))
	if errMsg := C.vm_loadstring(L, cStr); errMsg != nil {
		return nil, errors.New(C.GoString(errMsg))
	}
	if errMsg := C.vm_stringdump(L); errMsg != nil {
		return nil, errors.New(C.GoString(errMsg))
	}
	return dumpToBytes(L), nil
}

func CompileFromFile(srcFileName, outFileName, abiFileName string) error {
	cSrcFileName := C.CString(srcFileName)
	cOutFileName := C.CString(outFileName)
	cAbiFileName := C.CString(abiFileName)
	L := C.luac_vm_newstate()
	defer C.free(unsafe.Pointer(cSrcFileName))
	defer C.free(unsafe.Pointer(cOutFileName))
	defer C.free(unsafe.Pointer(cAbiFileName))
	defer C.luac_vm_close(L)

	if errMsg := C.vm_compile(L, cSrcFileName, cOutFileName, cAbiFileName); errMsg != nil {
		return errors.New(C.GoString(errMsg))
	}
	return nil
}

func DumpFromFile(srcFileName string) error {
	cSrcFileName := C.CString(srcFileName)
	L := C.luac_vm_newstate()
	defer C.free(unsafe.Pointer(cSrcFileName))
	defer C.luac_vm_close(L)

	if errMsg := C.vm_loadfile(L, cSrcFileName); errMsg != nil {
		return errors.New(C.GoString(errMsg))
	}
	if errMsg := C.vm_stringdump(L); errMsg != nil {
		return errors.New(C.GoString(errMsg))
	}

	fmt.Println(encoding.EncodeCode(dumpToBytes(L)))
	return nil
}

func DumpFromStdin() error {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return err
	}
	var buf []byte
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		buf, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		var bBuf bytes.Buffer
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			bBuf.WriteString(scanner.Text() + "\n")
		}
		if err = scanner.Err(); err != nil {
			return err
		}
		buf = bBuf.Bytes()
	}
	srcCode := C.CString(string(buf))
	L := C.luac_vm_newstate()
	defer C.free(unsafe.Pointer(srcCode))
	defer C.luac_vm_close(L)

	if errMsg := C.vm_loadstring(L, srcCode); errMsg != nil {
		return errors.New(C.GoString(errMsg))
	}
	if errMsg := C.vm_stringdump(L); errMsg != nil {
		return errors.New(C.GoString(errMsg))
	}
	fmt.Println(encoding.EncodeCode(dumpToBytes(L)))
	return nil
}

func dumpToBytes(L *C.lua_State) ByteCodeABI {
	var (
		c, a *C.char
		lc, la C.size_t
	)
	c = C.lua_tolstring(L, -2, &lc)
	a = C.lua_tolstring(L, -1, &la)
	return NewByteCodeABI(C.GoBytes(unsafe.Pointer(c), C.int(lc)), C.GoBytes(unsafe.Pointer(a), C.int(la)))
}

type ByteCodeABI []byte

const byteCodeLenLen = 4

func NewByteCodeABI(byteCode, abi []byte) ByteCodeABI {
	byteCodeLen := len(byteCode)
	code := make(ByteCodeABI, byteCodeLenLen+byteCodeLen+len(abi))
	binary.LittleEndian.PutUint32(code, uint32(byteCodeLen))
	copy(code[byteCodeLenLen:], byteCode)
	copy(code[byteCodeLenLen+byteCodeLen:], abi)
	return code
}

func (bc ByteCodeABI) ByteCodeLen() int {
	return int(binary.LittleEndian.Uint32(bc[:byteCodeLenLen]))
}

func (bc ByteCodeABI) ABI() []byte {
	return bc[byteCodeLenLen+bc.ByteCodeLen():]
}

func (bc ByteCodeABI) Len() int {
	return len(bc)
}

type DeployPayload []byte

const byteCodeAbiLenLen = 4

func NewDeployPayload(code ByteCodeABI, args []byte) DeployPayload {
	payload := make([]byte, byteCodeAbiLenLen+code.Len()+len(args))
	binary.LittleEndian.PutUint32(payload[0:], uint32(code.Len()+byteCodeAbiLenLen))
	copy(payload[byteCodeAbiLenLen:], code)
	copy(payload[byteCodeAbiLenLen+code.Len():], args)
	return payload
}

func (dp DeployPayload) headLen() int {
	return int(binary.LittleEndian.Uint32(dp[:byteCodeAbiLenLen]))
}

func (dp DeployPayload) Code() ByteCodeABI {
	return ByteCodeABI(dp[byteCodeAbiLenLen:dp.headLen()])
}

func (dp DeployPayload) CodeLen() int {
	return dp.headLen() - byteCodeAbiLenLen
}

func (dp DeployPayload) HasArgs() bool {
	return len(dp) > dp.headLen()
}

func (dp DeployPayload) Args() []byte {
	return dp[dp.headLen():]
}
