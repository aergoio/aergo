package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	"github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var addresses string

func init() {
	mempoolCmd := &cobra.Command{
		Use:               "mempool [flags] subcommand",
		Short:             "Mempool command",
		PersistentPreRun:  preConnectAergo,
		PersistentPostRun: disconnectAergo,
	}
	mempoolCmd.PersistentFlags().StringVarP(&sock, "sock", "s", "",
		"Unix domain socket file path to connect an aergo server (required)")
	mempoolCmd.MarkPersistentFlagRequired("sock")
	getCmd.Flags().StringVar(&addresses, "addresses", "", "comma separated address list")

	mempoolCmd.AddCommand(statCmd, getCmd)
	rootCmd.AddCommand(mempoolCmd)
}

var statCmd = &cobra.Command{
	Use:   "stat [flags]",
	Short: "Return mempool tx statistics",
	Run: func(cmd *cobra.Command, args []string) {
		r, err := admClient.MempoolTxStat(context.Background(), &types.Empty{})
		if err != nil {
			log.Fatalf("failed to execute: %v", err)
		}
		fmt.Println(string(r.Value))
	},
}

var getCmd = &cobra.Command{
	Use:   "get [flags]",
	Short: "Return address-wise transaction IDs from mempool",
	Run: func(cmd *cobra.Command, args []string) {
		getAccountList := func(in string) *types.AccountList {
			if len(in) == 0 {
				return &types.AccountList{}
			}
			addrs := strings.Split(in, ",")
			al := &types.AccountList{Accounts: make([]*types.Account, len(addrs))}
			for i, a := range addrs {
				addr, err := types.DecodeAddress(strings.TrimSpace(a))
				if err != nil {
					log.Printf("skip invalid address: %s", a)
					continue
				}
				al.Accounts[i] = &types.Account{Address: addr}
			}
			return al
		}

		r, err := admClient.MempoolTx(context.Background(), getAccountList(addresses))
		if err != nil {
			log.Fatalf("failed to execute: %v", err)
		}
		fmt.Println(string(r.Value))
	},
}

func newAergoAdminClient(sockPath string) types.AdminRPCServiceClient {
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024 * 256)),
		grpc.WithInsecure(),
	}
	return types.NewAdminRPCServiceClient(
		util.GetConn(fmt.Sprintf("unix:%s", sockPath), opts),
	)
}
