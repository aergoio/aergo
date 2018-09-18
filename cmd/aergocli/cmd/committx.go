/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"

	sha256 "github.com/minio/sha256-simd"

	"fmt"
	"io/ioutil"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var committxCmd = &cobra.Command{
	Use:   "committx",
	Short: "Send transaction",
	Args:  cobra.MinimumNArgs(0),
	Run:   execCommitTX,
}

var nonce uint64
var recipient string
var price int64

//var script string
var jsonTx string
var jsonPath string

func init() {
	rootCmd.AddCommand(committxCmd)
	committxCmd.Flags().StringVar(&jsonTx, "jsontx", "", "Transaction list json")
	committxCmd.Flags().StringVar(&jsonPath, "jsontxpath", "", "Transaction list json file path")
}

func execCommitTX(cmd *cobra.Command, args []string) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var client *util.ConnClient
	var ok bool
	if client, ok = util.GetClient(GetServerAddress(), opts).(*util.ConnClient); !ok {
		panic("Internal error. wrong RPC client type")
	}
	defer client.Close()

	var msg *types.CommitResultList
	if jsonPath != "" {
		b, readerr := ioutil.ReadFile(jsonPath)
		if readerr != nil {
			fmt.Printf("Failed: %s\n", readerr.Error())
			return
		}
		jsonTx = string(b)
	}

	if jsonTx != "" {
		txlist, err := util.ParseBase58Tx([]byte(jsonTx))
		if err != nil {
			fmt.Printf("Failed: %s\n", err.Error())
			return
		}
		msg, err = client.CommitTX(context.Background(), &types.TxList{Txs: txlist})
		if err != nil {
			fmt.Printf("Failed: %s\n", err.Error())
			return
		}
	}
	for i, r := range msg.Results {
		fmt.Println(i+1, ":", util.EncodeB64(r.Hash), r.Error)
	}
}

func convertBase58(in []byte) []byte {
	if in == nil {
		return nil
	}
	b64in := util.EncodeB64(in)
	out, err := base58.Decode(b64in)
	if err != nil {
		panic("could not convert input to base58")
	}
	return out
}

func caculateHash(body *types.TxBody) []byte {
	input := append(body.Recipient, body.Payload...)
	sum1 := sha256.Sum256(input)
	sum2 := sha256.Sum256(sum1[:])
	return sum2[:]
}
