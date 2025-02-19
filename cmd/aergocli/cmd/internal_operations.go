/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	aergorpc "github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
)

func init() {
	operationsCmd := &cobra.Command{
		Use:   "operations tx_hash",
		Short: "Get internal operations for a transaction",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Decode the transaction hash
			txHash, err := base58.Decode(args[0])
			if err != nil {
				log.Fatal(err)
			}

			// Retrieve the receipt to get the block height
			receipt, err := client.GetReceipt(context.Background(), &aergorpc.SingleBytes{Value: txHash})
			if err != nil {
				log.Fatal(err)
			}

			// Extract block height from the receipt
			blockHeight := receipt.BlockNo

			// Use block height to get internal operations
			msg, err := client.GetInternalOperations(context.Background(), &aergorpc.BlockNumberParam{BlockNo: blockHeight})
			if err != nil {
				log.Fatal(err)
			}

			// Extract the internal operations for the specific transaction
			var operations []map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Value), &operations); err != nil {
				log.Fatal(err)
			}

			// Print the internal operations for the specific transaction
			var ops string
			for _, op := range operations {
				if op["txhash"] == args[0] {
					// Remove the txhash field
					delete(op, "txhash")
					// Marshal the operation to JSON for better readability
					opJSON, err := json.MarshalIndent(op, "", "  ")
					if err != nil {
						log.Fatal(err)
					}
					ops = string(opJSON)
				}
			}
			if ops == "" {
				cmd.Println("No internal operations found for this transaction.")
			} else {
				cmd.Println(ops)
			}
		},
	}
	rootCmd.AddCommand(operationsCmd)
}
