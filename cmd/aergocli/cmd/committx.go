/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"errors"
	"io/ioutil"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var committxCmd = &cobra.Command{
	Use:   "committx",
	Short: "commit transaction to aergo server",
	Args:  cobra.MinimumNArgs(0),
	RunE:  execCommitTX,
}

var recipient string

var jsonTx string
var jsonPath string

func init() {
	rootCmd.AddCommand(committxCmd)

	committxCmd.Flags().StringVar(&jsonTx, "jsontx", "", "Transaction list json\n"+"Tx(Transaction) json example is\n"+`{
 "Hash": "Base58EncodedBytes",
 "Body": {
  "Nonce": 1,
  "Account": "AmLfhA2F82Nayuek17tvzechaQPe5cRQKBBJ8xfei7GejvufVRBp",
  "Recipient": "Amgf9vfcHKkC1ijGTMjxLoeTTutXgbaHHBznpHu5ugutU96iKSLW",
  "Amount": 0,
  "Payload": "Base58EncodedBytes",
  "Limit": 0,
  "Price": 0,
  "Type": 0,
  "Sign": "Base58EncodedBytes"
 }
}`)
	committxCmd.Flags().StringVar(&jsonPath, "jsontxpath", "", "Transaction list json file path")
}

func execCommitTX(cmd *cobra.Command, args []string) error {
	if jsonPath != "" {
		b, readerr := ioutil.ReadFile(jsonPath)
		if readerr != nil {
			return errors.New("Failed to read --jsontxpath\n" + readerr.Error())
		}
		jsonTx = string(b)
	}

	if jsonTx != "" {
		var msg *types.CommitResultList
		txlist, err := util.ParseBase58Tx([]byte(jsonTx))
		if err != nil {
			return errors.New("Failed to parse --jsontx\n" + err.Error())
		}
		msg, err = client.CommitTX(context.Background(), &types.TxList{Txs: txlist})
		if err != nil {
			return errors.New("Failed request to aergo server\n" + err.Error())
		}
		for i, r := range msg.Results {
			cmd.Println(i+1, ":", base58.Encode(r.Hash), r.Error, r.Detail)
		}
	}
	return nil
}
