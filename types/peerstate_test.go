package types

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/stretchr/testify/assert"
)

func TestPeerState_Get(t *testing.T) {
	wg := sync.WaitGroup{}

	concurrentCnt := 100
	epoch := 100
	logger := log.NewLogger("test")

	logger.Info().Msg("================= Without atomic ops.")
	target := STARTING
	wg.Add(concurrentCnt)
	for i := 0; i < concurrentCnt; i++ {
		go func(in int) {
			for j := 0; j < epoch; j++ {
				//				logger.Infof("#%03d old state value is %d", in, int32(target))
				target++
				//				logger.Infof("#%03d new state is set to %d", in, PeerState(target))
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	logger.Info().Msgf("Expected %d , actual = %d", concurrentCnt*epoch, int32(target))
	logger.Info().Msg("================= With atomic ops.")
	var ocounter, ncounter [5]int32

	target = STARTING
	wg.Add(concurrentCnt)
	for i := 0; i < concurrentCnt; i++ {
		go func(in int) {
			for j := 0; j < epoch; j++ {
				newS := PeerState(j % 5)
				oldS := target.SetAndGet(newS)
				atomic.AddInt32(&ocounter[oldS], 1)
				atomic.AddInt32(&ncounter[newS], 1)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	oldS := target.SetAndGet(STARTING)
	atomic.AddInt32(&ocounter[oldS], 1)
	atomic.AddInt32(&ncounter[STARTING], 1)
	for i := 0; i < 5; i++ {
		assert.Equal(t, ocounter[i], ncounter[i])
	}
	logger.Info().Msgf("Values %v", ocounter)
	logger.Info().Msg("Done")
}
