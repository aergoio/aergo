package cmd

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/mr-tron/base58/base58"
	"io/ioutil"
	"log"
	"os"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	client *util.ConnClient
	data   string
)

func init() {
	contractCmd := &cobra.Command{
		Use:               "contract [flags] subcommand",
		Short:             "contract command",
		PersistentPreRun:  connectAergo,
		PersistentPostRun: disconnectAergo,
	}

	deployCmd := &cobra.Command{
		Use:   "deploy [flags] creator [bcfile] [abifile]",
		Short: "deploy a contract",
		Args:  cobra.MinimumNArgs(1),
		Run:   runDeployCmd,
	}
	deployCmd.PersistentFlags().StringVar(&data, "payload", "", "result of compiling a contract")

	contractCmd.AddCommand(
		deployCmd,
		&cobra.Command{
			Use:   "call [flags] sender contract name [args]",
			Short: "call a contract function",
			Args:  cobra.MinimumNArgs(3),
			Run:   runCallCmd,
		},
		&cobra.Command{
			Use:   "abi [flags] contract",
			Short: "get ABI of the contract",
			Args:  cobra.MinimumNArgs(1),
			Run:   runGetABICmd,
		},
		&cobra.Command{
			Use:   "query [flags] contract fname [args]",
			Short: "query contract by executing read-only function",
			Args:  cobra.MinimumNArgs(2),
			Run:   runQueryCmd,
		},
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
		code, err := types.DecodeAddress(data)
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
	tx := &types.Tx{
		Body: &types.TxBody{
			Nonce:   state.GetNonce() + 1,
			Account: creator,
			Payload: payload,
		},
	}

	sign, err := client.SignTX(context.Background(), tx)
	if err != nil || sign == nil {
		log.Fatal(err)
	}
	txs := []*types.Tx{sign}
	commit, err := client.CommitTX(context.Background(), &types.TxList{Txs: txs})
	if err != nil {
		log.Fatal(err)
	}

	for i, r := range commit.Results {
		fmt.Println(i+1, ":", base58.Encode(r.Hash), r.Error)
	}
}

func runCallCmd(cmd *cobra.Command, args []string) {
	caller, err := types.DecodeAddress(args[0])
	if err != nil {
		log.Fatal(err)
	}
	state, err := client.GetState(context.Background(), &types.SingleBytes{Value: caller})
	if err != nil {
		log.Fatal(err)
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
	tx := &types.Tx{
		Body: &types.TxBody{
			Nonce:     state.GetNonce() + 1,
			Account:   caller,
			Recipient: contract,
			Payload:   payload,
		},
	}

	sign, err := client.SignTX(context.Background(), tx)
	if err != nil || sign == nil {
		log.Fatal(err)
	}
	txs := []*types.Tx{sign}
	commit, err := client.CommitTX(context.Background(), &types.TxList{Txs: txs})
	if err != nil {
		log.Fatal(err)
	}

	for i, r := range commit.Results {
		fmt.Println(i+1, ":", base58.Encode(r.Hash), r.Error)
	}
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
	fmt.Println(util.JSON(abi))
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
		ContractAddress: []byte(contract),
		Queryinfo:       callinfo,
	}

	ret, err := client.QueryContract(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ret)
}

func connectAergo(cmd *cobra.Command, args []string) {
	serverAddr := GetServerAddress()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var ok bool
	client, ok = util.GetClient(serverAddr, opts).(*util.ConnClient)
	if !ok {
		log.Fatal("internal error. wrong RPC client type")
	}
}

func disconnectAergo(cmd *cobra.Command, args []string) {
	if client != nil {
		client.Close()
	}
}
