package exec

import (
	"os"
)

func init() {
	registerExec(&exit{})
}

type exit struct{}

func (c *exit) Command() string {
	return "exit"
}

func (c *exit) Syntax() string {
	return ""
}

func (c *exit) Usage() string {
	return "exit"
}

func (c *exit) Describe() string {
	return "exit this application"
}

func (c *exit) Validate(args string) error {
	return nil
}

func (c *exit) Run(args string) (string, error) {

	os.Exit(0) // exit program
	return "", nil
}
