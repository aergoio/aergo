package cmd

import (
	"bytes"
	"os"
	"regexp"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
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

func TestNewAccount(t *testing.T) {
	const testDir = "test"
	output, err := executeCommand(rootCmd, "account", "new", "--password", "1", "--path", testDir)
	re := regexp.MustCompile(`\r?\n`)
	output = re.ReplaceAllString(output, "")

	addr, err := types.DecodeAddress(output)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(addr) != types.AddressLength {
		t.Errorf("Unexpected output: %v", output)
	}
	os.RemoveAll(testDir)
}
