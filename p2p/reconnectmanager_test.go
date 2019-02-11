package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_reconnectManager_AddJob(t *testing.T) {
	logger := log.NewLogger("test.p2p")
	// TODO: is it ok that this global var can be changed.
	durations = []time.Duration{
		time.Millisecond * 100,
		time.Millisecond * 120,
		time.Millisecond * 130,
		time.Millisecond * 150,
	}
	trials := len(durations)
	maxTrial = trials

	dummyMeta := p2pcommon.PeerMeta{ID: dummyPeerID}
	dummyMeta2 := p2pcommon.PeerMeta{ID: dummyPeerID2}
	dummyMeta3 := p2pcommon.PeerMeta{ID: dummyPeerID3}

	mockPm := &MockPeerManager{}
	dummyPeer := &remotePeerImpl{}
	mockPm.On("GetPeer", mock.MatchedBy(func(ID peer.ID) bool { return ID == dummyPeerID })).Return(nil, false)
	mockPm.On("GetPeer", mock.MatchedBy(func(ID peer.ID) bool { return ID != dummyPeerID2 })).Return(dummyPeer, true)
	mockPm.On("GetPeer", mock.MatchedBy(func(ID peer.ID) bool { return ID != dummyPeerID3 })).Return(nil, false)
	mockPm.On("AddNewPeer", mock.AnythingOfType("p2pcommon.PeerMeta"))
	mockPm2 := &MockPeerManager{}
	mockPm2.On("GetPeer", mock.MatchedBy(func(ID peer.ID) bool { return ID != dummyPeerID })).Return(dummyPeer, true)
	mockPm2.On("AddNewPeer", mock.AnythingOfType("p2pcommon.PeerMeta"))
	mockPm3 := &MockPeerManager{}
	mockPm3.On("GetPeer", mock.MatchedBy(func(ID peer.ID) bool { return ID != dummyPeerID })).Return(nil, false).Times(2)
	mockPm3.On("GetPeer", mock.MatchedBy(func(ID peer.ID) bool { return ID != dummyPeerID })).Return(dummyPeer, true).Once()
	mockPm3.On("AddNewPeer", mock.AnythingOfType("p2pcommon.PeerMeta"))

	tests := []struct {
		name string
		pm   PeerManager
	}{
		{"t1", mockPm},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := newReconnectManager(logger)
			rm.pm = mockPm
			rm.AddJob(dummyMeta)
			rm.AddJob(dummyMeta)
			rm.AddJob(dummyMeta)
			assert.Equal(t, 1, len(rm.jobs))
			rm.AddJob(dummyMeta2)
			rm.AddJob(dummyMeta3)
			assert.Equal(t, 3, len(rm.jobs))

			rm.CancelJob(dummyPeerID)
			rm.CancelJob(dummyPeerID)
			rm.CancelJob(dummyPeerID)
			rm.CancelJob(dummyPeerID)
			assert.Equal(t, 2, len(rm.jobs))
			rm.Stop()
			assert.Equal(t, 0, len(rm.jobs))

		})
	}

	// test stop
	t.Run("tstop", func(t *testing.T) {
		rm := newReconnectManager(logger)
		rm.pm = mockPm
		rm.AddJob(dummyMeta)
		assert.Equal(t, 1, len(rm.jobs))
		rm.AddJob(dummyMeta2)
		assert.Equal(t, 2, len(rm.jobs))
		rm.Stop()
		rm.AddJob(dummyMeta3)
		assert.Equal(t, 0, len(rm.jobs))
	})

}
