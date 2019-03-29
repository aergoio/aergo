/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"strings"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var revert bool
var election string

func init() {
	rootCmd.AddCommand(voteStatCmd)
	voteStatCmd.Flags().StringVar(&address, "address", "", "address of account")
	voteStatCmd.MarkFlagRequired("address")
	rootCmd.AddCommand(bpCmd)
	bpCmd.Flags().Uint64Var(&number, "count", 23, "the number of elected")
	rootCmd.AddCommand(paramCmd)
	paramCmd.Flags().StringVar(&election, "election", "bp", "block chain parameter")
}

var voteStatCmd = &cobra.Command{
	Use:   "votestat",
	Short: "show voting stat",
	Run:   execVoteStat,
}

var voteCmd = &cobra.Command{
	Use:   "vote",
	Short: "vote to BPs",
	Run:   execVote,
}

var bpCmd = &cobra.Command{
	Use:   "bp",
	Short: "show BP list",
	Run:   execBP,
}

var paramCmd = &cobra.Command{
	Use:   "param",
	Short: "show given parameter status",
	Run:   execParam,
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
	var ci types.CallInfo
	switch strings.ToLower(election) {
	case "bp":
		ci.Name = types.VoteBP
		err = json.Unmarshal([]byte(to), &ci.Args)
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}

		for i, v := range ci.Args {
			if i >= types.MaxCandidates {
				cmd.Println("too many candidates")
				return
			}
			candidate, err := base58.Decode(v.(string))
			if err != nil {
				cmd.Printf("Failed: %s (%s)\n", err.Error(), v)
				return
			}
			_, err = peer.IDFromBytes(candidate)
			if err != nil {
				cmd.Printf("Failed: %s (%s)\n", err.Error(), v)
				return
			}
		}
	case "numofbp",
		"gasprice",
		"nameprice",
		"minimumstaking":
		ci.Name = getVoteCmd(election)
		numberArg, ok := new(big.Int).SetString(to, 10)
		if !ok {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		ci.Args = append(ci.Args, numberArg.String())

	default:
		cmd.Printf("Failed: Wrong election\n")
		return
	}
	txs := make([]*types.Tx, 1)

	state, err := client.GetState(context.Background(),
		&types.SingleBytes{Value: account})
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	payload, err := json.Marshal(ci)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	txs[0] = &types.Tx{
		Body: &types.TxBody{
			Account:   account,
			Recipient: []byte(aergosystem),
			Payload:   payload,
			GasLimit:  0,
			Type:      types.TxType_GOVERNANCE,
			Nonce:     state.GetNonce() + 1,
		},
	}
	//cmd.Println(string(payload))
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
	cmd.Println(util.JSON(msg))
}

func execVoteStat(cmd *cobra.Command, args []string) {
	rawAddr, err := types.DecodeAddress(address)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	msg, err := client.GetAccountVotes(context.Background(), &types.AccountAddress{Value: rawAddr})
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	cmd.Println(util.JSON(msg))
}

func execBP(cmd *cobra.Command, args []string) {
	msg, err := client.GetVotes(context.Background(), &types.VoteParams{
		Count: uint32(number),
		Id:    types.VoteBP[2:],
	})
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	cmd.Println("[")
	comma := ","
	for i, r := range msg.GetVotes() {
		cmd.Printf("{\"" + base58.Encode(r.Candidate) + "\":" + r.GetAmountBigInt().String() + "}")
		if i+1 == len(msg.GetVotes()) {
			comma = ""
		}
		cmd.Println(comma)
	}
	cmd.Println("]")
}

func getVoteCmd(param string) string {
	numberVote := map[string]string{
		"numofbp":        types.VoteNumBP,
		"gasprice":       types.VoteGasPrice,
		"nameprice":      types.VoteNamePrice,
		"minimumstaking": types.VoteMinStaking,
	}
	return numberVote[election]
}

func execParam(cmd *cobra.Command, args []string) {
	id := getVoteCmd(election)
	if len(id) == 0 {
		cmd.Printf("Failed: unsupported parameter : %s\n", election)
		return
	}
	msg, err := client.GetVotes(context.Background(), &types.VoteParams{
		Count: uint32(number),
		Id:    id[2:],
	})
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	cmd.Println("[")
	comma := ","
	for i, r := range msg.GetVotes() {
		value, _ := new(big.Int).SetString(string(r.Candidate), 10)
		cmd.Printf("{\"" + value.String() + "\":" + r.GetAmountBigInt().String() + "}")
		if i+1 == len(msg.GetVotes()) {
			comma = ""
		}
		cmd.Println(comma)
	}
	cmd.Println("]")
}
