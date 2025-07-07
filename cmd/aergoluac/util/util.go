package util

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/cmd/aergoluac/encoding"
)


////////////////////////////////////////////////////////////////////////////////
// Pack/Bundle
////////////////////////////////////////////////////////////////////////////////

// Set to track imported files
var importedFiles = make(map[string]bool)

// ProcessLines processes a string containing contract code, handling imports
func processLines(input string, currentDir string) (string, error) {

	// Create a string builder for the output
	var output strings.Builder

	// Process each line
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()

		// Check if line starts with "import "
		if strings.HasPrefix(line, "import ") {
			// Extract the file name from import statement
			importLine := strings.TrimSpace(line[6:]) // Remove "import " prefix

			// Check if it has minimum length for a valid import
			if len(importLine) < 3 {
				return "", fmt.Errorf("invalid import format: %s", line)
			}

			// Check if it starts with a valid quote character
			quoteChar := importLine[0]
			if quoteChar != '"' && quoteChar != '\'' {
				return "", fmt.Errorf("import statement must use quotes: %s", line)
			}

			// Check if it ends with the same quote character
			if importLine[len(importLine)-1] != quoteChar {
				return "", fmt.Errorf("mismatched quotes in import: %s", line)
			}

			// Extract the file path between quotes
			importFile := importLine[1 : len(importLine)-1]

			// If it's not a URL and not an absolute path, make it relative to the current file's directory
			if !strings.HasPrefix(importFile, "http") && !filepath.IsAbs(importFile) {
				importFile = filepath.Join(currentDir, importFile)
			}

			// Get absolute path to check for circular imports
			absImportPath, err := filepath.Abs(importFile)
			if err != nil {
				return "", fmt.Errorf("error getting absolute path: %w", err)
			}

			// Skip if already imported
			if importedFiles[absImportPath] {
				continue
			}

			// Mark as imported
			importedFiles[absImportPath] = true

			// Read the imported file
			importContent, err := readContractFile(importFile)
			if err != nil {
				return "", fmt.Errorf("error importing file '%s': %w", importFile, err)
			}

			// Get the directory of the imported file for nested imports
			importedFileDir := filepath.Dir(importFile)

			// Process the imported content recursively
			processedImport, err := processLines(importContent, importedFileDir)
			if err != nil {
				return "", err
			}

			// Add the processed import to output
			output.WriteString(processedImport)
			output.WriteString("\n")
		} else {
			// Regular line, add to output
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error scanning input: %w", err)
	}

	return output.String(), nil
}

func readContractFile(filePath string) (string, error) {
	// if the file path is a url, read it from the web
	if strings.HasPrefix(filePath, "http") {
		// search in the web
		req, err := http.NewRequest("GET", filePath, nil)
		if err != nil {
			return "", err
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		fileBytes, _ := ioutil.ReadAll(resp.Body)
		return string(fileBytes), nil
	}

	// search in the local file system
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", err
	}
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(fileBytes), nil
}

// ReadContract reads a contract file and bundles the imports
func ReadContract(filePath string) (string, error) {
	// Reset imported files tracking for each new processing
	importedFiles = make(map[string]bool)

	// Get the absolute path of the main file
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %w", err)
	}

	// Get the directory of the main file
	mainFileDir := filepath.Dir(absPath)

	// Mark the main file as imported
	importedFiles[absPath] = true

	// read the contract file
	output, err := readContractFile(filePath)
	if err != nil {
		return "", err
	}

	// process the contract file for import statements, passing the main file's directory
	return processLines(output, mainFileDir)
}

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
