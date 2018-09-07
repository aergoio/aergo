package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/message"
	"github.com/btcsuite/btcd/btcec"
)

// KeyStore stucture of keystore
type KeyStore struct {
	addresses string
	storage   db.DB
}

// NewKeyStore make new instance of keystore
func NewKeyStore(storePath string) *KeyStore {
	const dbName = "account"
	dbPath := path.Join(storePath, dbName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	const addressFile = "addresses"
	addrPath := path.Join(storePath, addressFile)

	return &KeyStore{
		addresses: addrPath,
		storage:   db.NewDB(db.BadgerImpl, dbPath),
	}
}

// CreateKey make new key in keystore and return it's address
func (ks *KeyStore) CreateKey(pass string) (Address, error) {
	//gen new key
	privkey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	//gen new address
	address := generateAddress(&privkey.PublicKey)

	//save pass/address/key
	encryptkey := hashBytes(address, []byte(pass))
	encrypted, err := encrypt(address, encryptkey, privkey.Serialize())
	if err != nil {
		return nil, err
	}
	ks.storage.Set(hashBytes(address, encryptkey), encrypted)

	return address, nil
}

func (ks *KeyStore) SaveAddress(address Address) error {
	f, err := os.OpenFile(ks.addresses, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.Write(address); err != nil {
		return err
	}
	return nil
}

func (ks *KeyStore) GetAddresses() ([]Address, error) {
	const addressLength = 20
	b, err := ioutil.ReadFile(ks.addresses)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ret []Address
	for i := 0; i < len(b); i += addressLength {
		ret = append(ret, b[i:i+addressLength])
	}
	return ret, nil
}

func (ks *KeyStore) getKey(address []byte, passphrase string) ([]byte, error) {
	encryptkey := hashBytes(address, []byte(passphrase))
	key := ks.storage.Get(hashBytes(address, encryptkey))
	if cap(key) == 0 {
		return nil, message.ErrWrongAddressOrPassWord
	}
	return decrypt(address, encryptkey, key)
}

func hashBytes(b1 []byte, b2 []byte) []byte {
	h := sha256.New()
	h.Write(b1)
	h.Write(b2)
	return h.Sum(nil)
}

func encrypt(address, key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := address[:12]

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	cipherbytes := aesgcm.Seal(nil, nonce, data, nil)
	return cipherbytes, nil
}

func decrypt(address, key, data []byte) ([]byte, error) {
	nonce := address[:12]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainbytes, err := aesgcm.Open(nil, nonce, data, nil)

	if err != nil {
		return nil, err
	}
	return plainbytes, nil
}
