package cmd

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chaininfoCmd)
}

var chaininfoCmd = &cobra.Command{
	Use:   "chaininfo",
	Short: "Print current blockchain information",
	Run: func(cmd *cobra.Command, args []string) {
		msg, err := client.GetChainInfo(context.Background(), &types.Empty{})
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		cmd.Println(convChainInfoMsg(msg))
	},
}

type printChainId struct {
	Magic       string
	Public      bool
	Mainnet     bool
	CoinbaseFee string
	Consensus   string
}

type printChainInfo struct {
	Chainid        printChainId
	BpNumber       uint32
	MaxBlockSize   uint64
	MaxTokens      string
	StakingMinimum string
}

func convChainInfoMsg(msg *types.ChainInfo) string {
	out := &printChainInfo{}
	out.Chainid.Magic = msg.Chainid.Magic
	out.Chainid.Public = msg.Chainid.Public
	out.Chainid.Mainnet = msg.Chainid.Mainnet
	out.Chainid.CoinbaseFee = new(big.Int).SetBytes(msg.Chainid.Coinbasefee).String()
	out.Chainid.Consensus = msg.Chainid.Consensus
	out.BpNumber = msg.Bpnumber
	out.MaxBlockSize = msg.Maxblocksize
	out.MaxTokens = new(big.Int).SetBytes(msg.Maxtokens).String()
	out.StakingMinimum = new(big.Int).SetBytes(msg.Stakingminimum).String()
	jsonout, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		return ""
	}
	return string(jsonout)
}
