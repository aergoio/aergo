/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"errors"

	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

var nameCmd = &cobra.Command{
	Use:   "name [flags] subcommand",
	Short: "Name command",
}

func init() {
	rootCmd.AddCommand(nameCmd)
	newCmd := &cobra.Command{
		Use:                   "new",
		Short:                 "Create account name",
		RunE:                  execNameNew,
		DisableFlagsInUseLine: true,
	}
	newCmd.Flags().StringVar(&from, "from", "", "Sender account address")
	newCmd.MarkFlagRequired("from")
	newCmd.Flags().StringVar(&name, "name", "", "Name of account to create")
	newCmd.MarkFlagRequired("name")

	updateCmd := &cobra.Command{
		Use:                   "update",
		Short:                 "Update account name",
		RunE:                  execNameUpdate,
		DisableFlagsInUseLine: true,
	}
	updateCmd.Flags().StringVar(&from, "from", "", "Sender account address")
	updateCmd.MarkFlagRequired("from")
	updateCmd.Flags().StringVar(&to, "to", "", "Recipient account address")
	updateCmd.MarkFlagRequired("to")
	updateCmd.Flags().StringVar(&name, "name", "", "Name of account to create")
	updateCmd.MarkFlagRequired("name")

	ownerCmd := &cobra.Command{
		Use:                   "owner",
		Short:                 "Owner of account name",
		Run:                   execNameOwner,
		DisableFlagsInUseLine: true,
	}
	ownerCmd.Flags().StringVar(&name, "name", "", "Name of account to create")
	ownerCmd.MarkFlagRequired("name")

	nameCmd.AddCommand(newCmd, updateCmd, ownerCmd)
}

func execNameNew(cmd *cobra.Command, args []string) error {
	account, err := types.DecodeAddress(from)
	if err != nil {
		return errors.New("Wrong address in --from flag\n" + err.Error())
	}
	payload := []byte{'c'}
	if len(name) != types.NameLength {
		return errors.New("The name must be 12 alphabetic characters\n")
	}
	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   account,
			Recipient: []byte(types.AergoName),
			Payload:   append(payload, []byte(name)...),
			Limit:     0,
			Type:      types.TxType_GOVERNANCE,
		},
	}
	msg, err := client.SendTX(context.Background(), tx)
	if err != nil {
		cmd.Printf("Failed request to aergo sever\n" + err.Error())
		return nil
	}
	cmd.Println(base58.Encode(msg.Hash), msg.Error, msg.Detail)
	return nil
}

func execNameUpdate(cmd *cobra.Command, args []string) error {
	account, err := types.DecodeAddress(from)
	if err != nil {
		return errors.New("Wrong address in --from flag\n" + err.Error())
	}
	to, err := types.DecodeAddress(to)
	if err != nil {
		return errors.New("Wrong address in --from flag\n" + err.Error())
	}
	if len(name) != types.NameLength {
		return errors.New("The name must be 12 alphabetic characters\n")
	}
	payload := []byte{'u'}
	payload = append(payload, []byte(name)...)
	payload = append(payload, ',')
	payload = append(payload, to...)

	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   account,
			Recipient: []byte(types.AergoName),
			Payload:   payload,
			Limit:     0,
			Type:      types.TxType_GOVERNANCE,
		},
	}

	msg, err := client.SendTX(context.Background(), tx)
	if err != nil {
		cmd.Printf("Failed request to aergo sever\n" + err.Error())
		return nil
	}
	cmd.Println(base58.Encode(msg.Hash), msg.Error, msg.Detail)
	return nil
}

func execNameOwner(cmd *cobra.Command, args []string) {
	msg, err := client.GetNameInfo(context.Background(), &types.Name{Name: name})
	if err != nil {
		cmd.Println(err.Error())
		return
	}
	owner := msg.Owner
	if len(owner) > types.NameLength {
		cmd.Println(msg.Name.Name, types.EncodeAddress(owner))
	} else {
		cmd.Println(msg.Name.Name, string(msg.Owner))
	}
}
