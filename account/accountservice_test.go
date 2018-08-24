package account

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
)

var (
	configName     = "server"
	configPaths    = []string{".", "./bin/", "$HOME/.aergo/"}
	configRequired = false

	configFilePath string
	as             *AccountService
	conf           *config.Config
)

func initTest() {
	serverCtx := config.NewServerContext("", "")
	conf := serverCtx.GetDefaultConfig().(*config.Config)
	conf.DataDir, _ = ioutil.TempDir("", "test")
	as = NewAccountService(conf)
	as.testConfig = true
	as.BeforeStart()
}

func deinitTest() {
	as.BeforeStop()
}

func TestNewAccountAndGet(t *testing.T) {
	initTest()
	defer deinitTest()
	var testaccounts []*types.Account
	testsize := 10
	for i := 0; i < testsize; i++ {
		passphrase := fmt.Sprintf("test%d", i)
		account, err := as.createAccount(passphrase)
		if err != nil {
			t.Errorf("failed to create account:%s", err)
		}
		testaccounts = append(testaccounts, account)
	}
	getlist := as.getAccounts()
	var resultlist []*types.Account
	for _, a := range getlist {
		for _, t := range testaccounts {
			if EncodeB64(a.Address) == EncodeB64(t.Address) {
				resultlist = append(resultlist, t)
				break
			}
		}
	}
	if len(resultlist) != len(testaccounts) {
		t.Error("failed to get account")
	}
}

const AddressLength = 20

func TestNewAccountAndUnlockLock(t *testing.T) {
	initTest()
	defer deinitTest()
	var testaccounts []*types.Account
	testsize := 10
	if len(as.unlocked) != 0 {
		t.Errorf("account service already has unlocked account:%d", len(as.unlocked))
	}
	for i := 0; i < testsize; i++ {
		passphrase := fmt.Sprintf("test%d", i)
		account, err := as.createAccount(passphrase)
		if err != nil {
			t.Errorf("failed to create account:%s", err)
		}
		if len(account.Address) != AddressLength {
			t.Errorf("invalid account len:%d", len(account.Address))
		}
		testaccounts = append(testaccounts, account)
	}
	for i := 0; i < testsize; i++ {
		passphrase := fmt.Sprintf("test%d", i)
		account, err := as.unlockAccount(testaccounts[i].Address, passphrase)
		if err != nil || account == nil {
			t.Errorf("failed to unlock account[%d]:%s", i, err)
		}
	}
	if len(testaccounts) != len(as.unlocked) {
		t.Error("failed to unlock account")
	}
	for i := 0; i < testsize; i++ {
		passphrase := fmt.Sprintf("test%d", i)
		account, err := as.lockAccount(testaccounts[i].Address, passphrase)
		if err != nil || account == nil {
			t.Errorf("failed to lock account[%d]:%s", i, err)
		}
	}
	if len(as.unlocked) != 0 {
		t.Errorf("account service remain unlocked account:%d", len(as.unlocked))
	}
}

func TestNewAccountAndUnlockFail(t *testing.T) {
	initTest()
	defer deinitTest()
	var testaccounts []*types.Account
	testsize := 10
	if len(as.unlocked) != 0 {
		t.Errorf("account service already has unlocked account:%d", len(as.unlocked))
	}
	for i := 0; i < testsize; i++ {
		passphrase := fmt.Sprintf("test%d", i)
		account, err := as.createAccount(passphrase)
		if err != nil {
			t.Errorf("failed to create account:%s", err)
		}
		if len(account.Address) != AddressLength {
			t.Errorf("invalid account len:%d", len(account.Address))
		}
		testaccounts = append(testaccounts, account)
	}
	for i := 0; i < testsize; i++ {
		passphrase := fmt.Sprintf("test_Error%d", i)
		account, err := as.unlockAccount(testaccounts[i].Address, passphrase)
		if err == nil || account != nil {
			t.Errorf("should not unlock the account[%d]:%s", i, err)
		}
		if err != message.ErrWrongAddressOrPassWord {
			t.Errorf("should return proper error code expect = %s, return = %s", message.ErrWrongAddressOrPassWord, err)
		}
	}
	if len(as.unlocked) != 0 {
		t.Error("unlock account with wrong pass")
	}
}
func TestNewAccountUnlockSignVerfiy(t *testing.T) {
	initTest()
	defer deinitTest()
	passphrase := "test"
	account, err := as.createAccount(passphrase)
	if err != nil {
		t.Errorf("failed to create account:%s", err)
	}
	unlockedAccount, err := as.unlockAccount(account.Address, passphrase)
	if err != nil || unlockedAccount == nil {
		t.Errorf("failed to unlock account:%s", err)
		t.FailNow()
	}
	tx := types.Tx{Body: &types.TxBody{Account: account.Address}}
	signer := NewSigner(as.Logger, as.unlocked[EncodeB64(account.Address)])
	err = signer.SignTx(&tx)
	if err != nil {
		t.Fatalf("failed to sign: %s", err)
	}
	if tx.Body.Sign == nil {
		t.Fatalf("failed to sign: %s", err)
	}
	err = as.verifyTx(&tx)
	if err != nil {
		t.Fatalf("failed to verify: %s", err)
	}
}

func TestVerfiyFail(t *testing.T) {
	initTest()
	defer deinitTest()
	passphrase := "test"
	account, err := as.createAccount(passphrase)
	if err != nil {
		t.Errorf("failed to create account:%s", err)
	}
	unlockedAccount, err := as.unlockAccount(account.Address, passphrase)
	if err != nil || unlockedAccount == nil {
		t.Errorf("failed to unlock account:%s", err)
		t.FailNow()
	}
	tx := types.Tx{Body: &types.TxBody{Account: account.Address}}
	signer := NewSigner(as.Logger, as.unlocked[EncodeB64(account.Address)])
	err = signer.SignTx(&tx)
	if err != nil {
		t.Fatalf("failed to sign: %s", err)
	}
	if tx.Body.Sign == nil {
		t.Fatalf("failed to sign: %s", err)
	}
	//edit tx after sign
	tx.Body.Amount = 0xff
	err = as.verifyTx(&tx)
	if err != message.ErrSignNotMatch {
		t.Errorf("should return :%s", message.ErrSignNotMatch)
	}
	if err == nil {
		t.Fatal("should not success to verify")
	}
}
