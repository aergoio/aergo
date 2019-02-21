/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"log"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	aergorpc "github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var contractAddress string
var eventName string
var argFilter string
var start uint64
var end uint64
var desc bool

func init() {
	eventCmd := &cobra.Command{
		Use:   "event [flags] subcommand",
		Short: "Get event",
	}

	listCmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "list event",
		Args:  cobra.MinimumNArgs(0),
		Run:   execListEvent,
	}
	listCmd.Flags().Uint64Var(&start, "start", 0, "start block number")
	listCmd.Flags().Uint64Var(&end, "end", 0, "end block number")
	listCmd.Flags().StringVarP(&eventName, "event", "", "", "Event Name")
	listCmd.Flags().StringVarP(&contractAddress, "address", "", "", "Contract Address")
	listCmd.Flags().BoolVar(&desc, "desc", false, "descending order")
	listCmd.Flags().StringVarP(&argFilter, "argfilter", "", "", "argument filter")
	listCmd.MarkFlagRequired("address")

	streamCmd := &cobra.Command{
		Use:   "stream [flags]",
		Short: "stream event",
		Args:  cobra.MinimumNArgs(0),
		Run:   execStreamEvent,
	}
	streamCmd.Flags().StringVarP(&contractAddress, "address", "", "", "Contract Address")
	streamCmd.Flags().StringVarP(&eventName, "event", "", "", "Event Name")
	streamCmd.Flags().StringVarP(&argFilter, "argfilter", "", "", "argument filter")
	streamCmd.MarkFlagRequired("address")

	eventCmd.AddCommand(
		listCmd,
		streamCmd,
	)
	rootCmd.AddCommand(eventCmd)
}

func execListEvent(cmd *cobra.Command, args []string) {
	ba, err := aergorpc.DecodeAddress(contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	filter := &aergorpc.FilterInfo{
		Blockfrom:       start,
		Blockto:         end,
		ContractAddress: ba,
		EventName:       eventName,
		Desc:            desc,
		ArgFilter:       []byte(argFilter),
	}

	events, err := client.ListEvents(context.Background(), filter)
	if err != nil {
		cmd.Printf("Failed: %s\n", err.Error())
		return
	}
	for _, ev := range events.GetEvents() {
		cmd.Println(util.JSON(ev))
	}
}

func execStreamEvent(cmd *cobra.Command, args []string) {
	ba, err := aergorpc.DecodeAddress(contractAddress)
	if err != nil {
		log.Fatal(err)
	}
	filter := &aergorpc.FilterInfo{
		ContractAddress: ba,
		EventName:       eventName,
		ArgFilter:       []byte(argFilter),
	}

	stream, err := client.ListEventStream(context.Background(), filter)
	if err != nil {
		cmd.Printf("Failed: %s", err.Error())
		return
	}
	for {
		ev, err := stream.Recv()
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		cmd.Println(util.JSON(ev))
	}
}
