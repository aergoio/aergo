package syncer

import (
	"time"

	"github.com/pkg/errors"
)

type StubRequester struct {
	sendCh chan interface{}
	recvCh chan StubRequestResult
}

type StubRequestResult struct {
	result interface{}
	err    error
}

var (
	ErrHubFutureTimeOut = errors.New("stub compRequester request future timeout")
)

func (stubResult *StubRequestResult) Result() (interface{}, error) {
	return stubResult.result, stubResult.err
}

func NewStubRequester() *StubRequester {
	compRequester := &StubRequester{}

	compRequester.sendCh = make(chan interface{}, 1000)
	compRequester.recvCh = make(chan StubRequestResult, 1000)

	return compRequester
}

// handle requestFuture requset
// this api must not use parallel. TODO use lock
func (compRequester *StubRequester) RequestToFutureResult(targetName string, message interface{}, timeout time.Duration, tip string) (interface{}, error) {
	compRequester.sendCh <- message

	logger.Debug().Msg("stubcompRequester request future req")

	var res StubRequestResult
	select {
	case res = <-compRequester.recvCh:
		break
	case <-time.After(timeout):
		return nil, ErrHubFutureTimeOut
	}

	logger.Debug().Msg("stubcompRequester request future done")
	return res.result, res.err
}

func (compRequester *StubRequester) RequestTo(targetCompName string, message interface{}) {
	logger.Debug().Msg("stubcompRequester request")
	compRequester.sendCh <- message
}

func (compRequester *StubRequester) TellTo(targetName string, message interface{}) {
	logger.Debug().Msg("stubcompRequester tell")
	compRequester.sendCh <- message
}

//act like p2p or chain or syncer
func (compRequester *StubRequester) recvMessage() interface{} {
	select {
	case msg := <-compRequester.sendCh:
		logger.Debug().Msg("compRequester received message")
		return msg
	}
}

//act like p2p or chain or syncer
func (compRequester *StubRequester) sendReply(reply StubRequestResult) {
	compRequester.recvCh <- reply
}
