package cmd

import (
	"context"
	"fmt"
	"syscall"

	"github.com/mr-tron/base58/base58"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func init() {
	rootCmd.AddCommand(newAccountCmd)
	rootCmd.AddCommand(getAccountsCmd)
	rootCmd.AddCommand(lockAccountsCmd)
	rootCmd.AddCommand(unlockAccountsCmd)
}

var newAccountCmd = &cobra.Command{
	Use:   "newaccount",
	Short: "Create new account in the node",
	Run: func(cmd *cobra.Command, args []string) {
		serverAddr := GetServerAddress()
		opts := []grpc.DialOption{grpc.WithInsecure()}
		var client *util.ConnClient
		var ok bool
		if client, ok = util.GetClient(serverAddr, opts).(*util.ConnClient); !ok {
			panic("Internal error. wrong RPC client type")
		}
		defer client.Close()

		var param types.Personal
		var err error
		if len(args) > 0 {
			param.Passphrase = args[0]
		} else {
			param.Passphrase, err = getPasswd()
			if err != nil {
				fmt.Printf("Failed: %s\n", err.Error())
				return
			}
		}
		msg, err := client.CreateAccount(context.Background(), &param)
		if nil == err {
			fmt.Println(base58.Encode(msg.GetAddress()))
		} else {
			fmt.Printf("Failed: %s\n", err.Error())
		}
	},
}

var getAccountsCmd = &cobra.Command{
	Use:   "getaccounts",
	Short: "Get account list in the node",
	Run: func(cmd *cobra.Command, args []string) {
		serverAddr := GetServerAddress()
		opts := []grpc.DialOption{grpc.WithInsecure()}
		var client *util.ConnClient
		var ok bool
		if client, ok = util.GetClient(serverAddr, opts).(*util.ConnClient); !ok {
			panic("Internal error. wrong RPC client type")
		}
		defer client.Close()

		msg, err := client.GetAccounts(context.Background(), &types.Empty{})
		if nil == err {
			out := fmt.Sprintf("%s", "[")
			addresslist := msg.GetAccounts()
			for _, a := range addresslist {
				out = fmt.Sprintf("%s%s, ", out, base58.Encode(a.Address))
			}
			if addresslist != nil {
				out = out[:len(out)-2]
			}
			out = fmt.Sprintf("%s%s", out, "]")
			fmt.Println(out)
		} else {
			fmt.Printf("Failed: %s\n", err.Error())
		}
	},
}

var lockAccountsCmd = &cobra.Command{
	Use:   "lockaccount",
	Short: "Lock account in the node",
	Run: func(cmd *cobra.Command, args []string) {
		serverAddr := GetServerAddress()
		opts := []grpc.DialOption{grpc.WithInsecure()}
		var client *util.ConnClient
		var ok bool
		if client, ok = util.GetClient(serverAddr, opts).(*util.ConnClient); !ok {
			panic("Internal error. wrong RPC client type")
		}
		defer client.Close()
		param, err := parsePersonalParam(args)
		if err != nil {
			return
		}
		msg, err := client.LockAccount(context.Background(), param)
		if err == nil {
			fmt.Println(base58.Encode(msg.GetAddress()))
		} else {
			fmt.Printf("Failed: %s\n", err.Error())
		}
	},
}

var unlockAccountsCmd = &cobra.Command{
	Use:   "unlockaccount",
	Short: "Unlock account in the node",
	Run: func(cmd *cobra.Command, args []string) {
		serverAddr := GetServerAddress()
		opts := []grpc.DialOption{grpc.WithInsecure()}
		var client *util.ConnClient
		var ok bool
		if client, ok = util.GetClient(serverAddr, opts).(*util.ConnClient); !ok {
			panic("Internal error. wrong RPC client type")
		}
		defer client.Close()

		param, err := parsePersonalParam(args)
		if err != nil {
			return
		}
		msg, err := client.UnlockAccount(context.Background(), param)
		if nil == err {
			fmt.Println(base58.Encode(msg.GetAddress()))
		} else {
			fmt.Printf("Failed: %s\n", err.Error())
		}
	},
}

func parsePersonalParam(args []string) (*types.Personal, error) {
	var err error
	param := &types.Personal{Account: &types.Account{}}
	if len(args) > 1 {
		param.Account.Address, err = base58.Decode(args[0])
		param.Passphrase = args[1]
	} else {
		param.Account.Address, err = base58.Decode(args[0])
		param.Passphrase, err = getPasswd()
	}
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return nil, err
	}
	return param, nil
}

func getPasswd() (string, error) {
	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	return string(bytePassword), err
}
