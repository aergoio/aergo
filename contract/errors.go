/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package contract

import (
	"errors"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance for transfer")
)

type VmError error

