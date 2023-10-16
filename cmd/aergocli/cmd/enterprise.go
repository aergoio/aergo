/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	"github.com/aergoio/aergo/v2/cmd/aergocli/util/encoding/json"
	"github.com/aergoio/aergo/v2/contract/enterprise"
	"github.com/aergoio/aergo/v2/types"
	aergorpc "github.com/aergoio/aergo/v2/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var (
	ccBlockNo uint64
	timeout   uint64

	ErrNotExecutedConfChange = errors.New("change cluster request may be not proposed")
)

func init() {
	rootCmd.AddCommand(enterpriseCmd)

	enterpriseTxCmd.Flags().Uint64VarP(&timeout, "timeout", "t", 30, "timeout(second) of geting status of enterprise transaction")

	enterpriseCmd.AddCommand(enterpriseKeyCmd)
	enterpriseCmd.AddCommand(enterpriseTxCmd)
}

var enterpriseCmd = &cobra.Command{
	Use:   "enterprise subcommand",
	Short: "Enterprise command",
}

type outConf struct {
	Key    string
	On     *bool
	Values []string
}

var enterpriseKeyCmd = &cobra.Command{
	Use:   "query (admins | <config key>)",
	Short: "Print config values of enterprise",
	Long:  "'permissions' show everything you can set as special config key",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		msg, err := client.GetEnterpriseConfig(context.Background(), &aergorpc.EnterpriseConfigKey{Key: args[0]})
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		var out outConf
		out.Key = msg.Key
		out.Values = msg.Values
		if strings.ToUpper(args[0]) != "PERMISSIONS" {
			out.On = &msg.On //it's for print false
		}
		cmd.Println(util.B58JSON(out))
	},
}

func getConfChangeBlockNo(blockHash []byte) (aergorpc.BlockNo, error) {
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
	Payload  string
	TxStatus *aergorpc.EnterpriseTxStatus
}

func (occ *OutConfChange) ToString() string {
	d, err := json.Marshal(occ)
	if err != nil {
		return "failed to marshaling ConfChangeProgressOutput"
	}
	ret := string(d)
	return ret
}

func isTimeouted(timer *time.Timer) bool {
	if timer == nil {
		return true
	}

	select {
	case <-timer.C:
		return true
	default:
		return false
	}
}

func getChangeClusterStatus(cmd *cobra.Command, blockHash []byte, timer *time.Timer) (*aergorpc.ConfChangeProgress, error) {
	var (
		err               error
		cycle             = time.Duration(3) * time.Second
		msgConfChangeProg *aergorpc.ConfChangeProgress
	)

	// get ccBlockNo
	if ccBlockNo, err = getConfChangeBlockNo(blockHash); err != nil {
		return nil, err
	}
	// get conf chagne status with reqid
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, ccBlockNo)

	for {
		if isTimeouted(timer) {
			break
		}

		msgConfChangeProg, err := client.GetConfChangeProgress(context.Background(), &aergorpc.SingleBytes{Value: b})
		if err != nil {
			continue
		}

		if msgConfChangeProg.State == aergorpc.ConfChangeState_CONF_CHANGE_STATE_APPLIED {
			return msgConfChangeProg, nil
		}

		time.Sleep(cycle)
	}

	if msgConfChangeProg == nil {
		return nil, ErrNotExecutedConfChange
	}
	//cmd.Printf(msgConfChangeProg.ToJsonString())
	return msgConfChangeProg, nil
}

var enterpriseTxCmd = &cobra.Command{
	Use:   "tx <tx hash>",
	Short: "Print transaction for enterprise",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		txHashDecode, err := base58.Decode(args[0])
		if err != nil {
			cmd.Println("Failed: invalid tx hash")
			return
		}

		var (
			tx         *aergorpc.Tx
			msgblock   *aergorpc.TxInBlock
			output     OutConfChange
			cycle      = time.Duration(3) * time.Second
			ci         types.CallInfo
			confChange *aergorpc.ConfChangeProgress
			timer      *time.Timer
		)

		if timeout > 0 {
			timer = time.NewTimer(time.Duration(timeout) * time.Second)
		}

		getTxTimeout := func() (*aergorpc.Tx, *aergorpc.TxInBlock, error) {
			for {
				tx, err = client.GetTX(context.Background(), &aergorpc.SingleBytes{Value: txHashDecode})
				if err == nil {
					// tx is not excuted yet
					if isTimeouted(timer) {
						return tx, nil, nil
					}
					time.Sleep(cycle)
					continue
				}

				if err != nil {
					msgblock, err = client.GetBlockTX(context.Background(), &aergorpc.SingleBytes{Value: txHashDecode})
					if err != nil {
						return nil, nil, fmt.Errorf("failed to get tx from block (err=%s)", err.Error())

					}
					tx = msgblock.Tx

					return tx, msgblock, nil
				}

			}
		}

		if tx, msgblock, err = getTxTimeout(); err != nil {
			cmd.Printf("Error: %s", err.Error())
			return
		}

		if tx != nil {
			output.Payload = string(tx.GetBody().Payload)

			if err := json.Unmarshal(tx.GetBody().Payload, &ci); err != nil {
				cmd.Printf("Error: tx payload is not json: %s", err.Error())
				return
			}
		}

		if msgblock != nil {
			receipt, err := client.GetReceipt(context.Background(), &aergorpc.SingleBytes{Value: txHashDecode})
			if err != nil {
				cmd.Println("Error: failed to get receipt")
				cmd.Println(output.ToString())
				return
			}
			output.TxStatus = &aergorpc.EnterpriseTxStatus{
				Status: receipt.GetStatus(),
				Ret:    receipt.GetRet(),
			}

			if ci.Name == enterprise.ChangeCluster {
				if confChange, err = getChangeClusterStatus(cmd, msgblock.TxIdx.BlockHash, timer); err != nil {
					output.TxStatus.CCStatus = &types.ChangeClusterStatus{Error: err.Error()}
				}

				if confChange != nil {
					output.TxStatus.CCStatus = confChange.ToPrintable()
				}
			}

			cmd.Println(output.ToString())
		}

		return
	},
}
