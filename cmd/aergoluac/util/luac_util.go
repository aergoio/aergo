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

	"github.com/aergoio/aergo/v2/cmd/aergoluac/encoding"
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

func Compile(L *C.lua_State, code string) (LuaCode, error) {
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

func dumpToBytes(L *C.lua_State) LuaCode {
	var (
		c, a   *C.char
		lc, la C.size_t
	)
	c = C.lua_tolstring(L, -2, &lc)
	a = C.lua_tolstring(L, -1, &la)
	return NewLuaCode(C.GoBytes(unsafe.Pointer(c), C.int(lc)), C.GoBytes(unsafe.Pointer(a), C.int(la)))
}

type LuaCode []byte

const byteCodeLenLen = 4

func NewLuaCode(byteCode, abi []byte) LuaCode {
	byteCodeLen := len(byteCode)
	code := make(LuaCode, byteCodeLenLen+byteCodeLen+len(abi))
	binary.LittleEndian.PutUint32(code, uint32(byteCodeLen))
	copy(code[byteCodeLenLen:], byteCode)
	copy(code[byteCodeLenLen+byteCodeLen:], abi)
	return code
}

func (c LuaCode) ByteCode() []byte {
	if !c.IsValidFormat() {
		return nil
	}
	return c[byteCodeLenLen : byteCodeLenLen+c.byteCodeLen()]
}

func (c LuaCode) byteCodeLen() int {
	if c.Len() < byteCodeLenLen {
		return 0
	}
	return int(binary.LittleEndian.Uint32(c[:byteCodeLenLen]))
}

func (c LuaCode) ABI() []byte {
	if !c.IsValidFormat() {
		return nil
	}
	return c[byteCodeLenLen+c.byteCodeLen():]
}

func (c LuaCode) Len() int {
	return len(c)
}

func (c LuaCode) IsValidFormat() bool {
	if c.Len() <= byteCodeLenLen {
		return false
	}
	return byteCodeLenLen+c.byteCodeLen() < c.Len()
}

func (c LuaCode) Bytes() []byte {
	return c
}

type LuaCodePayload []byte

const codeLenLen = 4

func NewLuaCodePayload(code LuaCode, args []byte) LuaCodePayload {
	payload := make([]byte, codeLenLen+code.Len()+len(args))
	binary.LittleEndian.PutUint32(payload[0:], uint32(code.Len()+codeLenLen))
	copy(payload[codeLenLen:], code.Bytes())
	copy(payload[codeLenLen+code.Len():], args)
	return payload
}

func (p LuaCodePayload) headLen() int {
	if p.Len() < codeLenLen {
		return 0
	}
	return int(binary.LittleEndian.Uint32(p[:codeLenLen]))
}

func (p LuaCodePayload) Code() LuaCode {
	if v, _ := p.IsValidFormat(); !v {
		return nil
	}
	return LuaCode(p[codeLenLen:p.headLen()])
}

func (p LuaCodePayload) HasArgs() bool {
	if v, _ := p.IsValidFormat(); !v {
		return false
	}
	return len(p) > p.headLen()
}

func (p LuaCodePayload) Args() []byte {
	if v, _ := p.IsValidFormat(); !v {
		return nil
	}
	return p[p.headLen():]
}

func (p LuaCodePayload) Len() int {
	return len(p)
}

func (p LuaCodePayload) IsValidFormat() (bool, error) {
	if p.Len() <= codeLenLen {
		return false, fmt.Errorf("invalid code (%d bytes is too short)", p.Len())
	}
	if p.Len() < p.headLen() {
		return false, fmt.Errorf("invalid code (expected %d bytes, actual %d bytes)", p.headLen(), p.Len())
	}
	return true, nil
}
