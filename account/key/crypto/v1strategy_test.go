/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/assert"
)

func TestDecrypt(t *testing.T) {
	encrypted, err := ioutil.ReadFile("./test/AmMzdvQMremMBwFqrVWc4Urq1VaFPmbAGY7VBtfDnCXE93vSct7H__keystore.txt")
	if nil != err {
		assert.FailNow(t, "Could not read keystore file", err)
	}

	strategy := NewV1Strategy()
	password := "password"
	decrypted, err := strategy.Decrypt(encrypted, password)
	if nil != err {
		assert.FailNow(t, "Could not decrypt private key", err)
	}
	assert.NotNil(t, decrypted)
}

func TestEncryptAndDecrypt(t *testing.T) {
	dir, _ := ioutil.TempDir("", "tmp")
	defer os.RemoveAll(dir)

	for i := 0; i < 2; i++ {
		expected, err := btcec.NewPrivateKey()
		if nil != err {
			assert.FailNow(t, "Could not create private key", err)
		}

		strategy := NewV1Strategy()
		password := "password"
		encrypted, err := strategy.Encrypt(expected, password)
		if nil != err {
			assert.FailNow(t, "Could not save private key", err)
		}

		actual, err := strategy.Decrypt(encrypted, password)
		if nil != err {
			assert.FailNow(t, "Could not decrypt private key", err)
		}

		assert.Equalf(t, *expected, *actual, "Decrypted one is different with origin one")
	}
}
