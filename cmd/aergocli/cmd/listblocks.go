/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	"github.com/aergoio/aergo/v2/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var listblockheadersCmd = &cobra.Command{
	Use:   "listblocks",
	Short: "Get block headers list",
	Run:   execListBlockHeaders,
}
var gbhHash string
var gbhHeight int32
var gbhSize int
var gbhOffset int
var gbhAsc bool

func init() {
	rootCmd.AddCommand(listblockheadersCmd)

	listblockheadersCmd.Flags().StringVar(&gbhHash, "hash", "", "Block hash")
	listblockheadersCmd.Flags().Int32Var(&gbhHeight, "height", int32(-1), "Block height")
	listblockheadersCmd.Flags().IntVar(&gbhSize, "size", 20, "Max list size")
	listblockheadersCmd.Flags().IntVar(&gbhOffset, "offset", 0, "Offset")
	listblockheadersCmd.Flags().BoolVar(&gbhAsc, "asc", false, "Order by")

}

func execListBlockHeaders(cmd *cobra.Command, args []string) {
	var blockHash []byte
	var err error

	if cmd.Flags().Changed("hash") == true {
		blockHash, err = base58.Decode(gbhHash)
		if err != nil {
			cmd.Printf("Failed: %s", err.Error())
			return
		}
	} else if cmd.Flags().Changed("height") == false {
		cmd.Printf("Error: required flag(s) \"hash\" or \"height\"not set")
		return
	}

	uparams := &types.ListParams{
		Hash:   blockHash,
		Height: uint64(gbhHeight),
		Size:   uint32(gbhSize),
		Offset: uint32(gbhOffset),
		Asc:    gbhAsc,
	}

	msg, err := client.ListBlockHeaders(context.Background(), uparams)
	if err != nil {
		cmd.Printf("Failed: %s", err.Error())
		return
	}
	cmd.Println(util.JSON(msg))
}
