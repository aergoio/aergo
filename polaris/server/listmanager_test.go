/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/contract/enterprise"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const testAuthDir = "/tmp"

func Test_polarisListManager_saveListFile(t *testing.T) {
	eIDIP,_ := enterprise.NewWhiteListEntry(`{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"172.21.3.35" }`)
	eIDIR,_ := enterprise.NewWhiteListEntry(`{"peerid":"16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD", "cidr":"172.21.3.35/16" }`)
	eID,_ := enterprise.NewWhiteListEntry(`{"peerid":"16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR" }`)
	eIR,_ := enterprise.NewWhiteListEntry(`{"cidr":"211.5.3.123/16" }`)
	eIP6,_ := enterprise.NewWhiteListEntry(`{"address":"2001:0db8:0123:4567:89ab:cdef:1234:5678" }`)
	eIR6,_ := enterprise.NewWhiteListEntry(`{"cidr":"2001:0db8:0123:4567:89ab:cdef:1234:5678/96" }`)
	entries := []enterprise.WhiteListEntry{eIDIP,eIDIR, eID, eIR, eIP6, eIR6}

	logger := log.NewLogger("polaris.test")
	conf := config.AuthConfig{EnableLocalConf: true}
	lm := NewPolarisListManager(&conf, testAuthDir, logger)
	lm.entries = entries
	lm.saveListFile()
	defer func() {
		os.Remove(filepath.Join(testAuthDir, localListFile))
	}()

	lm2 := NewPolarisListManager(&conf, testAuthDir, logger)
	lm2.loadListFile()
	if len(lm2.entries) != len(lm.entries) {
		t.Errorf("polarisListManager.loadListFile() entry count %v, want %v", len(lm2.entries),len(lm.entries))
	}

	for i, e := range lm.entries {
		e2 := lm.entries[i]

		if !reflect.DeepEqual(e,e2) {
			t.Errorf("polarisListManager.loadListFile() entry %v, %v", e,e2)
		}
	}
}
