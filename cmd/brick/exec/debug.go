//go:build Debug
// +build Debug

package exec

import (
	"fmt"
	"strconv"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/contract"
	"github.com/aergoio/aergo/v2/types"
)

func init() {
	registerExec(&setb{})
	registerExec(&delb{})
	registerExec(&listb{})
	registerExec(&resetb{})
	registerExec(&setw{})
	registerExec(&delw{})
	registerExec(&listw{})
	registerExec(&resetw{})
}

// =====================================
//             Breakpoint
// =====================================

// =========== setb ==============
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

func (c *setb) Run(args string) (string, uint64, []*types.Event, error) {
	line, contractIDHex, _ := c.parse(args)

	err := contract.SetBreakPoint(contractIDHex, line)
	if err != nil {
		return "", 0, nil, err
	}
	addr, err := contract.HexAddrToBase58Addr(contractIDHex)
	if err != nil {
		return "", 0, nil, err
	}

	return "set breakpoint: " + fmt.Sprintf("%s:%d", addr, line), 0, nil, nil
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

func (c *delb) Run(args string) (string, uint64, []*types.Event, error) {
	line, contractIDHex, _ := c.parse(args)

	err := contract.DelBreakPoint(contractIDHex, line)
	if err != nil {
		return "", 0, nil, err
	}
	addr, err := contract.HexAddrToBase58Addr(contractIDHex)
	if err != nil {
		return "", 0, nil, err
	}

	return "del breakpoint: " + fmt.Sprintf("%s:%d", addr, line), 0, nil, nil
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

func (c *listb) Run(args string) (string, uint64, []*types.Event, error) {
	contract.PrintBreakPoints()

	return "list breakpoints", 0, nil, nil
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

func (c *resetb) Run(args string) (string, uint64, []*types.Event, error) {
	contract.ResetBreakPoints()

	return "reset breakpoints", 0, nil, nil
}

// =====================================
//             Watchpoint
// =====================================

// =========== setw ==============
type setw struct{}

func (c *setw) Command() string {
	return "setw"
}

func (c *setw) Syntax() string {
	return fmt.Sprintf("%s", "<watch_expr>")
}

func (c *setw) Usage() string {
	return "setw `<watch_expr>`"
}

func (c *setw) Describe() string {
	return "set watchpoint"
}

func (c *setw) Validate(args string) error {

	_, err := c.parse(args)

	return err
}

func (c *setw) parse(args string) (string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 1 {
		return "", fmt.Errorf("need an arguments. usage: %s", c.Usage())
	}

	return splitArgs[0].Text, nil
}

func (c *setw) Run(args string) (string, uint64, []*types.Event, error) {
	watch_expr, _ := c.parse(args)

	err := contract.SetWatchPoint(watch_expr)
	if err != nil {
		return "", 0, nil, err
	}

	return "set watchpoint: " + watch_expr, 0, nil, nil
}

// =========== delw ==============

type delw struct{}

func (c *delw) Command() string {
	return "delw"
}

func (c *delw) Syntax() string {
	return fmt.Sprintf("%s", "<index>")
}

func (c *delw) Usage() string {
	return "delw <index>"
}

func (c *delw) Describe() string {
	return "delete watchpoint"
}

func (c *delw) Validate(args string) error {

	_, err := c.parse(args)

	return err
}

func (c *delw) parse(args string) (uint64, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 1 {
		return 0, fmt.Errorf("need an arguments. usage: %s", c.Usage())
	}

	idx, err := strconv.ParseUint(splitArgs[0].Text, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("fail to parse number %s: %s", splitArgs[0].Text, err.Error())
	}

	return idx, nil
}

func (c *delw) Run(args string) (string, uint64, []*types.Event, error) {
	idx, _ := c.parse(args)

	err := contract.DelWatchPoint(idx)
	if err != nil {
		return "", 0, nil, err
	}

	return "del watchpoint: " + fmt.Sprintf("%d", idx), 0, nil, nil
}

// =========== listw ==============

type listw struct{}

func (c *listw) Command() string {
	return "listw"
}

func (c *listw) Syntax() string {
	return ""
}

func (c *listw) Usage() string {
	return "listw"
}

func (c *listw) Describe() string {
	return "list all watchpoints"
}

func (c *listw) Validate(args string) error {
	return nil
}

func (c *listw) Run(args string) (string, uint64, []*types.Event, error) {
	watchpoints := contract.ListWatchPoints()
	i := 0
	for e := watchpoints.Front(); e != nil; e = e.Next() {
		i++
		fmt.Printf("%d: %s\n", i, e.Value)
	}

	return "list watchpoints", 0, nil, nil
}

// =========== resetb ==============

type resetw struct{}

func (c *resetw) Command() string {
	return "resetw"
}

func (c *resetw) Syntax() string {
	return ""
}

func (c *resetw) Usage() string {
	return "resetw"
}

func (c *resetw) Describe() string {
	return "reset all watchpoints"
}

func (c *resetw) Validate(args string) error {
	return nil
}

func (c *resetw) Run(args string) (string, uint64, []*types.Event, error) {
	contract.ResetWatchPoints()

	return "reset watchpoints", 0, nil, nil
}

// =====================================
//             interfaces
// =====================================

func resetContractInfoInterface() {
	contract.ResetContractInfo()
}

func updateContractInfoInterface(contractName string, defPath string) {
	contract.UpdateContractInfo(
		contract.PlainStrToHexAddr(contractName), defPath)
}
