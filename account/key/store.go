/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"sync"
	"time"

	crypto "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec/v2"
)

type aergokey = btcec.PrivateKey

type keyPair struct {
	key   *aergokey
	timer *time.Timer
}

// Store stucture of keystore
type Store struct {
	sync.RWMutex
	timeout      time.Duration
	unlocked     map[string]*keyPair
	unlockedLock *sync.Mutex
	storage      Storage
}

// NewStore make new instance of keystore
func NewStore(storePath string, unlockTimeout uint) *Store {
	store := &Store{
		timeout:      time.Duration(unlockTimeout) * time.Second,
		unlocked:     map[string]*keyPair{},
		unlockedLock: &sync.Mutex{},
	}

	// FIXME: more elegant coding
	if storage, err := LoadBadgerStorage(storePath); nil == err {
		store.storage = storage
	} else {
		storage, err := NewAergoStorage(storePath)
		if nil != err {
			panic(err)
		}
		store.storage = storage
	}

	return store
}

// CloseStore locks all addresses and closes the storage
func (ks *Store) CloseStore() {
	ks.unlocked = nil
	ks.storage.Close()
}

// CreateKey make new key in keystore and return it's address
func (ks *Store) CreateKey(pass string) (Identity, error) {
	//gen new key
	privkey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, err
	}
	return ks.addKey(privkey, pass)
}

// ImportKey is to import encrypted key
func (ks *Store) ImportKey(imported []byte, oldpass string, newpass string) (Identity, error) {
	hash := hashBytes([]byte(oldpass), nil)
	rehash := hashBytes([]byte(oldpass), hash)
	key, err := decrypt(hash, rehash, imported)
	if err != nil {
		return nil, err
	}
	privkey, _ := btcec.PrivKeyFromBytes(key)
	idendity, err := ks.addKey(privkey, newpass)
	if err != nil {
		address := crypto.GenerateAddress(privkey.PubKey().ToECDSA())
		return address, err
	}
	return idendity, nil
}

// ExportKey is to export encrypted key
func (ks *Store) ExportKey(addr Identity, pass string) ([]byte, error) {
	key, err := ks.getKey(addr, pass)
	if key == nil {
		return nil, err
	}
	return EncryptKey(key.Serialize(), pass)
}

// EncryptKey encrypts a key with a given export for exporting
func EncryptKey(key []byte, pass string) ([]byte, error) {
	hash := hashBytes([]byte(pass), nil)
	rehash := hashBytes([]byte(pass), hash)
	return encrypt(hash, rehash, key)
}

// Unlock is to unlock account for signing
func (ks *Store) Unlock(addr Identity, pass string) (Identity, error) {
	pk, err := ks.getKey(addr, pass)
	if err != nil {
		return nil, err
	}
	addrKey := types.EncodeAddress(addr)

	ks.unlockedLock.Lock()
	defer ks.unlockedLock.Unlock()

	unlockedKeyPair, exist := ks.unlocked[addrKey]

	if ks.timeout == 0 {
		ks.unlocked[addrKey] = &keyPair{key: pk, timer: nil}
		return addr, nil
	}

	if exist {
		unlockedKeyPair.timer.Reset(ks.timeout)
	} else {
		lockTimer := time.AfterFunc(ks.timeout,
			func() {
				ks.Lock(addr, pass)
			},
		)
		ks.unlocked[addrKey] = &keyPair{key: pk, timer: lockTimer}
	}
	return addr, nil
}

// Lock locks an account
func (ks *Store) Lock(addr Identity, pass string) (Identity, error) {
	_, err := ks.getKey(addr, pass)
	if err != nil {
		return nil, err
	}
	b58addr := types.EncodeAddress(addr)

	ks.unlockedLock.Lock()
	defer ks.unlockedLock.Unlock()

	if _, exist := ks.unlocked[b58addr]; exist {
		ks.unlocked[b58addr] = nil
		delete(ks.unlocked, b58addr)
	}
	return addr, nil
}

// GetAddresses returns the list of stored addresses
func (ks *Store) GetAddresses() ([]Identity, error) {
	return ks.storage.List()
}

func (ks *Store) getKey(address []byte, pass string) (*aergokey, error) {
	return ks.storage.Load(address, pass)
}

func (ks *Store) GetKey(address []byte, pass string) (*aergokey, error) {
	return ks.getKey(address, pass)
}

func (ks *Store) addKey(key *btcec.PrivateKey, pass string) (Identity, error) {
	address := crypto.GenerateAddress(key.PubKey().ToECDSA())
	return ks.storage.Save(address, pass, key)
}

func (ks *Store) AddKey(key *btcec.PrivateKey, pass string) (Identity, error) {
	return ks.addKey(key, pass)
}
