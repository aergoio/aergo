/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"

	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var stakingCmd = &cobra.Command{
	Use:   "staking",
	Short: "Staking balance to aergo system",
	Run:   execStaking,
}

func execStaking(cmd *cobra.Command, args []string) {
	sendStaking(cmd, true)
}

var unstakingCmd = &cobra.Command{
	Use:   "unstaking",
	Short: "Unstaking balance from aergo system",
	Run:   execUnstaking,
}

func execUnstaking(cmd *cobra.Command, args []string) {
	sendStaking(cmd, false)
}

func sendStaking(cmd *cobra.Command, s bool) {
	account, err := types.DecodeAddress(address)
	if err != nil {
		cmd.Printf("Failed: (%s) %s\n", address, err.Error())
		return
	}
	payload := make([]byte, 1)
	if s {
		payload[0] = 's'
	} else {
		payload[0] = 'u'
	}
	if amount < types.StakingMinimum {
		cmd.Printf("Failed: minimum staking value is %d\n", types.StakingMinimum)
		return
	}

	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   account,
			Recipient: []byte(types.AergoSystem),
			Amount:    amount,
			Price:     0,
			Payload:   payload,
			Limit:     0,
			Type:      types.TxType_GOVERNANCE,
		},
	}
	msg, err := client.SendTX(context.Background(), tx)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	cmd.Println(base58.Encode(msg.Hash), msg.Error)
}
