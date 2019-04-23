package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	accountCmd := &cobra.Command{
		Use:               "account [flags] subcommand",
		Short:             "Account command",
		PersistentPreRun:  preConnectAergo,
		PersistentPostRun: disconnectAergo,
	}

	newCmd.Flags().StringVar(&pw, "password", "", "Password")
	newCmd.Flags().StringVar(&dataDir, "path", "$HOME/.aergo/data", "Path to data directory")

	listCmd.Flags().StringVar(&dataDir, "path", "$HOME/.aergo/data", "Path to data directory")

	unlockCmd.Flags().StringVar(&address, "address", "", "Address of account")
	unlockCmd.MarkFlagRequired("address")
	unlockCmd.Flags().StringVar(&pw, "password", "", "Password")

	lockCmd.Flags().StringVar(&address, "address", "", "Address of account")
	lockCmd.MarkFlagRequired("address")
	lockCmd.Flags().StringVar(&pw, "password", "", "Password")

	importCmd.Flags().StringVar(&importFormat, "if", "", "Base58 import format string")
	importCmd.MarkFlagRequired("if")
	importCmd.Flags().StringVar(&pw, "password", "", "Password when exporting")
	importCmd.Flags().StringVar(&to, "newpassword", "", "Password to be reset")
	importCmd.Flags().StringVar(&dataDir, "path", "$HOME/.aergo/data", "Path to data directory")

	exportCmd.Flags().StringVar(&address, "address", "", "Address of account")
	exportCmd.MarkFlagRequired("address")
	exportCmd.Flags().StringVar(&pw, "password", "", "Password")
	exportCmd.Flags().StringVar(&dataDir, "path", "$HOME/.aergo/data", "Path to data directory")

	voteCmd.Flags().StringVar(&address, "address", "", "Account address of voter")
	voteCmd.MarkFlagRequired("address")
	voteCmd.Flags().StringVar(&to, "to", "", "Json string array which has candidates or input file path")
	voteCmd.MarkFlagRequired("to")
	voteCmd.Flags().StringVar(&voteId, "id", types.VoteBP, "id to vote")

	stakeCmd.Flags().StringVar(&address, "address", "", "Account address")
	stakeCmd.MarkFlagRequired("address")
	stakeCmd.Flags().StringVar(&amount, "amount", "0", "Amount of staking")
	stakeCmd.MarkFlagRequired("amount")
	unstakeCmd.Flags().StringVar(&address, "address", "", "Account address")
	unstakeCmd.MarkFlagRequired("address")
	unstakeCmd.Flags().StringVar(&amount, "amount", "0", "Amount of staking")
	unstakeCmd.MarkFlagRequired("amount")

	accountCmd.AddCommand(newCmd, listCmd, unlockCmd, lockCmd, importCmd, exportCmd, voteCmd, stakeCmd, unstakeCmd)
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
			param.Passphrase, err = getPasswd(cmd, true)
			if err != nil {
				cmd.Printf("Failed get password: %s\n", err.Error())
				return
			}
		}
		var msg *types.Account
		var addr []byte
		if cmd.Flags().Changed("path") == false {
			msg, err = client.CreateAccount(context.Background(), &param)
		} else {
			dataEnvPath := os.ExpandEnv(dataDir)
			ks := key.NewStore(dataEnvPath, 0)
			defer ks.CloseStore()
			addr, err = ks.CreateKey(param.Passphrase)
			if err != nil {
				cmd.Printf("Failed: %s\n", err.Error())
				return
			}
			err = ks.SaveAddress(addr)
		}
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		if msg != nil {
			cmd.Println(types.EncodeAddress(msg.GetAddress()))
		} else {
			cmd.Println(types.EncodeAddress(addr))
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
		if cmd.Flags().Changed("path") == false {
			msg, err = client.GetAccounts(context.Background(), &types.Empty{})
		} else {
			dataEnvPath := os.ExpandEnv(dataDir)
			ks := key.NewStore(dataEnvPath, 0)
			defer ks.CloseStore()
			addrs, err = ks.GetAddresses()
		}
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
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
		cmd.Println(out)
	},
}

var lockCmd = &cobra.Command{
	Use:   "lock [flags]",
	Short: "Lock account in the node",
	Run: func(cmd *cobra.Command, args []string) {
		param, err := parsePersonalParam(cmd)
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		msg, err := client.LockAccount(context.Background(), param)
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		cmd.Println(types.EncodeAddress(msg.GetAddress()))
	},
}

var unlockCmd = &cobra.Command{
	Use:   "unlock [flags]",
	Short: "Unlock account in the node",
	Run: func(cmd *cobra.Command, args []string) {
		param, err := parsePersonalParam(cmd)
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		msg, err := client.UnlockAccount(context.Background(), param)
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		cmd.Println(types.EncodeAddress(msg.GetAddress()))
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
			cmd.Printf("Failed to decode input: %s\n", err.Error())
			return
		}
		wif := &types.ImportFormat{Wif: &types.SingleBytes{Value: importBuf}}
		if pw != "" {
			wif.Oldpass = pw
		} else {
			wif.Oldpass, err = getPasswd(cmd, false)
			if err != nil {
				cmd.Printf("Failed: %s\n", err.Error())
				return
			}
		}

		if to != "" {
			wif.Newpass = to
		} else {
			wif.Newpass = wif.Oldpass
		}

		if cmd.Flags().Changed("path") == false {
			msg, errRemote := client.ImportAccount(context.Background(), wif)
			if errRemote != nil {
				cmd.Printf("Failed: %s\n", errRemote.Error())
				return
			}
			address = msg.GetAddress()
		} else {
			dataEnvPath := os.ExpandEnv(dataDir)
			ks := key.NewStore(dataEnvPath, 0)
			defer ks.CloseStore()
			address, err = ks.ImportKey(importBuf, wif.Oldpass, wif.Newpass)
			if err != nil {
				cmd.Printf("Failed: %s\n", err.Error())
				return
			}
		}
		cmd.Println(types.EncodeAddress(address))
	},
}

var exportCmd = &cobra.Command{
	Use:   "export [flags]",
	Short: "Export account",
	Run: func(cmd *cobra.Command, args []string) {
		param, err := parsePersonalParam(cmd)
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		var result []byte
		if cmd.Flags().Changed("path") == false {
			msg, err := client.ExportAccount(context.Background(), param)
			if err != nil {
				cmd.Printf("Failed: %s\n", err.Error())
				return
			}
			result = msg.Value
		} else {
			dataEnvPath := os.ExpandEnv(dataDir)
			ks := key.NewStore(dataEnvPath, 0)
			defer ks.CloseStore()
			wif, err := ks.ExportKey(param.Account.Address, param.Passphrase)
			if err != nil {
				cmd.Printf("Failed: %s\n", err.Error())
				return
			}
			result = wif
		}
		cmd.Println(types.EncodePrivKey(result))
	},
}

func parsePersonalParam(cmd *cobra.Command) (*types.Personal, error) {
	var err error
	param := &types.Personal{Account: &types.Account{}}
	if address != "" {
		param.Account.Address, err = types.DecodeAddress(address)
		if err != nil {
			return nil, err
		}
		if pw != "" {
			param.Passphrase = pw
		} else {
			param.Passphrase, err = getPasswd(cmd, false)
			if err != nil {
				return nil, err
			}
		}
	}
	return param, nil
}

func getPasswd(cmd *cobra.Command, isNew bool) (string, error) {
	cmd.Print("Enter Password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	cmd.Println("")
	if err != nil {
		return "", err
	}
	if isNew {
		cmd.Print("Repeat Password: ")
		repeat, err := terminal.ReadPassword(int(syscall.Stdin))
		cmd.Println("")
		if err != nil {
			return "", err
		}
		if !bytes.Equal(password, repeat) {
			return "", errors.New("Password not matched")
		}
	}
	return string(password), err
}

func preConnectAergo(cmd *cobra.Command, args []string) {
	if cmd.Flags().Changed("path") == false {
		connectAergo(cmd, args)
	} else {
		client = nil
	}
}
