package message

import (
	"errors"

	"github.com/aergoio/aergo/types"
)

var (
	ErrSignNotMatch           = errors.New("signature not matched")
	ErrShouldUnlockAccount    = errors.New("should unlock account first")
	ErrWrongAddressOrPassWord = errors.New("address or password is incorrect")
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
	Tx *types.Tx
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
