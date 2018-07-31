/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"fmt"

	"github.com/mr-tron/base58/base58"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var getstateCmd = &cobra.Command{
	Use:   "getstate",
	Short: "Get account state",
	Run:   execGetState,
}

var address string

func init() {
	rootCmd.AddCommand(getstateCmd)
	getstateCmd.Flags().StringVar(&address, "address", "", "Get state from the address")
}

func execGetState(cmd *cobra.Command, args []string) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var client *util.ConnClient
	var ok bool
	if client, ok = util.GetClient(GetServerAddress(), opts).(*util.ConnClient); !ok {
		panic("Internal error. wrong RPC client type")
	}
	defer client.Close()
	param, err := base58.Decode(address)
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
	}
	msg, err := client.GetState(context.Background(),
		&types.SingleBytes{Value: param})
	if nil == err {
		fmt.Printf("{account:%s, nonce:%d, balance:%d}",
			base58.Encode(msg.GetAccount()), msg.GetNonce(), msg.GetBalance())
	} else {
		fmt.Printf("Failed: %s\n", err.Error())
	}
}
