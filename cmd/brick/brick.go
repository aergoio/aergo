package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/cmd/brick/context"
	"github.com/aergoio/aergo/cmd/brick/exec"
	prompt "github.com/c-bata/go-prompt"
	"github.com/mattn/go-colorable"
)

var logger = log.NewLogger("brick")

func completerBroker(d prompt.Document) []prompt.Suggest {
	var s []prompt.Suggest
	// parse first word. it is represent a command
	cmd, args := context.ParseFirstWord(d.Lines()[0])

	// find number of word before a current cursor location
	chunks := context.SplitSpaceAndAccent(args, true)
	chunkNum := len(chunks)

	// if there is nothing typed text or it is a first word
	if cmd == "" || (cmd != "" && chunkNum == 0 && !strings.HasSuffix(d.TextBeforeCursor(), " ")) {
		// suggest all commands avaialbe
		for _, executor := range exec.AllExecutors() {
			s = append(s, prompt.Suggest{Text: executor.Command(), Description: executor.Describe()})
		}

		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	} else if cmd == context.Comment {
		// ignore a special char; comment
		s = append(s, prompt.Suggest{Text: "<comment>"})

		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}

	// suggest consequence word using a syntax of command
	executor := exec.GetExecutor(cmd)
	if executor == nil {
		s = append(s, prompt.Suggest{Text: "<error>", Description: fmt.Sprintf("command not found: %s", cmd)})
	} else {
		syntax := executor.Syntax()
		syntaxSplit := strings.Fields(syntax)
		syntaxNum := len(syntaxSplit)
		var symbol string
		var current int

		// there exist a syntax
		if syntaxNum != 0 {
			// from the syntax, try to find a matched symbol of a current field
			if syntaxNum >= chunkNum {
				// when last char is a space, then skip to next symbol
				takeNextSymbol := strings.HasSuffix(d.TextBeforeCursor(), " ")
				if takeNextSymbol && syntaxNum != chunkNum {
					current = chunkNum
					symbol = syntaxSplit[current]
				} else if !takeNextSymbol {
					current = chunkNum - 1
					symbol = syntaxSplit[current]
				}
			}

			// search from index using symbol
			for text, descr := range exec.Candidates(cmd, chunks, current, symbol) {
				s = append(s, prompt.Suggest{Text: text, Description: descr})
			}
		}
	}

	// sort suggestions by text in ascending order
	sort.Slice(s, func(i, j int) bool {
		return strings.Compare(s[i].Text, s[j].Text) < 0
	})

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func main() {
	if len(os.Args) <= 1 {
		// cli mode
		p := prompt.New(
			exec.Broker,
			completerBroker,
			prompt.OptionPrefix(">"),
			prompt.OptionLivePrefix(context.LivePrefix),
			prompt.OptionTitle("Aergo Brick: Dummy Virtual Machine"),
		)
		p.Run()
	} else {
		// call batch executor
		cmd := "batch"
		args := os.Args[1]

		// set user-defined log level
		if len(os.Args) > 2 {
			if os.Args[2] == "-v" {
				exec.EnableVerbose()
			} else {
				logger.Fatal().Msg("Invalid Parameter. Usage: brick filename [-v]")
			}
		}

		exec.Execute(cmd, args)

		stdOut := colorable.NewColorableStdout()
		errCnt := exec.GetBatchErrorCount()
		if errCnt == 0 {
			fmt.Fprintf(stdOut, "\x1B[32;1mBatch is successfully finished\n")
		} else {
			fmt.Fprintf(stdOut, "\x1B[31;1mBatch is failed: Error %d\n", errCnt)
		}
	}
}
