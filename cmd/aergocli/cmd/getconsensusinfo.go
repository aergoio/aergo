/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	aergorpc "github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getConsensusInfoCmd)
}

func getConsensusInfo() ([]byte, error) {
	msg, err := client.GetConsensusInfo(context.Background(), &aergorpc.Empty{})
	if err != nil {
		return nil, err
	}

	type outInfo struct {
		Type string             `json:",omitempty"`
		Info *json.RawMessage   `json:",omitempty"`
		Bps  []*json.RawMessage `json:",omitempty"`
	}

	var out = &outInfo{}
	out.Type = msg.Type

	//cmd.Println(fmt.Sprintf("consensus:%s, leader:%d", msg.Type, uinfo.Leader))
	if len(msg.Info) > 0 {
		infoB := json.RawMessage(msg.Info)
		out.Info = &infoB
	}

	if len(msg.Bps) > 0 {
		out.Bps = make([]*json.RawMessage, len(msg.Bps))
		for i, bpstr := range msg.Bps {
			b := json.RawMessage([]byte(bpstr))
			out.Bps[i] = &b
		}
	}

	jsonout, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("decode consensus info: %v\n", err)
	}

	return jsonout, nil
}

var getConsensusInfoCmd = &cobra.Command{
	Use:   "getconsensusinfo",
	Short: "Print consensus info",
	Run: func(cmd *cobra.Command, args []string) {
		jsonout, err := getConsensusInfo()
		if err != nil {
			cmd.Printf("Failed: %s", err.Error())
			return
		}

		cmd.Println(string(jsonout))
	},
}
