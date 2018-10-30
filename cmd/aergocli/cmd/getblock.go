/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/binary"

	"github.com/mr-tron/base58/base58"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	aergorpc "github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
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
	fflags := cmd.Flags()
	if fflags.Changed("number") == false && fflags.Changed("hash") == false {
		cmd.Println("no block --hash or --number specified")
		return
	}
	var blockQuery []byte
	if hash == "" {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(number))
		blockQuery = b
	} else {
		decoded, err := base58.Decode(hash)
		if err != nil {
			cmd.Printf("decode error: %s", err.Error())
			return
		}
		if len(decoded) == 0 {
			cmd.Println("decode error:")
			return
		}
		blockQuery = decoded
	}

	msg, err := client.GetBlock(context.Background(), &aergorpc.SingleBytes{Value: blockQuery})
	if nil == err {
		cmd.Println(util.BlockConvBase58Addr(msg))
	} else {
		cmd.Printf("Failed: %s\n", err.Error())
	}
}
