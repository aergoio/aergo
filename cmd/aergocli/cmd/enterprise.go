/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/aergoio/aergo/cmd/aergocli/util/encoding/json"
	"github.com/mr-tron/base58/base58"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	aergorpc "github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var (
	enterpriseKey string

	txHash    string
	requestID uint64
)

func init() {
	rootCmd.AddCommand(enterpriseCmd)

	enterpriseCmd.Flags().StringVar(&enterpriseKey, "key", "all", "query the state of enterprise config by key")

	clusterCmd.Flags().StringVarP(&txHash, "tx", "t", "", "hash of changeCluster enterprise transaction")
	clusterCmd.Flags().Uint64VarP(&requestID, "reqid", "r", 0, "requestID of changeCluster enterprise transaction")

	enterpriseCmd.AddCommand(clusterCmd)
}

var enterpriseCmd = &cobra.Command{
	Use:   "enterprise",
	Short: "Print current enterprise status",
	Run: func(cmd *cobra.Command, args []string) {
		msg, err := client.GetEnterpriseConfig(context.Background(), &aergorpc.EnterpriseConfigKey{Key: enterpriseKey})
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		cmd.Println(util.JSON(msg))
	},
}

/*
func parseClusterChangePayload(tx *aergorpc.Tx) string {
	var (
		ci aergorpc.CallInfo
	)


	if err := json.Unmarshal(tx.GetBody().Payload, &ci); err != nil {
		return fmt.Sprintf("failed parse payload: %s", err.Error())
	}

	cc, err := aergorpc.ValidateChangeCluster(ci, 0)
	if err != nil {
		return fmt.Sprintf("failed parse payload: %s", err.Error())
	}

	return cc.(*aergorpc.MembershipChange).ToString()
}
*/

func getRequestID(blockHash []byte) (aergorpc.BlockNo, error) {
	if len(blockHash) == 0 {
		return 0, fmt.Errorf("failed to get block since blockhash is empty")
	}

	block, err := client.GetBlock(context.Background(), &aergorpc.SingleBytes{Value: blockHash})
	if err != nil {
		return 0, err
	}

	return block.BlockNo(), nil
}

type OutConfChange struct {
	Payload string
	Status  string
}

func (occ *OutConfChange) ToString() string {
	d, err := json.Marshal(occ)
	if err != nil {
		return "failed to marshaling ConfChangeProgressOutput"
	}
	ret := string(d)
	return ret
}

var clusterCmd = &cobra.Command{
	Use:   "cluster [flags]",
	Short: "Print status of change cluster transaction. This command can only be used for raft consensus.",
	Run: func(cmd *cobra.Command, args []string) {
		fflags := cmd.Flags()
		if fflags.Changed("tx") == false && fflags.Changed("reqid") == false {
			cmd.Println("no cluster --tx or --reqid specified")
			return
		}

		var (
			tx       *aergorpc.Tx
			msgblock *aergorpc.TxInBlock
			output   OutConfChange
		)

		// txHash -> getTx -> get BlockNo of tx -> reqid = blockNo
		if len(txHash) != 0 {
			txHashDecode, err := base58.Decode(txHash)
			if err != nil {
				cmd.Printf("Failed: invalid tx hash")
				return
			}
			tx, err = client.GetTX(context.Background(), &aergorpc.SingleBytes{Value: txHashDecode})
			if err == nil {
				cmd.Println("Failed: Tx doesn't executed yet")
				return
			} else {
				msgblock, err = client.GetBlockTX(context.Background(), &aergorpc.SingleBytes{Value: txHashDecode})
				if err != nil {
					cmd.Printf("Failed: to get block including tx %s", err.Error())
					return
				}

				tx = msgblock.Tx

				// get requestID
				if requestID, err = getRequestID(msgblock.TxIdx.BlockHash); err != nil {
					cmd.Printf("Failed to get request ID: %s", err.Error())
				}
			}

			// print payload
			//cmd.Printf(string(tx.GetBody().Payload))
			output.Payload = string(tx.GetBody().Payload)
		}

		// get conf chagne status with reqid
		if requestID != 0 {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, requestID)

			msgConfChangeProg, err := client.GetConfChangeProgress(context.Background(), &aergorpc.SingleBytes{Value: b})
			if err != nil {
				cmd.Printf("Failed to get progress: reqid=%d, %s", requestID, err.Error())
			}
			//cmd.Printf(msgConfChangeProg.ToJsonString())
			output.Status = msgConfChangeProg.ToJsonString()
		}

		cmd.Printf(output.ToString())
		return
	},
}
