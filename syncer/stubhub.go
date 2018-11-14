package syncer

import (
	"github.com/pkg/errors"
	"time"
)

type StubHub struct {
	sendCh chan interface{}
	recvCh chan StubHubResult
}

type StubHubResult struct {
	result interface{}
	err    error
}

var (
	ErrHubFutureTimeOut = errors.New("stub hub request future timeout")
)

func (hubResult *StubHubResult) Result() (interface{}, error) {
	return hubResult.result, hubResult.err
}

func NewStubHub() *StubHub {
	hub := &StubHub{}

	hub.sendCh = make(chan interface{})
	hub.recvCh = make(chan StubHubResult)

	return hub
}

func (hub *StubHub) RequestFutureResult(targetName string, message interface{}, timeout time.Duration, tip string) (interface{}, error) {
	hub.sendCh <- message

	logger.Debug().Msg("stubhub request future req")

	var res StubHubResult
	select {
	case res = <-hub.recvCh:
		break
	case <-time.After(timeout):
		return nil, ErrHubFutureTimeOut
	}

	logger.Debug().Msg("stubhub request future done")
	return res.result, res.err
}

func (hub *StubHub) Tell(targetName string, message interface{}) {
	hub.sendCh <- message
}

//act like p2p or chain or syncer
func (hub *StubHub) recvMessage() interface{} {
	select {
	case msg := <-hub.sendCh:
		return msg
	}
}

//act like p2p or chain or syncer
func (hub *StubHub) sendReply(reply StubHubResult) {
	hub.recvCh <- reply
}
