package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	luacEncoding "github.com/aergoio/aergo/cmd/aergoluac/encoding"
	luac "github.com/aergoio/aergo/cmd/aergoluac/util"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/types"
	aergorpc "github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var (
	client        *util.ConnClient
	data          string
	nonce         uint64
	toJson        bool
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

// Returns a function that runs callback and handles any errors
func runWithError(callback func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := callback(cmd, args); err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
	}
}

func init() {
	contractCmd := &cobra.Command{
		Use:   "contract [flags] subcommand",
		Short: "Contract command",
	}
	contractCmd.PersistentFlags().Uint64VarP(&gas, "gaslimit", "g", 0, "Gas limit")

	deployCmd := &cobra.Command{
		Use: `deploy [flags] --payload 'payload string' <creatorAddress> [args]
  aergocli contract deploy [flags] <creatorAddress> <bcfile> <abifile> [args]
  
  You can pass constructor arguments by passing a JSON string as the optional final parameter, e.g. "[1, 2, 3]".`,
		Short:                 "Deploy a compiled contract to the server",
		Args:                  nArgs([]int{1, 2, 3, 4}),
		Run:                   runWithError(runDeployCmd),
		DisableFlagsInUseLine: true,
	}
	deployCmd.PersistentFlags().StringVar(&data, "payload", "", "result of compiling a contract")
	deployCmd.PersistentFlags().StringVar(&amount, "amount", "0", "amount of token to send with deployment, in aer")
	deployCmd.PersistentFlags().StringVarP(&contractID, "redeploy", "r", "", "redeploy the contract")
	deployCmd.Flags().StringVar(&pw, "password", "", "password (optional, will be asked on the terminal if not given)")

	callCmd := &cobra.Command{
		Use: `call [flags] <sender> <contract> <funcname> [args]

  You can pass function arguments by passing a JSON string as the optional final parameter, e.g. "[1, 2, 3]".`,
		Short: "Call a contract function",
		Args:  nArgs([]int{3, 4}),
		Run:   runWithError(runCallCmd),
	}
	callCmd.PersistentFlags().Uint64Var(&nonce, "nonce", 0, "manually set a nonce (default: set nonce automatically)")
	callCmd.PersistentFlags().StringVar(&amount, "amount", "0", "amount of token to send with call, in aer")
	callCmd.PersistentFlags().StringVar(&chainIdHash, "chainidhash", "", "chain id hash value encoded by base58")
	callCmd.PersistentFlags().BoolVar(&toJson, "tojson", false, "display json transaction instead of sending to blockchain")
	callCmd.PersistentFlags().BoolVar(&gover, "governance", false, "setting type")
	callCmd.PersistentFlags().BoolVar(&feeDelegation, "delegation", false, "request fee delegation to contract")
	callCmd.Flags().StringVar(&pw, "password", "", "password (optional, will be asked on the terminal if not given)")

	stateQueryCmd := &cobra.Command{
		Use:   "statequery [flags] contract <varname> <varindex>",
		Short: "query the state of a contract with variable name and optional index",
		Args:  cobra.MinimumNArgs(2),
		Run:   runQueryStateCmd,
	}
	stateQueryCmd.Flags().StringVar(&stateroot, "root", "", "Query the state at a specified state root")
	stateQueryCmd.Flags().BoolVar(&compressed, "compressed", false, "Get a compressed proof for the state")

	contractCmd.AddCommand(
		deployCmd,
		callCmd,
		&cobra.Command{
			Use:   "abi [flags] <contractAddress>",
			Short: "Get ABI of the contract",
			Args:  cobra.MinimumNArgs(1),
			Run:   runGetABICmd,
		},
		&cobra.Command{
			Use:   "query [flags] <contractAddress> <funcname> [args]",
			Short: "Query contract by executing read-only function",
			Args:  cobra.MinimumNArgs(2),
			Run:   runQueryCmd,
		},
		stateQueryCmd,
	)
	rootCmd.AddCommand(contractCmd)
}

func runDeployCmd(cmd *cobra.Command, args []string) error {
	var err error
	var code []byte
	var deployArgs []byte

	creator, err := types.DecodeAddress(args[0])
	if err != nil {
		return fmt.Errorf("could not decode address: %v", err.Error())
	}
	state, err := client.GetState(context.Background(), &types.SingleBytes{Value: creator})
	if err != nil {
		return fmt.Errorf("failed to get creator account's state: %v", err.Error())
	}
	var payload []byte
	if len(data) == 0 {
		if len(args) < 3 {
			cmd.Usage()
			os.Exit(1)
		}
		code, err = ioutil.ReadFile(args[1])
		if err != nil {
			return fmt.Errorf("failed to read code file: %v", err.Error())
		}
		var abi []byte
		abi, err = ioutil.ReadFile(args[2])
		if err != nil {
			return fmt.Errorf("failed to read abi file: %v", err.Error())
		}
		if len(args) == 4 {
			var ci types.CallInfo
			err = json.Unmarshal([]byte(args[3]), &ci.Args)
			if err != nil {
				return fmt.Errorf("failed to parse JSON: %v", err.Error())
			}
			deployArgs = []byte(args[3])
		}
		payload = luac.NewLuaCodePayload(luac.NewLuaCode(code, abi), deployArgs)
	} else {
		if len(args) == 2 {
			var ci types.CallInfo
			err = json.Unmarshal([]byte(args[1]), &ci.Args)
			if err != nil {
				return fmt.Errorf("failed to parse JSON: %v", err.Error())
			}
			deployArgs = []byte(args[1])
		}
		code, err = luacEncoding.DecodeCode(data)
		if err != nil {
			return fmt.Errorf("failed to decode code: %v", err.Error())
		}
		payload = luac.NewLuaCodePayload(luac.LuaCode(code), deployArgs)
	}
	amountBigInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return fmt.Errorf("failed to parse amount: %v", err.Error())
	}
	txType := types.TxType_DEPLOY
	var contract []byte
	if len(contractID) > 0 {
		txType = types.TxType_REDEPLOY
		contract, err = types.DecodeAddress(contractID)
		if err != nil {
			return fmt.Errorf("failed to decode contract address: %v", err.Error())
		}
	}
	tx := &types.Tx{
		Body: &types.TxBody{
			Nonce:     state.GetNonce() + 1,
			Account:   creator,
			Payload:   payload,
			Amount:    amountBigInt.Bytes(),
			GasLimit:  gas,
			Type:      txType,
			Recipient: contract,
		},
	}

	cmd.Println(sendTX(cmd, tx, creator))
	return nil
}

func runCallCmd(cmd *cobra.Command, args []string) error {
	caller, err := types.DecodeAddress(args[0])
	if err != nil {
		return fmt.Errorf("could not decode sender address: %v", err.Error())
	}
	if nonce == 0 {
		state, err := client.GetState(context.Background(), &types.SingleBytes{Value: caller})
		if err != nil {
			return fmt.Errorf("failed to get creator account's state: %v", err.Error())
		}
		nonce = state.GetNonce() + 1
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

	if !toJson && !gover {
		abi, err := client.GetABI(context.Background(), &types.SingleBytes{Value: contract})
		if err != nil {
			return fmt.Errorf("failed to get abi: %v", err.Error())
		}
		if !abi.HasFunction(args[2]) {
			return fmt.Errorf("function %v not found in contract at address %s", args[2], args[1])
		}
	}

	amountBigInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return fmt.Errorf("failed to parse amount: %v", err.Error())
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
		if errStr := fillSign(tx, rootConfig.KeyStorePath, pw, caller); errStr != "" {
			return errors.New(errStr)
		}
	} else {
		sign, err := client.SignTX(context.Background(), tx)
		if err != nil || sign == nil {
			return fmt.Errorf("failed to sign tx: %v", err)
		}
		tx = sign
	}

	if toJson {
		cmd.Println(util.TxConvBase58Addr(tx))
	} else {
		txs := []*types.Tx{tx}
		var msgs *types.CommitResultList
		msgs, err = client.CommitTX(context.Background(), &types.TxList{Txs: txs})
		if err != nil {
			return fmt.Errorf("failed to commit tx: %v", err.Error())
		}
		cmd.Println(util.JSON(msgs))
	}
	return nil
}

func runGetABICmd(cmd *cobra.Command, args []string) {
	contract, err := types.DecodeAddress(args[0])
	if err != nil {
		cmd.PrintErrf("Could not decode address: %s\n", err.Error())
		return
	}
	abi, err := client.GetABI(context.Background(), &types.SingleBytes{Value: contract})
	if err != nil {
		log.Fatal(err)
	}
	cmd.Println(util.JSON(abi))
}

func runQueryCmd(cmd *cobra.Command, args []string) {
	contract, err := types.DecodeAddress(args[0])
	if err != nil {
		cmd.PrintErrf("Could not decode address: %s\n", err.Error())
		return
	}
	var ci types.CallInfo

	ci.Name = args[1]
	if len(args) > 2 {
		err = json.Unmarshal([]byte(args[2]), &ci.Args)
		if err != nil {
			log.Fatal(err)
		}
	}
	callinfo, err := json.Marshal(ci)
	if err != nil {
		log.Fatal(err)
	}

	query := &types.Query{
		ContractAddress: contract,
		Queryinfo:       callinfo,
	}

	ret, err := client.QueryContract(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	cmd.Println(ret)
}

func runQueryStateCmd(cmd *cobra.Command, args []string) {
	var root []byte
	var err error
	contract, err := types.DecodeAddress(args[0])
	if err != nil {
		cmd.PrintErrf("Could not decode address: %s\n", err.Error())
		return
	}
	if len(stateroot) != 0 {
		root, err = base58.Decode(stateroot)
		if err != nil {
			cmd.Printf("decode error: %s", err.Error())
			return
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
		log.Fatal(err)
	}
	cmd.Println(ret)
}

func fillChainId(tx *types.Tx) string {
	msg, err := client.Blockchain(context.Background(), &aergorpc.Empty{})
	if err != nil {
		return fmt.Sprintf("Failed: %s\n", err.Error())
	}
	tx.Body.ChainIdHash = msg.GetBestChainIdHash()
	return ""
}
