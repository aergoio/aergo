package account

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"os"
	"path"
	"sync"

	"github.com/aergoio/aergo-actor/actor"

	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/pkg/db"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
)

type aergokey = btcec.PrivateKey

type AccountService struct {
	*component.BaseComponent
	cfg         *cfg.Config
	accountLock sync.RWMutex
	accounts    []*types.Account
	unlocked    map[string]*aergokey
	storage     db.DB
	addrs       *Addresses
	testConfig  bool
}

var _ component.IComponent = (*AccountService)(nil)

//NewAccountService create account service
func NewAccountService(cfg *cfg.Config) *AccountService {
	return &AccountService{
		BaseComponent: component.NewBaseComponent(message.AccountsSvc, log.NewLogger(log.AccountsSvc), cfg.EnableDebugMsg),
		cfg:           cfg,
		accounts:      []*types.Account{},
		unlocked:      map[string]*aergokey{},
	}
}

func (as *AccountService) Start() {
	const dbName = "account"
	const addressFile = "addresses"
	if as.testConfig {

	} else {
		as.BaseComponent.Start(as)
	}

	//TODO: fix it store secure storage
	dbPath := path.Join(as.cfg.DataDir, dbName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	as.storage = db.NewDB(db.BadgerImpl, dbPath)

	addrPath := path.Join(as.cfg.DataDir, addressFile)
	as.addrs = NewAddresses(as.Logger, addrPath)
	as.accounts, _ = as.addrs.getAccounts()
}

func (as *AccountService) Stop() {
	as.accounts = nil
	as.unlocked = nil
	if as.storage != nil {
		as.storage.Close()
	}
	as.addrs = nil
	if as.testConfig {

	} else {
		as.BaseComponent.Stop()
	}
}

func (as *AccountService) Receive(context actor.Context) {
	as.BaseComponent.Receive(context)
	switch msg := context.Message().(type) {
	case *message.GetAccounts:
		accountList := as.getAccounts()
		context.Respond(&message.GetAccountsRsp{Accounts: &types.AccountList{Accounts: accountList}})
	case *message.CreateAccount:
		account, _ := as.createAccount(msg.Passphrase)
		context.Respond(&message.CreateAccountRsp{Account: account})
	case *message.LockAccount:
		account, err := as.lockAccount(msg.Account.Address, msg.Passphrase)
		context.Respond(&message.AccountRsp{Account: account, Err: err})
	case *message.UnlockAccount:
		account, err := as.unlockAccount(msg.Account.Address, msg.Passphrase)
		context.Respond(&message.AccountRsp{Account: account, Err: err})
	case *message.SignTx:
		err := as.signTx(msg.Tx)
		if err != nil {
			context.Respond(&message.SignTxRsp{Tx: nil, Err: err})
		} else {
			context.Respond(&message.SignTxRsp{Tx: msg.Tx, Err: nil})
		}
	case *message.VerifyTx:
		err := as.verifyTx(msg.Tx)
		if err != nil {
			context.Respond(&message.VerifyTxRsp{Tx: nil, Err: err})
		} else {
			context.Respond(&message.VerifyTxRsp{Tx: msg.Tx, Err: nil})
		}
	}
}

//TODO: refactoring util function
func EncodeB64(bs []byte) string {
	return base64.StdEncoding.EncodeToString(bs)
}

func DecodeB64(sb string) []byte {
	buf, _ := base64.StdEncoding.DecodeString(sb)
	return buf
}

func (as *AccountService) getAccounts() []*types.Account {
	as.accountLock.RLock()
	defer as.accountLock.RUnlock()
	return as.accounts
}

func (as *AccountService) createAccount(passphrase string) (*types.Account, error) {
	//gen new key
	privkey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		as.Errorf("could not generate key : %s", err)
		return nil, err
	}
	//gen new address
	address := generateAddress(&privkey.PublicKey)

	//save pass/address/key
	encryptkey := hashBytes(address, []byte(passphrase))
	encrypted, err := encrypt(address, encryptkey, privkey.Serialize())
	if err != nil {
		as.Errorf("could not encrypt key : %s", err)
		return nil, err
	}
	as.storage.Set(hashBytes(address, encryptkey), encrypted)
	account := types.NewAccount(address)

	//append list
	as.accountLock.Lock()
	//TODO: performance turning here
	as.addrs.addAddress(address)
	as.accounts = append(as.accounts, account)
	as.accountLock.Unlock()
	return account, nil
}

func generateAddress(pubkey *ecdsa.PublicKey) []byte {
	addr := new(bytes.Buffer)
	binary.Write(addr, binary.LittleEndian, pubkey.X.Bytes())
	binary.Write(addr, binary.LittleEndian, pubkey.Y.Bytes())
	return addr.Bytes()[:20] //TODO: ADDRESSLENGTH ?
}

func (as *AccountService) getKey(address []byte, passphrase string) ([]byte, error) {
	encryptkey := hashBytes(address, []byte(passphrase))
	key := as.storage.Get(hashBytes(address, encryptkey))
	if cap(key) == 0 {
		return nil, message.ErrWrongAddressOrPassWord
	}
	return decrypt(address, encryptkey, key)
}

func (as *AccountService) unlockAccount(address []byte, passphrase string) (*types.Account, error) {

	key, err := as.getKey(address, passphrase)
	if key == nil {
		as.Errorf("could not find the key : %s", err)
		return nil, err
	}
	as.unlocked[EncodeB64(address)], _ = btcec.PrivKeyFromBytes(btcec.S256(), key)
	return &types.Account{Address: address}, nil
}

func (as *AccountService) lockAccount(address []byte, passphrase string) (*types.Account, error) {
	key, err := as.getKey(address, passphrase)
	if key == nil {
		as.Errorf("could not load the key : %s", err)
		return nil, err
	}
	//TODO: zeroing key
	b64addr := EncodeB64(address)
	//TODO: lock
	as.unlocked[b64addr] = nil
	delete(as.unlocked, b64addr)

	return &types.Account{Address: address}, nil
}

func hashBytes(b1 []byte, b2 []byte) []byte {
	h := sha256.New()
	h.Write(b1)
	h.Write(b2)
	return h.Sum(nil)
}

func hashWithoutSign(txbody *types.TxBody) []byte {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, txbody.Nonce)
	h.Write(txbody.Account)
	h.Write(txbody.Recipient)
	binary.Write(h, binary.LittleEndian, txbody.Amount)
	h.Write(txbody.Payload)
	return h.Sum(nil)
}

func (as *AccountService) signTx(tx *types.Tx) error {
	txbody := tx.Body
	//hash tx
	hash := hashWithoutSign(txbody)
	//get key
	key, exist := as.unlocked[EncodeB64(txbody.Account)]
	if !exist {
		return message.ErrShouldUnlockAccount
	}
	//sign tx
	//TODO: race condition? if lock account at this moment ...
	sign, err := btcec.SignCompact(btcec.S256(), key, hash, true)
	if err != nil {
		as.Errorf("could not sign: %s", err)
		return err
	}
	txbody.Sign = sign
	//txbody.Sign = sign
	/*
		sign, _ := key.Sign(hash)
		txbody.Sign = sign.Serialize()
	*/
	tx.Hash = tx.CalculateTxHash()
	return nil
}

func (as *AccountService) verifyTx(tx *types.Tx) error {
	txbody := tx.Body
	hash := hashWithoutSign(txbody)

	pubkey, _, err := btcec.RecoverCompact(btcec.S256(), txbody.Sign, hash)
	if err != nil {
		as.Errorf("could not recover sign: %s", err)
		return err
	}
	address := generateAddress(pubkey.ToECDSA())
	if !bytes.Equal(address, txbody.Account) {
		return message.ErrSignNotMatch
	}
	return nil
}

func encrypt(address, key, data []byte) ([]byte, error) {
	// Load your secret key from a safe place and reuse it across multiple
	// Seal/Open calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	// When decoded the key should be 16 bytes (AES-128) or 32 (AES-256).
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
	// Load your secret key from a safe place and reuse it across multiple
	// Seal/Open calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	// When decoded the key should be 16 bytes (AES-128) or 32 (AES-256).
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
