package cmd

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/mr-tron/base58/base58"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/aergoio/aergo/account/keystore"
	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func init() {
	rootCmd.AddCommand(newAccountCmd)
	newAccountCmd.Flags().BoolVar(&remote, "remote", false, "choose account in the remote node or not")
	newAccountCmd.Flags().StringVar(&dataDir, "path", "$HOME/.aergo/data", "path to data directory")
	rootCmd.AddCommand(getAccountsCmd)
	getAccountsCmd.Flags().BoolVar(&remote, "remote", false, "choose account in the remote node or not")
	getAccountsCmd.Flags().StringVar(&dataDir, "path", "$HOME/.aergo/data", "path to data directory")
	rootCmd.AddCommand(lockAccountsCmd)
	rootCmd.AddCommand(unlockAccountsCmd)
}

var remote bool
var dataDir string
var newAccountCmd = &cobra.Command{
	Use:   "newaccount",
	Short: "Create new account in the node",
	Run: func(cmd *cobra.Command, args []string) {

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

		var msg *types.Account
		var addr []byte
		if remote {
			serverAddr := GetServerAddress()
			opts := []grpc.DialOption{grpc.WithInsecure()}
			var client *util.ConnClient
			var ok bool
			if client, ok = util.GetClient(serverAddr, opts).(*util.ConnClient); !ok {
				panic("Internal error. wrong RPC client type")
			}
			defer client.Close()

			msg, err = client.CreateAccount(context.Background(), &param)
		} else {
			dataEnvPath := os.ExpandEnv(dataDir)
			ks := keystore.NewKeyStore(dataEnvPath)
			addr, err = ks.CreateKey(param.Passphrase)
			if nil != err {
				fmt.Printf("Failed: %s\n", err.Error())
			}
			err = ks.SaveAddress(addr)
		}
		if nil != err {
			fmt.Printf("Failed: %s\n", err.Error())
		} else {
			if msg != nil {
				fmt.Println(base58.Encode(msg.GetAddress()))
			} else {
				fmt.Println(base58.Encode(addr))
			}
		}
	},
}

var getAccountsCmd = &cobra.Command{
	Use:   "getaccounts",
	Short: "Get account list in the node",
	Run: func(cmd *cobra.Command, args []string) {

		var err error
		var msg *types.AccountList
		var addrs [][]byte
		if remote {
			serverAddr := GetServerAddress()
			opts := []grpc.DialOption{grpc.WithInsecure()}
			var client *util.ConnClient
			var ok bool
			if client, ok = util.GetClient(serverAddr, opts).(*util.ConnClient); !ok {
				panic("Internal error. wrong RPC client type")
			}
			defer client.Close()

			msg, err = client.GetAccounts(context.Background(), &types.Empty{})

		} else {
			dataEnvPath := os.ExpandEnv(dataDir)
			ks := keystore.NewKeyStore(dataEnvPath)
			addrs, err = ks.GetAddresses()
		}
		if nil == err {
			out := fmt.Sprintf("%s", "[")
			if msg != nil {
				addresslist := msg.GetAccounts()
				for _, a := range addresslist {
					out = fmt.Sprintf("%s%s, ", out, base58.Encode(a.Address))
				}
				if addresslist != nil {
					out = out[:len(out)-2]
				}
			} else if addrs != nil {
				for _, a := range addrs {
					out = fmt.Sprintf("%s%s, ", out, base58.Encode(a))
				}
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
