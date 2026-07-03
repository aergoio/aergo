package sbp

import (
	"testing"
	"time"

	cchain "github.com/aergoio/aergo/consensus/chain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testFactory(interval time.Duration) *SimpleBlockFactory {
	return &SimpleBlockFactory{
		blockInterval: interval,
		bpTimeoutC:    make(chan struct{}, 1),
	}
}

func TestCheckBpTimeoutUsesChannel(t *testing.T) {
	s := testFactory(time.Second)
	assert.NoError(t, s.checkBpTimeout())

	s.bpTimeoutC <- struct{}{}
	err := s.checkBpTimeout()
	require.Error(t, err)
	assert.Equal(t, cchain.ErrTimeout{Kind: "block"}, err)
}

func TestNotifyBpTimeoutSignalsChannel(t *testing.T) {
	s := testFactory(50 * time.Millisecond)
	s.notifyBpTimeout()

	select {
	case <-s.bpTimeoutC:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("bp timeout was not signaled")
	}
}

func TestDrainBpTimeoutClearsStaleSignal(t *testing.T) {
	s := testFactory(time.Second)
	s.bpTimeoutC <- struct{}{}
	s.drainBpTimeout()
	assert.NoError(t, s.checkBpTimeout())
}

func TestDecoNilWithoutRejectedTx(t *testing.T) {
	s := testFactory(time.Second)
	assert.Nil(t, s.deco())
}

func TestQueueJobSkipsWhileConnecting(t *testing.T) {
	s := testFactory(time.Second)
	s.connecting = true
	jq := make(chan interface{}, 1)
	s.QueueJob(time.Now(), jq)
	assert.Len(t, jq, 0)
}

func TestConnectTimeoutScalesWithBlockInterval(t *testing.T) {
	s := testFactory(time.Second)
	assert.Equal(t, 300*time.Second, s.connectTimeout())

	s.blockInterval = 2 * time.Second
	assert.Equal(t, 600*time.Second, s.connectTimeout())
}

func TestBpProductionTimeoutIsHalfInterval(t *testing.T) {
	s := testFactory(2 * time.Second)
	assert.Equal(t, time.Second, s.bpProductionTimeout())
}
