package cmd

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/spf13/cobra"
)

type keyJson struct {
	Address string `json:"address"`
	PubKey  string `json:"pubkey"`
	PrivKey string `json:"privkey"`
	Id      string `json:"id"`
}

var (
	genPubkey bool
	genID     bool
	genJSON   bool
	password  string
)

func init() {
	//keygenCmd.Flags().StringVar(&prefix, "prefix", "nodekey", "prefix name of key file")
	//keygenCmd.Flags().BoolVar(&genPubkey, "genpubkey", true, "also generate public key")
	keygenCmd.Flags().BoolVar(&genJSON, "json", false, "also generate combined json file")
	keygenCmd.Flags().StringVar(&password, "password", "", "password for encrypted private key in json file")

	rootCmd.AddCommand(keygenCmd)
}

var keygenCmd = &cobra.Command{
	Use:   "keygen [flags] <prefix>",
	Short: "Generate private key",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Failed: no prefix")
			return
		}
		prefix := args[0]
		if prefix == "" {
			fmt.Printf("Failed: invalid prefix %s\n", prefix)
			return
		}
		if err := generateKey(prefix); err != nil {
			fmt.Printf("Failed: %s\n", err.Error())
			return
		}
	},
}

func generateKey(prefix string) error {
	priv, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	pid, _ := peer.IDFromPublicKey(pub)

	pkFile := prefix + ".key"
	pubFile := prefix + ".pub"
	idFile := prefix + ".id"
	jsonFile := prefix + ".json"

	// Write private key file
	pkf, err := os.Create(pkFile)
	if err != nil {
		return err
	}
	pkBytes, err := priv.Bytes()
	if err != nil {
		return err
	}
	pkf.Write(pkBytes)
	pkf.Sync()

	// Write public key file
	pubf, err := os.Create(pubFile)
	if err != nil {
		return err
	}
	pubBytes, err := pub.Bytes()
	if err != nil {
		return err
	}
	pubf.Write(pubBytes)
	pubf.Sync()

	// Write id file
	idf, err := os.Create(idFile)
	if err != nil {
		return err
	}
	idBytes := []byte(peer.IDB58Encode(pid))
	idf.Write(idBytes)
	idf.Sync()

	// Write combined json file
	if genJSON {
		if password == "" {
			fmt.Printf("Warning: private key in json file encrypted with empty password. Use command line parameter --password.\n")
		}
		jsonf, err := os.Create(jsonFile)
		if err != nil {
			return err
		}
		privKeyExport, err := key.EncryptKey(pkBytes, password)
		if err != nil {
			return err
		}
		_, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), pkBytes)
		address := key.GenerateAddress(pubkey.ToECDSA())
		addressEncoded := types.EncodeAddress(address)
		jsonMarshalled, err := json.Marshal(keyJson{
			Address: addressEncoded,
			PubKey:  b64.StdEncoding.EncodeToString(pubBytes),
			PrivKey: types.EncodePrivKey(privKeyExport),
			Id:      peer.IDB58Encode(pid),
		})
		if err != nil {
			return err
		}
		jsonf.Write(jsonMarshalled)
		jsonf.Sync()

		fmt.Printf("New account address: %s\n", addressEncoded)
		fmt.Printf("Generated key files %s.{key,pub,id,json}.\n", prefix)
	} else {
		fmt.Printf("Generated key files %s.{key,pub,id}.\n", prefix)
	}

	return nil
}
