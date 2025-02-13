package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
)

func TestCompile(t *testing.T) {
	StartVMPool(1)

	// Read the test Lua code
	code, err := os.ReadFile("vm_dummy/test_files/gas_op.lua")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Compile the code
	byteCodeAbi, err := Compile(string(code), false)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Calculate SHA256 hash of the compiled code
	hash := sha256.Sum256(byteCodeAbi)
	actualHash := hex.EncodeToString(hash[:])

	// Expected hash
	expectedHash := "c9a19d29ed80a0353c0128436946f773c612e8218997ad480d2b40f7759cbee8"

	if actualHash != expectedHash {
		t.Errorf("Compiled code hash mismatch\nExpected: %s\nGot: %s", expectedHash, actualHash)
	}
}
