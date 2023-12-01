/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"fmt"
	"io/ioutil"
	"sync"
	"testing"

	crypto "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/assert"
)

var (
	testDir string
	ks      *Store
)

func initTest() {
	testDir, _ = ioutil.TempDir("", "test")
	ks = NewStore(testDir, 0)
}

func deinitTest() {
	ks.CloseStore()
}
func TestCreateKey(t *testing.T) {
	initTest()
	defer deinitTest()
	const testSize = 5
	for i := 0; i < testSize; i++ {
		pass := fmt.Sprintf("%d", i)
		addr, err := ks.CreateKey(pass)
		if err != nil {
			t.Errorf("could not create key : %s", err.Error())
		}
		if len(addr) != types.AddressLength {
			t.Errorf("invalid address created : length = %d", len(addr))
		}
	}
}

func TestCreateKeyLongPass(t *testing.T) {
	initTest()
	defer deinitTest()
	const testSize = 3
	for i := 0; i < testSize; i++ {
		pass := fmt.Sprintf("%1024d", i)
		addr, err := ks.CreateKey(pass)
		if err != nil {
			t.Errorf("could not create key : %s", err.Error())
		}
		if len(addr) != types.AddressLength {
			t.Errorf("invalid address created : length = %d", len(addr))
		}
	}
}

func TestImportKey(t *testing.T) {
	initTest()
	defer deinitTest()
	const testSize = 3
	for i := 0; i < testSize; i++ {
		key, err := btcec.NewPrivateKey()
		addr := crypto.GenerateAddress(&(key.PublicKey))
		if err != nil {
			t.Errorf("could not create key : %s", err.Error())
		}
		pass := fmt.Sprintf("%d", i)
		encrypted, err := EncryptKey(key.Serialize(), pass)
		if err != nil {
			t.Errorf("could not encrypt key : %s", err.Error())
		}

		newPass := fmt.Sprintf("new%d", i)
		imported, err := ks.ImportKey(encrypted, pass, newPass)
		assert.NoError(t, err, "import")
		assert.Equal(t, addr, imported, "import result")
	}
}

func TestExportKey(t *testing.T) {
	initTest()
	defer deinitTest()
	const testSize = 3
	for i := 0; i < testSize; i++ {
		pass := fmt.Sprintf("%d", i)
		addr, err := ks.CreateKey(pass)
		if err != nil {
			t.Errorf("could not create key : %s", err.Error())
		}
		if len(addr) != types.AddressLength {
			t.Errorf("invalid address created : length = %d", len(addr))
		}
		exported, err := ks.ExportKey(addr, pass)
		if err != nil {
			t.Errorf("could not export key : %s", err.Error())
		}
		if len(exported) != 48 {
			t.Errorf("invalid exported address : length = %d", len(exported))
		}
	}
}

func TestSignTx(t *testing.T) {
	initTest()
	defer deinitTest()
	const testSize = 3
	for i := 0; i < testSize; i++ {
		pass := fmt.Sprintf("%32d", i)
		addr, err := ks.CreateKey(pass)
		if err != nil {
			t.Errorf("could not create key : %s", err.Error())
		}
		if len(addr) != types.AddressLength {
			t.Errorf("invalid address created : length = %d", len(addr))
		}
		unlocked, err := ks.Unlock(addr, pass)
		if err != nil {
			t.Errorf("could not unlock address: %s", err.Error())
		}
		if len(unlocked) != types.AddressLength {
			t.Errorf("invalid unlock address : length = %d", len(unlocked))
		}
		tx := &types.Tx{Body: &types.TxBody{Account: addr}}
		err = ks.SignTx(tx, nil) //TODO : improve
		if err != nil {
			t.Errorf("could not sign : %s", err.Error())
		}
		if tx.Body.Sign == nil {
			t.Errorf("sign is nil : %s", tx.String())
		}
	}
}

func TestSign(t *testing.T) {
	initTest()
	defer deinitTest()
	const testSize = 3
	for i := 0; i < testSize; i++ {
		pass := fmt.Sprintf("%32d", i)
		addr, err := ks.CreateKey(pass)
		if err != nil {
			t.Errorf("could not create key : %s", err.Error())
		}
		if len(addr) != types.AddressLength {
			t.Errorf("invalid address created : length = %d", len(addr))
		}
		hash := []byte(pass)
		_, err = ks.Sign(addr, pass, hash) //TODO : improve
		if err != nil {
			t.Errorf("could not sign : %s", err.Error())
		}
	}
}

func TestConcurrentUnlockAndLock(t *testing.T) {
	initTest()
	defer deinitTest()

	pass := "password"
	addr, err := ks.CreateKey(pass)
	if err != nil {
		t.Errorf("could not create key : %s", err.Error())
	}

	const testSize = 5
	var wg sync.WaitGroup
	for i := 0; i < testSize; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, id int) {
			defer wg.Done()
			ks.Unlock(addr, pass)
			ks.Lock(addr, pass)
		}(&wg, i)
	}
	wg.Wait()
}
