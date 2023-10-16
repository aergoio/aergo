/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/types"
	rtypes "github.com/aergoio/etcd/pkg/types"
)

func Test_rPeerStatus_activate(t *testing.T) {
	logger := log.NewLogger("raft.support.test")
	id := rtypes.ID(111)
	pid, _ := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")

	tests := []struct {
		name    string
		lastAct bool

		wantActive bool
	}{
		{"TActive", true, true},
		{"TInactive", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newPeerStatus(id, pid, logger)

			barrier := sync.WaitGroup{}
			barrier.Add(2)
			allFin := make(chan bool)
			startTime := time.Now()
			go func() {
				barrier.Wait()
				if tt.lastAct {
					s.activate()
				} else {
					s.deactivate("final")
				}
				close(allFin)
			}()
			// do activate
			go func() {
				for i := 0; i < 100; i++ {
					s.activate()
					time.Sleep(time.Microsecond * time.Duration(rand.Intn(10)))
				}
				barrier.Done()
			}()
			// do deactivate
			go func() {
				for i := 0; i < 100; i++ {
					s.deactivate("p " + strconv.Itoa(i))
					time.Sleep(time.Microsecond * time.Duration(rand.Intn(10)))
				}
				barrier.Done()
			}()

			<-allFin
			if s.isActive() != tt.wantActive {
				t.Errorf("rPeerStatus.isActive() = %v , want %v", s.isActive(), tt.wantActive)
			}
			if s.isActive() != (s.activeSince().After(startTime)) {
				t.Errorf("rPeerStatus.ActiveSince() = %v , want not ", s.activeSince())
			}
		})
	}
}
