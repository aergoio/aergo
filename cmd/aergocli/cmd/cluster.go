package cmd

import (
	"context"
	"github.com/aergoio/aergo/cmd/aergocli/util"
	aergorpc "github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var (
	nodename string
	nodeid   string
	url      string
	peerid   string
)

func init() {
	clusterCmd := &cobra.Command{
		Use:   "cluster [flags] subcommand",
		Short: "Cluster command for raft consensus",
	}

	addCmd.Flags().StringVar(&nodename, "name", "", "node name to add to the cluster")
	addCmd.MarkFlagRequired("name")
	addCmd.Flags().StringVar(&url, "url", "", "node url to add to the cluster")
	addCmd.MarkFlagRequired("url")
	addCmd.Flags().StringVar(&peerid, "peerid", "", "peer id of node to add to the cluster")
	addCmd.MarkFlagRequired("peerid")

	removeCmd.Flags().StringVar(&nodeid, "id", "", "node id to add to the cluster")
	removeCmd.MarkFlagRequired("id")

	clusterCmd.AddCommand(addCmd, removeCmd)
	rootCmd.AddCommand(clusterCmd)
}

var addCmd = &cobra.Command{
	Use:   "add [flags]",
	Short: "add new member node to cluster. This command can only be used for raft consensus.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(nodename) == 0 || len(url) == 0 || len(peerid) == 0 {
			cmd.Printf("Failed: name, len, peerid flag must have value\n")
			return
		}

		var changeReq = &aergorpc.MembershipChange{
			Type: aergorpc.MembershipChangeType_ADD_MEMBER,
			Cmd: &aergorpc.MembershipChange_Add{
				Add: &aergorpc.MemberAdd{Name: nodename, Url: url, PeerID: peerid},
			},
		}
		reply, err := client.ChangeMembership(context.Background(), changeReq)
		if err != nil {
			cmd.Printf("Failed to add member: %s\n", err.Error())
		}

		cmd.Println(util.JSON(reply))
		return
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove [flags]",
	Short: "Get account list in the node or cli",
	Run: func(cmd *cobra.Command, args []string) {
		if len(nodeid) == 0 {
			cmd.Printf("Failed: id flag must have value\n")
			return
		}

		changeReq := &aergorpc.MembershipChange{
			Type: aergorpc.MembershipChangeType_REMOVE_MEMBER,
			Cmd: &aergorpc.MembershipChange_Remove{
				Remove: &aergorpc.MemberRemove{NodeID: name},
			},
		}
		reply, err := client.ChangeMembership(context.Background(), changeReq)
		if err != nil {
			cmd.Printf("Failed to remove member: %s\n", err.Error())
		}

		cmd.Println(util.JSON(reply))
		return
	},
}
