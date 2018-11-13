package exec

import (
	"fmt"
	"strconv"

	"github.com/aergoio/aergo/cmd/brick/context"
	"github.com/aergoio/aergo/contract"
)

func init() {
	registerExec(&sendCoin{})
}

type sendCoin struct{}

func (c *sendCoin) Command() string {
	return "send"
}

func (c *sendCoin) Syntax() string {
	return fmt.Sprintf("%s %s %s", context.AccountSymbol,
		context.AccountSymbol, context.AmountSymbol)
}

func (c *sendCoin) Usage() string {
	return "send <sender_name> <receiver_name> <amount>"
}

func (c *sendCoin) Describe() string {
	return "send aergo from sender to receiver"
}

func (c *sendCoin) Validate(args string) error {

	// is chain is loaded?
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, _, _, err := c.parse(args)

	return err
}

func (c *sendCoin) parse(args string) (string, string, uint64, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 3 {
		return "", "", 0, fmt.Errorf("need 3 arguments. usage: %s", c.Usage())
	}

	amount, err := strconv.ParseUint(splitArgs[2], 10, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("fail to parse number %s: %s", splitArgs[1], err.Error())
	}

	return splitArgs[0],
		splitArgs[1],
		amount,
		nil
}

func (c *sendCoin) Run(args string) (string, error) {
	senderName, receiverName, amount, _ := c.parse(args)

	err := context.Get().ConnectBlock(
		contract.NewLuaTxSend(senderName, receiverName, amount),
	)

	if err != nil {
		return "", err
	}

	Index(context.AccountSymbol, receiverName)

	return "send aergo successfully", nil
}
