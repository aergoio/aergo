/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/jsonrpc"
	"github.com/spf13/cobra"
)

var evmDeployCmd = &cobra.Command{
	Use:   "evmdeploy",
	Short: "Invoke EVM",
	Args:  cobra.MinimumNArgs(0),
	RunE:  deployEVM,
}

var evmCallCmd = &cobra.Command{
	Use:   "evmcall",
	Short: "Invoke EVM",
	Args:  cobra.MinimumNArgs(0),
	RunE:  callEVM,
}

var queryEvmCmd = &cobra.Command{
	Use:   "evmquery [flags] <contractAddress> <payload> [args]",
	Short: "Query EVM contract by executing read-only function",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runQueryEVMCmd,
}

func init() {
	rootCmd.AddCommand(evmDeployCmd)
	evmDeployCmd.PersistentFlags().StringVar(&data, "payload", "", "result of compiling a contract")
	evmDeployCmd.Flags().StringVar(&from, "from", "", "Sender account address")
	evmDeployCmd.MarkFlagRequired("from")
	evmDeployCmd.Flags().Uint64Var(&nonce, "nonce", 0, "setting nonce manually")
	evmDeployCmd.Flags().Uint64VarP(&gas, "gaslimit", "g", 0, "Gas limit")
	evmDeployCmd.Flags().StringVar(&pw, "password", "", "Password")

	rootCmd.AddCommand(evmCallCmd)
	evmCallCmd.PersistentFlags().StringVar(&data, "payload", "", "result of compiling a contract")
	evmCallCmd.Flags().StringVar(&from, "from", "", "Sender account address")
	evmCallCmd.MarkFlagRequired("from")
	evmCallCmd.Flags().StringVar(&to, "to", "", "contract address")
	evmCallCmd.MarkFlagRequired("to")
	evmCallCmd.Flags().Uint64Var(&nonce, "nonce", 0, "setting nonce manually")
	evmCallCmd.Flags().Uint64VarP(&gas, "gaslimit", "g", 0, "Gas limit")
	evmCallCmd.Flags().StringVar(&pw, "password", "", "Password")

	rootCmd.AddCommand(queryEvmCmd)
}

func deployEVM(cmd *cobra.Command, args []string) error {
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
		Type:     types.TxType_EVMDEPLOY,
		Account:  account,
		Payload:  payload,
		Nonce:    nonce,
		GasLimit: gas,
	}}

	cmd.Println(sendEVMTX(cmd, tx, account))
	return nil
}

func callEVM(cmd *cobra.Command, args []string) error {
	account, err := types.DecodeAddress(from)
	if err != nil {
		return errors.New("Wrong address in --from flag\n" + err.Error())
	}

	contractAddress, err := hex.DecodeString(to)
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

	payload = append(contractAddress, payload...)

	cmd.Println(hex.EncodeToString(payload))

	tx := &types.Tx{Body: &types.TxBody{
		Type:     types.TxType_EVMCALL,
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
		return jsonrpc.MarshalJSON(msgs.Results[0])
	} else {
		msg, err := client.SendTX(context.Background(), tx)
		if err != nil {
			return "Failed request to aergo sever: " + err.Error()
		}
		return jsonrpc.MarshalJSON(msg)
	}
}

func runQueryEVMCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	contract, _ := hex.DecodeString(args[0])
	callinfo, _ := hex.DecodeString(args[1])

	query := &types.Query{
		ContractAddress: contract,
		Queryinfo:       callinfo,
	}

	ret, err := client.QueryEVMContract(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to query EVM contract: %v", err.Error())
	}
	cmd.Println(hex.EncodeToString(ret.Value))
	return nil
}
