// debugger

// set bp
// remove bp
// list bp
//

package exec

import (
	"fmt"
	"strconv"

	"github.com/aergoio/aergo/cmd/brick/context"
	"github.com/aergoio/aergo/contract"
)

func init() {
	registerExec(&setb{})
	registerExec(&delb{})
	registerExec(&listb{})
	registerExec(&resetb{})
}

type setb struct{}

func (c *setb) Command() string {
	return "setb"
}

func (c *setb) Syntax() string {
	return fmt.Sprintf("%s %s", "<line>", context.ContractSymbol)
}

func (c *setb) Usage() string {
	return "setb <line> <contract_name>"
}

func (c *setb) Describe() string {
	return "set breakpoint"
}

func (c *setb) Validate(args string) error {

	_, _, err := c.parse(args)

	return err
}

func (c *setb) parse(args string) (uint64, string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 2 {
		return 0, "", fmt.Errorf("need 2 arguments. usage: %s", c.Usage())
	}

	line, err := strconv.ParseUint(splitArgs[0].Text, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("fail to parse number %s: %s", splitArgs[1].Text, err.Error())
	}

	contractIDHex := contract.PlainStrToHexAddr(splitArgs[1].Text)

	return line, contractIDHex, nil
}

func (c *setb) Run(args string) (string, error) {
	line, contractIDHex, _ := c.parse(args)

	err := contract.SetBreakPoint(contractIDHex, line)
	if err != nil {
		return "", err
	}
	addr, err := contract.HexAddrToBase58Addr(contractIDHex)
	if err != nil {
		return "", err
	}

	return "set breakpoint: " + fmt.Sprintf("%s:%d", addr, line), nil
}

// =========== delb ==============

type delb struct{}

func (c *delb) Command() string {
	return "delb"
}

func (c *delb) Syntax() string {
	return fmt.Sprintf("%s %s", "<line>", context.ContractSymbol)
}

func (c *delb) Usage() string {
	return "delb <line> <contract_name>"
}

func (c *delb) Describe() string {
	return "delete breakpoint"
}

func (c *delb) Validate(args string) error {

	_, _, err := c.parse(args)

	return err
}

func (c *delb) parse(args string) (uint64, string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 2 {
		return 0, "", fmt.Errorf("need 2 arguments. usage: %s", c.Usage())
	}

	line, err := strconv.ParseUint(splitArgs[0].Text, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("fail to parse number %s: %s", splitArgs[1].Text, err.Error())
	}

	contractIDHex := contract.PlainStrToHexAddr(splitArgs[1].Text)

	return line, contractIDHex, nil
}

func (c *delb) Run(args string) (string, error) {
	line, contractIDHex, _ := c.parse(args)

	err := contract.DelBreakPoint(contractIDHex, line)
	if err != nil {
		return "", err
	}
	addr, err := contract.HexAddrToBase58Addr(contractIDHex)
	if err != nil {
		return "", err
	}

	return "del breakpoint: " + fmt.Sprintf("%s:%d", addr, line), nil
}

// =========== listb ==============

type listb struct{}

func (c *listb) Command() string {
	return "listb"
}

func (c *listb) Syntax() string {
	return ""
}

func (c *listb) Usage() string {
	return "listb"
}

func (c *listb) Describe() string {
	return "list all breakpoints"
}

func (c *listb) Validate(args string) error {
	return nil
}

func (c *listb) Run(args string) (string, error) {
	contract.PrintBreakPoints()

	return "list breakpoints", nil
}

// =========== resetb ==============

type resetb struct{}

func (c *resetb) Command() string {
	return "resetb"
}

func (c *resetb) Syntax() string {
	return ""
}

func (c *resetb) Usage() string {
	return "resetb"
}

func (c *resetb) Describe() string {
	return "reset all breakpoints"
}

func (c *resetb) Validate(args string) error {
	return nil
}

func (c *resetb) Run(args string) (string, error) {
	contract.ResetBreakPoints()

	return "reset breakpoints", nil
}
