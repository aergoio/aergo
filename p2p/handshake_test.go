/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
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
		{"TNorm3", args{func(done chan<- interface{}) {
			time.Sleep(time.Millisecond * 7)
			done <- "delayed"
		}, time.Millisecond * 10}, "delayed", false},
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
	mockPM := new(MockP2PService)

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
			mockReader := new(MockReader)
			mockWriter := new(MockWriter)
			rw := bufio.NewReadWriter(bufio.NewReader(mockReader), bufio.NewWriter(mockWriter))
			mockReader.On("Read", mock.Anything).After(tt.delay).Return(0, fmt.Errorf("must not reach"))
			mockWriter.On("Write", mock.Anything).After(tt.delay).Return(0, fmt.Errorf("must not reach"))

			got, err := h.handshakeOutboundPeerTimeout(rw, time.Millisecond*50)
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
	mockPM := new(MockP2PService)

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
		// {"TNormal", dummyStatusMsg, nil, nil, dummyStatusMsg, false},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandshaker(mockPM, mockActor, logger, samplePeerID)
			mockReader := new(MockReader)
			mockWriter := new(MockWriter)
			mockRW := bufio.NewReadWriter(bufio.NewReader(mockReader), bufio.NewWriter(mockWriter))
			containerMsg := &types.P2PMessage{Header: &types.MessageData{}, Data: statusBytes}
			containerBytes, _ := marshalMessage(containerMsg)
			mockReader.On("Read", mock.AnythingOfType("[]uint8")).Run(func(args mock.Arguments) {
				buf := args.Get(0).([]byte)
				copy(buf, containerBytes)
			}).Return(len(containerBytes), tt.readError)
			mockWriter.On("Write", mock.AnythingOfType("[]uint8")).Return(0, tt.writeError)

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
