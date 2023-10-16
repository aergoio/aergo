/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
)

var sampleEntries []types.WhiteListEntry

func init() {
	eIDIP, _ := types.ParseListEntry(`{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"172.21.3.35" }`)
	eIDIR, _ := types.ParseListEntry(`{"peerid":"16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD", "cidr":"172.21.3.35/16" }`)
	eID, _ := types.ParseListEntry(`{"peerid":"16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR" }`)
	eIR, _ := types.ParseListEntry(`{"cidr":"211.5.3.123/16" }`)
	eIP6, _ := types.ParseListEntry(`{"address":"2001:0db8:0123:4567:89ab:cdef:1234:5678" }`)
	eIR6, _ := types.ParseListEntry(`{"cidr":"2001:0db8:0123:4567:89ab:cdef:1234:5678/96" }`)
	sampleEntries = []types.WhiteListEntry{eIDIP, eIDIR, eID, eIR, eIP6, eIR6}
}
func Test_polarisListManager_saveListFile(t *testing.T) {
	tmpAuthDir, err := ioutil.TempDir("", "aergoTestPolaris")
	if err != nil {
		t.Fatalf("Failed to create temp directory to test: %v ", err.Error())
	} else {
		t.Logf("Create tmp directory on %v", tmpAuthDir)
		defer os.RemoveAll(tmpAuthDir)
	}

	logger := log.NewLogger("polaris.test")
	conf := config.PolarisConfig{EnableBlacklist: true}
	lm := NewPolarisListManager(&conf, tmpAuthDir, logger)
	lm.entries = sampleEntries
	lm.saveListFile()

	lm2 := NewPolarisListManager(&conf, tmpAuthDir, logger)
	lm2.loadListFile()
	if len(lm2.entries) != len(lm.entries) {
		t.Errorf("polarisListManager.loadListFile() entry count %v, want %v", len(lm2.entries), len(lm.entries))
	}

	for i, e := range lm.entries {
		e2 := lm.entries[i]

		if !reflect.DeepEqual(e, e2) {
			t.Errorf("polarisListManager.loadListFile() entry %v, %v", e, e2)
		}
	}
}

func Test_polarisListManager_RemoveEntry(t *testing.T) {
	logger := log.NewLogger("polaris.test")
	conf := &config.PolarisConfig{EnableBlacklist: true}

	tests := []struct {
		name string
		idx  int
		want bool
	}{
		{"TFirst", 0, true},
		{"TMid", 1, true},
		{"TLast", 5, true},
		{"TOverflow", 6, false},
		{"TNegative", -1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := NewPolarisListManager(conf, "temp", logger)
			lm.entries = sampleEntries
			if got := lm.RemoveEntry(tt.idx); got != tt.want {
				t.Errorf("RemoveEntry() = %v, want %v", got, tt.want)
			} else if got != (len(lm.entries) < 6) {
				t.Errorf("RemoveEntry() remain size = %v, want not", len(lm.entries))
			}

		})
	}
}

func Test_polarisListManager_AddEntry(t *testing.T) {
	logger := log.NewLogger("polaris.test")
	conf := &config.PolarisConfig{EnableBlacklist: true}

	tests := []struct {
		name     string
		args     []types.WhiteListEntry
		wantSize int
	}{
		{"TFirst", sampleEntries[:0], 0},
		{"T1", sampleEntries[:1], 1},
		{"TAll", sampleEntries, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := NewPolarisListManager(conf, "temp", logger)
			for _, e := range tt.args {
				lm.AddEntry(e)
			}

			if len(lm.ListEntries()) != tt.wantSize {
				t.Errorf("AddEntry() entries size = %v, want %v", len(lm.ListEntries()), tt.wantSize)
			}
		})
	}
}

func Test_polarisListManager_IsBanned(t *testing.T) {
	conf := config.NewServerContext("", "").GetDefaultPolarisConfig()
	conf.EnableBlacklist = true
	logger := log.NewLogger("polaris.test")

	addr1 := "123.45.67.89"
	id1 := types.RandomPeerID()
	addrother := "8.8.8.8"
	idother := types.RandomPeerID()
	thirdAddr := "222.8.8.8"
	thirdID := types.RandomPeerID()

	IDOnly, e1 := types.ParseListEntry(`{"peerid":"` + id1.Pretty() + `"}`)
	AddrOnly, e2 := types.ParseListEntry(`{"address":"` + addr1 + `"}`)
	IDAddr, e3 := types.ParseListEntry(`{"peerid":"` + idother.Pretty() + `", "address":"` + addrother + `"}`)
	if e1 != nil || e2 != nil || e3 != nil {
		t.Fatalf("Inital entry value failure %v , %v , %v", e1, e2, e3)
	}
	listCfg := []types.WhiteListEntry{IDOnly, AddrOnly, IDAddr}
	emptyCfg := []types.WhiteListEntry{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		addr string
		pid  types.PeerID
	}
	tests := []struct {
		name   string
		preset []types.WhiteListEntry
		args   args
		want   bool
	}{
		{"TFoundBoth", listCfg, args{addr1, id1}, true},
		{"TIDOnly", listCfg, args{addrother, id1}, true},
		{"TIDOnly2", listCfg, args{thirdAddr, id1}, true},
		{"TIDOnlyFail", listCfg, args{thirdAddr, idother}, false},
		{"TAddrOnly1", listCfg, args{addr1, idother}, true},
		{"TAddrOnly2", listCfg, args{addr1, thirdID}, true},
		{"TIDAddrSucc", listCfg, args{addrother, idother}, true},
		{"TIDAddrFail", listCfg, args{addrother, thirdID}, false},
		{"TIDAddrFail2", listCfg, args{thirdAddr, idother}, false},

		// if config have nothing. everything is allowed
		{"TEmpFoundBoth", emptyCfg, args{addr1, id1}, false},
		{"TEmpIDOnly", emptyCfg, args{addrother, id1}, false},
		{"TEmpIDOnly2", emptyCfg, args{thirdAddr, id1}, false},
		{"TEmpIDOnly2", emptyCfg, args{thirdAddr, id1}, false},
		{"TEmpAddrOnly1", emptyCfg, args{addr1, idother}, false},
		{"TEmpAddrOnly2", emptyCfg, args{addr1, thirdID}, false},
		{"TEmpIDAddrSucc", emptyCfg, args{addrother, idother}, false},
		{"TEmpIDAddrFail", emptyCfg, args{addrother, id1}, false},
		{"TEmpIDAddrFail2", emptyCfg, args{thirdAddr, idother}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewPolarisListManager(conf, "/tmp", logger)
			b.entries = tt.preset

			if got, _ := b.IsBanned(tt.args.addr, tt.args.pid); got != tt.want {
				t.Errorf("listManagerImpl.IsBanned() = %v, want %v", got, tt.want)
			}
		})
	}
}
