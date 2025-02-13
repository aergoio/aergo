package util

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/cmd/aergoluac/encoding"
)


////////////////////////////////////////////////////////////////////////////////
// Decode
////////////////////////////////////////////////////////////////////////////////

// Decode decodes the payload from a hex string or a base58 string or a JSON string
// and writes the bytecode, abi and deploy arguments to files
func Decode(srcFileName string, payload string) error {
	var decoded []byte
	var err error

	// check if the payload is in hex format
	if hex.IsHexString(payload) {
		// the data is expected to be copied from aergoscan view of
		// the transaction that deployed the contract
		decoded, err = hex.Decode(payload)
	} else {
		// the data is the output of aergoluac
		decoded, err = encoding.DecodeCode(payload)
		if err != nil {
			// the data is extracted from JSON transaction from aergocli
			decoded, err = base58.Decode(payload)
		}
	}
	if err != nil {
		return fmt.Errorf("failed to decode payload 1: %v", err.Error())
	}

	err = os.WriteFile(srcFileName + "-raw", decoded, 0644);
	if err != nil {
		return fmt.Errorf("failed to write raw file: %v", err.Error())
	}

	data := LuaCodePayload(decoded)
	_, err = data.IsValidFormat()
	if err != nil {
		return fmt.Errorf("failed to decode payload 2: %v", err.Error())
	}

	contract := data.Code()
	if !contract.IsValidFormat() {
		// the data is the output of aergoluac, so it does not contain deploy arguments
		contract = LuaCode(decoded)
		data = NewLuaCodePayload(contract, []byte{})
	}

	err = os.WriteFile(srcFileName + "-bytecode", contract.ByteCode(), 0644);
	if err != nil {
		return fmt.Errorf("failed to write bytecode file: %v", err.Error())
	}

	err = os.WriteFile(srcFileName + "-abi", contract.ABI(), 0644);
	if err != nil {
		return fmt.Errorf("failed to write ABI file: %v", err.Error())
	}

	var deployArgs []byte
	if data.HasArgs() {
		deployArgs = data.Args()
	}
	err = os.WriteFile(srcFileName + "-deploy-arguments", deployArgs, 0644);
	if err != nil {
		return fmt.Errorf("failed to write deploy-arguments file: %v", err.Error())
	}

	fmt.Println("done.")
	return nil
}

func DecodeFromFile(srcFileName string) error {
	payload, err := os.ReadFile(srcFileName)
	if err != nil {
		return fmt.Errorf("failed to read payload file: %v", err.Error())
	}
	return Decode(srcFileName, string(payload))
}

func DecodeFromStdin() error {
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
	return Decode("contract", string(buf))
}


////////////////////////////////////////////////////////////////////////////////
// LuaCode and LuaCodePayload
// used to store bytecode, abi and deploy arguments
////////////////////////////////////////////////////////////////////////////////

type LuaCode []byte

func NewLuaCode(byteCode, abi []byte) LuaCode {
	byteCodeLen := len(byteCode)
	code := make(LuaCode, 4+byteCodeLen+len(abi))
	binary.LittleEndian.PutUint32(code, uint32(byteCodeLen))
	copy(code[4:], byteCode)
	copy(code[4+byteCodeLen:], abi)
	return code
}

func (c LuaCode) ByteCode() []byte {
	if !c.IsValidFormat() {
		return nil
	}
	return c[4:4+c.byteCodeLen()]
}

func (c LuaCode) byteCodeLen() int {
	if c.Len() < 4 {
		return 0
	}
	return int(binary.LittleEndian.Uint32(c[:4]))
}

func (c LuaCode) ABI() []byte {
	if !c.IsValidFormat() {
		return nil
	}
	return c[4+c.byteCodeLen():]
}

func (c LuaCode) Len() int {
	return len(c)
}

func (c LuaCode) IsValidFormat() bool {
	if c.Len() <= 4 {
		return false
	}
	return 4 + c.byteCodeLen() < c.Len()
}

func (c LuaCode) Bytes() []byte {
	return c
}

//------------------------------------------------------------------------------

type LuaCodePayload []byte

func NewLuaCodePayload(code LuaCode, args []byte) LuaCodePayload {
	payload := make([]byte, 4+code.Len()+len(args))
	binary.LittleEndian.PutUint32(payload[0:], uint32(4+code.Len()))
	copy(payload[4:], code.Bytes())
	copy(payload[4+code.Len():], args)
	return payload
}

func (p LuaCodePayload) headLen() int {
	if p.Len() < 4 {
		return 0
	}
	return int(binary.LittleEndian.Uint32(p[:4]))
}

func (p LuaCodePayload) Code() LuaCode {
	if v, _ := p.IsValidFormat(); !v {
		return nil
	}
	return LuaCode(p[4:p.headLen()])
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
	if p.Len() <= 4 {
		return false, fmt.Errorf("invalid code (%d bytes is too short)", p.Len())
	}
	if p.Len() < p.headLen() {
		return false, fmt.Errorf("invalid code (expected %d bytes, actual %d bytes)", p.headLen(), p.Len())
	}
	return true, nil
}
