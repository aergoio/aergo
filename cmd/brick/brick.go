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
)

var logger = log.NewLogger("brick")

func completerBroker(d prompt.Document) []prompt.Suggest {
	var s []prompt.Suggest
	// parse first word. it is represent a command
	cmd, args := context.ParseFirstWord(d.Lines()[0])

	// find number of word before a current cursor location
	cusorWords := strings.Fields(d.TextBeforeCursor())
	cusorWordSeq := len(cusorWords)

	// if there is nothing typed text or it is a first word
	if cusorWordSeq == 0 || (cusorWordSeq == 1 && !strings.HasSuffix(d.TextBeforeCursor(), " ")) {
		// suggest all commands avaialbe
		for _, executor := range exec.AllExecutors() {
			s = append(s, prompt.Suggest{Text: executor.Command(), Description: executor.Describe()})
		}
	} else if cmd == context.Comment {
		// ignore a special char; comment
		s = append(s, prompt.Suggest{Text: "<comment>"})
	} else {
		// suggest consequence word using a syntax of command
		executor := exec.GetExecutor(cmd)
		if executor == nil {
			s = append(s, prompt.Suggest{Text: "<error>", Description: fmt.Sprintf("command not found: %s", cmd)})
		} else {
			syntax := executor.Syntax()
			syntaxSplit := strings.Fields(syntax)
			var symbol string

			// there exist a syntax
			if len(syntaxSplit) != 1 {
				// from the syntax, try to find a matched symbol of a current field
				if len(syntaxSplit) >= cusorWordSeq {
					takeNextSymbol := strings.HasSuffix(d.TextBeforeCursor(), " ")
					if takeNextSymbol && len(syntaxSplit) != cusorWordSeq {
						symbol = syntaxSplit[cusorWordSeq]
					} else if !takeNextSymbol {
						symbol = syntaxSplit[cusorWordSeq-1]
					}
				}

				// search from index using symbol
				for text, descr := range exec.Candidates(cmd, args, symbol) {
					s = append(s, prompt.Suggest{Text: text, Description: descr})
				}
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

		exec.Execute(cmd, args)
	}
}
