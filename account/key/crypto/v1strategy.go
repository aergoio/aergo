/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"reflect"

	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"golang.org/x/crypto/scrypt"
)

const (
	// version
	version = "1"

	// cipher
	cipherAlgorithm = "aes-128-ctr"

	// kdf
	kdfAlgorithm = "scrypt"
	// Recommended values according to https://godoc.org/golang.org/x/crypto/scrypt
	scryptN     = 1 << 15
	scryptP     = 1
	scryptR     = 8
	scryptDKLen = 32
)

type v1KeyStoreFormat struct {
	Address string       `json:"aergo_address"`
	Version string       `json:"ks_version"`
	Cipher  v1CipherJSON `json:"cipher"`
	Kdf     v1KdfJson    `json:"kdf"`
}

type v1CipherJSON struct {
	Algorithm  string             `json:"algorithm"`
	Params     v1CipherParamsJSON `json:"params"`
	Ciphertext string             `json:"ciphertext"`
}

type v1CipherParamsJSON struct {
	Iv string `json:"iv"`
}

type v1KdfJson struct {
	Algorithm string          `json:"algorithm"`
	Params    v1KdfParamsJSON `json:"params"`
	Mac       string          `json:"mac"`
}

type v1KdfParamsJSON struct {
	Dklen int    `json:"dklen"`
	N     int    `json:"n"`
	P     int    `json:"p"`
	R     int    `json:"r"`
	Salt  string `json:"salt"`
}

type v1Strategy struct {
}

func NewV1Strategy() *v1Strategy {
	return &v1Strategy{}
}

func (ks *v1Strategy) Encrypt(key *PrivateKey, passphrase string) ([]byte, error) {
	// derive key
	salt, err := newSalt()
	if nil != err {
		return nil, err
	}
	derivedKey, err := scrypt.Key([]byte(passphrase), salt, scryptN, scryptR, scryptP, scryptDKLen)
	if err != nil {
		return nil, err
	}

	// encrypt
	iv, err := newIV()
	if err != nil {
		return nil, err
	}
	encryptKey := derivedKey[:16]
	plaintext := key.Serialize()
	ciphertext, err := aesCTRXOR(encryptKey, plaintext, iv)
	if err != nil {
		return nil, err
	}

	// mac
	mac := generateMac(derivedKey, ciphertext)

	// json: cipher
	cipher := v1CipherJSON{
		Algorithm: cipherAlgorithm,
		Params: v1CipherParamsJSON{
			Iv: hex.Encode(iv),
		},
		Ciphertext: hex.Encode(ciphertext),
	}
	// json: kdf
	kdf := v1KdfJson{
		Algorithm: kdfAlgorithm,
		Params: v1KdfParamsJSON{
			Dklen: scryptDKLen,
			N:     scryptN,
			P:     scryptP,
			R:     scryptR,
			Salt:  hex.Encode(salt),
		},
		Mac: hex.Encode(mac),
	}
	rawAddress := GenerateAddress(&(key.ToECDSA().PublicKey))
	encodedAddress := types.EncodeAddress(rawAddress)
	keyFormat := v1KeyStoreFormat{
		Address: encodedAddress,
		Version: version,
		Cipher:  cipher,
		Kdf:     kdf,
	}

	return json.Marshal(keyFormat)
}

func (ks *v1Strategy) Decrypt(encrypted []byte, passphrase string) (*PrivateKey, error) {
	keyFormat := new(v1KeyStoreFormat)
	err := json.Unmarshal(encrypted, keyFormat)
	if nil != err {
		return nil, err
	}

	err = checkKeyFormat(keyFormat)
	if nil != err {
		return nil, err
	}

	cipher := keyFormat.Cipher
	kdf := keyFormat.Kdf

	// derive decrypt key
	derivedKey, err := deriveCipherKey([]byte(passphrase), kdf)
	if nil != err {
		return nil, err
	}

	// check mac
	mac, err := hex.Decode(kdf.Mac)
	if nil != err {
		return nil, err
	}
	cipherText, err := hex.Decode(cipher.Ciphertext)
	if nil != err {
		return nil, err
	}
	calculatedMac := generateMac(derivedKey, cipherText)
	if false == reflect.DeepEqual(mac, calculatedMac) {
		return nil, types.ErrWrongAddressOrPassWord
	}

	// decrypt
	decryptKey := derivedKey[:16]
	iv, err := hex.Decode(cipher.Params.Iv)
	if nil != err {
		return nil, err
	}
	plaintext, err := aesCTRXOR(decryptKey, cipherText, iv)
	if nil != err {
		return nil, err
	}

	privateKey, _ := btcec.PrivKeyFromBytes(plaintext)

	rawAddress := GenerateAddress(&(privateKey.ToECDSA().PublicKey))
	encodedAddress := types.EncodeAddress(rawAddress)
	if encodedAddress != keyFormat.Address {
		return nil, errors.New("Invalid matching address")
	}

	return privateKey, nil
}

func checkKeyFormat(keyFormat *v1KeyStoreFormat) error {
	if version != keyFormat.Version {
		return errors.New("Keystore version type")
	}

	if cipherAlgorithm != keyFormat.Cipher.Algorithm {
		return errors.New("Cipher algorithm must be " + cipherAlgorithm)
	}

	if kdfAlgorithm != keyFormat.Kdf.Algorithm {
		return errors.New("Kdf algorithm must be " + kdfAlgorithm)
	}

	return nil
}

func deriveCipherKey(passphrase []byte, kdf v1KdfJson) ([]byte, error) {
	salt, err := hex.Decode(kdf.Params.Salt)
	if err != nil {
		return nil, err
	}

	dkLen := kdf.Params.Dklen
	n := kdf.Params.N
	r := kdf.Params.R
	p := kdf.Params.P
	return scrypt.Key(passphrase, salt, n, r, p, dkLen)
}

func generateMac(derivedKey []byte, message []byte) []byte {
	concated := append(derivedKey[16:32], message[:]...)
	return hash(concated)
}

func hash(message []byte) []byte {
	h := sha256.New()
	h.Write(message)
	return h.Sum(nil)
}

func newSalt() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func newIV() ([]byte, error) {
	iv := make([]byte, aes.BlockSize) // 16
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	return iv, nil
}

func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	// AES-128 is selected due to size of encryptKey.
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}
