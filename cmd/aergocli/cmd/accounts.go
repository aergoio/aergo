package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/aergoio/aergo/v2/account/key"
	"github.com/aergoio/aergo/v2/types"
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

	newCmd.Flags().StringVar(&pw, "password", "", "password (optional, will be asked on the terminal if not given)")

	unlockCmd.Flags().StringVar(&address, "address", "", "address of account")
	unlockCmd.MarkFlagRequired("address")
	unlockCmd.Flags().StringVar(&pw, "password", "", "password (optional, will be asked on the terminal if not given)")

	lockCmd.Flags().StringVar(&address, "address", "", "address of account")
	lockCmd.MarkFlagRequired("address")
	lockCmd.Flags().StringVar(&pw, "password", "", "password (optional, will be asked on the terminal if not given)")

	importCmd.Flags().StringVar(&importFormat, "if", "", "Base58 import format string")
	importCmd.Flags().StringVar(&pw, "password", "", "password used when exporting")
	importCmd.Flags().StringVar(&to, "newpassword", "", "new password for storing account")
	importCmd.Flags().StringVar(&importFilePath, "path", "", "path to import keystore file")

	exportCmd.Flags().StringVar(&address, "address", "", "Address of account")
	exportCmd.MarkFlagRequired("address")
	exportCmd.Flags().BoolVar(&exportAsWif, "wif", false, "export as encrypted string instead of keystore format")
	exportCmd.Flags().StringVar(&pw, "password", "", "password (optional, will be asked on the terminal if not given)")

	voteCmd.Flags().StringVar(&address, "address", "", "account address of voter")
	voteCmd.MarkFlagRequired("address")
	voteCmd.Flags().StringVar(&to, "to", "", "json string array which has candidates, or input file path")
	voteCmd.MarkFlagRequired("to")
	voteCmd.Flags().StringVar(&voteId, "id", types.OpvoteBP.Cmd(), "id to vote (e.g. bpcount, stakingmin, gasprice, nameprice)")
	voteCmd.Flags().StringVar(&pw, "password", "", "password (optional, will be asked on the terminal if not given)")

	stakeCmd.Flags().StringVar(&address, "address", "", "account address")
	stakeCmd.MarkFlagRequired("address")
	stakeCmd.Flags().StringVar(&amount, "amount", "0", "amount to stake")
	stakeCmd.MarkFlagRequired("amount")
	stakeCmd.Flags().StringVar(&pw, "password", "", "password (optional, will be asked on the terminal if not given)")

	unstakeCmd.Flags().StringVar(&address, "address", "", "account address")
	unstakeCmd.MarkFlagRequired("address")
	unstakeCmd.Flags().StringVar(&amount, "amount", "0", "amount to unstake")
	unstakeCmd.MarkFlagRequired("amount")
	unstakeCmd.Flags().StringVar(&pw, "password", "", "password (optional, will be asked on the terminal if not given)")

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
				cmd.PrintErrf("Failed get password: %s\n", err.Error())
				return
			}
		}
		var msg *types.Account
		var addr []byte
		if rootConfig.KeyStorePath == "" {
			msg, err = client.CreateAccount(context.Background(), &param)
			if msg != nil {
				addr = msg.GetAddress()
			}
		} else {
			dataEnvPath := os.ExpandEnv(rootConfig.KeyStorePath)
			ks := key.NewStore(dataEnvPath, 0)
			defer ks.CloseStore()
			addr, err = ks.CreateKey(param.Passphrase)
			if err != nil {
				cmd.PrintErrf("Failed: %s\n", err.Error())
				return
			}
		}
		if err != nil {
			cmd.PrintErrf("Failed: %s\n", err.Error())
			return
		}
		cmd.Println(types.EncodeAddress(addr))
	},
}

var listCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "Get account list in the node or cli",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var msg *types.AccountList
		var addrs [][]byte
		if rootConfig.KeyStorePath == "" {
			msg, err = client.GetAccounts(context.Background(), &types.Empty{})
		} else {
			dataEnvPath := os.ExpandEnv(rootConfig.KeyStorePath)
			ks := key.NewStore(dataEnvPath, 0)
			defer ks.CloseStore()
			addrs, err = ks.GetAddresses()
		}
		if err != nil {
			cmd.PrintErrf("Failed: %s\n", err.Error())
			return
		}
		out := fmt.Sprintf("%s", "[")
		if msg != nil {
			addresslist := msg.GetAccounts()
			for _, a := range addresslist {
				out = fmt.Sprintf("%s\"%s\", ", out, types.EncodeAddress(a.Address))
			}
			if addresslist != nil {
				out = out[:len(out)-2]
			}
		} else if addrs != nil && len(addrs) > 0 {
			for _, a := range addrs {
				out = fmt.Sprintf("%s\"%s\", ", out, types.EncodeAddress(a))
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
		if rootConfig.KeyStorePath != "" {
			cmd.PrintErrf("There is no lock on the local keystore.\n")
			return
		}
		param, err := parsePersonalParam(cmd)
		if err != nil {
			cmd.PrintErrf("Failed: %s\n", err.Error())
			return
		}
		msg, err := client.LockAccount(context.Background(), param)
		if err != nil {
			cmd.PrintErrf("Failed: %s\n", err.Error())
			return
		}
		cmd.Println(types.EncodeAddress(msg.GetAddress()))
	},
}

var unlockCmd = &cobra.Command{
	Use:   "unlock [flags]",
	Short: "Unlock account in the node",
	Run: func(cmd *cobra.Command, args []string) {
		if rootConfig.KeyStorePath != "" {
			cmd.PrintErrf("Local keystore do not need to be unlocked.\n")
			return
		}
		param, err := parsePersonalParam(cmd)
		msg, err := client.UnlockAccount(context.Background(), param)
		if err != nil {
			cmd.PrintErrf("Failed: %s\n", err.Error())
			return
		}
		cmd.Println(types.EncodeAddress(msg.GetAddress()))
	},
}

// import account using WIF (legacy)
func importWif(cmd *cobra.Command) ([]byte, error) {
	var err error
	var address []byte
	importBuf, err := types.DecodePrivKey(importFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to decode input: %s", err.Error())
	}
	wif := &types.ImportFormat{Wif: &types.SingleBytes{Value: importBuf}}
	if pw != "" {
		wif.Oldpass = pw
	} else {
		wif.Oldpass, err = getPasswd(cmd, false)
		if err != nil {
			return nil, err
		}
	}

	if to != "" {
		wif.Newpass = to
	} else {
		wif.Newpass = wif.Oldpass
	}

	if rootConfig.KeyStorePath == "" {
		msg, errRemote := client.ImportAccount(context.Background(), wif)
		address = msg.GetAddress()
		if errRemote != nil {
			return address, fmt.Errorf("node request returned error: %s", errRemote.Error())
		}
	} else {
		dataEnvPath := os.ExpandEnv(rootConfig.KeyStorePath)
		ks := key.NewStore(dataEnvPath, 0)
		defer ks.CloseStore()
		address, err = ks.ImportKey(importBuf, wif.Oldpass, wif.Newpass)
		if err != nil {
			return address, err
		}
	}
	return address, nil
}

// import account using keystore
func importKeystore(cmd *cobra.Command) ([]byte, error) {
	var err error
	var address []byte
	absPath, err := filepath.Abs(importFilePath)
	if err != nil {
		return nil, err
	}
	keystore, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	wif := &types.ImportFormat{Keystore: &types.SingleBytes{Value: keystore}}
	if pw != "" {
		wif.Oldpass = pw
	} else {
		wif.Oldpass, err = getPasswd(cmd, false)
		if err != nil {
			return nil, err
		}
	}

	if to != "" {
		wif.Newpass = to
	} else {
		wif.Newpass = wif.Oldpass
	}

	if rootConfig.KeyStorePath == "" {
		msg, errRemote := client.ImportAccount(context.Background(), wif)
		if errRemote != nil {
			return nil, errRemote
		}
		address = msg.GetAddress()
	} else {
		dataEnvPath := os.ExpandEnv(rootConfig.KeyStorePath)
		ks := key.NewStore(dataEnvPath, 0)
		defer ks.CloseStore()
		privateKey, err := key.LoadKeystore(keystore, wif.Oldpass)
		if err != nil {
			return nil, err
		}
		address, err = ks.AddKey(privateKey, wif.Newpass)
		if err != nil {
			return nil, err
		}
	}

	return address, nil
}

var importCmd = &cobra.Command{
	Use:   "import [flags]",
	Short: "Import account",
	Run: func(cmd *cobra.Command, args []string) {
		var address []byte
		var err error
		if importFormat != "" {
			address, err = importWif(cmd)
		} else if importFilePath != "" {
			address, err = importKeystore(cmd)
		} else {
			cmd.Help()
		}

		if err != nil {
			cmd.PrintErrln(err)
		}
		if address != nil {
			cmd.Println(types.EncodeAddress(address))
		}
	},
}

// export account as WIF (legacy)
func exportWif(cmd *cobra.Command, param *types.Personal) ([]byte, error) {
	if rootConfig.KeyStorePath == "" {
		msg, err := client.ExportAccount(context.Background(), param)
		if err != nil {
			return nil, fmt.Errorf("node request returned error: %s", err.Error())
		}
		return msg.Value, nil
	} else {
		dataEnvPath := os.ExpandEnv(rootConfig.KeyStorePath)
		ks := key.NewStore(dataEnvPath, 0)
		defer ks.CloseStore()
		wif, err := ks.ExportKey(param.Account.Address, param.Passphrase)
		if err != nil {
			return nil, err
		}
		return wif, nil
	}
}

// export account as keystore
func exportKeystore(cmd *cobra.Command, param *types.Personal) ([]byte, error) {
	if rootConfig.KeyStorePath == "" {
		msg, err := client.ExportAccountKeystore(context.Background(), param)
		if err != nil {
			return nil, fmt.Errorf("node request returned error: %s", err.Error())
		}
		return msg.Value, nil
	} else {
		dataEnvPath := os.ExpandEnv(rootConfig.KeyStorePath)
		ks := key.NewStore(dataEnvPath, 0)
		defer ks.CloseStore()

		privateKey, err := ks.GetKey(param.Account.Address, param.Passphrase)
		if err != nil {
			return nil, err
		}

		wif, err := key.GetKeystore(privateKey, param.Passphrase)
		if err != nil {
			return nil, err
		}
		return wif, nil
	}
}

var exportCmd = &cobra.Command{
	Use:   "export [flags]",
	Short: "Export account",
	Run: func(cmd *cobra.Command, args []string) {
		var result []byte
		var err error

		param, err := parsePersonalParam(cmd)
		if err != nil {
			cmd.PrintErr(err.Error())
			return
		}

		if exportAsWif {
			result, err = exportWif(cmd, param)
			if result != nil {
				cmd.Println(types.EncodePrivKey(result))
			}
		} else {
			result, err = exportKeystore(cmd, param)
			if result != nil {
				cmd.Println(string(result))
			}
		}

		if err != nil {
			cmd.PrintErrln(err)
		}
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

// Try to open /dev/tty for read/write, with stdin/stdout as fallback
// This is so we can use aergocli in a pipe and still read passwords from tty
func getTerminalReaderWriter() (struct {
	io.Reader
	io.Writer
}, int) {
	path := "/dev/tty"
	fallback := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
	in, err := os.Open(path)
	if err != nil {
		return fallback, 0
	}
	out, err := os.OpenFile(path, syscall.O_WRONLY, 0)
	if err != nil {
		return fallback, 0
	}
	return struct {
		io.Reader
		io.Writer
	}{in, out}, int(out.Fd())
}

func getPasswd(cmd *cobra.Command, isNew bool) (string, error) {
	screen, fd := getTerminalReaderWriter()
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer terminal.Restore(0, oldState)
	term := terminal.NewTerminal(screen, "")
	password, err := term.ReadPassword("Enter password: ")
	if err != nil {
		return "", err
	}
	if len(password) == 0 {
		return "", errors.New("empty password")
	}
	if isNew {
		repeat, err := term.ReadPassword("Repeat password: ")
		if err != nil {
			return "", err
		}
		if password != repeat {
			return "", errors.New("Passwords don't match")
		}
	}
	return password, err
}

func preConnectAergo(cmd *cobra.Command, args []string) {
	if rootConfig.KeyStorePath == "" {
		connectAergo(cmd, args)
	} else {
		client = nil
	}

	if len(sock) > 0 {
		if admClient = newAergoAdminClient(sock); admClient == nil {
			log.Fatal("can't create admin connection")
		}
	}
}
