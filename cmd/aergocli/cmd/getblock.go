/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	aergorpc "github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var getblockCmd = &cobra.Command{
	Use:   "getblock",
	Short: "Get block information",
	Args:  cobra.MinimumNArgs(0),
	Run:   execGetBlock,
}

var number uint64
var hash string

func init() {
	rootCmd.AddCommand(getblockCmd)
	getblockCmd.Flags().Uint64VarP(&number, "number", "n", 0, "Block height")
	getblockCmd.Flags().StringVarP(&hash, "hash", "", "", "Block hash")
}

func execGetBlock(cmd *cobra.Command, args []string) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var client *util.ConnClient
	var ok bool
	if client, ok = util.GetClient(GetServerAddress(), opts).(*util.ConnClient); !ok {
		panic("Internal error. wrong RPC client type")
	}
	defer client.Close()

	var blockQuery []byte
	if hash == "" {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(number))
		blockQuery = b
	} else {

		if len(hash)%4 > 0 {
			toAdd := 4 - len(hash)%4
			for toAdd > 0 {
				hash = hash + "="
				toAdd--
			}
			fmt.Printf("Trying to append change input to %s by appending filling char =\n", hash)
		}
		decoded, err := util.DecodeB64(hash)
		if err != nil {
			fmt.Printf("decode error: %s", err.Error())
			return
		}
		if len(decoded) == 0 {
			fmt.Println("decode error:")
			return
		}
		blockQuery = decoded
	}

	msg, err := client.GetBlock(context.Background(), &aergorpc.SingleBytes{Value: blockQuery})
	if nil == err {
		fmt.Println(util.JSON(msg))
	} else {
		fmt.Printf("Failed: %s\n", err.Error())
	}
}
