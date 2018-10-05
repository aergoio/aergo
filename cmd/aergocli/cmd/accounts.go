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

	listCmd.Flags().BoolVar(&remote, "remote", true, "choose account in the remote node or not")
	listCmd.Flags().StringVar(&dataDir, "path", "$HOME/.aergo/data", "path to data directory")

	unlockCmd.Flags().StringVar(&address, "address", "", "address of account")
	unlockCmd.MarkFlagRequired("address")
	unlockCmd.Flags().StringVar(&pw, "password", "", "password")

	lockCmd.Flags().StringVar(&address, "address", "", "address of account")
	lockCmd.MarkFlagRequired("address")
	lockCmd.Flags().StringVar(&pw, "password", "", "password")

	importCmd.Flags().StringVar(&importFormat, "if", "", "base58 import format string")
	importCmd.MarkFlagRequired("if")
	importCmd.Flags().StringVar(&pw, "password", "", "password when exporting")
	importCmd.Flags().StringVar(&to, "newpassword", "", "password to be reset")
	importCmd.Flags().BoolVar(&remote, "remote", true, "choose account in the remote node or not")
	importCmd.Flags().StringVar(&dataDir, "path", "$HOME/.aergo/data", "path to data directory")

	exportCmd.Flags().StringVar(&address, "address", "", "address of account")
	exportCmd.MarkFlagRequired("address")
	exportCmd.Flags().StringVar(&pw, "password", "", "password")
	exportCmd.Flags().BoolVar(&remote, "remote", true, "choose account in the remote node or not")
	exportCmd.Flags().StringVar(&dataDir, "path", "$HOME/.aergo/data", "path to data directory")

	voteCmd.Flags().StringVar(&address, "address", "", "base58 address of voter")
	voteCmd.MarkFlagRequired("address")
	voteCmd.Flags().StringVar(&to, "to", "", "json array which has base58 address of candidates or input file path")
	voteCmd.MarkFlagRequired("to")

	stakingCmd.Flags().StringVar(&address, "address", "", "base58 address")
	stakingCmd.MarkFlagRequired("address")
	stakingCmd.Flags().Uint64Var(&amount, "amount", 0, "amount of staking")
	stakingCmd.MarkFlagRequired("amount")
	unstakingCmd.Flags().StringVar(&address, "address", "", "base58 address")
	unstakingCmd.MarkFlagRequired("address")
	unstakingCmd.Flags().Uint64Var(&amount, "amount", 0, "amount of staking")
	unstakingCmd.MarkFlagRequired("amount")

	accountCmd.AddCommand(newCmd, listCmd, unlockCmd, lockCmd, importCmd, exportCmd, voteCmd, stakingCmd, unstakingCmd)
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
			ks.CloseStore()
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

var listCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "Get account list in the node or cli",
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
			ks.CloseStore()
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

var importCmd = &cobra.Command{
	Use:   "import [flags]",
	Short: "Import account",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var address []byte
		importBuf, err := types.DecodePrivKey(importFormat)
		if err != nil {
			fmt.Printf("Failed to decode input: %s\n", err.Error())
			return
		}
		wif := &types.ImportFormat{Wif: &types.SingleBytes{Value: importBuf}}
		if pw != "" {
			wif.Oldpass = pw
		} else {
			wif.Oldpass, err = getPasswd()
			if err != nil {
				fmt.Printf("Failed: %s\n", err.Error())
				return
			}
		}

		if to != "" {
			wif.Newpass = to
		} else {
			wif.Newpass = pw
		}

		if remote {
			msg, errRemote := client.ImportAccount(context.Background(), wif)
			if nil == errRemote {
				fmt.Println(types.EncodeAddress(msg.GetAddress()))
				return
			}
			err = errRemote
		} else {
			dataEnvPath := os.ExpandEnv(dataDir)
			ks := key.NewStore(dataEnvPath)
			address, err = ks.ImportKey(importBuf, wif.Oldpass, wif.Newpass)
			if nil == err {
				fmt.Println(types.EncodeAddress(address))
				return
			}
			ks.CloseStore()
		}
		fmt.Printf("Failed to import account: %s\n", err.Error())

	},
}

var exportCmd = &cobra.Command{
	Use:   "export [flags]",
	Short: "Export account",
	Run: func(cmd *cobra.Command, args []string) {
		param, err := parsePersonalParam()
		if err != nil {
			return
		}
		var result []byte
		if remote {
			msg, err := client.ExportAccount(context.Background(), param)
			if err != nil {
				fmt.Printf("Failed: %s\n", err.Error())
				return
			}
			result = msg.Value
		} else {
			dataEnvPath := os.ExpandEnv(dataDir)
			ks := key.NewStore(dataEnvPath)
			wif, err := ks.ExportKey(param.Account.Address, param.Passphrase)
			if err != nil {
				fmt.Printf("Failed: %s\n", err.Error())
				return
			}
			result = wif
			ks.CloseStore()
		}
		fmt.Println(types.EncodePrivKey(result))
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
