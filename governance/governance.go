package governance

import (
	"github.com/aergoio/aergo-lib/log"

	"github.com/aergoio/aergo/v2/contract/system"
)

type Governance struct {
	log *log.Logger

	// immutable params
	cfg *Config

	// snapshot params
	// init : update prev, next to default
	// commit : update prev to next
	// revert : update next to prev
	// reorg : update prev, next to db(reorg)
	ctx  *system.SystemContext
	init *SystemSnapshot
	prev *SystemSnapshot
	next *SystemSnapshot
}

func (s *Governance) Init(consensus string) {
	s.log = log.NewLogger("system")
	s.cfg = NewSystemImmutable(consensus)
}

func (s *Governance) Commit() {
	s.prev = s.next.Copy()
}

func (s *Governance) Revert() {
	s.next = s.prev.Copy()
}

func (s *Governance) Reorg() {
	var Reorg *SystemSnapshot
	// TODO : init reorg from db

	s.prev = Reorg.Copy()
	s.next = Reorg.Copy()
}
