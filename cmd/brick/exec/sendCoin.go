package exec

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/aergoio/aergo/cmd/brick/context"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/types"
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

func (c *sendCoin) parse(args string) (string, string, *big.Int, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 3 {
		return "", "", nil, fmt.Errorf("need 3 arguments. usage: %s", c.Usage())
	}

	amountStr := context.ParseDecimalAmount(splitArgs[2].Text, 18)
	amount, success := new(big.Int).SetString(amountStr, 10)
	if success == false {
		return "", "", nil, fmt.Errorf("fail to parse number %s", splitArgs[2].Text)
	}

	return splitArgs[0].Text,
		splitArgs[1].Text,
		amount,
		nil
}

func (c *sendCoin) Run(args string) (string, uint64, []*types.Event, error) {
	senderName, receiverName, amount, _ := c.parse(args)

	// assuming target is contract
	var tx contract.LuaTxTester
	tx = contract.NewLuaTxCallBig(senderName, receiverName, amount, "")
	err := context.Get().ConnectBlock(tx)

	if err != nil && strings.HasPrefix(err.Error(), "not found contract") {
		// retry to normal address
		tx = contract.NewLuaTxSendBig(senderName, receiverName, amount)
		err := context.Get().ConnectBlock(tx)
		if err != nil {
			return "", 0, nil, err
		}
	} else if err != nil {
		return "", 0, nil, err
	}

	Index(context.AccountSymbol, receiverName)

	return "send aergo successfully",
		context.Get().GetReceipt(tx.Hash()).GasUsed,
		nil,
		nil
}
