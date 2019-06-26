/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var revert bool
var voteId string
var voteVersion string

func init() {
	rootCmd.AddCommand(voteStatCmd)
	voteStatCmd.Flags().StringVar(&address, "address", "", "address of account")
	voteStatCmd.MarkFlagRequired("address")
	rootCmd.AddCommand(bpCmd)
	bpCmd.Flags().Uint64Var(&number, "count", 0, "the number of elected")
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
	if strings.ToLower(voteId) == strings.ToLower(types.OpvoteBP.Cmd()) {
		ci.Name = types.OpvoteBP.Cmd()
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
			_, err = types.IDFromBytes(candidate)
			if err != nil {
				cmd.Printf("Failed: %s (%s)\n", err.Error(), v)
				return
			}
		}
	} else {
		ci.Name = types.OpvoteProposal.Cmd()
		err := json.Unmarshal([]byte(to), &ci.Args)
		if err != nil {
			cmd.Printf("Failed: %s (%s)\n", err.Error(), to)
			return
		}
		ci.Args = append([]interface{}{voteId}, ci.Args...)
	}
	payload, err := json.Marshal(ci)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	//cmd.Println(string(payload))
	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   account,
			Recipient: []byte(aergosystem),
			Payload:   payload,
			GasLimit:  0,
			Type:      types.TxType_GOVERNANCE,
		},
	}
	//TODO : support local
	msg, err := client.SendTX(context.Background(), tx)
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
		Id:    types.OpvoteBP.Name(),
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
