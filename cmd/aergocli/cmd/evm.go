/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var evmCmd = &cobra.Command{
	Use:   "evm",
	Short: "Invoke EVM",
	Args:  cobra.MinimumNArgs(0),
	RunE:  execEVM,
}

func init() {
	rootCmd.AddCommand(evmCmd)
	evmCmd.PersistentFlags().StringVar(&data, "payload", "", "result of compiling a contract")
	evmCmd.Flags().StringVar(&from, "from", "", "Sender account address")
	evmCmd.MarkFlagRequired("from")
	evmCmd.Flags().Uint64Var(&nonce, "nonce", 0, "setting nonce manually")
	evmCmd.Flags().Uint64VarP(&gas, "gaslimit", "g", 0, "Gas limit")
	evmCmd.Flags().StringVar(&pw, "password", "", "Password")
}

func execEVM(cmd *cobra.Command, args []string) error {
	account, err := types.DecodeAddress(from)
	if err != nil {
		return errors.New("Wrong address in --from flag\n" + err.Error())
	}

	var payload []byte
	// process payload
	// FIXME: use hex encoded bytecode to follow ethereum convention for now
	payload, err = hex.DecodeString(data)
	if err != nil {
		return errors.New("failed to parse payload")
	}

	tx := &types.Tx{Body: &types.TxBody{
		Type:     types.TxType_EVM,
		Account:  account,
		Payload:  payload,
		Nonce:    nonce,
		GasLimit: gas,
	}}

	cmd.Println(sendEVMTX(cmd, tx, account))
	return nil
}

func sendEVMTX(cmd *cobra.Command, tx *types.Tx, account []byte) string {
	if rootConfig.KeyStorePath != "" {
		var err error
		if pw == "" {
			pw, err = getPasswd(cmd, false)
			if err != nil {
				return "Failed get password:" + err.Error()
			}
		}
		if tx.GetBody().GetChainIdHash() == nil {
			if errStr := fillChainId(tx); errStr != "" {
				return errStr
			}
		}
		if tx.GetBody().GetNonce() == 0 {
			state, err := client.GetState(context.Background(), &types.SingleBytes{Value: account})
			if err != nil {
				return err.Error()
			}
			tx.GetBody().Nonce = state.GetNonce() + 1
		}
		if errStr := fillSign(tx, rootConfig.KeyStorePath, pw, account); errStr != "" {
			return "Failed to sign: " + errStr
		}
		txs := []*types.Tx{tx}
		var msgs *types.CommitResultList
		msgs, err = client.CommitTX(context.Background(), &types.TxList{Txs: txs})
		if err != nil {
			return "Failed request to aergo server: " + err.Error()
		}
		return util.JSON(msgs.Results[0])
	} else {
		msg, err := client.SendTX(context.Background(), tx)
		if err != nil {
			return "Failed request to aergo sever: " + err.Error()
		}
		return util.JSON(msg)
	}
}
