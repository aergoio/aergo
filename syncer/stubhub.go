package syncer

import (
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

func (hubResult *StubHubResult) Result() (interface{}, error) {
	return hubResult.result, hubResult.err
}

func NewStubHub() *StubHub {
	hub := &StubHub{}

	hub.sendCh = make(chan interface{})

	return hub
}

func (hub *StubHub) RequestFutureResult(targetName string, message interface{}, timeout time.Duration, tip string) (interface{}, error) {
	hub.sendCh <- message

	var res StubHubResult
	res = <-hub.recvCh

	return res.result, res.err
}

func (hub *StubHub) Tell(targetName string, message interface{}) {
	hub.sendCh <- message
}

//get msg from recvCh
func (hub *StubHub) GetMessage() interface{} {
	select {
	case msg := <-hub.sendCh:
		return msg
	}
}
