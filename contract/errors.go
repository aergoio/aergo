/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package contract

type VmError struct {
	error
}

func newVmError(e error) error {
	return &VmError{e}
}

type DbSystemError struct {
	error
}

func newDbSystemError(e error) error {
	return &DbSystemError{e}
}