/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"sync"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	rtypes "github.com/aergoio/etcd/pkg/types"
)

type rPeerStatus struct {
	logger *log.Logger
	id     rtypes.ID
	pid    types.PeerID
	mu     sync.Mutex // protect variables below
	active bool
	since  time.Time
}

func newPeerStatus(id rtypes.ID, pid types.PeerID, logger *log.Logger) *rPeerStatus {
	return &rPeerStatus{
		id: id, pid: pid, logger: logger,
	}
}

func (s *rPeerStatus) activate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active {
		s.logger.Info().Stringer(p2putil.LogPeerID, types.LogPeerShort(s.pid)).Str("raftID", s.id.String()).Msgf("peer became active")
		s.active = true
		s.since = time.Now()
	} else {
		s.logger.Debug().Stringer(p2putil.LogPeerID, types.LogPeerShort(s.pid)).Str("raftID", s.id.String()).Msgf("activate called to already active peer")
	}
}

func (s *rPeerStatus) deactivate(cause string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active {
		s.logger.Info().Str("cause", cause).Stringer(p2putil.LogPeerID, types.LogPeerShort(s.pid)).Str("raftID", s.id.String()).Msgf("peer became inactive")

		s.active = false
		s.since = time.Time{}
	} else {
		s.logger.Debug().Str("cause", cause).Stringer(p2putil.LogPeerID, types.LogPeerShort(s.pid)).Str("raftID", s.id.String()).Msgf("deactivate called to already inactive peer")
	}
}

func (s *rPeerStatus) isActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

func (s *rPeerStatus) activeSince() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.since
}
