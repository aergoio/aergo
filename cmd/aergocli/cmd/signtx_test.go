package cmd

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/mr-tron/base58/base58"

	"github.com/aergoio/aergo/cmd/aergocli/util/encoding/json"

	"github.com/aergoio/aergo/types"
)

func TestSignWithKey(t *testing.T) {
	const testAddr = "AmNBjtxomk1uaFrwj8rEKVxYEJ1nzy73dsGrNZzkqs88q8Mkv8GN"
	const signLength = 70
	output, err := executeCommand(rootCmd, "signtx", "--key", "12345678", "--jsontx", "{}")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	outputline := strings.Split(output, "\n")

	addr, err := types.DecodeAddress(outputline[0])
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(addr) != types.AddressLength {
		t.Errorf("Unexpected output: %v", output)
	}
	ouputjson := strings.Join(outputline[1:], "")
	var tx util.InOutTx
	err = json.Unmarshal([]byte(ouputjson), &tx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	sign, err := base58.Decode(tx.Body.Sign)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(sign) != signLength {
		t.Errorf("invalid sign length: %s", tx.Body.Sign)
	}
}

func TestSignWithPath(t *testing.T) {
	const testDir = "signwithpathtest"

	addr, err := executeCommand(rootCmd, "account", "new", "--password", "1", "--path", testDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	re := regexp.MustCompile(`\r?\n`)
	addr = re.ReplaceAllString(addr, "")

	rawaddr, err := types.DecodeAddress(addr)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(rawaddr) != types.AddressLength {
		t.Errorf("Unexpected output: %v", addr)
	}

	_, err = executeCommand(rootCmd, "signtx", "--path", testDir, "--jsontx", "{}", "--password", "1", "--address", addr)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	os.RemoveAll(testDir)
}
