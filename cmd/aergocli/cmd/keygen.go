package cmd

import (
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/spf13/cobra"
)

var (
	genPubkey bool
	genID     bool
)

func init() {
	//keygenCmd.Flags().StringVar(&prefix, "prefix", "nodekey", "prefix name of key file")
	//keygenCmd.Flags().BoolVar(&genPubkey, "genpubkey", true, "also generate public key")
	//keygenCmd.Flags().BoolVar(&genID, "genid", true, "also generate id")

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
	fmt.Printf("Key file %s is generated.\n", pkFile)

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

	idf, err := os.Create(idFile)
	idBytes := []byte(peer.IDB58Encode(pid))
	if err != nil {
		return err
	}
	idf.Write(idBytes)
	idf.Sync()

	return nil
}
