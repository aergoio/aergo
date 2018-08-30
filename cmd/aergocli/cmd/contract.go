package cmd

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
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
	rootCmd.AddCommand(contractCmd)

	deployCmd := &cobra.Command{
		Use:   "deploy [flags] creator [bcfile] [abifile]",
		Short: "deploy a contract",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			creator, err := base58.Decode(args[0])
			if err != nil {
				log.Fatal(err)
			}
			state, err := client.GetState(context.Background(), &types.SingleBytes{Value: creator})
			if err != nil {
				log.Fatal(err)
			}
			var payload []byte
			if len(data) == 0 {
				if len(args) != 3 {
					fmt.Fprint(os.Stderr, "Usage: aergoluac contract deploy <creator> <bcfile> <abifile>")
					os.Exit(1)
				}
				code, err := ioutil.ReadFile(args[1])
				if err != nil {
					log.Fatal(err)
				}
				abi, err := ioutil.ReadFile(args[2])
				if err != nil {
					log.Fatal(err)
				}
				payload = make([]byte, 4+len(code)+len(abi))

				binary.LittleEndian.PutUint32(payload[0:], uint32(len(code)))
				copy(payload[4:], code)
				copy(payload[4+len(code):], abi)
			} else {
				payload, err = base58.Decode(data)
				if err != nil {
					fmt.Fprint(os.Stderr, err)
					os.Exit(1)
				}
			}
			tx := &types.Tx{
				Body: &types.TxBody{
					Nonce:   state.GetNonce() + 1,
					Account: []byte(creator),
					Payload: payload,
				},
			}

			sign, err := client.SignTX(context.Background(), tx)
			if err != nil || sign == nil {
				log.Fatal(err)
			}
			txs := []*types.Tx{sign}
			commit, err := client.CommitTX(context.Background(), &types.TxList{Txs: txs})

			for i, r := range commit.Results {
				fmt.Println(i+1, ":", util.EncodeB64(r.Hash), r.Error)
			}
		},
	}
	deployCmd.PersistentFlags().StringVar(&data, "payload", "", "result of compiling a contract")
	contractCmd.AddCommand(deployCmd)
	contractCmd.AddCommand(
		&cobra.Command{
			Use:   "call [flags] sender contract name [args]",
			Short: "call a contract function",
			Args:  cobra.MinimumNArgs(3),
			Run: func(cmd *cobra.Command, args []string) {
				caller, err := base58.Decode(args[0])
				if err != nil {
					log.Fatal(err)
				}
				state, err := client.GetState(context.Background(), &types.SingleBytes{Value: caller})
				if err != nil {
					log.Fatal(err)
				}
				contract, err := base58.Decode(args[1])
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
						Account:   []byte(caller),
						Recipient: []byte(contract),
						Payload:   payload,
					},
				}

				sign, err := client.SignTX(context.Background(), tx)
				if err != nil || sign == nil {
					log.Fatal(err)
				}
				txs := []*types.Tx{sign}
				commit, err := client.CommitTX(context.Background(), &types.TxList{Txs: txs})

				for i, r := range commit.Results {
					fmt.Println(i+1, ":", util.EncodeB64(r.Hash), r.Error)
				}
			},
		},
	)
	contractCmd.AddCommand(
		&cobra.Command{
			Use:   "abi [flags] contract",
			Short: "get ABI of the contract",
			Args:  cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				contract, err := base58.Decode(args[0])
				if err != nil {
					log.Fatal(err)
				}
				abi, err := client.GetABI(context.Background(), &types.SingleBytes{Value: contract})
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(util.JSON(abi))
			},
		},
	)
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
