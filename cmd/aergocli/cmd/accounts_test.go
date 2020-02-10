package cmd

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func TestAccountWithPath(t *testing.T) {
	const testDir = "test"
	const testDir2 = "test2"
	outputNew, err := executeCommand(rootCmd, "account", "new", "--password", "1", "--path", testDir)
	assert.NoError(t, err, "should be success")
	re := regexp.MustCompile(`\r?\n`)
	outputNew = re.ReplaceAllString(outputNew, "")

	addr, err := types.DecodeAddress(outputNew)
	assert.NoError(t, err, "should be success")
	assert.Equalf(t, types.AddressLength, len(addr), "wrong address length value = %s", outputNew)

	outputList, err := executeCommand(rootCmd, "account", "list", "--path", testDir)
	assert.NoError(t, err, "should be success")
	outputAddress := strings.Split(outputList[2:], "\"]")[0]

	outputList = re.ReplaceAllString(outputList, "")
	assert.Equalf(t, len(outputNew)+4, len(outputList), "wrong address list length value = %s", outputList)

	outputExport, err := executeCommand(rootCmd, "account", "export", "--address", outputAddress, "--password", "1", "--path", testDir)
	assert.NoError(t, err, "should be success")
	importFormat := strings.Split(outputExport, "\n")[0]

	outputImport, err := executeCommand(rootCmd, "account", "import", "--if", importFormat, "--password", "1", "--path", testDir)
	assert.Equalf(t, "Failed: already exists\n", outputImport, "should return duplicate = %s", outputImport)

	outputImport, err = executeCommand(rootCmd, "account", "import", "--if", importFormat, "--password", "1", "--path", testDir2)
	assert.Equal(t, outputAddress+"\n", outputImport)
	os.RemoveAll(testDir)
	os.RemoveAll(testDir2)
}
