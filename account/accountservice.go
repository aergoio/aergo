package account

import (
	"sync"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/account/key"
	cfg "github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/contract/name"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

type AccountService struct {
	*component.BaseComponent
	cfg         *cfg.Config
	sdb         *state.ChainStateDB
	ks          *key.Store
	accountLock sync.RWMutex
	accounts    []*types.Account
	testConfig  bool
}

// NewAccountService create account service
func NewAccountService(cfg *cfg.Config, sdb *state.ChainStateDB) *AccountService {
	actor := &AccountService{
		cfg: cfg,
		sdb: sdb,
	}
	actor.BaseComponent = component.NewBaseComponent(message.AccountsSvc, actor, log.NewLogger("account"))

	return actor
}

func (as *AccountService) BeforeStart() {
	as.ks = key.NewStore(as.cfg.DataDir, as.cfg.Account.UnlockTimeout)

	as.accounts = []*types.Account{}
	addresses, err := as.ks.GetAddresses()
	if err != nil {
		as.Logger.Error().Err(err).Msg("could not open addresses")
	}
	for _, v := range addresses {
		as.accounts = append(as.accounts, &types.Account{Address: v})
	}
}

func (as *AccountService) AfterStart() {}

func (as *AccountService) BeforeStop() {
	as.ks.CloseStore()
	as.accounts = nil
}

func (as *AccountService) Statistics() *map[string]interface{} {
	return &map[string]interface{}{
		"totalaccounts": len(as.accounts),
		"personal":      as.cfg.Personal,
		"config":        as.cfg.Account,
	}
}
func (as *AccountService) resolveName(namedAddress []byte) ([]byte, error) {
	scs, err := as.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte(types.AergoName)))
	if err != nil {
		return nil, err
	}
	return name.GetAddress(scs, namedAddress), nil
}

func (as *AccountService) Receive(context actor.Context) {

	switch msg := context.Message().(type) {
	case *message.GetAccounts:
		accountList := as.getAccounts()
		context.Respond(&message.GetAccountsRsp{Accounts: &types.AccountList{Accounts: accountList}})
	case *message.CreateAccount:
		account, _ := as.createAccount(msg.Passphrase)
		context.Respond(&message.CreateAccountRsp{Account: account})
	case *message.LockAccount:
		actualAddress := msg.Account.Address
		var err error
		if len(actualAddress) == types.NameLength {
			actualAddress, err = as.resolveName(actualAddress)
			if err != nil {
				context.Respond(&message.AccountRsp{
					Account: &types.Account{Address: actualAddress},
					Err:     err,
				})
			}
		}
		account, err := as.lockAccount(actualAddress, msg.Passphrase)
		context.Respond(&message.AccountRsp{Account: account, Err: err})
	case *message.UnlockAccount:
		actualAddress := msg.Account.Address
		var err error
		if len(actualAddress) == types.NameLength {
			actualAddress, err = as.resolveName(actualAddress)
			if err != nil {
				context.Respond(&message.AccountRsp{
					Account: &types.Account{Address: actualAddress},
					Err:     err,
				})
			}
		}
		account, err := as.unlockAccount(actualAddress, msg.Passphrase)
		context.Respond(&message.AccountRsp{Account: account, Err: err})
	case *message.ImportAccount:
		var account *types.Account
		var err error
		if msg.Wif != nil {
			account, err = as.importAccount(msg.Wif, msg.OldPass, msg.NewPass)
		} else {
			account, err = as.importAccountFromKeystore(msg.Keystore, msg.OldPass, msg.NewPass)
		}
		context.Respond(&message.ImportAccountRsp{Account: account, Err: err})
	case *message.ExportAccount:
		var wif []byte
		var err error
		if msg.AsKeystore {
			wif, err = as.exportAccountKeystore(msg.Account.Address, msg.Pass)
		} else {
			wif, err = as.exportAccount(msg.Account.Address, msg.Pass)
		}
		context.Respond(&message.ExportAccountRsp{Wif: wif, Err: err})
	case *message.SignTx:
		var err error
		actualAddress := msg.Tx.GetBody().GetAccount()
		if len(actualAddress) == types.NameLength {
			actualAddress, err = as.resolveName(msg.Tx.GetBody().GetAccount())
			if err != nil {
				context.Respond(&message.SignTxRsp{Tx: nil, Err: err})
			}
			msg.Requester = actualAddress
		}
		err = as.signTx(context, msg)
		if err != nil {
			context.Respond(&message.SignTxRsp{Tx: nil, Err: err})
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

func (as *AccountService) getAccounts() []*types.Account {
	as.accountLock.RLock()
	defer as.accountLock.RUnlock()
	return as.accounts
}

func (as *AccountService) addAccount(account *types.Account) {
	as.accountLock.Lock()
	defer as.accountLock.Unlock()
	as.accounts = append(as.accounts, account)
}

func (as *AccountService) createAccount(passphrase string) (*types.Account, error) {
	address, err := as.ks.CreateKey(passphrase)
	if err != nil {
		return nil, err
	}
	account := types.NewAccount(address)
	as.addAccount(account)
	return account, nil
}

func (as *AccountService) importAccount(wif []byte, old string, new string) (*types.Account, error) {
	address, err := as.ks.ImportKey(wif, old, new)
	if err != nil {
		return nil, err
	}
	account := &types.Account{Address: address}
	as.addAccount(account)
	return account, nil
}

func (as *AccountService) importAccountFromKeystore(keystore []byte, old string, new string) (*types.Account, error) {
	privateKey, err := key.LoadKeystore(keystore, old)
	if err != nil {
		return nil, err
	}
	address, err := as.ks.AddKey(privateKey, new)
	if err != nil {
		return nil, err
	}
	account := &types.Account{Address: address}
	as.addAccount(account)
	return account, nil
}

func (as *AccountService) exportAccount(address []byte, pass string) ([]byte, error) {
	wif, err := as.ks.ExportKey(address, pass)
	if err != nil {
		return nil, err
	}
	return wif, nil
}

func (as *AccountService) exportAccountKeystore(address []byte, pass string) ([]byte, error) {
	privateKey, err := as.ks.GetKey(address, pass)
	if err != nil {
		return nil, err
	}
	wif, err := key.GetKeystore(privateKey, pass)
	if err != nil {
		return nil, err
	}
	return wif, nil
}

func (as *AccountService) unlockAccount(address []byte, passphrase string) (*types.Account, error) {
	addr, err := as.ks.Unlock(address, passphrase)
	if err != nil {
		as.Warn().Err(err).Msg("could not find the key")
		return nil, err
	}
	return &types.Account{Address: addr}, nil
}

func (as *AccountService) lockAccount(address []byte, passphrase string) (*types.Account, error) {
	addr, err := as.ks.Lock(address, passphrase)
	if err != nil {
		as.Warn().Err(err).Msg("could not load the key")
		return nil, err
	}
	return &types.Account{Address: addr}, nil
}

func (as *AccountService) signTx(c actor.Context, msg *message.SignTx) error {
	//sign tx
	prop := actor.FromInstance(NewSigner(as.ks))
	signer := c.Spawn(prop)
	signer.Request(msg, c.Sender())
	return nil
}

func (as *AccountService) verifyTx(tx *types.Tx) error {
	return as.ks.VerifyTx(tx)
}
