/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var nameCmd = &cobra.Command{
	Use:   "name [flags] subcommand",
	Short: "Name command",
}
var spending string

func init() {
	rootCmd.AddCommand(nameCmd)
	newCmd := &cobra.Command{
		Use:                   "new",
		Short:                 "Create account name. It spend at least 1 aergo",
		RunE:                  execNameNew,
		DisableFlagsInUseLine: true,
	}
	newCmd.Flags().StringVar(&from, "from", "", "Sender account address")
	newCmd.MarkFlagRequired("from")
	newCmd.Flags().StringVar(&name, "name", "", "Name of account to create")
	newCmd.MarkFlagRequired("name")
	newCmd.Flags().StringVar(&spending, "amount", "1aergo", "Spending for create name. at least 1 aergo")

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
	updateCmd.Flags().StringVar(&spending, "amount", "1aergo", "Spending for create name. at least 1 aergo")

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

	if len(name) != types.NameLength {
		return errors.New("The name must be 12 alphabetic characters\n")
	}
	amount, err := util.ParseUnit(spending)
	if err != nil {
		return errors.New("Wrong value in --amount flag\n" + err.Error())
	}
	var ci types.CallInfo
	ci.Name = types.NameCreate
	err = json.Unmarshal([]byte("[\""+name+"\"]"), &ci.Args)
	if err != nil {
		log.Fatal(err)
	}
	payload, err := json.Marshal(ci)
	if err != nil {
		log.Fatal(err)
	}
	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   account,
			Recipient: []byte(types.AergoName),
			Amount:    amount.Bytes(),
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
	cmd.Println(util.JSON(msg))
	return nil
}

func execNameUpdate(cmd *cobra.Command, args []string) error {
	account, err := types.DecodeAddress(from)
	if err != nil {
		return errors.New("Wrong address in --from flag\n" + err.Error())
	}
	_, err = types.DecodeAddress(to)
	if err != nil {
		return errors.New("Wrong address in --from flag\n" + err.Error())
	}
	if len(name) != types.NameLength {
		return errors.New("The name must be 12 alphabetic characters\n")
	}
	amount, err := util.ParseUnit(spending)
	if err != nil {
		return errors.New("Wrong value in --amount flag\n" + err.Error())
	}
	var ci types.CallInfo
	ci.Name = types.NameUpdate
	err = json.Unmarshal([]byte("[\""+name+"\",\""+to+"\"]"), &ci.Args)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal([]byte("[\""+name+"\",\""+to+"\"]"), &ci.Args)
	if err != nil {
		log.Fatal(err)
	}
	payload, err := json.Marshal(ci)
	if err != nil {
		log.Fatal(err)
	}
	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   account,
			Recipient: []byte(types.AergoName),
			Amount:    amount.Bytes(),
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
	cmd.Println(util.JSON(msg))
	return nil
}

func execNameOwner(cmd *cobra.Command, args []string) {
	msg, err := client.GetNameInfo(context.Background(), &types.Name{Name: name})
	if err != nil {
		cmd.Println(err.Error())
		return
	}
	cmd.Println("{\n \"" + msg.Name.Name + "\": {\n  " +
		"\"Owner\": \"" + types.EncodeAddress(msg.Owner) + "\",\n  " +
		"\"Destination\": \"" + types.EncodeAddress(msg.Destination) + "\"\n  }\n}")
}
