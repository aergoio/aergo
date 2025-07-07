package cmd

import (
	"bytes"
	"context"
	"math/big"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	luacEncoding "github.com/aergoio/aergo/v2/cmd/aergoluac/encoding"
	luaUtil "github.com/aergoio/aergo/v2/cmd/aergoluac/util"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/types"
	aergorpc "github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/jsonrpc"
	"github.com/spf13/cobra"
)

var (
	client        *ConnClient
	admClient     types.AdminRPCServiceClient
	data          string
	nonce         uint64
	toJSON        bool
	gover         bool
	feeDelegation bool
	contractID    string
	gas           uint64
)

func intListToString(ns []int, word string) string {
	if ns == nil || len(ns) == 0 {
		return ""
	}
	if len(ns) == 1 {
		return strconv.Itoa(ns[0])
	}
	slice := ns[:len(ns)-1]
	end := strconv.Itoa(ns[len(ns)-1])
	ret := ""
	for idx, n := range slice {
		ret += strconv.Itoa(n)
		if idx < len(ns)-2 {
			ret += ", "
		}
	}
	return fmt.Sprintf("%s %s %s", ret, word, end)
}

func nArgs(ns []int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, n := range ns {
			if n == len(args) {
				return nil
			}
		}
		return fmt.Errorf("requires exactly %s args but received %d", intListToString(ns, "or"), len(args))
	}
}

func init() {
	contractCmd := &cobra.Command{
		Use:   "contract [flags] subcommand",
		Short: "Contract command",
	}
	contractCmd.PersistentFlags().Uint64VarP(&gas, "gaslimit", "g", 0, "Gas limit")

	deployCmd := &cobra.Command{
		Use: `deploy [flags] <creatorAddress> <path-to-lua-file> [args]
  aergocli contract deploy [flags] <creatorAddress> --payload 'payload string' [args]
  
  You can pass arguments to the constructor() function by passing a JSON string as the optional final parameter, e.g. '[1, "test"]'`,
		Short:                 "Deploy a contract to the server",
		Args:                  nArgs([]int{1, 2, 3}),
		RunE:                  runDeployCmd,
		DisableFlagsInUseLine: true,
	}
	deployCmd.PersistentFlags().Uint64Var(&nonce, "nonce", 0, "manually set a nonce (default: set nonce automatically)")
	deployCmd.PersistentFlags().StringVar(&data, "payload", "", "result of compiling a contract with aergoluac")
	deployCmd.PersistentFlags().StringVar(&amount, "amount", "0", "amount of token to send with deployment, in aer")
	deployCmd.PersistentFlags().StringVarP(&contractID, "redeploy", "r", "", "redeploy the contract")
	deployCmd.Flags().StringVar(&pw, "password", "", "password (optional, will be asked on the terminal if not given)")

	callCmd := &cobra.Command{
		Use: `call [flags] <sender> <contract> <funcname> [args]

  You can pass function arguments by passing a JSON string as the optional final parameter, e.g. "[1, 2, 3]".`,
		Short: "Call a contract function",
		Args:  nArgs([]int{3, 4}),
		RunE:  runCallCmd,
	}
	callCmd.PersistentFlags().Uint64Var(&nonce, "nonce", 0, "manually set a nonce (default: set nonce automatically)")
	callCmd.PersistentFlags().StringVar(&amount, "amount", "0", "amount of token to send with call, in aer")
	callCmd.PersistentFlags().StringVar(&chainIdHash, "chainidhash", "", "chain id hash value encoded by base58")
	callCmd.PersistentFlags().BoolVar(&toJSON, "tojson", false, "display json transaction instead of sending to blockchain")
	callCmd.PersistentFlags().BoolVar(&gover, "governance", false, "setting type")
	callCmd.PersistentFlags().BoolVar(&feeDelegation, "delegation", false, "request fee delegation to contract")
	callCmd.Flags().StringVar(&pw, "password", "", "password (optional, will be asked on the terminal if not given)")

	multicallCmd := &cobra.Command{
		Use: `multicall [flags] <sender> <script>

  The script is a JSON array of arrays, enclosed as a string, containing the commands`,
		Short: "Calls multiple contracts / functions",
		Args:  nArgs([]int{2}),
		RunE:  runMulticallCmd,
	}
	multicallCmd.PersistentFlags().Uint64Var(&nonce, "nonce", 0, "manually set a nonce (default: set nonce automatically)")
	multicallCmd.PersistentFlags().StringVar(&chainIdHash, "chainidhash", "", "chain id hash value encoded by base58")
	multicallCmd.PersistentFlags().BoolVar(&toJSON, "tojson", false, "display json transaction instead of sending to blockchain")
	multicallCmd.Flags().StringVar(&pw, "password", "", "password (optional, will be asked on the terminal if not given)")

	stateQueryCmd := &cobra.Command{
		Use:   "statequery [flags] <contractAddress> <varname> [varindex]",
		Short: "query the state of a contract with variable name and optional index",
		Args:  cobra.MinimumNArgs(2),
		RunE:  runQueryStateCmd,
	}
	stateQueryCmd.Flags().StringVar(&stateroot, "root", "", "Query the state at a specified state root")
	stateQueryCmd.Flags().BoolVar(&compressed, "compressed", false, "Get a compressed proof for the state")

	contractCmd.AddCommand(
		deployCmd,
		callCmd,
		multicallCmd,
		&cobra.Command{
			Use:   "abi [flags] <contractAddress>",
			Short: "Get ABI of the contract",
			Args:  cobra.ExactArgs(1),
			RunE:  runGetABICmd,
		},
		&cobra.Command{
			Use:   "query [flags] <contractAddress> <funcname> [args]",
			Short: "Query contract by executing read-only function",
			Args:  cobra.MinimumNArgs(2),
			RunE:  runQueryCmd,
		},
		stateQueryCmd,
	)
	rootCmd.AddCommand(contractCmd)
}

func runDeployCmd(cmd *cobra.Command, args []string) error {
	var err error
	var code []byte
	var deployArgs []byte

	cmd.SilenceUsage = true

	creator, err := types.DecodeAddress(args[0])
	if err != nil {
		return fmt.Errorf("could not decode address: %v", err.Error())
	}

	if nonce == 0 {
		state, err := client.GetState(context.Background(), &types.SingleBytes{Value: creator})
		if err != nil {
			return fmt.Errorf("failed to get creator account's state: %v", err.Error())
		}
		nonce = state.GetNonce() + 1
	}

	chainInfo, err := client.GetChainInfo(context.Background(), &types.Empty{})
	if err != nil {
		return fmt.Errorf("could not retrieve chain info: %v", err.Error())
	}

	var payload []byte
	if len(data) == 0 {
		if chainInfo.Id.Version < 4 {
			cmd.SilenceUsage = false
			return errors.New("for old hardforks use aergoluac and --payload method instead")
		}
		if len(args) < 2 {
			cmd.SilenceUsage = false
			return errors.New("not enough arguments")
		}
		codeString, err := luaUtil.ReadContract(args[1])
		if err != nil {
			return fmt.Errorf("failed to read code file: %v", err.Error())
		}
		code = []byte(codeString)
		if len(args) == 3 {
			var ci types.CallInfo
			err = json.Unmarshal([]byte(args[2]), &ci.Args)
			if err != nil {
				return fmt.Errorf("failed to parse arguments (JSON): %v", err.Error())
			}
			deployArgs = []byte(args[2])
		}
		payload = luaUtil.NewLuaCodePayload(luaUtil.LuaCode(code), deployArgs)
	} else {
		if chainInfo.Id.Version >= 4 {
			cmd.SilenceUsage = false
			return errors.New("this chain only accepts deploy in plain source code\nuse the other method instead")
		}
		if len(args) == 2 {
			var ci types.CallInfo
			err = json.Unmarshal([]byte(args[1]), &ci.Args)
			if err != nil {
				return fmt.Errorf("failed to parse JSON: %v", err.Error())
			}
			deployArgs = []byte(args[1])
		}
		// check if the data is in hex format
		if hex.IsHexString(data) {
			if deployArgs != nil {
				cmd.SilenceUsage = false
				return errors.New("the call arguments are expected to be already on the hex data")
			}
			// the data is expected to be copied from aergoscan view of
			// the transaction that deployed the contract
			payload, err = hex.Decode(data)
		} else {
			// the data is the output of aergoluac
			code, err = luacEncoding.DecodeCode(data)
			if err != nil {
				return fmt.Errorf("failed to decode code: %v", err.Error())
			}
			payload = luaUtil.NewLuaCodePayload(luaUtil.LuaCode(code), deployArgs)
		}
	}

	amountBigInt, err := jsonrpc.ParseUnit(amount)
	if err != nil {
		return fmt.Errorf("failed to parse amount: %v", err.Error())
	}

	txType := types.TxType_DEPLOY
	var recipient []byte
	if len(contractID) > 0 {
		txType = types.TxType_REDEPLOY
		recipient, err = types.DecodeAddress(contractID)
		if err != nil {
			return fmt.Errorf("failed to decode contract address: %v", err.Error())
		}
	}

	tx := &types.Tx{
		Body: &types.TxBody{
			Nonce:     nonce,
			Account:   creator,
			Payload:   payload,
			Amount:    amountBigInt.Bytes(),
			GasLimit:  gas,
			Type:      txType,
			Recipient: recipient,
		},
	}
	cmd.Println(sendTX(cmd, tx, creator))
	return nil
}

func runCallCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	caller, err := types.DecodeAddress(args[0])
	if err != nil {
		return fmt.Errorf("could not decode sender address: %v", err.Error())
	}

	contract, err := types.DecodeAddress(args[1])
	if err != nil {
		return fmt.Errorf("could not decode contract address: %v", err.Error())
	}

	var ci types.CallInfo
	ci.Name = args[2]
	if len(args) > 3 {
		err = json.Unmarshal([]byte(args[3]), &ci.Args)
		if err != nil {
			return fmt.Errorf("failed to parse JSON: %v", err.Error())
		}
	}
	payload, err := json.Marshal(ci)
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err.Error())
	}

	if !toJSON && !gover {
		abi, err := client.GetABI(context.Background(), &types.SingleBytes{Value: contract})
		if err != nil {
			return fmt.Errorf("failed to get abi: %v", err.Error())
		}
		if !abi.HasFunction(args[2]) {
			return fmt.Errorf("function %v not found in contract at address %s", args[2], args[1])
		}
	}

	amountBigInt, err := jsonrpc.ParseUnit(amount)
	if err != nil {
		return fmt.Errorf("failed to parse amount: %v", err)
	}

	var txType types.TxType
	if gover {
		txType = types.TxType_GOVERNANCE
	} else if feeDelegation {
		txType = types.TxType_FEEDELEGATION
	} else {
		txType = types.TxType_CALL
	}

	tx := &types.Tx{
		Body: &types.TxBody{
			Nonce:     nonce,
			Account:   caller,
			Recipient: contract,
			Payload:   payload,
			Amount:    amountBigInt.Bytes(),
			GasLimit:  gas,
			Type:      txType,
		},
	}

	return sendCallTx(cmd, tx, caller)
}

func runMulticallCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	caller, err := types.DecodeAddress(args[0])
	if err != nil {
		return fmt.Errorf("could not decode sender address: %v", err.Error())
	}

	script := args[1]

	tx := &types.Tx{
		Body: &types.TxBody{
			Nonce:     nonce,
			Account:   caller,
			Recipient: []byte{},
			Payload:   []byte(script),
			Amount:    big.NewInt(0).Bytes(),
			GasLimit:  gas,
			Type:      types.TxType_MULTICALL,
		},
	}

	return sendCallTx(cmd, tx, caller)
}

func sendCallTx(cmd *cobra.Command, tx *types.Tx, sender []byte) error {
	var err error

	if tx.GetBody().GetNonce() == 0 {
		state, err := client.GetState(context.Background(), &types.SingleBytes{Value: sender})
		if err != nil {
			return fmt.Errorf("failed to get sender account's state: %v", err.Error())
		}
		tx.GetBody().Nonce = state.GetNonce() + 1
	}

	if chainIdHash != "" {
		rawCidHash, err := base58.Decode(chainIdHash)
		if err != nil {
			return fmt.Errorf("failed to parse chainidhash: %v", err.Error())
		}
		tx.Body.ChainIdHash = rawCidHash
	} else {
		if errStr := fillChainId(tx); errStr != "" {
			return errors.New(errStr)
		}
	}

	if pw == "" {
		pw, err = getPasswd(cmd, false)
		if err != nil {
			return err
		}
	}

	if rootConfig.KeyStorePath != "" {
		if errStr := fillSign(tx, rootConfig.KeyStorePath, pw, sender); errStr != "" {
			return errors.New(errStr)
		}
	} else {
		sign, err := client.SignTX(context.Background(), tx)
		if err != nil || sign == nil {
			return fmt.Errorf("failed to sign tx: %v", err)
		}
		tx = sign
	}

	if toJSON {
		res := jsonrpc.ConvTx(tx, jsonrpc.Base58)
		cmd.Println(jsonrpc.MarshalJSON(res))
	} else {
		txs := []*types.Tx{tx}
		var msgs *types.CommitResultList
		msgs, err = client.CommitTX(context.Background(), &types.TxList{Txs: txs})
		if err != nil {
			return fmt.Errorf("failed to commit tx: %v", err.Error())
		}
		res := jsonrpc.ConvCommitResult(msgs.Results[0])
		cmd.Println(jsonrpc.MarshalJSON(res))
	}
	return nil
}

func runGetABICmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	contract, err := types.DecodeAddress(args[0])
	if err != nil {
		return fmt.Errorf("failed to decode address: %v", err.Error())
	}
	abi, err := client.GetABI(context.Background(), &types.SingleBytes{Value: contract})
	if err != nil {
		return fmt.Errorf("failed to get abi: %v", err.Error())
	}
	res := jsonrpc.ConvAbi(abi)
	cmd.Println(jsonrpc.MarshalJSON(res))
	return nil
}

func runQueryCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	contract, err := types.DecodeAddress(args[0])
	if err != nil {
		return fmt.Errorf("failed to decode address: %v", err.Error())
	}
	var ci types.CallInfo

	ci.Name = args[1]
	if len(args) > 2 {
		err = json.Unmarshal([]byte(args[2]), &ci.Args)
		if err != nil {
			return fmt.Errorf("failed to parse JSON: %v", err.Error())
		}
	}
	callinfo, err := json.Marshal(ci)
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err.Error())
	}

	query := &types.Query{
		ContractAddress: contract,
		Queryinfo:       callinfo,
	}

	ret, err := client.QueryContract(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to query contract: %v", err.Error())
	}
	cmd.Println(ret)
	return nil
}

func runQueryStateCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	var root []byte
	var err error
	contract, err := types.DecodeAddress(args[0])
	if err != nil {
		return fmt.Errorf("failed to decode address: %v", err.Error())
	}
	if len(stateroot) != 0 {
		root, err = base58.Decode(stateroot)
		if err != nil {
			return fmt.Errorf("failed to decode stateroot: %v", err.Error())
		}
	}
	storageKeyPlain := bytes.NewBufferString("_sv_")
	storageKeyPlain.WriteString(args[1])
	if len(args) > 2 {
		storageKeyPlain.WriteString("-")
		storageKeyPlain.WriteString(args[2])
	}
	storageKey := common.Hasher([]byte(storageKeyPlain.Bytes()))
	stateQuery := &types.StateQuery{
		ContractAddress: contract,
		StorageKeys:     [][]byte{storageKey},
		Root:            root,
		Compressed:      compressed,
	}
	ret, err := client.QueryContractState(context.Background(), stateQuery)
	if err != nil {
		return fmt.Errorf("failed to query contract state: %v", err.Error())
	}
	cmd.Println(ret)
	return nil
}

func fillChainId(tx *types.Tx) string {
	msg, err := client.Blockchain(context.Background(), &aergorpc.Empty{})
	if err != nil {
		return fmt.Sprintf("Failed: %s\n", err.Error())
	}
	tx.Body.ChainIdHash = msg.GetBestChainIdHash()
	return ""
}
