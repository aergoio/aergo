/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package contract

type ErrSystem interface {
	System() bool
}

func isSystemError(err error) bool {
	sErr, ok := err.(ErrSystem)
	return ok && sErr.System()
}

type vmStartError struct{}

func (e *vmStartError) Error() string {
	return "cannot start a VM"
}

func (e *vmStartError) System() bool {
	return e != nil
}

var ErrVmStart = &vmStartError{}

type DbSystemError struct {
	error
}

func newDbSystemError(e error) error {
	return &DbSystemError{e}
}

func (e *DbSystemError) System() bool {
	return e != nil
}

type VmSystemError struct {
	error
}

func newVmSystemError(e error) error {
	return &VmSystemError{e}
}

func (e *VmSystemError) System() bool {
	return e != nil
}

type VmTimeoutError struct{}

func (e *VmTimeoutError) Error() string {
	return "contract timeout of VmTimeoutError"
}

func (e *VmTimeoutError) System() bool {
	return e != nil
}

type ErrRuntime interface {
	Runtime() bool
}

func IsRuntimeError(err error) bool {
	rErr, ok := err.(ErrRuntime)
	return ok && rErr.Runtime()
}

type vmError struct {
	error
}

func newVmError(err error) error {
	return &vmError{err}
}

func (e *vmError) Runtime() bool {
	return e != nil
}

//Governance Errors

type ErrGovEnt interface {
	GovEnt() bool
}

type GovEntErr struct {
	error
}

func NewGovEntErr(e error) error {
	return &GovEntErr{e}
}

func (e *GovEntErr) GovEnt() bool {
	return e != nil
}

func (e *GovEntErr) Runtime() bool {
	return e != nil
}

func IsGovEntErr(err error) bool {
	rErr, ok := err.(ErrGovEnt)
	return ok && rErr.GovEnt()
}
