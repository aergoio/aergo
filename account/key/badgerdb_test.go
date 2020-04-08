/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	crypto "github.com/aergoio/aergo/account/key/crypto"
	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/assert"
)

func TestSaveAndLoadOnBadger(t *testing.T) {
	dir, _ := ioutil.TempDir("", "tmp")
	defer os.RemoveAll(dir)

	storage, _ := NewBadgerStorage(dir)
	expected, err := btcec.NewPrivateKey(btcec.S256())
	if nil != err {
		assert.FailNow(t, "Could not create private key", err)
	}

	identity := crypto.GenerateAddress(&expected.PublicKey)
	password := "password"
	saved, err := storage.Save(identity, password, expected)
	if nil != err {
		assert.FailNow(t, "Could not save private key", err)
	}
	assert.Equalf(t, identity, saved, "Returned one isn't same")

	actual, err := storage.Load(identity, password)
	if nil != err {
		assert.FailNow(t, "Could not load private key", err)
	}

	assert.Equalf(t, *expected, *actual, "Wrong exported one")
}

func TestSaveAndListOnBadger(t *testing.T) {
	dir, _ := ioutil.TempDir("", "tmp")
	defer os.RemoveAll(dir)

	storage, _ := NewBadgerStorage(dir)
	expected, err := btcec.NewPrivateKey(btcec.S256())
	if nil != err {
		assert.FailNow(t, "Could not create private key", err)
	}

	identity := crypto.GenerateAddress(&expected.PublicKey)
	password := "password"
	saved, err := storage.Save(identity, password, expected)
	if nil != err {
		assert.FailNow(t, "Could not save private key", err)
	}
	assert.Equalf(t, identity, saved, "Returned one isn't same")

	list, err := storage.List()
	if nil != err {
		assert.FailNow(t, "Could list stored addresses", err)
	}

	found := false
	for _, i := range list {
		if reflect.DeepEqual(identity, i) {
			found = true
			break
		}
	}
	assert.True(t, found, "Cound not found stored identity")
}

func TestDeleteOnBadgerDb(t *testing.T) {
	dir, _ := ioutil.TempDir("", "tmp")
	defer os.RemoveAll(dir)

	storage, err := NewBadgerStorage(dir)
	if nil != err {
		assert.FailNow(t, "Could not create storage", err)
	}

	// generate private keys and store them
	password := "password"
	idLen := 3
	var idArr = make([]Identity, idLen)
	for i := 0; i < idLen; i++ {
		privKey, err := btcec.NewPrivateKey(btcec.S256())
		if nil != err {
			assert.FailNow(t, "Could not create private key", err)
		}

		idArr[i] = crypto.GenerateAddress(&privKey.PublicKey)
		saved, err := storage.Save(idArr[i], password, privKey)
		if nil != err {
			assert.FailNow(t, "Could not save private key", err)
		}
		assert.Equalf(t, idArr[i], saved, "Returned one isn't same")
	}

	for i := 0; i < idLen; i++ {
		list, err := storage.List()
		if nil != err {
			assert.FailNow(t, "Could list stored addresses", err)
		}
		assert.Equal(t, len(list), idLen - i)

		storage.Delete(idArr[i], password)
	}
}
