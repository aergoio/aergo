/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/binary"
	"encoding/json"

	"github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
)

var (
	serverinfoCmd = &cobra.Command{
		Use:   "serverinfo",
		Short: "Show configs and status of server",
		Args:  cobra.MinimumNArgs(0),
		Run:   execServerInfo,
	}
)

func init() {
	rootCmd.AddCommand(serverinfoCmd)
}

func execServerInfo(cmd *cobra.Command, args []string) {
	var b []byte
	var params types.KeyParams

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(number))

	msg, err := client.GetServerInfo(context.Background(), &params)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	buf, err := json.MarshalIndent(msg, "", " ")
	if err != nil {
		cmd.Printf("Failed: invalid server response %s\n", err.Error())
		return
	}
	cmd.Printf("%s\n", string(buf))
}
