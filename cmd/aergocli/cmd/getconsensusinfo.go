/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/json"
	aergorpc "github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getConsensusInfoCmd)
}

var getConsensusInfoCmd = &cobra.Command{
	Use:   "getconsensusinfo",
	Short: "Print consensus info",
	Run: func(cmd *cobra.Command, args []string) {
		msg, err := client.GetConsensusInfo(context.Background(), &aergorpc.Empty{})
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}

		type outInfo struct {
			Type string
			Info *json.RawMessage
			Bps  []*json.RawMessage
		}

		var out = &outInfo{}
		out.Type = msg.Type

		//cmd.Println(fmt.Sprintf("consensus:%s, leader:%d", msg.Type, uinfo.Leader))
		infoB := json.RawMessage(msg.Info)
		out.Info = &infoB

		out.Bps = make([]*json.RawMessage, len(msg.Bps))
		for i, bpstr := range msg.Bps {
			b := json.RawMessage([]byte(bpstr))
			out.Bps[i] = &b
		}

		jsonout, err := json.Marshal(out)
		if err != nil {
			cmd.Println("failed decode consensus info")
			return
		}

		cmd.Println(string(jsonout))
	},
}
