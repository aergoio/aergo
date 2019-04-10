package cmd

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var (
	client *util.ConnClient
	data   string
	nonce  uint64
	toJson bool
	gover  bool
)

func init() {
	contractCmd := &cobra.Command{
		Use:   "contract [flags] subcommand",
		Short: "Contract command",
	}

	deployCmd := &cobra.Command{
		Use:                   "deploy [flags] --payload 'payload string' creator\n  aergocli contract deploy [flags] creator bcfile abifile",
		Short:                 "Deploy a compiled contract to the server",
		Args:                  cobra.MinimumNArgs(1),
		Run:                   runDeployCmd,
		DisableFlagsInUseLine: true,
	}
	deployCmd.PersistentFlags().StringVar(&data, "payload", "", "result of compiling a contract")
	deployCmd.PersistentFlags().StringVar(&amount, "amount", "0", "setting amount")

	callCmd := &cobra.Command{
		Use:   "call [flags] sender contract funcname '[argument...]'",
		Short: "Call a contract function",
		Args:  cobra.MinimumNArgs(3),
		Run:   runCallCmd,
	}
	callCmd.PersistentFlags().Uint64Var(&nonce, "nonce", 0, "setting nonce manually")
	callCmd.PersistentFlags().StringVar(&amount, "amount", "0", "setting amount")
	callCmd.PersistentFlags().StringVar(&chainIdHash, "chainidhash", "", "chain id hash value encoded by base58")
	callCmd.PersistentFlags().BoolVar(&toJson, "tojson", false, "get jsontx")
	callCmd.PersistentFlags().BoolVar(&gover, "governance", false, "setting type")

	stateQueryCmd := &cobra.Command{
		Use:   "statequery [flags] contract varname varindex",
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
			Use:   "abi [flags] contract",
			Short: "Get ABI of the contract",
			Args:  cobra.MinimumNArgs(1),
			Run:   runGetABICmd,
		},
		&cobra.Command{
			Use:   "query [flags] contract funcname '[argument...]'",
			Short: "Query contract by executing read-only function",
			Args:  cobra.MinimumNArgs(2),
			Run:   runQueryCmd,
		},
		stateQueryCmd,
	)
	rootCmd.AddCommand(contractCmd)
}

func runDeployCmd(cmd *cobra.Command, args []string) {
	var err error
	creator, err := types.DecodeAddress(args[0])
	if err != nil {
		log.Fatal(err)
	}
	state, err := client.GetState(context.Background(), &types.SingleBytes{Value: creator})
	if err != nil {
		log.Fatal(err)
	}
	var payload []byte
	if len(data) == 0 {
		if len(args) < 3 {
			fmt.Fprint(os.Stderr, "Usage: aergocli contract deploy <creator> <bcfile> <abifile> [args]")
			os.Exit(1)
		}
		var code []byte
		var argLen int
		code, err = ioutil.ReadFile(args[1])
		if err != nil {
			log.Fatal(err)
		}
		var abi []byte
		abi, err = ioutil.ReadFile(args[2])
		if err != nil {
			log.Fatal(err)
		}
		if len(args) == 4 {
			var ci types.CallInfo
			err = json.Unmarshal([]byte(args[3]), &ci.Args)
			if err != nil {
				log.Fatal(err)
			}
			argLen = len(args[3])
		}
		payload = make([]byte, 8+len(code)+len(abi)+argLen)
		binary.LittleEndian.PutUint32(payload[0:], uint32(len(code)+len(abi)+8))
		binary.LittleEndian.PutUint32(payload[4:], uint32(len(code)))
		codeLen := copy(payload[8:], code)
		abiLen := copy(payload[8+codeLen:], abi)
		if argLen != 0 {
			copy(payload[8+codeLen+abiLen:], args[3])
		}
	} else {
		var argLen int

		if len(args) == 2 {
			var ci types.CallInfo
			err = json.Unmarshal([]byte(args[1]), &ci.Args)
			if err != nil {
				log.Fatal(err)
			}
			argLen = len(args[1])
		}
		code, err := util.DecodeCode(data)
		payload = make([]byte, 4+len(code)+argLen)
		binary.LittleEndian.PutUint32(payload[0:], uint32(len(code)+4))
		codeLen := copy(payload[4:], code)
		if argLen != 0 {
			copy(payload[4+codeLen:], args[1])
		}

		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	}
	amountBigInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		fmt.Fprint(os.Stderr, "failed to parse --amount flags")
		os.Exit(1)
	}
	tx := &types.Tx{
		Body: &types.TxBody{
			Nonce:   state.GetNonce() + 1,
			Account: creator,
			Payload: payload,
			Amount:  amountBigInt.Bytes(),
		},
	}

	msg, err := client.SendTX(context.Background(), tx)
	if err != nil || msg == nil {
		log.Fatal(err)
	}
	cmd.Println(util.JSON(msg))
}

func runCallCmd(cmd *cobra.Command, args []string) {
	caller, err := types.DecodeAddress(args[0])
	if err != nil {
		log.Fatal(err)
	}
	if nonce == 0 {
		state, err := client.GetState(context.Background(), &types.SingleBytes{Value: caller})
		if err != nil {
			log.Fatal(err)
		}
		nonce = state.GetNonce() + 1
	}
	contract, err := types.DecodeAddress(args[1])
	if err != nil {
		log.Fatal(err)
	}

	var ci types.CallInfo
	ci.Name = args[2]
	if len(args) > 3 {
		err = json.Unmarshal([]byte(args[3]), &ci.Args)
		if err != nil {
			log.Fatal(err)
		}
	}
	payload, err := json.Marshal(ci)
	if err != nil {
		log.Fatal(err)
	}

	if !toJson && !gover {
		abi, err := client.GetABI(context.Background(), &types.SingleBytes{Value: contract})
		if err != nil {
			log.Fatal(err)
		}
		var found bool
		for _, fn := range abi.Functions {
			if fn.GetName() == args[2] {
				found = true
				break
			}
		}
		if !found {
			log.Fatal(args[2], " function not found in contract :", args[1])
		}
	}

	amountBigInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		fmt.Fprint(os.Stderr, "failed to parse --amount flags")
		os.Exit(1)
	}
	txType := types.TxType_NORMAL
	if gover {
		txType = types.TxType_GOVERNANCE
	}

	tx := &types.Tx{
		Body: &types.TxBody{
			Nonce:     nonce,
			Account:   caller,
			Recipient: contract,
			Payload:   payload,
			Amount:    amountBigInt.Bytes(),
			Type:      txType,
		},
	}

	if chainIdHash != "" {
		rawCidHash, err := base58.Decode(chainIdHash)
		if err != nil {
			fmt.Fprint(os.Stderr, "failed to parse --chainidhash flags\n")
			os.Exit(1)
		}
		tx.Body.ChainIdHash = rawCidHash
	}

	if toJson {
		if chainIdHash == "" {
			status, err := client.Blockchain(context.Background(), &types.Empty{})
			if err != nil {
				cmd.Printf("Failed: %s\n", err.Error())
				return
			}
			tx.Body.ChainIdHash = status.BestChainIdHash
		}
		sign, err := client.SignTX(context.Background(), tx)
		if err != nil || sign == nil {
			log.Fatal(err)
		}
		fmt.Println(util.TxConvBase58Addr(sign))
		return
	}
	msg, err := client.SendTX(context.Background(), tx)
	if err != nil || msg == nil {
		log.Fatal(err)
	}
	cmd.Println(util.JSON(msg))
}

func runGetABICmd(cmd *cobra.Command, args []string) {
	contract, err := types.DecodeAddress(args[0])
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
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
		log.Fatal(err)
	}
	if len(stateroot) != 0 {
		root, err = base58.Decode(stateroot)
		if err != nil {
			cmd.Printf("decode error: %s", err.Error())
			return
		}
	}
	storageKey := bytes.NewBufferString("_sv_")
	storageKey.WriteString(args[1])
	if len(args) > 2 {
		storageKey.WriteString("-")
		storageKey.WriteString(args[2])
	}
	stateQuery := &types.StateQuery{
		ContractAddress: contract,
		StorageKeys:     []string{storageKey.String()},
		Root:            root,
		Compressed:      compressed,
	}
	ret, err := client.QueryContractState(context.Background(), stateQuery)
	if err != nil {
		log.Fatal(err)
	}
	cmd.Println(ret)
}
