package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringVar(&jsonTx, "jsontx", "", "transaction json to sign")
	signCmd.Flags().StringVar(&dataDir, "path", "$HOME/.aergo/data/cli", "path to data directory")
	signCmd.Flags().StringVar(&address, "address", "1", "address of account to use for signing")
	signCmd.Flags().BoolVar(&remote, "remote", true, "indicate account in the remote node or not")
	signCmd.Flags().StringVar(&pw, "password", "", "local account password")
	signCmd.Flags().StringVar(&privKey, "key", "", "base58 encoded key for sign")
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().StringVar(&jsonTx, "jsontx", "", "transaction list json to verify")
	verifyCmd.Flags().BoolVar(&remote, "remote", true, "choose verify in the remote node or not")
}

var signCmd = &cobra.Command{
	Use:    "signtx",
	Short:  "Sign transaction",
	Args:   cobra.MinimumNArgs(0),
	PreRun: preConnectAergo,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if jsonTx == "" {
			fmt.Printf("need to transaction json input")
			return
		}
		param, err := util.ParseBase58TxBody([]byte(jsonTx))
		if err != nil {
			fmt.Printf("Failed: %s\n", err.Error())
			return
		}

		var msg *types.Tx
		if privKey != "" {
			rawKey, err := base58.Decode(privKey)
			if err != nil {
				fmt.Printf("Failed: %s\n", err.Error())
				return
			}
			tx := &types.Tx{Body: param}
			signKey, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), rawKey)
			err = key.SignTx(tx, signKey)
			if err != nil {
				fmt.Printf("Failed: %s\n", err.Error())
				return
			}
			fmt.Println(types.EncodeAddress(key.GenerateAddress(pubkey.ToECDSA())))
			msg = tx
		} else if remote {
			msg, err = client.SignTX(context.Background(), &types.Tx{Body: param})
		} else {
			tx := &types.Tx{Body: param}
			if tx.Body.Sign != nil {
				tx.Body.Sign = nil
			}
			hash := key.CalculateHashWithoutSign(param)

			dataEnvPath := os.ExpandEnv(dataDir)
			ks := key.NewStore(dataEnvPath)
			addr, err := types.DecodeAddress(address)
			if err != nil {
				fmt.Printf("Failed: %s\n", err.Error())
				return
			}
			tx.Body.Sign, err = ks.Sign(addr, pw, hash)
			if err != nil {
				fmt.Printf("Failed: %s\n", err.Error())
				return
			}
			tx.Hash = tx.CalculateTxHash()
			msg = tx
		}

		if nil == err && msg != nil {
			fmt.Println(util.TxConvBase58Addr(msg))
		} else {
			fmt.Printf("Failed: %s\n", err.Error())
		}
	},
}

var verifyCmd = &cobra.Command{
	Use:    "verifytx",
	Short:  "Verify transaction",
	PreRun: preConnectAergo,
	Run: func(cmd *cobra.Command, args []string) {
		if jsonTx == "" {
			fmt.Printf("need to transaction json input")
			return
		}
		param, err := util.ParseBase58Tx([]byte(jsonTx))
		if err != nil {
			fmt.Printf("Failed: %s\n", err.Error())
			return
		}
		if remote {
			msg, err := client.VerifyTX(context.Background(), param[0])
			if nil == err {
				if msg.Tx != nil {
					fmt.Println(util.TxConvBase58Addr(msg.Tx))
				} else {
					fmt.Println(msg.Error)
				}
			} else {
				fmt.Printf("Failed: %s\n", err.Error())
			}
		} else {
			err := key.VerifyTx(param[0])
			if err != nil {
				fmt.Printf("Failed: %s\n", err.Error())
				return
			}
		}
	},
}
