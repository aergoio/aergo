/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"reflect"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
)

func TestNewHCM(t *testing.T) {
	type args struct {
		mapService *PeerMapService
		nt         p2pcommon.NetworkTransport
	}
	tests := []struct {
		name string
		args args
		want *healthCheckManager
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHCM(tt.args.mapService, tt.args.nt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHCM() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_healthCheckManager_Start(t *testing.T) {
	type fields struct {
		logger    *log.Logger
		ms        *PeerMapService
		nt        p2pcommon.NetworkTransport
		finish    chan interface{}
		workerCnt int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hcm := &healthCheckManager{
				logger:    tt.fields.logger,
				ms:        tt.fields.ms,
				nt:        tt.fields.nt,
				finish:    tt.fields.finish,
				workerCnt: tt.fields.workerCnt,
			}
			hcm.Start()
		})
	}
}

func Test_healthCheckManager_Stop(t *testing.T) {
	type fields struct {
		logger    *log.Logger
		ms        *PeerMapService
		nt        p2pcommon.NetworkTransport
		finish    chan interface{}
		workerCnt int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hcm := &healthCheckManager{
				logger:    tt.fields.logger,
				ms:        tt.fields.ms,
				nt:        tt.fields.nt,
				finish:    tt.fields.finish,
				workerCnt: tt.fields.workerCnt,
			}
			hcm.Stop()
		})
	}
}

func Test_healthCheckManager_runHCM(t *testing.T) {
	type fields struct {
		logger    *log.Logger
		ms        *PeerMapService
		nt        p2pcommon.NetworkTransport
		finish    chan interface{}
		workerCnt int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hcm := &healthCheckManager{
				logger:    tt.fields.logger,
				ms:        tt.fields.ms,
				nt:        tt.fields.nt,
				finish:    tt.fields.finish,
				workerCnt: tt.fields.workerCnt,
			}
			hcm.runHCM()
		})
	}
}

func Test_healthCheckManager_checkPeers(t *testing.T) {
	type fields struct {
		logger    *log.Logger
		ms        *PeerMapService
		nt        p2pcommon.NetworkTransport
		finish    chan interface{}
		workerCnt int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hcm := &healthCheckManager{
				logger:    tt.fields.logger,
				ms:        tt.fields.ms,
				nt:        tt.fields.nt,
				finish:    tt.fields.finish,
				workerCnt: tt.fields.workerCnt,
			}
			hcm.checkPeers()
		})
	}
}
