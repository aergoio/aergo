/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	aergorpc "github.com/aergoio/aergo/v2/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var getblockCmd = &cobra.Command{
	Use:   "getblock",
	Short: "Get block information",
	Args:  cobra.NoArgs,
	RunE:  execGetBlock,
}

var stream bool
var number uint64
var hash string

func init() {
	rootCmd.AddCommand(getblockCmd)
	getblockCmd.Flags().Uint64VarP(&number, "number", "n", 0, "block height")
	getblockCmd.Flags().StringVarP(&hash, "hash", "", "", "block hash")
	getblockCmd.Flags().BoolVar(&stream, "stream", false, "continiously stream new blocks as they get created")
}

func streamBlocks(cmd *cobra.Command) error {
	bs, err := client.ListBlockStream(context.Background(), &aergorpc.Empty{})
	if err != nil {
		return fmt.Errorf("failed to connect stream: %v", err)
	}
	for {
		b, err := bs.Recv()
		if err != nil {
			return fmt.Errorf("failed to receive block: %v", err)
		}
		cmd.Println(util.BlockConvBase58Addr(b))
	}
}

func getSingleBlock(cmd *cobra.Command) error {
	var blockQuery []byte
	if hash == "" {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(number))
		blockQuery = b
	} else {
		decoded, err := base58.Decode(hash)
		if err != nil {
			return fmt.Errorf("failed to decode block hash: %v", err)
		}
		if len(decoded) == 0 {
			return fmt.Errorf("decoded block hash is empty")
		}
		blockQuery = decoded
	}

	msg, err := client.GetBlock(context.Background(), &aergorpc.SingleBytes{Value: blockQuery})
	if err != nil {
		return fmt.Errorf("failed to get block: %v", err)
	}
	cmd.Println(util.BlockConvBase58Addr(msg))
	return nil
}

func execGetBlock(cmd *cobra.Command, args []string) error {
	fflags := cmd.Flags()
	if !stream && fflags.Changed("number") == false && fflags.Changed("hash") == false {
		return errors.New("no block --hash or --number specified")
	}
	cmd.SilenceUsage = true

	if stream {
		return streamBlocks(cmd)
	}
	return getSingleBlock(cmd)
}
