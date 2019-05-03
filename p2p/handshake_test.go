/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/p2p/p2pkey"
	peer "github.com/libp2p/go-libp2p-peer"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
)

const (
	sampleKeyFile = "../test/sample.key"
)

var (
	// sampleID matches the key defined in test config file
	sampleID peer.ID
)

func init() {
	sampleID = "16Uiu2HAmP2iRDpPumUbKhNnEngoxAUQWBmCyn7FaYUrkaDAMXJPJ"
	baseCfg := &config.BaseConfig{AuthDir: "test"}
	p2pCfg := &config.P2PConfig{NPKey: sampleKeyFile}
	p2pkey.InitNodeInfo(baseCfg, p2pCfg, "0.0.1-test", logger)
}

func TestPeerHandshaker_handshakeOutboundPeerTimeout(t *testing.T) {
	var myChainID = &types.ChainID{Magic:"itSmain1"}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger = log.NewLogger("test")
	// dummyStatusMsg := &types.Status{}
	tests := []struct {
		name    string
		delay   time.Duration
		want    *types.Status
		wantErr bool
	}{
		// {"TNormal", time.Millisecond, dummyStatusMsg, false},
		{"TTimeout", time.Millisecond * 200, nil, true},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockCA := p2pmock.NewMockChainAccessor(ctrl)
			mockPM.EXPECT().SelfMeta().Return(dummyMeta).Times(2)
			mockActor.EXPECT().GetChainAccessor().Return(mockCA)
			mockCA.EXPECT().GetBestBlock().Return(dummyBestBlock, nil)

			h := newHandshaker(mockPM, mockActor, logger, myChainID, samplePeerID)
			mockReader := p2pmock.NewMockReader(ctrl)
			mockWriter := p2pmock.NewMockWriter(ctrl)
			mockReader.EXPECT().Read(gomock.Any()).DoAndReturn(func(p interface{}) (int, error) {
				time.Sleep(tt.delay)
				return 0, fmt.Errorf("must not reach")
			}).AnyTimes()
			mockWriter.EXPECT().Write(gomock.Any()).DoAndReturn(func(p interface{}) (int, error) {
				time.Sleep(tt.delay)
				return -1, fmt.Errorf("must not reach")
			})
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
			defer cancel()
			_, got, err := h.handshakeOutboundPeer(ctx, mockReader, mockWriter)
			//_, got, err := h.handshakeOutboundPeerTimeout(mockReader, mockWriter, time.Millisecond*50)
			if !strings.Contains(err.Error(),"context deadline exceeded") {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() error = %v, wantErr %v", err, "context deadline exceeded")
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerHandshaker_Select(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger = log.NewLogger("test")
	mockActor := p2pmock.NewMockActorService(ctrl)
	mockPM := p2pmock.NewMockPeerManager(ctrl)

	tests := []struct {
		name     string
		hsheader p2pcommon.HSHeader
		wantErr  bool
	}{
		{"TVer030", p2pcommon.HSHeader{p2pcommon.MAGICMain, p2pcommon.P2PVersion030}, false},
		{"Tver020", p2pcommon.HSHeader{p2pcommon.MAGICMain, 0x00000200}, true},
		{"TInavlid", p2pcommon.HSHeader{p2pcommon.MAGICMain, 0x000001}, true},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockReader := p2pmock.NewMockReader(ctrl)
			mockWriter := p2pmock.NewMockWriter(ctrl)

			h := newHandshaker(mockPM, mockActor, logger, nil, samplePeerID)

			actual, err := h.selectProtocolVersion(test.hsheader, bufio.NewReader(mockReader),
				bufio.NewWriter(mockWriter))
			assert.Equal(t, test.wantErr, err != nil)
			if !test.wantErr {
				assert.NotNil(t, actual)
			}
		})
	}
}

func TestHSHeader_Marshal(t *testing.T) {
	tests := []struct {
		name            string
		input           []byte
		expectedNewwork uint32
		expectedVersion uint32
	}{
		{"TMain030", []byte{0x047, 0x041, 0x68, 0x41, 0, 0, 3, 0}, p2pcommon.MAGICMain, p2pcommon.P2PVersion030},
		{"TMain020", []byte{0x02e, 0x041, 0x54, 0x29, 0, 1, 3, 5}, p2pcommon.MAGICTest, 0x010305},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hs := p2pcommon.HSHeader{}
			hs.Unmarshal(test.input)
			assert.Equal(t, test.expectedNewwork, hs.Magic)
			assert.Equal(t, test.expectedVersion, hs.Version)

			actualBytes := hs.Marshal()
			assert.True(t, bytes.Equal(test.input, actualBytes))
		})
	}
}
