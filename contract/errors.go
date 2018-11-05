/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package contract

import "errors"

type VmError error

type DbSystemError error

func newDbSystemError(text string) error {
	return DbSystemError(errors.New(text))
}