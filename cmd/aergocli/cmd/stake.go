/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"errors"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var stakeCmd = &cobra.Command{
	Use:   "stake",
	Short: "Stake balance to aergo system",
	RunE:  execStake,
}

func execStake(cmd *cobra.Command, args []string) error {
	return sendStake(cmd, true)
}

var unstakeCmd = &cobra.Command{
	Use:   "unstake",
	Short: "Unstake balance from aergo system",
	RunE:  execUnstake,
}

func execUnstake(cmd *cobra.Command, args []string) error {
	return sendStake(cmd, false)
}

func sendStake(cmd *cobra.Command, s bool) error {
	account, err := types.DecodeAddress(address)
	if err != nil {
		return errors.New("Failed to parse --address flag (" + address + ")\n" + err.Error())
	}
	payload := make([]byte, 1)
	if s {
		payload[0] = 's'
	} else {
		payload[0] = 'u'
	}
	amountBigInt, err := util.ParseUnit(amount)
	if err != nil {
		return errors.New("Failed to parse --amount flag\n" + err.Error())
	}
	if amountBigInt.Cmp(types.StakingMinimum) < 0 {
		return errors.New("Failed: minimum stake value is " + types.StakingMinimum.String())
	}

	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   account,
			Recipient: []byte(types.AergoSystem),
			Amount:    amountBigInt.Bytes(),
			Payload:   payload,
			Limit:     0,
			Type:      types.TxType_GOVERNANCE,
		},
	}
	msg, err := client.SendTX(context.Background(), tx)
	if err != nil {
		return err
	}
	cmd.Println(base58.Encode(msg.Hash), msg.Error, msg.Detail)
	return nil
}
