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
	"strconv"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var proposalCmd = &cobra.Command{
	Use:   "proposal subcommand",
	Short: "proposal in blockchain",
}

var createProposalCmd = &cobra.Command{
	Use:   "create",
	Short: "create to proposal in blockchain",
	Run:   execProposal,
}

var execProposalCmd = &cobra.Command{
	Use:   "show",
	Short: "show all proposal in blockchain",
	Run:   execProposalShow,
}

var proposal string
var proposalId string
var proposalVersion string

func init() {
	createProposalCmd.Flags().StringVar(&address, "address", "", "An account address of proposal creator")
	createProposalCmd.MarkFlagRequired("address")
	createProposalCmd.Flags().StringVar(&proposal, "json", "", "An proposal in json form")
	createProposalCmd.MarkFlagRequired("json")
	execProposalCmd.Flags().StringVar(&proposalId, "id", "", "An proposal in json form")
	proposalCmd.AddCommand(createProposalCmd)
	proposalCmd.AddCommand(execProposalCmd)
	rootCmd.AddCommand(proposalCmd)
}

func execProposalShow(cmd *cobra.Command, args []string) {
	msg, err := client.GetVotes(context.Background(), &types.VoteParams{
		Id:    string(types.GenProposalKey(proposalId)),
		Count: uint32(number),
	})
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	cmd.Println("[")
	comma := ","
	for i, r := range msg.GetVotes() {
		cmd.Printf("{\"" + string(r.Candidate) + "\":" + r.GetAmountBigInt().String() + "}")
		if i+1 == len(msg.GetVotes()) {
			comma = ""
		}
		cmd.Println(comma)
	}
	cmd.Println("]")
}

func execProposal(cmd *cobra.Command, args []string) {
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

	var arg types.Proposal
	err = json.Unmarshal([]byte(proposal), &arg)
	if err != nil {
		cmd.Printf("Failed: %s (%s)\n", err.Error(), proposal)
		return
	}
	var ci types.CallInfo
	ci.Name = types.CreateProposal
	ci.Args = append(ci.Args, arg.Id,
		strconv.FormatUint(arg.Blockfrom, 10),
		strconv.FormatUint(arg.Blockto, 10),
		strconv.FormatUint(uint64(arg.Maxvote), 10),
		arg.Description, arg.Candidates)
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
