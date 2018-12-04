/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var revert bool

func init() {
	rootCmd.AddCommand(voteStatCmd)
	voteStatCmd.Flags().StringVar(&address, "address", "", "address of account")
	voteStatCmd.Flags().Uint64Var(&number, "count", 1, "the number of elected")
}

var voteCmd = &cobra.Command{
	Use:   "vote",
	Short: "vote to BPs",
	Run:   execVote,
}

const PeerIDLength = 39

func execVote(cmd *cobra.Command, args []string) {
	account, err := types.DecodeAddress(address)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	_, err = os.Stat(to)
	if err == nil {
		b, readerr := ioutil.ReadFile(to)
		if readerr != nil {
			cmd.Printf("Failed: %s\n", readerr.Error())
			return
		}
		to = string(b)
	}
	var candidates []string
	err = json.Unmarshal([]byte(to), &candidates)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}

	payload := make([]byte, (len(candidates)*PeerIDLength)+1)
	payload[0] = 'v'
	for i, v := range candidates {
		candidate, err := base58.Decode(v)
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		_, err = peer.IDFromBytes(candidate)
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		copy(payload[1+(i*PeerIDLength):], candidate)
	}

	txs := make([]*types.Tx, 1)

	state, err := client.GetState(context.Background(),
		&types.SingleBytes{Value: account})
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}

	txs[0] = &types.Tx{
		Body: &types.TxBody{
			Account:   account,
			Recipient: []byte(aergosystem),
			Payload:   payload,
			Limit:     0,
			Type:      types.TxType_GOVERNANCE,
			Nonce:     state.GetNonce() + 1,
		},
	}
	//TODO : support local
	tx, err := client.SignTX(context.Background(), txs[0])
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	txs[0] = tx

	msg, err := client.CommitTX(context.Background(), &types.TxList{Txs: txs})
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}

	for _, r := range msg.Results {
		cmd.Println("voting hash :", base58.Encode(r.Hash), r.Error, r.Detail)
		return
	}
}

var voteStatCmd = &cobra.Command{
	Use:   "votestat",
	Short: "show voting stat",
	Run:   execVoteStat,
}

func execVoteStat(cmd *cobra.Command, args []string) {
	var rawAddr []byte
	var err error
	if address != "" {
		rawAddr, err = types.DecodeAddress(address)
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
	}

	var msg *types.VoteList
	if rawAddr != nil {
		msg, err = client.GetVotes(context.Background(), &types.SingleBytes{Value: rawAddr})
	} else {
		var voteQuery []byte
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(number))
		voteQuery = b
		msg, err = client.GetVotes(context.Background(), &types.SingleBytes{Value: voteQuery})
	}
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	for i, r := range msg.GetVotes() {
		cmd.Println(i+1, " : ", base58.Encode(r.Candidate), " : ", r.GetAmountBigInt().String())
	}
}
