package types

import "errors"

var (
	//ErrTxNotFound is returned by MemPool Service if transaction does not exists
	ErrTxNotFound = errors.New("tx not found in mempool")

	//ErrTxHasInvalidHash is returned by MemPool Service if transaction does have invalid hash
	ErrTxHasInvalidHash = errors.New("tx has invalid hash")

	//ErrTxAlreadyInMempool is returned by MemPool Service if transaction which already exists
	ErrTxAlreadyInMempool = errors.New("tx already in mempool")

	//ErrTxFormatInvalid is returned by MemPool Service if transaction does not exists ErrTxFormatInvalid = errors.New("tx invalid format")
	ErrTxFormatInvalid = errors.New("tx invalid format")

	//ErrInsufficientBalance is returned by MemPool Service if account has not enough balance
	ErrInsufficientBalance = errors.New("not enough balance")

	//ErrTxNonceTooLow is returned by MemPool Service if transaction's nonce is already existed in block
	ErrTxNonceTooLow = errors.New("nonce is too low")

	//ErrTxNonceToohigh is for internal use only
	ErrTxNonceToohigh = errors.New("nonce is too high")

	ErrSignNotMatch = errors.New("signature not matched")

	ErrCouldNotRecoverPubKey = errors.New("could not recover pubkey from sign")

	ErrShouldUnlockAccount = errors.New("should unlock account first")

	ErrWrongAddressOrPassWord = errors.New("address or password is incorrect")

	//ErrInvalidRecipient
	ErrInvalidRecipient = errors.New("invalid recipient")

	//ErrStakeBeforeVote
	ErrMustStakeBeforeVote = errors.New("must stake before vote")

	//ErrLessTimeHasPassed
	ErrLessTimeHasPassed = errors.New("less time has passed")

	//ErrTooSmallAmount
	ErrTooSmallAmount = errors.New("too small amount to influence")

	//ErrMustStakeBeforeUnstake
	ErrMustStakeBeforeUnstake = errors.New("must stake before unstake")
)
