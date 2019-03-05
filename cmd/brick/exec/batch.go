package exec

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aergoio/aergo/cmd/brick/context"
	"github.com/fsnotify/fsnotify"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
)

var (
	verboseBatch    = false
	letBatchKnowErr error
	batchErrorCount = 0
	enableWatch     = false
	watchFileList   []string
	watcher         *fsnotify.Watcher
)

func EnableVerbose() {
	verboseBatch = true
}

func GetBatchErrorCount() int {
	return batchErrorCount
}

func EnableWatch() {
	enableWatch = true
}

func init() {
	registerExec(&batch{})
}

type batch struct {
	level int
}

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
	stdOut := colorable.NewColorableStdout()

	var err error

	for {
		if c.level == 0 && enableWatch {
			watcher, err = fsnotify.NewWatcher()
			if err != nil {
				return "", err
			}
			// clear screnn
			fmt.Fprintf(stdOut, "\033[H\033[2J")
		}

		prefix := ""

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

		c.level++

		// set highest log level to turn off verbose
		if false == verboseBatch {
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
			fmt.Fprintf(stdOut, "> %s\n", batchFile.Name())
		} else if verboseBatch && c.level != 1 {
			prefix = fmt.Sprintf("%d-", c.level-1)
			fmt.Fprintf(stdOut, "\n<<<<<<< %s\n", batchFile.Name())
		}

		for i, line := range cmdLines {
			lineNum := i + 1

			cmd, args := context.ParseFirstWord(line)
			if len(cmd) == 0 {
				if verboseBatch {
					fmt.Fprintf(stdOut, "\x1B[0;37m%s%d\x1B[0m\n", prefix, lineNum)
				}
				continue
			} else if context.Comment == cmd {
				if verboseBatch {
					fmt.Fprintf(stdOut, "\x1B[0;37m%s%d \x1B[32m%s\x1B[0m\n", prefix, lineNum, line)
				}
				continue
			}
			if verboseBatch {
				fmt.Fprintf(stdOut, "\x1B[0;37m%s%d \x1B[34;1m%s \x1B[0m%s\n", prefix, lineNum, cmd, args)
			}

			Broker(line)

			if letBatchKnowErr != nil {
				// if there is error during execution, then print line for error trace
				fmt.Fprintf(stdOut, "\x1B[0;37m%s:%d \x1B[34;1m%s \x1B[0m%s\n\n", batchFile.Name(), lineNum, cmd, args)
				letBatchKnowErr = nil
			}
		}

		if c.level != 1 && verboseBatch {
			fmt.Fprintf(stdOut, ">>>>>>> %s\n", batchFile.Name())
		}

		c.level--

		// print final result and reset params
		if c.level == 0 {
			if batchErrorCount == 0 {
				fmt.Fprintf(stdOut, "\x1B[32;1mBatch is successfully finished\x1B[0m\n")
			} else {
				fmt.Fprintf(stdOut, "\x1B[31;1mBatch is failed: Error %d\x1B[0m\n", batchErrorCount)
			}
			// reset params
			batchErrorCount = 0
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		// add file to watch list
		if enableWatch {
			absPath, _ := filepath.Abs(batchFile.Name())
			watcher.Add(absPath)
		}

		if c.level == 0 && enableWatch {
			defer watcher.Close()
			// wait and check file changes
		fileWatching:
			for {
				select {
				case <-watcher.Events:
					break fileWatching
				case err, _ := <-watcher.Errors:
					if err != nil {
						fmt.Fprintf(stdOut, "\x1B[0;37mWatching File %s Error: %s\x1B[0m\n", batchFile.Name(), err.Error())
					}
					break fileWatching
				}
			}
			continue
		}
		break
	}

	return "batch exec is finished", nil
}
