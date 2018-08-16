package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var client *util.ConnClient

func init() {
	contractCmd := &cobra.Command{
		Use:               "contract [flags] subcommand",
		Short:             "contract command",
		PersistentPreRun:  connectAergo,
		PersistentPostRun: disconnectAergo,
	}
	rootCmd.AddCommand(contractCmd)

	contractCmd.AddCommand(
		&cobra.Command{
			Use:   "deploy [flags] address bcfile",
			Short: "deploy a contract",
			Args:  cobra.MinimumNArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				creator, err := base58.Decode(args[0])
				if err != nil {
					log.Fatal(err)
				}
				state, err := client.GetState(context.Background(), &types.SingleBytes{Value: creator})
				if err != nil {
					log.Fatal(err)
				}
				code, err := ioutil.ReadFile(args[1])
				if err != nil {
					log.Fatal(err)
				}
				tx := &types.Tx{
					Body: &types.TxBody{
						Nonce:   state.GetNonce() + 1,
						Account: []byte(creator),
						Payload: []byte(code),
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
			Use:   "call [flags] sender contract name args",
			Short: "deploy contract",
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
				var abi types.ABI
				abi.Name = args[2]
				if len(args) > 3 {
					err = json.Unmarshal([]byte(args[3]), &abi.Args)
					if err != nil {
						log.Fatal(err)
					}
				}
				payload, err := json.Marshal(abi)
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
