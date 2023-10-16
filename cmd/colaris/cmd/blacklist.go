/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"strconv"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	"github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
)

var confCommand = &cobra.Command{
	Use:   "config <command>",
	Short: "commands for config",
}

var blacklistCmd = &cobra.Command{
	Use:   "blacklist <subcommand>",
	Short: "blacklist command",
}

var listSubCommand = &cobra.Command{
	Use:   "show",
	Short: "list entries subcommand",
	Run:   listBlacklistEntries,
}

var addSubCommand = &cobra.Command{
	Use:   "add <flags>",
	Short: "add list entry",
	Run:   addBlacklistEntry,
}

var removeSubCommand = &cobra.Command{
	Use:   "rm <idx>",
	Short: "remove list entry",
	Run:   removeBlacklistEntry,
}

func init() {
	rootCmd.AddCommand(confCommand)
	confCommand.AddCommand(blacklistCmd)

	blacklistCmd.AddCommand(listSubCommand)
	blacklistCmd.AddCommand(addSubCommand)
	blacklistCmd.AddCommand(removeSubCommand)

	addSubCommand.Flags().StringVarP(&addAddr, "address", "A", "", "ip address to block")
	addSubCommand.Flags().StringVarP(&addCidr, "cidr", "C", "", "cidr formatted ip range to block")
	addSubCommand.Flags().StringVarP(&addPid, "peerid", "I", "", "peer id to block")
}

func listBlacklistEntries(cmd *cobra.Command, args []string) {
	var err error
	uparams := &types.Empty{}

	msg, err := client.ListBLEntries(context.Background(), uparams)
	if err != nil {
		cmd.Printf("Failed: %s", err.Error())
		return
	}

	cmd.Println(util.JSON(msg))
}

func addBlacklistEntry(cmd *cobra.Command, args []string) {
	var err error

	if len(addPid) == 0 && len(addCidr) == 0 && len(addAddr) == 0 {
		cmd.Printf("Failed: at least one flags is required. ")
		return
	} else if len(addCidr) > 0 && len(addAddr) > 0 {
		cmd.Printf("Failed: either address or cidr is allowed, not both ")
		return
	}
	uparams := &types.AddEntryParams{
		PeerID:  addPid,
		Address: addAddr,
		Cidr:    addCidr,
	}

	msg, err := client.AddBLEntry(context.Background(), uparams)
	if err != nil {
		desc := "unknown error"
		if msg != nil {
			desc = msg.Value
		}
		cmd.Printf("Failed: %s, %s", err.Error(), desc)
		return
	}

	cmd.Println("Success ")
}

func removeBlacklistEntry(cmd *cobra.Command, args []string) {
	var err error

	if len(args) == 0 {
		cmd.Printf("Failed: index is required.")
		return
	}

	idx, err := strconv.Atoi(args[0])
	if err != nil {
		cmd.Printf("Failed: invalid index: %v", args[0])
		return
	}
	uparams := &types.RmEntryParams{
		Index: uint32(idx),
	}

	_, err = client.RemoveBLEntry(context.Background(), uparams)
	if err != nil {
		cmd.Printf("Failed: %s \n", err.Error())
		return
	}

	cmd.Println("Success ")
}
