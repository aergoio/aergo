package exec

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/types"
	"github.com/fsnotify/fsnotify"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
)

var (
	verboseBatch        = false
	letBatchKnowErr     error
	batchErrorCount     = 0
	lastBatchErrorCount = 0
	enableWatch         = false
	watchFileList       []string
	watcher             *fsnotify.Watcher
)

func EnableVerbose() {
	verboseBatch = true
}

func GetBatchErrorCount() int {
	return lastBatchErrorCount
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

func (c *batch) readBatchFile(batchFilePath string) ([]string, error) {
	if strings.HasPrefix(batchFilePath, "http") {
		// search in the web
		req, err := http.NewRequest("GET", batchFilePath, nil)
		if err != nil {
			return nil, err
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var cmdLines []string
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			cmdLines = append(cmdLines, scanner.Text())
		}

		return cmdLines, nil
	}

	batchFile, err := os.Open(batchFilePath)
	if err != nil {
		return nil, err
	}
	defer batchFile.Close()

	var commands []string
	var command string
	var line_no int = 0
	var isOpen bool = false
	scanner := bufio.NewScanner(batchFile)
	for scanner.Scan() {
		line := scanner.Text()
		line_no += 1
		command += line
		if len(line) > 0 && line[0:1] != "#" {
			isOpen, err = context.IsCompleteCommand(line, line_no, isOpen)
			if err != nil {
				return nil, err
			}
		}
		if !isOpen {
			commands = append(commands, command)
			command = ""
		}
	}

	return commands, nil
}

func (c *batch) parse(args string) (string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) != 1 {
		return "", fmt.Errorf("invalid format. usage: %s", c.Usage())
	}

	batchFilePath := splitArgs[0].Text

	if _, err := c.readBatchFile(batchFilePath); err != nil {
		return "", fmt.Errorf("fail to read a brick batch file %s: %s", batchFilePath, err.Error())
	}

	return batchFilePath, nil
}

func (c *batch) Run(args string) (string, uint64, []*types.Event, error) {
	stdOut := colorable.NewColorableStdout()

	var err error

	for {
		if c.level == 0 && enableWatch {
			watcher, err = fsnotify.NewWatcher()
			if err != nil {
				return "", 0, nil, err
			}
			// clear screen
			fmt.Fprintf(stdOut, "\033[H\033[2J")
		}

		prefix := ""

		batchFilePath, _ := c.parse(args)
		cmdLines, err := c.readBatchFile(batchFilePath)
		if err != nil {
			return "", 0, nil, err
		}

		c.level++

		// set highest log level to turn off verbose
		if false == verboseBatch {
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
			fmt.Fprintf(stdOut, "> %s\n", batchFilePath)
		} else if verboseBatch && c.level != 1 {
			prefix = fmt.Sprintf("%d-", c.level-1)
			fmt.Fprintf(stdOut, "\n<<<<<<< %s\n", batchFilePath)
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
				fmt.Fprintf(stdOut, "\x1B[0;37m%s:%d \x1B[34;1m%s \x1B[0m%s\n\n", batchFilePath, lineNum, cmd, args)
				letBatchKnowErr = nil
			}
		}

		if c.level != 1 && verboseBatch {
			fmt.Fprintf(stdOut, ">>>>>>> %s\n", batchFilePath)
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
			lastBatchErrorCount = batchErrorCount
			batchErrorCount = 0
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		// add file to watch list
		if enableWatch && !strings.HasPrefix(batchFilePath, "http") {
			absPath, _ := filepath.Abs(batchFilePath)
			watcher.Add(absPath)
		}

		if c.level == 0 && enableWatch {
			// wait and check file changes
		fileWatching:
			for {
				select {
				case <-watcher.Events:
					break fileWatching
				case err, _ := <-watcher.Errors:
					if err != nil {
						fmt.Fprintf(stdOut, "\x1B[0;37mWatching File %s Error: %s\x1B[0m\n", batchFilePath, err.Error())
					}
					break fileWatching
				}
			}
			watcher.Close()
			continue
		}
		break
	}

	return "batch exec is finished", 0, nil, nil
}
