package cmd

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	"github.com/aergoio/aergo/v2/cmd/aergocli/util/encoding/json"
	"github.com/aergoio/aergo/v2/types"
	"github.com/mr-tron/base58/base58"
	"github.com/stretchr/testify/assert"
)

func TestSignWithKey(t *testing.T) {
	const testAddr = "AmNBjtxomk1uaFrwj8rEKVxYEJ1nzy73dsGrNZzkqs88q8Mkv8GN"
	const signLength = 71
	output, err := executeCommand(rootCmd, "signtx", "--key", "12345678", "--jsontx", "{}")
	assert.NoError(t, err, "should be success")

	outputline := strings.Split(output, "\n")

	addr, err := types.DecodeAddress(outputline[0])
	assert.NoError(t, err, "should be success")
	assert.Equalf(t, types.AddressLength, len(addr), "wrong address length value = %s", output)

	ouputjson := strings.Join(outputline[1:], "")
	var tx util.InOutTx
	err = json.Unmarshal([]byte(ouputjson), &tx)
	assert.NoError(t, err, "should be success")

	sign, err := base58.Decode(tx.Body.Sign)
	assert.NoError(t, err, "should be success")
	assert.Equalf(t, len(sign), signLength, "wrong sign length value = %s", tx.Body.Sign)
}

func TestSignWithPath(t *testing.T) {
	const testDir = "signwithpathtest"

	addr, err := executeCommand(rootCmd, "account", "new", "--password", "1", "--keystore", testDir)
	assert.NoError(t, err, "should be success")

	re := regexp.MustCompile(`\r?\n`)
	addr = re.ReplaceAllString(addr, "")

	rawaddr, err := types.DecodeAddress(addr)
	assert.NoError(t, err, "should be success")
	assert.Equalf(t, types.AddressLength, len(rawaddr), "wrong address length from %s", addr)

	_, err = executeCommand(rootCmd, "signtx", "--keystore", testDir, "--jsontx", "{}", "--password", "1", "--address", addr)
	assert.NoError(t, err, "should be success")

	os.RemoveAll(testDir)
}
