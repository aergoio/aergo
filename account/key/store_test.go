package key

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aergoio/aergo/types"
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
	const testSize = 10
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
	const testSize = 10
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

func TestExportKey(t *testing.T) {
	initTest()
	defer deinitTest()
	const testSize = 10
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
	const testSize = 10
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
	const testSize = 10
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
