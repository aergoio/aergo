package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aergoio/aergo/v2/account/key"
	keycrypto "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/aergoio/aergo/v2/internal/enc/base64"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/spf13/cobra"
)

type keyJson struct {
	Address string `json:"address"`
	PubKey  string `json:"pubkey"`
	PrivKey string `json:"privkey"`
	Id      string `json:"id"`
}

var (
	fromPK     bool
	genPubkey  bool
	genID      bool
	genJSON    bool
	genAddress bool
	password   string
)

func init() {
	//keygenCmd.Flags().StringVar(&prefix, "prefix", "nodekey", "prefix name of key file")
	//keygenCmd.Flags().BoolVar(&genPubkey, "genpubkey", true, "also generate public key")
	keygenCmd.Flags().BoolVar(&fromPK, "fromKey", false, "generate files from existing private key file")
	keygenCmd.Flags().BoolVar(&genJSON, "json", false, "output combined json object instead of generating files")
	keygenCmd.Flags().StringVar(&password, "password", "", "password for encrypted private key in json file")
	keygenCmd.Flags().BoolVar(&genAddress, "addr", false, "generate prefix.addr for wallet address")

	rootCmd.AddCommand(keygenCmd)
}

var keygenCmd = &cobra.Command{
	Use:   "keygen [flags] <prefix>",
	Short: "Generate private key",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if genJSON {
			if fromPK {
				if len(args) < 1 {
					fmt.Println("Failed: no keyfile")
					return
				}
				priv, pub, err := p2putil.LoadKeyFile(args[0])
				if err == nil {
					err = generateKeyJson(priv, pub)
				}
			} else {
				priv, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
				err = generateKeyJson(priv, pub)
			}
		} else {
			if len(args) < 1 {
				fmt.Println("Failed: no prefix")
				return
			}
			prefix := args[0]
			if prefix == "" {
				fmt.Printf("Failed: invalid prefix %s\n", prefix)
				return
			}
			if fromPK {
				err = loadPKAndGenerateKeyFiles(prefix)
			} else {
				err = generateKeyFiles(prefix)
			}
		}
		if err != nil {
			fmt.Printf("Failed: %s\n", err.Error())
			return
		}
	},
}

func loadPKAndGenerateKeyFiles(pkFile string) error {
	priv, pub, err := p2putil.LoadKeyFile(pkFile)
	if err != nil {
		return err
	}
	pkExt := filepath.Ext(pkFile)
	if pkExt == ".pub" || pkExt == ".id" || pkExt == ".addr" {
		return fmt.Errorf("invalid pk extension %s", pkExt)
	}
	prefix := strings.TrimSuffix(pkFile, pkExt)
	return saveFilesFromKeys(priv, pub, prefix)
}

func generateKeyFiles(prefix string) error {
	priv, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)

	pkFile := prefix + ".key"

	// Write private key file
	err := savePrivKeyFile(pkFile, priv)
	if err != nil {
		return err
	}

	return saveFilesFromKeys(priv, pub, prefix)
}

func saveFilesFromKeys(priv crypto.PrivKey, pub crypto.PubKey, prefix string) error {
	pubFile := prefix + ".pub"
	idFile := prefix + ".id"
	// Write public key file
	err := savePubKeyFile(pubFile, pub)
	if err != nil {
		return err
	}

	// Write id file
	pid, _ := types.IDFromPublicKey(pub)
	idBytes := []byte(types.IDB58Encode(pid))
	saveBytesToFile(idFile, idBytes)

	if genAddress {
		pkBytes, err := crypto.MarshalPrivateKey(priv)
		if err != nil {
			return err
		}
		addrFile := prefix + ".addr"
		addrf, err := os.Create(addrFile)
		if err != nil {
			return err
		}
		_, pubkey := btcec.PrivKeyFromBytes(pkBytes)
		address := keycrypto.GenerateAddress(pubkey.ToECDSA())
		addrf.WriteString(types.EncodeAddress(address))
		addrf.Sync()

		fmt.Printf("Wrote files %s.{key,pub,id,addr}.\n", prefix)
	} else {
		fmt.Printf("Wrote files %s.{key,pub,id}.\n", prefix)
	}
	return nil
}

func savePrivKeyFile(pkFile string, priv crypto.PrivKey) error {
	pkf, err := os.Create(pkFile)
	if err != nil {
		return err
	}
	pkBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return err
	}
	_, err = pkf.Write(pkBytes)
	if err != nil {
		return err
	}
	return pkf.Sync()
}

func savePubKeyFile(pkFile string, pub crypto.PubKey) error {
	pkf, err := os.Create(pkFile)
	if err != nil {
		return err
	}
	pkBytes, err := crypto.MarshalPublicKey(pub)
	if err != nil {
		return err
	}
	_, err = pkf.Write(pkBytes)
	if err != nil {
		return err
	}
	return pkf.Sync()
}

func saveBytesToFile(fileName string, bytes []byte) error {
	pkf, err := os.Create(fileName)
	if err != nil {
		return err
	}
	_, err = pkf.Write(bytes)
	if err != nil {
		return err
	}
	return pkf.Sync()
}
func generateKeyJson(priv crypto.PrivKey, pub crypto.PubKey) error {
	btcPK := p2putil.ConvertPKToBTCEC(priv)
	pkBytes := btcPK.Serialize()
	pubBytes, err := crypto.MarshalPublicKey(pub)
	pid, _ := types.IDFromPublicKey(pub)
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
	address := keycrypto.GenerateAddress(btcPK.PubKey().ToECDSA())
	addressEncoded := types.EncodeAddress(address)
	jsonMarshalled, err := json.MarshalIndent(keyJson{
		Address: addressEncoded,
		PubKey:  base64.Encode(pubBytes),
		PrivKey: types.EncodePrivKey(privKeyExport),
		Id:      types.IDB58Encode(pid),
	}, "", "    ")
	if err != nil {
		return err
	}

	fmt.Printf(string(jsonMarshalled))

	return nil
}
