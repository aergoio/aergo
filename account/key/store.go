package key

import (
	"crypto/aes"
	"crypto/cipher"
	"os"
	"path"

	sha256 "github.com/minio/sha256-simd"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
)

type aergokey = btcec.PrivateKey

// Store stucture of keystore
type Store struct {
	unlocked  map[string]*aergokey
	addresses string
	storage   db.DB
}

// NewStore make new instance of keystore
func NewStore(storePath string) *Store {
	const dbName = "account"
	dbPath := path.Join(storePath, dbName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	const addressFile = "addresses"
	addrPath := path.Join(storePath, addressFile)

	return &Store{
		unlocked:  map[string]*aergokey{},
		addresses: addrPath,
		storage:   db.NewDB(db.BadgerImpl, dbPath),
	}
}
func (ks *Store) DestroyStore() {
	ks.addresses = ""
	ks.storage.Close()
}

// CreateKey make new key in keystore and return it's address
func (ks *Store) CreateKey(pass string) (Address, error) {
	//gen new key
	privkey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	//gen new address
	address := GenerateAddress(&privkey.PublicKey)

	//save pass/address/key
	encryptkey := hashBytes(address, []byte(pass))
	encrypted, err := encrypt(address, encryptkey, privkey.Serialize())
	if err != nil {
		return nil, err
	}
	ks.storage.Set(hashBytes(address, encryptkey), encrypted)

	return address, nil
}

func (ks *Store) Unlock(addr Address, pass string) (Address, error) {
	key, err := ks.getKey(addr, pass)
	if key == nil {
		return nil, err
	}
	ks.unlocked[types.EncodeAddress(addr)], _ = btcec.PrivKeyFromBytes(btcec.S256(), key)
	return addr, nil
}

func (ks *Store) Lock(addr Address, pass string) (Address, error) {
	key, err := ks.getKey(addr, pass)
	if key == nil {
		return nil, err
	}
	b58addr := types.EncodeAddress(addr)
	ks.unlocked[b58addr] = nil
	delete(ks.unlocked, b58addr)
	return addr, nil
}

func (ks *Store) getKey(address []byte, passphrase string) ([]byte, error) {
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
