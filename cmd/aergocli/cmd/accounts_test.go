package cmd

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestAccountWithPath(t *testing.T) {
	const testDir = "test"
	const testDir2 = "test2"
	const testDir3 = "test3"

	defer func() {
		os.RemoveAll(testDir)
		os.RemoveAll(testDir2)
		os.RemoveAll(testDir3)
	}()

	// New account
	outputNew, err := executeCommand(rootCmd, "account", "new", "--password", "1", "--keystore", testDir)
	assert.NoError(t, err, "should be success")
	re := regexp.MustCompile(`\r?\n`)
	outputNew = re.ReplaceAllString(outputNew, "")
	addr, err := types.DecodeAddress(outputNew)
	assert.NoError(t, err, "should be success")
	assert.Equalf(t, types.AddressLength, len(addr), "wrong address length value = %s", outputNew)

	// List accounts
	outputList, err := executeCommand(rootCmd, "account", "list", "--keystore", testDir)
	assert.NoError(t, err, "should be success")
	outputAddress := strings.Split(outputList[2:], "\"]")[0]
	outputList = re.ReplaceAllString(outputList, "")
	assert.Equalf(t, len(outputNew)+4, len(outputList), "wrong address list length value = %s", outputList)

	// Export using WIF (legacy)
	outputExport, err := executeCommand(rootCmd, "account", "export", "--wif", "--address", outputAddress, "--password", "1", "--keystore", testDir)
	assert.NoError(t, err, "should be success")
	importFormat := strings.TrimSpace(outputExport)

	// Import again, should fail (duplicate)
	outputImport, err := executeCommand(rootCmd, "account", "import", "--if", importFormat, "--password", "1", "--keystore", testDir)
	assert.Contains(t, outputImport, "already exists", "should return duplicate = %s", outputImport)

	// Import again in another path, should succeed
	outputImport, err = executeCommand(rootCmd, "account", "import", "--if", importFormat, "--password", "1", "--keystore", testDir2)
	assert.Equal(t, outputAddress+"\n", outputImport)

	// Export using Keystore
	outputExportKeystore, err := executeCommand(rootCmd, "account", "export", "--address", outputAddress, "--password", "1", "--keystore", testDir)
	assert.NoError(t, err, "should be success")
	keystore := strings.TrimSpace(outputExportKeystore)

	// Import again in another path, should succeed
	outputImport, err = executeCommand(rootCmd, "account", "import", "--path", keystore, "--password", "1", "--keystore", testDir3)
	assert.Equal(t, outputAddress+"\n", outputImport)
}
