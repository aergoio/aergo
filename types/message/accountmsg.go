package message

import (
	"github.com/aergoio/aergo/v2/types"
)

const AccountsSvc = "AccountsSvc"

type CreateAccount struct {
	Passphrase string
}

type CreateAccountRsp struct {
	Account *types.Account
}

type LockAccount struct {
	Account    *types.Account
	Passphrase string
}

type UnlockAccount struct {
	Account    *types.Account
	Passphrase string
}

type AccountRsp struct {
	Account *types.Account
	Err     error
}

/*
TODO: is it good? or bad?
type LockAccountRsp struct {
	Account *types.Account
	Err     error
}

type UnlockAccountRsp struct {
	Account *types.Account
	Err     error
}
*/

type SignTx struct {
	Tx        *types.Tx
	Requester []byte
}
type SignTxRsp struct {
	Tx  *types.Tx
	Err error
}

type VerifyTx struct {
	Tx *types.Tx
}
type VerifyTxRsp struct {
	Tx  *types.Tx
	Err error
}

type GetAccounts struct{}
type GetAccountsRsp struct {
	Accounts *types.AccountList
}

type ImportAccount struct {
	Wif      []byte
	OldPass  string
	NewPass  string
	Keystore []byte
}
type ImportAccountRsp struct {
	Account *types.Account
	Err     error
}

type ExportAccount struct {
	Account    *types.Account
	Pass       string
	AsKeystore bool
}

type ExportAccountRsp struct {
	Wif []byte
	Err error
}
