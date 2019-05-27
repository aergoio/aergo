/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
)

func Test_defaultVersionManager_FindBestP2PVersion(t *testing.T) {
	type fields struct {
		pm           p2pcommon.PeerManager
		actor        p2pcommon.ActorService
		logger       *log.Logger
		localChainID *types.ChainID
	}
	type args struct {
		versions []p2pcommon.P2PVersion
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   p2pcommon.P2PVersion
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &defaultVersionManager{
				pm:           tt.fields.pm,
				actor:        tt.fields.actor,
				logger:       tt.fields.logger,
				localChainID: tt.fields.localChainID,
			}
			if got := vm.FindBestP2PVersion(tt.args.versions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultVersionManager.FindBestP2PVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultVersionManager_GetVersionedHandshaker(t *testing.T) {
	type fields struct {
		pm           p2pcommon.PeerManager
		actor        p2pcommon.ActorService
		logger       *log.Logger
		localChainID *types.ChainID
	}
	type args struct {
		version p2pcommon.P2PVersion
		peerID  types.PeerID
		r       io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    p2pcommon.VersionedHandshaker
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &defaultVersionManager{
				pm:           tt.fields.pm,
				actor:        tt.fields.actor,
				logger:       tt.fields.logger,
				localChainID: tt.fields.localChainID,
			}
			w := &bytes.Buffer{}
			got, err := h.GetVersionedHandshaker(tt.args.version, tt.args.peerID, tt.args.r, w)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultVersionManager.GetVersionedHandshaker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultVersionManager.GetVersionedHandshaker() = %v, want %v", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("defaultVersionManager.GetVersionedHandshaker() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
