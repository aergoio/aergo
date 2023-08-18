/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"reflect"
	"testing"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
)

func Test_defaultVersionManager_FindBestP2PVersion(t *testing.T) {

	dummyChainID := &types.ChainID{}

	type args struct {
		versions []p2pcommon.P2PVersion
	}
	tests := []struct {
		name string
		args args

		want p2pcommon.P2PVersion
	}{
		{"TSingle", args{[]p2pcommon.P2PVersion{p2pcommon.P2PVersion033}}, p2pcommon.P2PVersion033},
		{"TMulti", args{[]p2pcommon.P2PVersion{p2pcommon.P2PVersion031, p2pcommon.P2PVersion033}}, p2pcommon.P2PVersion033},
		{"TOld", args{[]p2pcommon.P2PVersion{p2pcommon.P2PVersion030}}, p2pcommon.P2PVersionUnknown},
		{"TUnknown", args{[]p2pcommon.P2PVersion{9999999, 9999998}}, p2pcommon.P2PVersionUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			is := p2pmock.NewMockInternalService(ctrl)
			pm := p2pmock.NewMockPeerManager(ctrl)
			actor := p2pmock.NewMockActorService(ctrl)
			ca := p2pmock.NewMockChainAccessor(ctrl)
			vm := newDefaultVersionManager(is, actor, pm, ca, logger, dummyChainID)

			if got := vm.FindBestP2PVersion(tt.args.versions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultVersionManager.FindBestP2PVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultVersionManager_GetVersionedHandshaker(t *testing.T) {
	dummyChainID := &types.ChainID{}
	if chain.Genesis == nil {
		chain.Genesis = &types.Genesis{ID: *dummyChainID}
	}

	type args struct {
		version p2pcommon.P2PVersion
	}
	tests := []struct {
		name string
		args args

		wantErr bool
	}{
		//
		{"TRecent", args{p2pcommon.P2PVersion033}, false},
		{"TLegacy", args{p2pcommon.P2PVersion031}, false},
		{"TOld", args{p2pcommon.P2PVersion030}, true},
		{"TUnknown", args{9999999}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			is := p2pmock.NewMockInternalService(ctrl)
			pm := p2pmock.NewMockPeerManager(ctrl)
			actor := p2pmock.NewMockActorService(ctrl)
			ca := p2pmock.NewMockChainAccessor(ctrl)
			r := p2pmock.NewMockReadWriteCloser(ctrl)
			sampleID := types.RandomPeerID()

			ca.EXPECT().ChainID(gomock.Any()).Return(dummyChainID).MaxTimes(1)

			h := newDefaultVersionManager(is, actor, pm, ca, logger, dummyChainID)

			got, err := h.GetVersionedHandshaker(tt.args.version, sampleID, r)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultVersionManager.GetVersionedHandshaker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got != nil) == tt.wantErr {
				t.Errorf("defaultVersionManager.GetVersionedHandshaker() returns nil == %v , want %v", (got != nil), tt.wantErr)
				return
			}
		})
	}
}
