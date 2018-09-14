/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/binary"
	"fmt"

	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/mr-tron/base58/base58"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var from string
var to string
var revert bool

func init() {
	rootCmd.AddCommand(voteCmd)
	voteCmd.Flags().StringVar(&from, "from", "", "base58 address of voter")
	voteCmd.MarkFlagRequired("from")
	voteCmd.Flags().StringVar(&to, "to", "", "base58 address of candidate")
	voteCmd.MarkFlagRequired("to")
	voteCmd.Flags().Uint64Var(&amount, "amount", 0, "amount address")
	voteCmd.Flags().BoolVar(&revert, "revert", false, "revert to vote")

	rootCmd.AddCommand(voteStatCmd)
	voteStatCmd.Flags().Uint64Var(&number, "count", 1, "the number of elected")
}

var voteCmd = &cobra.Command{
	Use:   "vote",
	Short: "vote to agenda",
	Run:   execVote,
}

func execVote(cmd *cobra.Command, args []string) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var client *util.ConnClient
	var ok bool
	if client, ok = util.GetClient(GetServerAddress(), opts).(*util.ConnClient); !ok {
		panic("Internal error. wrong RPC client type")
	}
	defer client.Close()

	account, err := base58.Decode(from)
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return
	}
	candidate, err := base58.Decode(to)
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return
	}
	_, err = peer.IDFromBytes(candidate)
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return
	}

	payload := make([]byte, len(candidate)+1)
	if revert {
		payload[0] = 'r'
	} else {
		payload[0] = 'v'
	}
	copy(payload[1:], candidate)

	txs := make([]*types.Tx, 1)

	state, err := client.GetState(context.Background(),
		&types.SingleBytes{Value: account})
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return
	}

	txs[0] = &types.Tx{
		Body: &types.TxBody{
			Account:   account,
			Recipient: []byte("aergo.bp"),
			Amount:    uint64(amount),
			Price:     0,
			Payload:   payload,
			Limit:     0,
			Type:      types.TxType_GOVERNANCE,
			Nonce:     state.GetNonce() + 1,
		},
	}

	tx, err := client.SignTX(context.Background(), txs[0])
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return
	}
	txs[0] = tx

	msg, err := client.CommitTX(context.Background(), &types.TxList{Txs: txs})
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return
	}

	for _, r := range msg.Results {
		fmt.Println("voting hash :", util.EncodeB64(r.Hash), r.Error)
		return
	}
}

var voteStatCmd = &cobra.Command{
	Use:   "votestat",
	Short: "show voting stat",
	Run:   execVoteStat,
}

func execVoteStat(cmd *cobra.Command, args []string) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var client *util.ConnClient
	var ok bool
	if client, ok = util.GetClient(GetServerAddress(), opts).(*util.ConnClient); !ok {
		panic("Internal error. wrong RPC client type")
	}
	defer client.Close()
	var voteQuery []byte
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(number))
	voteQuery = b
	msg, err := client.GetVotes(context.Background(), &types.SingleBytes{Value: voteQuery})
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return
	}
	for i, r := range msg.GetVotes() {
		fmt.Println(i+1, " : ", base58.Encode(r.Candidate), " : ", r.Amount)
	}
}
