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
	keygenCmd.Flags().BoolVar(&genJSON, "json", false, "output combined json object instead of generating files")
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
		var err error
		if genJSON {
			err = generateKeyJson()
		} else {
			err = generateKeyFiles(prefix)
		}
		if err != nil {
			fmt.Printf("Failed: %s\n", err.Error())
			return
		}
	},
}

func generateKeyFiles(prefix string) error {
	priv, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)

	pkFile := prefix + ".key"
	pubFile := prefix + ".pub"
	idFile := prefix + ".id"

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
	pid, _ := peer.IDFromPublicKey(pub)
	idBytes := []byte(peer.IDB58Encode(pid))
	idf.Write(idBytes)
	idf.Sync()
	return nil
}

func generateKeyJson() error {
	priv, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	pkBytes, err := priv.Bytes()
	pubBytes, err := pub.Bytes()
	pid, _ := peer.IDFromPublicKey(pub)
	if err != nil {
		return err
	}
	if password == "" {
		fmt.Printf("Warning: private key in json file encrypted with empty password. Use command line parameter --password.\n")
	}
	privKeyExport, err := key.EncryptKey(pkBytes, password)
	if err != nil {
		return err
	}
	_, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), pkBytes)
	address := key.GenerateAddress(pubkey.ToECDSA())
	addressEncoded := types.EncodeAddress(address)
	jsonMarshalled, err := json.MarshalIndent(keyJson{
		Address: addressEncoded,
		PubKey:  b64.StdEncoding.EncodeToString(pubBytes),
		PrivKey: types.EncodePrivKey(privKeyExport),
		Id:      peer.IDB58Encode(pid),
	}, "", "    ")
	if err != nil {
		return err
	}

	fmt.Printf(string(jsonMarshalled))

	return nil
}
