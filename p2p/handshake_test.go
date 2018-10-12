/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/mock"
)

func Test_runFuncTimeout(t *testing.T) {
	type args struct {
		m   targetFunc
		ttl time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{"Tnorm", args{func(done chan<- interface{}) {
			done <- "success"
		}, time.Millisecond * 10}, "success", false},
		{"Tnorm2", args{func(done chan<- interface{}) {
			done <- -3
		}, time.Millisecond * 10}, -3, false},
		{"Ttimeout1", args{func(done chan<- interface{}) {
			time.Sleep(time.Millisecond * 11)
		}, time.Millisecond * 10}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runFuncTimeout(tt.args.m, tt.args.ttl)
			if (err != nil) != tt.wantErr {
				t.Errorf("runFuncTimeout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("runFuncTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerHandshaker_handshakeOutboundPeerTimeout(t *testing.T) {
	logger = log.NewLogger("test")
	mockActor := new(MockActorService)
	mockPM := new(MockPeerManager)

	dummyMeta := PeerMeta{ID: dummyPeerID}
	mockPM.On("SelfMeta").Return(dummyMeta)
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	dummyBlkRsp := message.GetBestBlockRsp{Block: dummyBlock}
	mockActor.On("CallRequest", mock.Anything, mock.AnythingOfType("*message.GetBestBlock")).Return(dummyBlkRsp, nil)

	// dummyStatusMsg := &types.Status{}
	tests := []struct {
		name    string
		delay   time.Duration
		want    *types.Status
		wantErr bool
	}{
		// {"TNormal", time.Millisecond, dummyStatusMsg, false},
		{"TTimeout", time.Millisecond * 60, nil, true},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandshaker(mockPM, mockActor, logger, samplePeerID)
			mockRW := new(MockMsgReadWriter)
			mockRW.On("ReadMsg").After(tt.delay).Return(0, fmt.Errorf("must not reach"))
			mockRW.On("WriteMsg", mock.Anything).After(tt.delay).Return(fmt.Errorf("must not reach"))

			got, err := h.handshakeOutboundPeerTimeout(mockRW, time.Millisecond*50)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerHandshaker_handshakeOutboundPeer(t *testing.T) {
	logger = log.NewLogger("test")
	mockActor := new(MockActorService)
	mockPM := new(MockPeerManager)

	dummyMeta := PeerMeta{ID: dummyPeerID}
	mockPM.On("SelfMeta").Return(dummyMeta)
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	dummyBlkRsp := message.GetBestBlockRsp{Block: dummyBlock}
	mockActor.On("CallRequest", mock.Anything, mock.AnythingOfType("*message.GetBestBlock")).Return(dummyBlkRsp, nil)

	dummyStatusMsg := &types.Status{}
	statusBytes, _ := marshalMessage(dummyStatusMsg)
	tests := []struct {
		name       string
		readReturn *types.Status
		readError  error
		writeError error
		want       *types.Status
		wantErr    bool
	}{
		{"TSuccess", dummyStatusMsg, nil, nil, dummyStatusMsg, false},
		{"TUnexpMsg", nil, nil, nil, nil, true},
		{"TRFail", dummyStatusMsg, fmt.Errorf("failed"), nil, nil, true},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandshaker(mockPM, mockActor, logger, samplePeerID)
			mockRW := new(MockMsgReadWriter)
			containerMsg := &V030Message{payload:statusBytes}
			if tt.readReturn != nil {
				containerMsg.subProtocol = StatusRequest
			} else {
				containerMsg.subProtocol = AddressesRequest
			}

			mockRW.On("ReadMsg").Return(containerMsg, tt.readError)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeError)

			got, err := h.handshakeOutboundPeer(mockRW)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerHandshaker_handshakeInboundPeer(t *testing.T) {
	// t.SkipNow()
	logger = log.NewLogger("test")
	mockActor := new(MockActorService)
	mockPM := new(MockPeerManager)

	dummyMeta := PeerMeta{ID: dummyPeerID}
	mockPM.On("SelfMeta").Return(dummyMeta)
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	dummyBlkRsp := message.GetBestBlockRsp{Block: dummyBlock}
	mockActor.On("CallRequest", mock.Anything, mock.AnythingOfType("*message.GetBestBlock")).Return(dummyBlkRsp, nil)

	dummyStatusMsg := &types.Status{}
	statusBytes, _ := marshalMessage(dummyStatusMsg)
	tests := []struct {
		name       string
		readReturn *types.Status
		readError  error
		writeError error
		want       *types.Status
		wantErr    bool
	}{
		{"TSuccess", dummyStatusMsg, nil, nil, dummyStatusMsg, false},
		{"TUnexpMsg", nil, nil, nil, nil, true},
		{"TRFail", dummyStatusMsg, fmt.Errorf("failed"), nil, nil, true},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandshaker(mockPM, mockActor, logger, samplePeerID)
			mockRW := new(MockMsgReadWriter)
			containerMsg := &V030Message{payload:statusBytes}
			if tt.readReturn != nil {
				containerMsg.subProtocol = StatusRequest
			} else {
				containerMsg.subProtocol = AddressesRequest
			}

			mockRW.On("ReadMsg").Return(containerMsg, tt.readError)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeError)

			got, err := h.handshakeInboundPeer(mockRW)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}
