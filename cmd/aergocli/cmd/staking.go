/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"fmt"

	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var stakingCmd = &cobra.Command{
	Use:   "staking",
	Short: "",
	Run:   execStaking,
}

func execStaking(cmd *cobra.Command, args []string) {
	sendStaking(true)
}

var unstakingCmd = &cobra.Command{
	Use:   "unstaking",
	Short: "",
	Run:   execUnstaking,
}

func execUnstaking(cmd *cobra.Command, args []string) {
	sendStaking(false)
}

func sendStaking(s bool) {
	account, err := types.DecodeAddress(address)
	if err != nil {
		fmt.Printf("Failed: (%s) %s\n", address, err.Error())
		return
	}
	payload := make([]byte, 1)
	if s {
		payload[0] = 's'
	} else {
		payload[0] = 'u'
	}
	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   account,
			Recipient: []byte(aergosystem),
			Amount:    amount,
			Price:     0,
			Payload:   payload,
			Limit:     0,
			Type:      types.TxType_GOVERNANCE,
		},
	}
	msg, err := client.SendTX(context.Background(), tx)
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return
	}
	fmt.Println(base58.Encode(msg.Hash), msg.Error)
}
