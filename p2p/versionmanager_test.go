/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/golang/mock/gomock"
	"reflect"
	"testing"

	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
)

func Test_defaultVersionManager_FindBestP2PVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dummyChainID := &types.ChainID{}


	type args struct {
		versions []p2pcommon.P2PVersion
	}
	tests := []struct {
		name   string
		args   args

		want   p2pcommon.P2PVersion
	}{
		{"TSingle", args{[]p2pcommon.P2PVersion{p2pcommon.P2PVersion030}}, p2pcommon.P2PVersion030},
		{"TMulti", args{[]p2pcommon.P2PVersion{p2pcommon.P2PVersion031, p2pcommon.P2PVersion030}}, p2pcommon.P2PVersion031},
		{"TUnknown", args{[]p2pcommon.P2PVersion{9999999, 9999998}}, p2pcommon.P2PVersionUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := p2pmock.NewMockPeerManager(ctrl)
			actor := p2pmock.NewMockActorService(ctrl)
			ca := p2pmock.NewMockChainAccessor(ctrl)
			vm := newDefaultVersionManager(pm, actor, ca, logger, dummyChainID)

			if got := vm.FindBestP2PVersion(tt.args.versions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultVersionManager.FindBestP2PVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultVersionManager_GetVersionedHandshaker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dummyChainID := &types.ChainID{}


	type args struct {
		version p2pcommon.P2PVersion
	}
	tests := []struct {
		name   string
		args   args

		wantErr bool
	}{
		//
		{"TRecent", args{p2pcommon.P2PVersion031}, false},
		{"TLegacy", args{p2pcommon.P2PVersion030}, false},
		{"TUnknown", args{9999999}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := p2pmock.NewMockPeerManager(ctrl)
			actor := p2pmock.NewMockActorService(ctrl)
			ca := p2pmock.NewMockChainAccessor(ctrl)

			r := p2pmock.NewMockReadWriteCloser(ctrl)

			h := newDefaultVersionManager(pm, actor, ca, logger, dummyChainID)

			got, err := h.GetVersionedHandshaker(tt.args.version, sampleID, r)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultVersionManager.GetVersionedHandshaker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got != nil) == tt.wantErr {
				t.Errorf("defaultVersionManager.GetVersionedHandshaker() returns nil == %v , want %v", (got != nil), tt.wantErr )
				return
			}
		})
	}
}
