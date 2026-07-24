package syncer

import (
	"errors"
	"testing"
	"time"

	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleSyncStart_BusyWhileRunning covers the orphan/tip path that keeps
// requesting SyncStart while catch-up is already running.
//
// Before the busy-path fix, handleSyncStart returned ErrSyncerBusy, which made
// handleMessage log "SyncStart failed" (and SyncChain abort when NotifyC was
// set). Busy must still notify waiters, but must not surface as a SyncStart
// handler error.
func TestHandleSyncStart_BusyWhileRunning(t *testing.T) {
	localChain := chain.InitStubBlockChain(nil, 10)
	s := NewSyncer(nil, localChain, SyncerCfg)
	s.isRunning = true
	seqBefore := s.GetSeq()

	t.Run("without NotifyC (orphan-like)", func(t *testing.T) {
		err := s.handleSyncStart(&message.SyncStart{
			PeerID:   targetPeerID,
			TargetNo: 1000,
		})
		assert.NoError(t, err, "busy SyncStart must not fail the handler")
		assert.True(t, s.isRunning)
		assert.Equal(t, seqBefore, s.GetSeq(), "must not start a second sync")
	})

	t.Run("with NotifyC (consensus SyncChain)", func(t *testing.T) {
		notiC := make(chan error, 1)
		err := s.handleSyncStart(&message.SyncStart{
			PeerID:   targetPeerID,
			TargetNo: 1000,
			NotifyC:  notiC,
		})
		assert.NoError(t, err, "busy SyncStart must not fail the handler")

		select {
		case nerr := <-notiC:
			assert.True(t, errors.Is(nerr, ErrSyncerBusy), "got %v", nerr)
		case <-time.After(time.Second):
			t.Fatal("expected ErrSyncerBusy on NotifyC")
		}

		assert.True(t, s.isRunning)
		assert.Equal(t, seqBefore, s.GetSeq(), "must not start a second sync")
	})
}

func TestHandleSyncStart_BusyDoesNotClearRunning(t *testing.T) {
	localChain := chain.InitStubBlockChain(nil, 10)
	s := NewSyncer(nil, localChain, SyncerCfg)
	require.False(t, s.isRunning)

	// First request should start sync (stub chain best is below target).
	err := s.handleSyncStart(&message.SyncStart{
		PeerID:   targetPeerID,
		TargetNo: uint64(localChain.Best + 10),
	})
	require.NoError(t, err)
	require.True(t, s.isRunning)
	seq := s.GetSeq()

	// Second request while running: same contract as orphan spam during catch-up.
	err = s.handleSyncStart(&message.SyncStart{
		PeerID:   targetPeerID,
		TargetNo: uint64(localChain.Best + 20),
	})
	assert.NoError(t, err)
	assert.True(t, s.isRunning)
	assert.Equal(t, seq, s.GetSeq())
}
