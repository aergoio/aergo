package account

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

var (
	configName     = "server"
	configPaths    = []string{".", "./bin/", "$HOME/.aergo/"}
	configRequired = false

	configFilePath string
	as             *AccountService
	conf           *config.Config
)

var sdb *state.ChainStateDB

func initTest() {
	serverCtx := config.NewServerContext("", "")
	conf := serverCtx.GetDefaultConfig().(*config.Config)
	conf.DataDir, _ = ioutil.TempDir("", "test")

	sdb = state.NewChainStateDB()
	testmode := true
	sdb.Init(string(db.BadgerImpl), conf.DataDir, nil, testmode)

	as = NewAccountService(conf, sdb)
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
	testsize := 5
	for i := 0; i < testsize; i++ {
		passphrase := fmt.Sprintf("test%d", i)
		account, err := as.createAccount(passphrase)
		assert.NoError(t, err, "failed to create account")
		assert.Equalf(t, types.AddressLength, len(account.Address), "wrong address length : %s", account.Address)

		testaccounts = append(testaccounts, account)
	}
	getlist := as.getAccounts()
	var resultlist []*types.Account
	for _, a := range getlist {
		for _, t := range testaccounts {
			if types.EncodeAddress(a.Address) == types.EncodeAddress(t.Address) {
				resultlist = append(resultlist, t)
				break
			}
		}
	}
	assert.Len(t, resultlist, len(testaccounts), "failed to get account")
}

func TestNewAccountAndUnlockLock(t *testing.T) {
	initTest()
	defer deinitTest()
	var testaccounts []*types.Account
	testsize := 3
	for i := 0; i < testsize; i++ {
		passphrase := fmt.Sprintf("test%d", i)
		account, err := as.createAccount(passphrase)
		assert.NoError(t, err, "failed to create account")
		assert.Equalf(t, types.AddressLength, len(account.Address), "wrong address length : %s", account.Address)

		testaccounts = append(testaccounts, account)
	}
	for i := 0; i < testsize; i++ {
		passphrase := fmt.Sprintf("test%d", i)
		account, err := as.unlockAccount(testaccounts[i].Address, passphrase)
		if err != nil || account == nil {
			t.Errorf("failed to unlock account[%d]:%s", i, err)
		}
	}
	for i := 0; i < testsize; i++ {
		passphrase := fmt.Sprintf("test%d", i)
		account, err := as.lockAccount(testaccounts[i].Address, passphrase)
		if err != nil || account == nil {
			t.Errorf("failed to lock account[%d]:%s", i, err)
		}
	}
}

func TestNewAccountAndUnlockFail(t *testing.T) {
	initTest()
	defer deinitTest()
	var testaccounts []*types.Account
	testsize := 3
	for i := 0; i < testsize; i++ {
		passphrase := fmt.Sprintf("test%d", i)
		account, err := as.createAccount(passphrase)
		assert.NoError(t, err, "failed to create account")
		assert.Equalf(t, types.AddressLength, len(account.Address), "wrong address length : %s", account.Address)

		testaccounts = append(testaccounts, account)
	}
	for i := 0; i < testsize; i++ {
		passphrase := fmt.Sprintf("test_Error%d", i)
		account, err := as.unlockAccount(testaccounts[i].Address, passphrase)
		if err == nil || account != nil {
			t.Errorf("should not unlock the account[%d]:%s", i, err)
		}
		if err != types.ErrWrongAddressOrPassWord {
			t.Errorf("should return proper error code expect = %s, return = %s", types.ErrWrongAddressOrPassWord, err)
		}
	}
}

func TestNewAccountUnlockSignVerfiy(t *testing.T) {
	initTest()
	defer deinitTest()
	passphrase := "test"
	account, err := as.createAccount(passphrase)
	assert.NoError(t, err, "failed to create account")
	assert.Equalf(t, types.AddressLength, len(account.Address), "wrong address length : %s", account.Address)

	unlockedAccount, err := as.unlockAccount(account.Address, passphrase)
	if err != nil || unlockedAccount == nil {
		t.Errorf("failed to unlock account:%s", err)
		t.FailNow()
	}
	tx := &types.Tx{Body: &types.TxBody{Account: account.Address}}
	err = as.ks.SignTx(tx, nil)
	assert.NoError(t, err, "failed to sign")
	assert.NotNil(t, tx.Body.Sign, "failed to sign")

	err = as.ks.VerifyTx(tx)
	assert.NoError(t, err, "failed to verify")
}

func TestVerfiyFail(t *testing.T) {
	initTest()
	defer deinitTest()
	passphrase := "test"
	account, err := as.createAccount(passphrase)
	assert.NoError(t, err, "failed to create account")
	assert.Equalf(t, types.AddressLength, len(account.Address), "wrong address length : %s", account.Address)

	unlockedAccount, err := as.unlockAccount(account.Address, passphrase)
	if err != nil || unlockedAccount == nil {
		t.Errorf("failed to unlock account:%s", err)
		t.FailNow()
	}

	tx := &types.Tx{Body: &types.TxBody{Account: account.Address}}
	err = as.ks.SignTx(tx, nil)
	assert.NoError(t, err, "failed to sign")
	assert.NotNil(t, tx.Body.Sign, "failed to sign")

	//edit tx after sign
	tx.Body.Amount = new(big.Int).SetUint64(0xff).Bytes()
	err = as.ks.VerifyTx(tx)
	assert.Error(t, err, types.ErrSignNotMatch, "failed to verfiy")
}

func TestBase58CheckEncoding(t *testing.T) {
	initTest()
	defer deinitTest()
	addr := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33}

	encoded := types.EncodeAddress(addr)
	expected := "AmJaNDXoPbBRn9XHh9onKbDKuAzj88n5Bzt7KniYA78qUEc5EwBd"
	assert.Equal(t, expected, encoded, "incorrectly encoded address")

	decoded, _ := types.DecodeAddress(encoded)
	assert.Equal(t, addr, decoded, "incorrectly decoded address")

	_, err := types.DecodeAddress("AmJaNDXoPbBRn9XHh9onKbDKuAzj88n5Bzt7KniYA78qUEc5EwBA")
	assert.NotEmpty(t, err, "decoding address with wrong checksum")
}
