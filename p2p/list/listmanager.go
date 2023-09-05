/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package list

import (
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/contract/enterprise"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
)

// variables that are used internally
var (
	NotFoundError = errors.New("ban status not found")
	UndefinedTime = time.Unix(0, 0)
	FarawayFuture = time.Date(99999, 1, 1, 0, 0, 0, 0, time.UTC)
)

const (
	localListFile = "list.json"
)

type listManagerImpl struct {
	logger    *log.Logger
	chainAcc  types.ChainAccessor
	prm       p2pcommon.PeerRoleManager
	publicNet bool

	entries []types.WhiteListEntry
	enabled bool
	rwLock  sync.RWMutex
	authDir string

	stopScheduler chan interface{}
}

func NewListManager(conf *config.AuthConfig, authDir string, chainAcc types.ChainAccessor, prm p2pcommon.PeerRoleManager, logger *log.Logger, publicNet bool) p2pcommon.ListManager {
	bm := &listManagerImpl{
		logger:    logger,
		chainAcc:  chainAcc,
		prm:       prm,
		publicNet: publicNet,

		authDir:       authDir,
		stopScheduler: make(chan interface{}),
	}

	return bm
}

func (lm *listManagerImpl) Start() {
	lm.logger.Debug().Msg("starting up list manager")

	lm.RefineList()
}

func (lm *listManagerImpl) Stop() {
	lm.logger.Debug().Msg("stopping list manager")
}

func (lm *listManagerImpl) IsBanned(addr string, pid types.PeerID) (bool, time.Time) {
	// empty entry is
	if len(lm.entries) == 0 {
		return false, FarawayFuture
	}

	// malformed ip address is banned
	ip := net.ParseIP(addr)
	if ip == nil {
		return true, FarawayFuture
	}

	// bps are automatically allowed
	if lm.prm.GetRole(pid) == types.PeerRole_Producer {
		return false, FarawayFuture
	}

	// finally check peer is in list
	for _, ent := range lm.entries {
		if ent.Contains(ip, pid) {
			return false, FarawayFuture
		}
	}
	return true, FarawayFuture
}

func (lm *listManagerImpl) RefineList() {
	if lm.publicNet {
		lm.logger.Info().Msg("network is public, apply default policy instead (allow all)")
		lm.entries = make([]types.WhiteListEntry, 0)
		lm.enabled = false
		return
	}

	wl, err := lm.chainAcc.GetEnterpriseConfig(enterprise.P2PWhite)
	if err != nil {
		lm.logger.Info().Msg("error while getting whitelist config. apply default policy instead (allow all)")
		//ent, _ := ParseListEntry(":")
		//lm.entries = append(lm.entries, ent)
		lm.entries = make([]types.WhiteListEntry, 0)
		lm.enabled = false
		return
	}
	lm.enabled = wl.GetOn()
	if !wl.GetOn() {
		lm.logger.Debug().Msg("whitelist conf is disabled. apply default policy instead (allow all)")
		lm.entries = make([]types.WhiteListEntry, 0)
	} else if len(wl.Values) == 0 {
		lm.logger.Debug().Msg("no whitelist found. apply default policy instead (allow all)")
		//ent, _ := ParseListEntry(":")
		//lm.entries = append(lm.entries, ent)
		lm.entries = make([]types.WhiteListEntry, 0)
	} else {
		entries := make([]types.WhiteListEntry, 0, len(wl.Values))
		for _, v := range wl.Values {
			ent, err := types.ParseListEntry(v)
			if err != nil {
				panic("invalid whitelist entry " + v)
			}
			entries = append(entries, ent)
		}
		lm.entries = entries
		lm.logger.Debug().Str("entries", strings.Join(wl.Values, " , ")).Msg("loaded whitelist entries")
	}

}

func (lm *listManagerImpl) Summary() map[string]interface{} {
	// There can be a little error
	sum := make(map[string]interface{})
	entries := make([]string, 0, len(lm.entries))
	for _, e := range lm.entries {
		entries = append(entries, e.String())
	}
	sum["whitelist"] = entries
	sum["whitelist_on"] = lm.enabled

	return sum
}
