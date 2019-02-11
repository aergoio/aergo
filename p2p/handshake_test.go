/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/stretchr/testify/assert"
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
	mockCA := new(MockChainAccessor)
	dummyBestBlock := &types.Block{Header:&types.BlockHeader{}}
	dummyMeta := p2pcommon.PeerMeta{ID: dummyPeerID}
	mockPM.On("SelfMeta").Return(dummyMeta)
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	dummyBlkRsp := message.GetBestBlockRsp{Block: dummyBlock}
	mockActor.On("CallRequest", mock.Anything, mock.AnythingOfType("*message.GetBestBlock")).Return(dummyBlkRsp, nil)
	mockActor.On("GetChainAccessor").Return(mockCA)
	mockCA.On("GetBestBlock").Return(dummyBestBlock, nil)
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
			h := newHandshaker(mockPM, mockActor, logger, myChainID, samplePeerID)
			mockReader := new(MockReader)
			mockWriter := new(MockWriter)
			mockReader.On("Read", mock.Anything).After(tt.delay).Return(0, fmt.Errorf("must not reach"))
			mockWriter.On("Write", mock.Anything).After(tt.delay).Return(-1, fmt.Errorf("must not reach"))

			_, got, err := h.handshakeOutboundPeerTimeout(mockReader, mockWriter, time.Millisecond*50)
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

func TestPeerHandshaker_Select(t *testing.T) {
	logger = log.NewLogger("test")
	mockActor := new(MockActorService)
	mockPM := new(MockPeerManager)

	tests := []struct {
		name string
		hsheader HSHeader
		wantErr bool
	}{
		{"TVer030", HSHeader{MAGICMain, P2PVersion030}, false},
		{"Tver020", HSHeader{MAGICMain, 0x00000200}, true},
		{"TInavlid", HSHeader{MAGICMain, 0x000001}, true},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockReader := new(MockReader)
			mockWriter := new(MockWriter)
			h := newHandshaker(mockPM, mockActor, logger, nil, samplePeerID)

			actual, err := h.selectProtocolVersion(test.hsheader, bufio.NewReader(mockReader),
				bufio.NewWriter(mockWriter))
			assert.Equal(t, test.wantErr, err != nil )
			if !test.wantErr {
				assert.NotNil(t, actual)
			}
		})
	}
}

func TestHSHeader_Marshal(t *testing.T) {
	tests := []struct {
		name string
		input []byte
		expectedNewwork uint32
		expectedVersion uint32
	}{
		{"TMain030", []byte{0x047, 0x041, 0x68,0x41, 0,0,3,0}, MAGICMain, P2PVersion030},
		{"TMain020", []byte{0x02e, 0x041, 0x54,0x29, 0,1,3,5}, MAGICTest, 0x010305},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hs := HSHeader{}
			hs.Unmarshal(test.input)
			assert.Equal(t, test.expectedNewwork, hs.Magic)
			assert.Equal(t, test.expectedVersion, hs.Version)

			actualBytes := hs.Marshal()
			assert.True(t, bytes.Equal(test.input, actualBytes))
		})
	}
}
