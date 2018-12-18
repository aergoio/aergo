package exec

import (
	"bufio"
	"fmt"
	"os"

	"github.com/aergoio/aergo/cmd/brick/context"
	"github.com/mattn/go-colorable"
)

func init() {
	registerExec(&batch{})
}

type batch struct{}

func (c *batch) Command() string {
	return "batch"
}

func (c *batch) Syntax() string {
	return fmt.Sprintf("%s", context.PathSymbol)
}

func (c *batch) Usage() string {
	return fmt.Sprintf("batch `<batch_file_path>`")
}

func (c *batch) Describe() string {
	return "batch run"
}

func (c *batch) Validate(args string) error {

	_, err := c.parse(args)

	return err
}

func (c *batch) parse(args string) (string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) != 1 {
		return "", fmt.Errorf("invalid format. usage: %s", c.Usage())
	}

	batchFilePath := splitArgs[0]

	if _, err := os.Stat(batchFilePath.Text); os.IsNotExist(err) {
		return "", fmt.Errorf("fail to read a brick batch file %s: %s", batchFilePath.Text, err.Error())
	}

	return batchFilePath.Text, nil
}

func (c *batch) Run(args string) (string, error) {

	batchFilePath, _ := c.parse(args)

	batchFile, err := os.Open(batchFilePath)
	if err != nil {
		return "", err
	}

	var cmdLines []string
	scanner := bufio.NewScanner(batchFile)
	for scanner.Scan() {
		cmdLines = append(cmdLines, scanner.Text())
	}

	batchFile.Close()

	for i, line := range cmdLines {
		stdOut := colorable.NewColorableStdout()
		cmd, args := context.ParseFirstWord(line)
		if len(cmd) == 0 {
			fmt.Fprintf(stdOut, "\x1B[0;37m%3d\x1B[0m\n", i)
			continue
		} else if context.Comment == cmd {
			fmt.Fprintf(stdOut, "\x1B[0;37m%3d \x1B[32m%s\x1B[0m\n", i, line)
			continue
		}

		fmt.Fprintf(stdOut, "\x1B[0;37m%3d \x1B[34;1m%s \x1B[0m%s\n", i, cmd, args)
		Broker(line)

	}

	return "batch exec is finished", nil
}
