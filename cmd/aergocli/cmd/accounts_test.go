package cmd

import (
	"bytes"
	"os"
	"regexp"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOutput(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

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
