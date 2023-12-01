package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/aergoio/aergo/v2/account/key"
	crypto "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/jsonrpc"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringVar(&jsonTx, "jsontx", "", "transaction json to sign")
	signCmd.Flags().StringVar(&jsonPath, "jsontxpath", "", "transaction json file path to sign")
	signCmd.Flags().StringVar(&address, "address", "1", "address of account to use for signing")
	signCmd.Flags().StringVar(&pw, "password", "", "local account password")
	signCmd.Flags().StringVar(&privKey, "key", "", "base58 encoded key for sign")
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().StringVar(&jsonTx, "jsontx", "", "transaction list json to verify")
	verifyCmd.Flags().StringVar(&jsonPath, "jsontxpath", "", "transaction json file path to verify")
	verifyCmd.Flags().BoolVar(&remote, "remote", false, "verify in the node")
}

var signCmd = &cobra.Command{
	Use:    "signtx",
	Short:  "Sign transaction",
	Args:   cobra.MinimumNArgs(0),
	PreRun: preConnectAergo,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if jsonTx == "" && jsonPath == "" {
			cmd.Printf("need to transaction json input")
			return
		}
		if jsonTx == "" {
			b, readerr := os.ReadFile(jsonPath)
			if readerr != nil {
				cmd.Printf("Failed to read --jsontxpath\n" + readerr.Error())
				return
			}
			jsonTx = string(b)
		}
		param, err := jsonrpc.ParseBase58TxBody([]byte(jsonTx))
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}

		var msg *types.Tx
		if privKey != "" {
			rawKey, err := base58.Decode(privKey)
			if err != nil {
				cmd.Printf("Failed: %s\n", err.Error())
				return
			}
			tx := &types.Tx{Body: param}
			signKey, pubkey := btcec.PrivKeyFromBytes(rawKey)
			err = key.SignTx(tx, signKey)
			if err != nil {
				cmd.Printf("Failed: %s\n", err.Error())
				return
			}
			cmd.Println(types.EncodeAddress(crypto.GenerateAddress(pubkey.ToECDSA())))
			msg = tx
		} else if rootConfig.KeyStorePath == "" {
			msg, err = client.SignTX(context.Background(), &types.Tx{Body: param})
		} else {
			if cmd.Flags().Changed("address") == false {
				cmd.Print("Error: required flag(s) \"address\" not set")
				return
			}
			addr, err := types.DecodeAddress(address)
			if err != nil {
				cmd.Printf("Failed: %s\n", err.Error())
				return
			}
			tx := &types.Tx{Body: param}
			if tx.Body.Sign != nil {
				tx.Body.Sign = nil
			}
			if pw == "" {
				pw, err = getPasswd(cmd, false)
				if err != nil {
					cmd.Println("Failed get password:" + err.Error())
					return
				}
			}
			if errStr := fillSign(tx, rootConfig.KeyStorePath, pw, addr); errStr != "" {
				cmd.Printf("Failed: %s\n", errStr)
				return
			}
			msg = tx
		}

		if nil == err && msg != nil {
			res := jsonrpc.ConvTx(msg, jsonrpc.Base58)
			cmd.Println(jsonrpc.MarshalJSON(res))
		} else {
			cmd.Printf("Failed: %s\n", err.Error())
		}
	},
}

var verifyCmd = &cobra.Command{
	Use:    "verifytx",
	Short:  "Verify transaction",
	PreRun: preConnectAergo,
	Run: func(cmd *cobra.Command, args []string) {
		if jsonTx == "" && jsonPath == "" {
			cmd.Printf("need to transaction json input")
			return
		}
		if jsonTx == "" {
			b, readerr := os.ReadFile(jsonPath)
			if readerr != nil {
				cmd.Printf("Failed to read --jsontxpath\n" + readerr.Error())
				return
			}
			jsonTx = string(b)
		}

		param, err := jsonrpc.ParseBase58Tx([]byte(jsonTx))
		if err != nil {
			cmd.Printf("Failed: %s\n", err.Error())
			return
		}
		if remote {
			msg, err := client.VerifyTX(context.Background(), param[0])
			if err != nil {
				cmd.Printf("Failed: %s\n", err.Error())
				return
			}
			if msg.Tx != nil {
				res := jsonrpc.ConvTx(msg.Tx, jsonrpc.Base58)
				cmd.Println(jsonrpc.MarshalJSON(res))
			} else {
				cmd.Println(msg.Error)
			}
		} else {
			err := key.VerifyTx(param[0])
			if err != nil {
				cmd.Printf("Failed: %s\n", err.Error())
				return
			}
			res := jsonrpc.ConvTx(param[0], jsonrpc.Base58)
			cmd.Println(jsonrpc.MarshalJSON(res))
		}
	},
}

func fillSign(tx *types.Tx, dataDir, pw string, account []byte) string {
	hash := key.CalculateHashWithoutSign(tx.Body)
	dataEnvPath := os.ExpandEnv(dataDir)
	ks := key.NewStore(dataEnvPath, 0)
	defer ks.CloseStore()
	var err error
	tx.Body.Sign, err = ks.Sign(account, pw, hash)
	if err != nil {
		return fmt.Sprintf("Failed: %s\n", err.Error())
	}
	tx.Hash = tx.CalculateTxHash()
	return ""
}
