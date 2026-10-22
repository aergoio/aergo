/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	"github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
)

var nameCmd = &cobra.Command{
	Use:   "name [flags] subcommand",
	Short: "Name command",
}
var spending string
var blockNo uint64

func init() {
	rootCmd.AddCommand(nameCmd)
	newCmd := &cobra.Command{
		Use:                   "new",
		Short:                 "Create account name. It spend at least an amount of aergo according to nameprice",
		RunE:                  execNameNew,
		DisableFlagsInUseLine: true,
	}
	newCmd.Flags().StringVar(&from, "from", "", "sender account address")
	newCmd.MarkFlagRequired("from")
	newCmd.Flags().StringVar(&name, "name", "", "name of account to create")
	newCmd.MarkFlagRequired("name")
	newCmd.Flags().StringVar(&spending, "amount", "20aergo", "spending for create name. Must be set to the current nameprice")
	newCmd.Flags().StringVar(&pw, "password", "", "password")

	updateCmd := &cobra.Command{
		Use:                   "update",
		Short:                 "Update account name",
		RunE:                  execNameUpdate,
		DisableFlagsInUseLine: true,
	}
	updateCmd.Flags().StringVar(&from, "from", "", "sender account address")
	updateCmd.MarkFlagRequired("from")
	updateCmd.Flags().StringVar(&to, "to", "", "recipient account address")
	updateCmd.MarkFlagRequired("to")
	updateCmd.Flags().StringVar(&name, "name", "", "name of account to create")
	updateCmd.MarkFlagRequired("name")
	updateCmd.Flags().StringVar(&spending, "amount", "20aergo", "spending for update name. Must be set to the current nameprice")
	updateCmd.Flags().StringVar(&pw, "password", "", "password")

	ownerCmd := &cobra.Command{
		Use:                   "owner",
		Short:                 "Owner of account name",
		Run:                   execNameOwner,
		DisableFlagsInUseLine: true,
	}
	ownerCmd.Flags().StringVar(&name, "name", "", "name of account to create")
	ownerCmd.MarkFlagRequired("name")
	ownerCmd.Flags().Uint64VarP(&blockNo, "blockno", "n", 0, "Block height")

	nameCmd.AddCommand(newCmd, updateCmd, ownerCmd)
}

func execNameNew(cmd *cobra.Command, args []string) error {
	account, err := types.DecodeAddress(from)
	if err != nil {
		return fmt.Errorf("wrong address in --from flag: %v", err.Error())
	}

	if len(name) != types.NameLength {
		return errors.New("the name must be 12 alphabetic characters")
	}
	amount, err := util.ParseUnit(spending)
	if err != nil {
		return fmt.Errorf("wrong value in --amount flag: %v", err.Error())
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
			GasLimit:  0,
			Type:      types.TxType_GOVERNANCE,
		},
	}
	cmd.Println(sendTX(cmd, tx, account))
	return nil
}

func execNameUpdate(cmd *cobra.Command, args []string) error {
	account, err := types.DecodeAddress(from)
	if err != nil {
		return fmt.Errorf("Wrong address in --from flag: %v", err.Error())
	}
	_, err = types.DecodeAddress(to)
	if err != nil {
		return fmt.Errorf("Wrong address in --to flag: %v", err.Error())
	}
	amount, err := util.ParseUnit(spending)
	if err != nil {
		return fmt.Errorf("Wrong value in --amount flag: %v", err.Error())
	}
	var ci types.CallInfo
	if name == types.AergoName {
		ci.Name = types.SetContractOwner
		err = json.Unmarshal([]byte("[\""+to+"\"]"), &ci.Args)
		if err != nil {
			log.Fatal(err)
		}
		amount = big.NewInt(0)
	} else {
		ci.Name = types.NameUpdate
		if len(name) != types.NameLength {
			return errors.New("the name must be 12 alphabetic characters")
		}
		err = json.Unmarshal([]byte("[\""+name+"\",\""+to+"\"]"), &ci.Args)
		if err != nil {
			log.Fatal(err)
		}
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
			GasLimit:  0,
			Type:      types.TxType_GOVERNANCE,
		},
	}

	cmd.Println(sendTX(cmd, tx, account))
	return nil
}

func execNameOwner(cmd *cobra.Command, args []string) {
	msg, err := client.GetNameInfo(context.Background(), &types.Name{Name: name, BlockNo: blockNo})
	if err != nil {
		cmd.Println(err.Error())
		return
	}
	cmd.Println("{\n \"" + msg.Name.Name + "\": {\n  " +
		"\"Owner\": \"" + types.EncodeAddress(msg.Owner) + "\",\n  " +
		"\"Destination\": \"" + types.EncodeAddress(msg.Destination) + "\"\n  }\n}")
}
