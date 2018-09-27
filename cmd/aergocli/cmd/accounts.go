package cmd

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

func init() {
	accountCmd := &cobra.Command{
		Use:               "account [flags] subcommand",
		Short:             "account command",
		PersistentPreRun:  preConnectAergo,
		PersistentPostRun: disconnectAergo,
	}

	newCmd.Flags().StringVar(&pw, "password", "", "password")
	newCmd.Flags().BoolVar(&remote, "remote", true, "choose account in the remote node or not")
	newCmd.Flags().StringVar(&dataDir, "path", "$HOME/.aergo/data", "path to data directory")

	allCmd.Flags().BoolVar(&remote, "remote", true, "choose account in the remote node or not")
	allCmd.Flags().StringVar(&dataDir, "path", "$HOME/.aergo/data", "path to data directory")

	unlockCmd.Flags().StringVar(&address, "address", "", "address of account")
	unlockCmd.MarkFlagRequired("address")
	unlockCmd.Flags().StringVar(&pw, "password", "", "password")

	lockCmd.Flags().StringVar(&address, "address", "", "address of account")
	lockCmd.MarkFlagRequired("address")
	lockCmd.Flags().StringVar(&pw, "password", "", "password")

	accountCmd.AddCommand(newCmd, allCmd, unlockCmd, lockCmd)
	rootCmd.AddCommand(accountCmd)
}

var newCmd = &cobra.Command{
	Use:   "new [flags]",
	Short: "Create new account in the node or cli",
	Run: func(cmd *cobra.Command, args []string) {
		var param types.Personal
		var err error
		if pw != "" {
			param.Passphrase = pw
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
			msg, err = client.CreateAccount(context.Background(), &param)
		} else {
			dataEnvPath := os.ExpandEnv(dataDir)
			ks := key.NewStore(dataEnvPath)
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
				fmt.Println(types.EncodeAddress(msg.GetAddress()))
			} else {
				fmt.Println(types.EncodeAddress(addr))
			}
		}
	},
}

var allCmd = &cobra.Command{
	Use:   "all [flags]",
	Short: "Get all account list in the node or cli",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var msg *types.AccountList
		var addrs [][]byte
		if remote {
			msg, err = client.GetAccounts(context.Background(), &types.Empty{})
		} else {
			dataEnvPath := os.ExpandEnv(dataDir)
			ks := key.NewStore(dataEnvPath)
			addrs, err = ks.GetAddresses()
		}
		if nil == err {
			out := fmt.Sprintf("%s", "[")
			if msg != nil {
				addresslist := msg.GetAccounts()
				for _, a := range addresslist {
					out = fmt.Sprintf("%s%s, ", out, types.EncodeAddress(a.Address))
				}
				if addresslist != nil {
					out = out[:len(out)-2]
				}
			} else if addrs != nil {
				for _, a := range addrs {
					out = fmt.Sprintf("%s%s, ", out, types.EncodeAddress(a))
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

var lockCmd = &cobra.Command{
	Use:   "lock [flags]",
	Short: "Lock account in the node",
	Run: func(cmd *cobra.Command, args []string) {
		param, err := parsePersonalParam()
		if err != nil {
			return
		}
		msg, err := client.LockAccount(context.Background(), param)
		if err == nil {
			fmt.Println(types.EncodeAddress(msg.GetAddress()))
		} else {
			fmt.Printf("Failed: %s\n", err.Error())
		}
	},
}

var unlockCmd = &cobra.Command{
	Use:   "unlock [flags]",
	Short: "Unlock account in the node",
	Run: func(cmd *cobra.Command, args []string) {
		param, err := parsePersonalParam()
		if err != nil {
			return
		}
		msg, err := client.UnlockAccount(context.Background(), param)
		if nil == err {
			fmt.Println(types.EncodeAddress(msg.GetAddress()))
		} else {
			fmt.Printf("Failed: %s\n", err.Error())
		}
	},
}

func parsePersonalParam() (*types.Personal, error) {
	var err error
	param := &types.Personal{Account: &types.Account{}}
	if address != "" {
		param.Account.Address, err = types.DecodeAddress(address)
		if err != nil {
			fmt.Printf("Failed: %s\n", err.Error())
			return nil, err
		}
		if pw != "" {
			param.Passphrase = pw
		} else {
			param.Passphrase, err = getPasswd()
			if err != nil {
				fmt.Printf("Failed: %s\n", err.Error())
				return nil, err
			}
		}
	}
	return param, nil
}

func getPasswd() (string, error) {
	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	return string(bytePassword), err
}

func preConnectAergo(cmd *cobra.Command, args []string) {
	if remote {
		connectAergo(cmd, args)
	}
}
