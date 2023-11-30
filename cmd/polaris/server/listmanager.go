/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/types"
	"github.com/rs/zerolog"
)

// variables that are used internally
var (
	NotFoundError = errors.New("ban status not found")
	UndefinedTime = time.Unix(0, 0)
	FarawayFuture = time.Date(99999, 1, 1, 0, 0, 0, 0, time.UTC)
)

const (
	localListFile = "blacklist.json"
)

type polarisListManager struct {
	logger *log.Logger
	mutex  sync.Mutex

	entries []types.WhiteListEntry
	enabled bool
	rwLock  sync.RWMutex
	authDir string

	stopScheduler chan interface{}
}

func NewPolarisListManager(conf *config.PolarisConfig, authDir string, logger *log.Logger) *polarisListManager {
	lm := &polarisListManager{
		logger:  logger,
		enabled: conf.EnableBlacklist,

		authDir:       authDir,
		stopScheduler: make(chan interface{}),
	}

	return lm
}

func (lm *polarisListManager) loadListFile() {
	blFile := filepath.Join(lm.authDir, localListFile)

	jsonFile, err := os.Open(blFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		lm.logger.Info().Err(err).Str("file", blFile).Msg("Failed to read blacklist file")
		return
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	entries, err := types.ReadEntries(byteValue)
	if err != nil {
		lm.logger.Info().Err(err).Str("file", blFile).Msg("Failed to parse blacklist file")
		return
	}
	lm.logger.Info().Array("entry", ListEntriesMarshaller{entries, 10}).Msg("Loaded blacklist file")
	lm.entries = entries
}

func (lm *polarisListManager) saveListFile() {
	blFile := filepath.Join(lm.authDir, localListFile)
	lm.logger.Debug().Str("file", blFile).Msg("Saving local blacklist file")
	jsonFile, err := os.OpenFile(blFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		lm.logger.Info().Err(err).Str("file", blFile).Msg("Failed to open blacklist file for writing")
		return
	}
	defer jsonFile.Close()
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	err = types.WriteEntries(lm.entries, jsonFile)
	if err != nil {
		lm.logger.Info().Err(err).Str("file", blFile).Msg("Failed to write blacklist file")
		return
	}
	lm.logger.Info().Array("entry", ListEntriesMarshaller{lm.entries, 10}).Msg("Saved blacklist file")
}

func (lm *polarisListManager) ListEntries() []types.WhiteListEntry {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	newEntry := make([]types.WhiteListEntry, len(lm.entries))
	copy(newEntry, lm.entries)
	return newEntry
}
func (lm *polarisListManager) AddEntry(entry types.WhiteListEntry) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	newEntry := make([]types.WhiteListEntry, 0, len(lm.entries)+1)
	newEntry = append(newEntry, lm.entries...)
	newEntry = append(newEntry, entry)
	lm.entries = newEntry
	lm.logger.Info().Str("entry", entry.String()).Msg("Added blacklist entry")
}

func (lm *polarisListManager) RemoveEntry(idx int) bool {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	if idx < 0 || idx >= len(lm.entries) {
		return false
	}
	toRemove := lm.entries[idx]
	newEntries := make([]types.WhiteListEntry, 0, len(lm.entries))
	newEntries = append(newEntries, lm.entries...)
	newEntries = append(newEntries[:idx], newEntries[idx+1:]...)
	lm.entries = newEntries
	lm.logger.Info().Str("entry", toRemove.String()).Msg("Removed blacklist entry")
	return true
}

func (lm *polarisListManager) Start() {
	lm.logger.Debug().Msg("starting up list manager")
	if lm.enabled {
		lm.loadListFile()
	}
}

func (lm *polarisListManager) Stop() {
	lm.logger.Debug().Msg("stopping list manager")
	if lm.enabled {
		lm.saveListFile()
	}
}

func (lm *polarisListManager) IsBanned(addr string, pid types.PeerID) (bool, time.Time) {
	// polaris is blaklist
	// empty entry means no blacklist
	if !lm.enabled || len(lm.entries) == 0 {
		return false, FarawayFuture
	}

	// malformed ip address is banned
	ip := net.ParseIP(addr)
	if ip == nil {
		return true, FarawayFuture
	}

	// finally check peer is in list
	for _, ent := range lm.entries {
		if ent.Contains(ip, pid) {
			return true, FarawayFuture
		}
	}
	return false, FarawayFuture
}

func (lm *polarisListManager) RefineList() {
	// no refine is needed
}

func (lm *polarisListManager) Summary() map[string]interface{} {
	// There can be a little error
	sum := make(map[string]interface{})
	entries := make([]string, 0, len(lm.entries))
	for _, e := range lm.entries {
		entries = append(entries, e.String())
	}
	sum["blacklist"] = entries
	sum["blacklist_on"] = lm.enabled

	return sum
}

type ListEntriesMarshaller struct {
	arr   []types.WhiteListEntry
	limit int
}

func (m ListEntriesMarshaller) MarshalZerologArray(a *zerolog.Array) {
	size := len(m.arr)
	if size > m.limit {
		for i := 0; i < m.limit-1; i++ {
			a.Str(m.arr[i].String())
		}
		a.Str(fmt.Sprintf("(and %d more)", size-m.limit+1))
	} else {
		for _, element := range m.arr {
			a.Str(element.String())
		}
	}
}
