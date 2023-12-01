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

	crypto "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/assert"
)

func TestSaveAndLoadOnBadger(t *testing.T) {
	dir, _ := ioutil.TempDir("", "tmp")
	defer os.RemoveAll(dir)

	storage, _ := NewBadgerStorage(dir)
	expected, err := btcec.NewPrivateKey()
	if nil != err {
		assert.FailNow(t, "Could not create private key", err)
	}

	identity := crypto.GenerateAddress(expected.PubKey().ToECDSA())
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
	expected, err := btcec.NewPrivateKey()
	if nil != err {
		assert.FailNow(t, "Could not create private key", err)
	}

	identity := crypto.GenerateAddress(expected.PubKey().ToECDSA())
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
