package governance

import "errors"

var (
	ErrNotSupportedConsensus = errors.New("not supported by this consensus")
)

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
