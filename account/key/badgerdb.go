/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec"
)

type BadgerStorage struct {
	sync.RWMutex
	db db.DB
}

const dbName = "account"

var addresses = []byte("ADDRESSES")

func NewBadgerStorage(storePath string) (*BadgerStorage, error) {
	dbPath := path.Join(storePath, dbName)

	db, err := openBadgerDb(dbPath)
	if nil != err {
		return nil, err
	}

	return &BadgerStorage{
		db: db,
	}, nil
}

func LoadBadgerStorage(storePath string) (*BadgerStorage, error) {
	dbPath := path.Join(storePath, dbName)

	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Not badger storage in %s\n", dbPath)
	}

	db, err := openBadgerDb(dbPath)
	if nil != err {
		return nil, err
	}

	return &BadgerStorage{
		db: db,
	}, nil
}

func (ks *BadgerStorage) Save(identity Identity, password string, key *PrivateKey) (Identity, error) {
	encryptkey := hashBytes(identity, []byte(password))
	storeKey := hashBytes(identity, encryptkey)

	encrypted, err := encrypt(identity, encryptkey, key.Serialize())
	if err != nil {
		return nil, err
	}

	existing := ks.db.Get(storeKey)
	if cap(existing) != 0 {
		return nil, errors.New("already exists")
	}

	listErr := ks.saveToList(identity)
	if listErr != nil {
		return nil, listErr
	}

	ks.db.Set(storeKey, encrypted)
	return identity, nil
}

func (ks *BadgerStorage) Load(identity Identity, password string) (*PrivateKey, error) {
	encryptkey := hashBytes(identity, []byte(password))
	storeKey := hashBytes(identity, encryptkey)

	encrypted := ks.db.Get(storeKey)
	if cap(encrypted) == 0 {
		return nil, types.ErrWrongAddressOrPassWord
	}

	decrypted, err := decrypt(identity, encryptkey, encrypted)
	if err != nil {
		return nil, types.ErrWrongAddressOrPassWord
	}

	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), decrypted)
	return privateKey, nil
}

func (ks *BadgerStorage) List() ([]Identity, error) {
	ks.RWMutex.RLock()
	defer ks.RWMutex.RUnlock()

	b := ks.db.Get(addresses)
	var ret []Identity
	for i := 0; i < len(b); i += types.AddressLength {
		ret = append(ret, b[i:i+types.AddressLength])
	}
	return ret, nil
}

func (ks *BadgerStorage) Close() {
	ks.db.Close()
}

func openBadgerDb(path string) (opened db.DB, err error) {
	defer func() {
		failed := recover()
		if nil != failed {
			err = fmt.Errorf("%s", failed)
		}
	}()
	opened = db.NewDB(db.LevelImpl, path)
	return opened, err
}

func (ks *BadgerStorage) saveToList(identity Identity) error {
	if len(identity) != types.AddressLength {
		return errors.New("invalid address length")
	}

	ks.RWMutex.Lock()
	defer ks.RWMutex.Unlock()

	identities := ks.db.Get(addresses)
	newIdentities := append(identities, identity...)
	ks.db.Set(addresses, newIdentities)
	return nil
}
