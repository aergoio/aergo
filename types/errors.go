package types

import "errors"

var (
	//ErrTxNotFound is returned by MemPool Service if transaction does not exists
	ErrTxNotFound = errors.New("tx not found in mempool")

	//ErrTxHasInvalidHash is returned by MemPool Service if transaction does have invalid hash
	ErrTxHasInvalidHash = errors.New("tx has invalid hash")

	//ErrTxAlreadyInMempool is returned by MemPool Service if exact same transaction is already exists
	ErrTxAlreadyInMempool = errors.New("tx is already in mempool")

	//ErrSameNonceInMempool is returned by MemPool Service if transaction which has same nonce is already exists
	ErrSameNonceAlreadyInMempool = errors.New("tx with same nonce is already in mempool")

	//ErrTxFormatInvalid is returned by MemPool Service if transaction does not exists ErrTxFormatInvalid = errors.New("tx invalid format")
	ErrTxFormatInvalid = errors.New("tx invalid format")

	//ErrInsufficientBalance is returned by MemPool Service if account has not enough balance
	ErrInsufficientBalance = errors.New("not enough balance")

	//ErrTxNonceTooLow is returned by MemPool Service if transaction's nonce is already existed in block
	ErrTxNonceTooLow = errors.New("nonce is too low")

	//ErrTxNonceToohigh is for internal use only
	ErrTxNonceToohigh = errors.New("nonce is too high")

	ErrTxInvalidType = errors.New("tx invalid type")

	ErrTxInvalidAccount = errors.New("tx invalid account")

	ErrTxNotAllowedAccount = errors.New("tx not allowed account")

	ErrTxInvalidChainIdHash = errors.New("tx invalid chain id hash")

	//ErrInvalidRecipient
	ErrTxInvalidRecipient = errors.New("tx invalid recipient")

	ErrTxInvalidAmount = errors.New("tx invalid amount")

	ErrTxInvalidPrice = errors.New("tx invalid price")

	ErrTxInvalidPayload = errors.New("tx invalid payload")

	ErrTxInvalidSize = errors.New("size of tx exceeds max length")

	ErrTxOnlySupportedInPriv = errors.New("tx only supported in private")

	ErrSignNotMatch = errors.New("signature not matched")

	ErrCouldNotRecoverPubKey = errors.New("could not recover pubkey from sign")

	ErrShouldUnlockAccount = errors.New("should unlock account first")

	ErrWrongAddressOrPassWord = errors.New("address or password is incorrect")

	//ErrStakeBeforeVote
	ErrMustStakeBeforeVote = errors.New("must stake before vote")

	//ErrLessTimeHasPassed
	ErrLessTimeHasPassed = errors.New("less time has passed")

	//ErrTooSmallAmount
	ErrTooSmallAmount = errors.New("too small amount to influence")

	ErrNameNotFound = errors.New("could not find name")

	//ErrMustStakeBeforeUnstake
	ErrMustStakeBeforeUnstake = errors.New("must stake before unstake")

	//ErrTooSmallAmount
	ErrExceedAmount = errors.New("request amount exceeds")

	ErrCreatorNotMatch = errors.New("creator not matched")

	ErrNotAllowedFeeDelegation = errors.New("fee delegation is not allowed")

	ErrNotEnoughGas = errors.New("not enough gas")
)

type InternalError struct {
	Reason string
}

func (e *InternalError) Error() string {
	return "internal error: " + e.Reason
}
