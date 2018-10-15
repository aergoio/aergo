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

func TestAccountWithPath(t *testing.T) {
	const testDir = "test"
	outputNew, err := executeCommand(rootCmd, "account", "new", "--password", "1", "--path", testDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	re := regexp.MustCompile(`\r?\n`)
	outputNew = re.ReplaceAllString(outputNew, "")

	addr, err := types.DecodeAddress(outputNew)
	if len(addr) != types.AddressLength {
		t.Errorf("Unexpected output: %v", outputNew)
	}
	outputList, err := executeCommand(rootCmd, "account", "list", "--path", testDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	outputList = re.ReplaceAllString(outputList, "")
	if len(outputList) != len(outputNew)+2 {
		t.Errorf("Unexpected output: %v", outputList)
	}
	os.RemoveAll(testDir)
}
