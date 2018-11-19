package cmd

import (
	"os"
	"regexp"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func TestAccountWithPath(t *testing.T) {
	const testDir = "test"
	outputNew, err := executeCommand(rootCmd, "account", "new", "--password", "1", "--path", testDir)
	assert.NoError(t, err, "should be success")
	re := regexp.MustCompile(`\r?\n`)
	outputNew = re.ReplaceAllString(outputNew, "")

	addr, err := types.DecodeAddress(outputNew)
	assert.NoError(t, err, "should be success")
	assert.Equalf(t, types.AddressLength, len(addr), "wrong address length value = %s", outputNew)

	outputList, err := executeCommand(rootCmd, "account", "list", "--path", testDir)
	assert.NoError(t, err, "should be success")

	outputList = re.ReplaceAllString(outputList, "")
	assert.Equalf(t, len(outputNew)+2, len(outputList), "wrong address list length value = %s", outputList)

	os.RemoveAll(testDir)
}
