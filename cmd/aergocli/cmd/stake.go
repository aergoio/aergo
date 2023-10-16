/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"encoding/json"
	"errors"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	"github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
)

var stakeCmd = &cobra.Command{
	Use:    "stake",
	Short:  "Stake balance to aergo system",
	RunE:   execStake,
	PreRun: connectAergo,
}

func execStake(cmd *cobra.Command, args []string) error {
	return sendStake(cmd, true)
}

var unstakeCmd = &cobra.Command{
	Use:    "unstake",
	Short:  "Unstake balance from aergo system",
	RunE:   execUnstake,
	PreRun: connectAergo,
}

func execUnstake(cmd *cobra.Command, args []string) error {
	return sendStake(cmd, false)
}

func sendStake(cmd *cobra.Command, s bool) error {
	account, err := types.DecodeAddress(address)
	if err != nil {
		return errors.New("Failed to parse --address flag (" + address + ")\n" + err.Error())
	}
	var ci types.CallInfo
	if s {
		ci.Name = types.Opstake.Cmd()
	} else {
		ci.Name = types.Opunstake.Cmd()
	}
	amountBigInt, err := util.ParseUnit(amount)
	if err != nil {
		return errors.New("Failed to parse --amount flag\n" + err.Error())
	}
	payload, err := json.Marshal(ci)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return nil
	}
	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   account,
			Recipient: []byte(types.AergoSystem),
			Amount:    amountBigInt.Bytes(),
			Payload:   payload,
			GasLimit:  0,
			Type:      types.TxType_GOVERNANCE,
		},
	}

	cmd.Println(sendTX(cmd, tx, account))
	return nil
}
