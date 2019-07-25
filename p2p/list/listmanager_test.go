/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package list

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/contract/enterprise"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"testing"
)

func TestListManagerImpl_Start(t *testing.T) {
	conf := config.NewServerContext("", "").GetDefaultAuthConfig()
	logger := log.NewLogger("p2p.list.test")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	IDOnly := "16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt:"
	AddrOnly := ":122.1.3.4"
	IDAddr := "16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt:122.1.3.4"

	tests := []struct {
		name string

		confs     []string
		wantPanic bool
	}{
		{"TEmpty", nil, false},
		{"TSingle", []string{IDOnly}, false},
		{"TMulti", []string{IDOnly, AddrOnly, IDAddr}, false},
		{"TWrong", []string{IDOnly, ":e23dgvsdvz.32@", IDAddr}, true},
		{"TWrong2", []string{IDOnly, "e23dgvsd!v32@:", IDAddr}, true},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ecfg := &types.EnterpriseConfig{Key: enterprise.P2PWhite, On: true, Values: tt.confs}
			mockCA := p2pmock.NewMockChainAccessor(ctrl)
			mockCA.EXPECT().GetEnterpriseConfig(enterprise.P2PWhite).Return(ecfg, nil)
			got := NewListManager(conf, "", mockCA, nil, logger, false).(*listManagerImpl)
			func() {
				defer checkPanic(t, tt.wantPanic)
				got.Start()
			}()
			if tt.wantPanic {
				return
			}
			if got.entries == nil {
				t.Errorf("NewListManager() fields not initialized %v", "addrMap")
			}
			wantSize := len(tt.confs)
			if len(got.entries) != wantSize {
				t.Errorf("NewListManager() len(ListManager.entries) = %v, want %v", len(got.entries), wantSize)
			}
		})
	}
}

func checkPanic(t *testing.T, wantPanic bool) {
	if r := recover(); (r != nil) != wantPanic {
		t.Errorf("panic of NewListManager() %v, want %v", r != nil, wantPanic)
	}
}

func Test_blacklistManagerImpl_IsBanned(t *testing.T) {
	conf := config.NewServerContext("", "").GetDefaultAuthConfig()
	addr1 := "123.45.67.89"
	id1, _ := p2putil.RandomPeerID()
	addrother := "8.8.8.8"
	idother, _ := p2putil.RandomPeerID()
	thirdAddr := "222.8.8.8"
	thirdID, _ := p2putil.RandomPeerID()

	IDOnly := id1.Pretty() + ":"
	AddrOnly := ":" + addr1
	IDAddr := idother.Pretty() + ":" + addrother

	logger := log.NewLogger("p2p.list.test")
	listCfg := &types.EnterpriseConfig{Key: enterprise.P2PWhite, On: true, Values: []string{IDOnly, AddrOnly, IDAddr}}
	emptyCfg := &types.EnterpriseConfig{Key: enterprise.P2PWhite, On: true, Values: nil}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		addr string
		pid  types.PeerID
	}
	tests := []struct {
		name string
		cfg  *types.EnterpriseConfig
		args args
		want bool
	}{
		{"TFoundBoth", listCfg, args{addr1, id1}, false},
		{"TIDOnly", listCfg, args{addrother, id1}, false},
		{"TIDOnly2", listCfg, args{thirdAddr, id1}, false},
		{"TIDOnlyFail", listCfg, args{thirdAddr, idother}, true},
		{"TAddrOnly1", listCfg, args{addr1, idother}, false},
		{"TAddrOnly2", listCfg, args{addr1, thirdID}, false},
		{"TIDAddrSucc", listCfg, args{addrother, idother}, false},
		{"TIDAddrFail", listCfg, args{addrother, thirdID}, true},
		{"TIDAddrFail2", listCfg, args{thirdAddr, idother}, true},

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
			mockCA := p2pmock.NewMockChainAccessor(ctrl)
			mockCA.EXPECT().GetEnterpriseConfig(enterprise.P2PWhite).Return(tt.cfg, nil)
			mockPRM := p2pmock.NewMockPeerRoleManager(ctrl)
			mockPRM.EXPECT().GetRole(gomock.Any()).Return(p2pcommon.Watcher).AnyTimes()

			b := NewListManager(conf, "", mockCA, mockPRM, logger, false).(*listManagerImpl)
			b.Start()
			if got, _ := b.IsBanned(tt.args.addr, tt.args.pid); got != tt.want {
				t.Errorf("listManagerImpl.IsBanned() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_blacklistManagerImpl_IsBanned2(t *testing.T) {
	conf := config.NewServerContext("", "").GetDefaultAuthConfig()
	ent := []string{
		":192.168.1.14",
		"16Uiu2HAkvbHmK1Ke1hqAHmahwTGE4ndkdMdXJeXFE3kgBs17k2oQ:",
		"16Uiu2HAmNxKsrFQ4Wez4DYHW6o72y2Jpy6RMv5TuqAvjcQ5QPZWw:192.168.1.13",
		"16Uiu2HAmDFV41vku39rsMtXBaFT1MFUDyHxXiDJrUDt7gJycSKnX:192.168.1.12",
	}

	addr1 := "192.168.1.13"
	addr2 := "192.168.1.12"
	id1, _ := types.IDB58Decode("16Uiu2HAmQn3nFBGhJM7TnZRguLhgUx1HnpNL2easdt2JrxdbFjtb")
	id2, _ := types.IDB58Decode("16Uiu2HAmAnQ5jjk7huhepfFtDFFCreuJ21nHYBApVpg8G7EBdwme")
	id3, _ := types.IDB58Decode("16Uiu2HAkvbHmK1Ke1hqAHmahwTGE4ndkdMdXJeXFE3kgBs17k2oQ")
	id4, _ := types.IDB58Decode("16Uiu2HAkw9ZZ61iq8uWbrQrmNEXFbrbkWupdqiHSKkCuCFLTM6gF")
	id5, _ := types.IDB58Decode("16Uiu2HAmUkoPDPHrYYC8J4sVvaVRho8UxfWPLDgZS8gu5bsGSRSA")
	id6, _ := types.IDB58Decode("16Uiu2HAmNxKsrFQ4Wez4DYHW6o72y2Jpy6RMv5TuqAvjcQ5QPZWw")
	id7, _ := types.IDB58Decode("16Uiu2HAmDFV41vku39rsMtXBaFT1MFUDyHxXiDJrUDt7gJycSKnX")

	logger := log.NewLogger("p2p.list.test")
	listCfg := &types.EnterpriseConfig{Key: enterprise.P2PWhite, On: true, Values: ent}
	disabledCfg := &types.EnterpriseConfig{Key: enterprise.P2PWhite, On: false, Values: ent}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		addr string
		pid  types.PeerID
	}
	tests := []struct {
		name string
		cfg  *types.EnterpriseConfig
		role p2pcommon.PeerRole

		args args
		want bool
	}{
		{"T1", listCfg, p2pcommon.Watcher, args{addr1, id1}, true},
		{"T2", listCfg, p2pcommon.Watcher, args{addr1, id2}, true},
		{"T3", listCfg, p2pcommon.Watcher, args{addr1, id3}, false},
		{"T4", listCfg, p2pcommon.Watcher, args{addr2, id4}, true},
		{"T5", listCfg, p2pcommon.Watcher, args{addr2, id5}, true},
		{"T6", listCfg, p2pcommon.Watcher, args{addr2, id6}, true},
		{"T7", listCfg, p2pcommon.Watcher, args{addr2, id7}, false},

		// bp is always allowed
		{"T1", listCfg, p2pcommon.BlockProducer, args{addr1, id1}, false},
		{"T2", listCfg, p2pcommon.BlockProducer, args{addr1, id2}, false},
		{"T3", listCfg, p2pcommon.BlockProducer, args{addr1, id3}, false},
		{"T4", listCfg, p2pcommon.BlockProducer, args{addr2, id4}, false},
		{"T5", listCfg, p2pcommon.BlockProducer, args{addr2, id5}, false},
		{"T6", listCfg, p2pcommon.BlockProducer, args{addr2, id6}, false},
		{"T7", listCfg, p2pcommon.BlockProducer, args{addr2, id7}, false},

		// disabling conf will allow all connection
		{"TDis1", disabledCfg, p2pcommon.Watcher, args{addr1, id1}, false},
		{"TDis2", disabledCfg, p2pcommon.Watcher, args{addr1, id2}, false},
		{"TDis3", disabledCfg, p2pcommon.Watcher, args{addr1, id3}, false},
		{"TDis4", disabledCfg, p2pcommon.Watcher, args{addr2, id4}, false},
		{"TDis5", disabledCfg, p2pcommon.Watcher, args{addr2, id5}, false},
		{"TDis6", disabledCfg, p2pcommon.Watcher, args{addr2, id6}, false},
		{"TDis7", disabledCfg, p2pcommon.Watcher, args{addr2, id7}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCA := p2pmock.NewMockChainAccessor(ctrl)
			mockCA.EXPECT().GetEnterpriseConfig(enterprise.P2PWhite).Return(tt.cfg, nil)
			mockPRM := p2pmock.NewMockPeerRoleManager(ctrl)
			mockPRM.EXPECT().GetRole(gomock.Any()).Return(tt.role).AnyTimes()

			b := NewListManager(conf, "", mockCA, mockPRM, logger, false).(*listManagerImpl)
			b.Start()
			if got, _ := b.IsBanned(tt.args.addr, tt.args.pid); got != tt.want {
				t.Errorf("listManagerImpl.IsBanned() = %v, want %v", got, tt.want)
			}
		})
	}
}
