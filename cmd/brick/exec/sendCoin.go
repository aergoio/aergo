package exec

import (
	"fmt"
	"math/big"
	"strconv"
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

	_, _, _, _, err := c.parse(args)

	return err
}

func (c *sendCoin) parse(args string) (int32, string, string, *big.Int, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 4 {
		return 0, "", "", nil, fmt.Errorf("need 4 arguments. usage: %s", c.Usage())
	}

	version, err := strconv.ParseInt(splitArgs[0].Text, 10, 32)
	if err != nil {
		return 0, "", "", nil, fmt.Errorf("fail to parse version %s", splitArgs[0].Text)
	}

	amount, success := new(big.Int).SetString(splitArgs[3].Text, 10)
	if success == false {
		return 0, "", "", nil, fmt.Errorf("fail to parse number %s", splitArgs[3].Text)
	}

	return int32(version),
		splitArgs[1].Text,
		splitArgs[2].Text,
		amount,
		nil
}

func (c *sendCoin) Run(args string) (string, uint64, []*types.Event, error) {
	version, senderName, receiverName, amount, _ := c.parse(args)

	// assuming target is contract
	var tx contract.LuaTxTester
	tx = contract.NewLuaTxCallBig(senderName, receiverName, amount, "")
	err := context.Get().ConnectBlock(version, tx)

	if err != nil && strings.HasPrefix(err.Error(), "not found contract") {
		// retry to normal address
		tx = contract.NewLuaTxSendBig(senderName, receiverName, amount)
		err := context.Get().ConnectBlock(version, tx)
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
