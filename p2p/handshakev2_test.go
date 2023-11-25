/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

func Test_baseWireHandshaker_writeWireHSRequest(t *testing.T) {
	tests := []struct {
		name     string
		args     p2pcommon.HSHeadReq
		wantErr  bool
		wantSize int
		wantErr2 bool
	}{
		{"TEmpty", p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: nil}, false, 8, true},
		{"TSingle", p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: []p2pcommon.P2PVersion{p2pcommon.P2PVersion033}}, false, 12, false},
		{"TMulti", p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: []p2pcommon.P2PVersion{0x033333, 0x092fa10, p2pcommon.P2PVersion033, p2pcommon.P2PVersion032}}, false, 24, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &baseWireHandshaker{}
			buffer := bytes.NewBuffer(nil)
			//wr := bufio.NewWriter(buffer)
			err := h.writeWireHSRequest(tt.args, buffer)
			if (err != nil) != tt.wantErr {
				t.Errorf("baseWireHandshaker.writeWireHSRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if buffer.Len() != tt.wantSize {
				t.Errorf("baseWireHandshaker.writeWireHSRequest() error = %v, wantErr %v", buffer.Len(), tt.wantSize)
			}

			got, err2 := h.readWireHSRequest(buffer)
			if (err2 != nil) != tt.wantErr2 {
				t.Errorf("baseWireHandshaker.readWireHSRequest() error = %v, wantErr %v", err2, tt.wantErr2)
			}
			if !reflect.DeepEqual(tt.args, got) {
				t.Errorf("baseWireHandshaker.readWireHSRequest() = %v, want %v", got, tt.args)
			}
			if buffer.Len() != 0 {
				t.Errorf("baseWireHandshaker.readWireHSRequest() error = %v, wantErr %v", buffer.Len(), 0)
			}

		})
	}
}

func Test_baseWireHandshaker_writeWireHSResponse(t *testing.T) {
	tests := []struct {
		name     string
		args     p2pcommon.HSHeadResp
		wantErr  bool
		wantSize int
		wantErr2 bool
	}{
		{"TSingle", p2pcommon.HSHeadResp{Magic: p2pcommon.MAGICMain, RespCode: p2pcommon.P2PVersion030.Uint32()}, false, 8, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &baseWireHandshaker{}
			buffer := bytes.NewBuffer(nil)
			err := h.writeWireHSResponse(tt.args, buffer)
			if (err != nil) != tt.wantErr {
				t.Errorf("baseWireHandshaker.writeWireHSRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if buffer.Len() != tt.wantSize {
				t.Errorf("baseWireHandshaker.writeWireHSRequest() error = %v, wantErr %v", buffer.Len(), tt.wantSize)
			}

			got, err2 := h.readWireHSResp(buffer)
			if (err2 != nil) != tt.wantErr2 {
				t.Errorf("baseWireHandshaker.readWireHSRequest() error = %v, wantErr %v", err2, tt.wantErr2)
			}
			if !reflect.DeepEqual(tt.args, got) {
				t.Errorf("baseWireHandshaker.readWireHSRequest() = %v, want %v", got, tt.args)
			}
			if buffer.Len() != 0 {
				t.Errorf("baseWireHandshaker.readWireHSRequest() error = %v, wantErr %v", buffer.Len(), 0)
			}

		})
	}
}

func TestInboundWireHandshker_handleInboundPeer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleChainID := &types.ChainID{}
	sampleResult := &p2pcommon.HandshakeResult{}
	logger := log.NewLogger("p2p.test")
	sampleEmptyHSReq := p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: nil}
	sampleEmptyHSResp := p2pcommon.HSHeadResp{Magic: p2pcommon.HSError, RespCode: p2pcommon.HSCodeWrongHSReq}

	type args struct {
		r []byte
	}
	tests := []struct {
		name string
		in   []byte

		bestVer   p2pcommon.P2PVersion
		ctxCancel int  // 0 is not , 1 is during read, 2 is during write
		vhErr     bool // version handshaker failed

		wantW   []byte // sent header
		wantErr bool
	}{
		// All valid
		{"TCurrentVersion", p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: []p2pcommon.P2PVersion{p2pcommon.P2PVersion033, p2pcommon.P2PVersion032, 0x000101}}.Marshal(), p2pcommon.P2PVersion033, 0, false, p2pcommon.HSHeadResp{Magic: p2pcommon.MAGICMain, RespCode: p2pcommon.P2PVersion033.Uint32()}.Marshal(), false},
		{"TOldVersion", p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: []p2pcommon.P2PVersion{0x000010, p2pcommon.P2PVersion032, 0x000101}}.Marshal(), p2pcommon.P2PVersion032, 0, false, p2pcommon.HSHeadResp{Magic: p2pcommon.MAGICMain, RespCode: p2pcommon.P2PVersion032.Uint32()}.Marshal(), false},
		// wrong io read
		{"TWrongRead", sampleEmptyHSReq.Marshal()[:7], p2pcommon.P2PVersion033, 0, false, sampleEmptyHSResp.Marshal(), true},
		// empty version
		{"TEmptyVersion", sampleEmptyHSReq.Marshal(), p2pcommon.P2PVersion033, 0, false, sampleEmptyHSResp.Marshal(), true},
		// wrong io write
		// {"TWrongWrite", sampleEmptyHSReq.Marshal()[:7], sampleEmptyHSResp.Marshal(), true },
		// wrong magic
		{"TWrongMagic", p2pcommon.HSHeadReq{Magic: 0x0001, Versions: []p2pcommon.P2PVersion{p2pcommon.P2PVersion033}}.Marshal(), p2pcommon.P2PVersion033, 0, false, sampleEmptyHSResp.Marshal(), true},
		// not supported version (or wrong version)
		{"TNoVersion", p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: []p2pcommon.P2PVersion{0x000010, 0x030405, 0x000101}}.Marshal(), p2pcommon.P2PVersionUnknown, 0, false, p2pcommon.HSHeadResp{Magic: p2pcommon.HSError, RespCode: p2pcommon.HSCodeNoMatchedVersion}.Marshal(), true},
		// protocol handshake failed
		{"TVersionHSFailed", p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: []p2pcommon.P2PVersion{p2pcommon.P2PVersion033, p2pcommon.P2PVersion032, 0x000101}}.Marshal(), p2pcommon.P2PVersion033, 0, true, p2pcommon.HSHeadResp{Magic: p2pcommon.MAGICMain, RespCode: p2pcommon.P2PVersion033.Uint32()}.Marshal(), true},

		// timeout while read, no reply to remote
		{"TTimeoutRead", p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: []p2pcommon.P2PVersion{p2pcommon.P2PVersion033, p2pcommon.P2PVersion032, 0x000101}}.Marshal(), p2pcommon.P2PVersion033, 1, false, []byte{}, true},
		// timeout while writing, sent but remote not receiving fast
		{"TTimeoutWrite", p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: []p2pcommon.P2PVersion{p2pcommon.P2PVersion033, p2pcommon.P2PVersion032, 0x000101}}.Marshal(), p2pcommon.P2PVersion033, 2, false, p2pcommon.HSHeadResp{Magic: p2pcommon.MAGICMain, RespCode: p2pcommon.P2PVersion033.Uint32()}.Marshal(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockVM := p2pmock.NewMockVersionedManager(ctrl)
			mockVH := p2pmock.NewMockVersionedHandshaker(ctrl)

			mockCtx := NewContextTestDouble(tt.ctxCancel)
			wbuf := bytes.NewBuffer(nil)
			dummyReader := &RWCWrapper{bytes.NewBuffer(tt.in), wbuf, nil}
			dummyMsgRW := p2pmock.NewMockMsgReadWriter(ctrl)

			mockVM.EXPECT().FindBestP2PVersion(gomock.Any()).Return(tt.bestVer).MaxTimes(1)
			mockVM.EXPECT().GetVersionedHandshaker(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockVH, nil).MaxTimes(1)
			if !tt.vhErr {
				mockVH.EXPECT().DoForInbound(mockCtx).Return(sampleResult, nil).MaxTimes(1)
				mockVH.EXPECT().GetMsgRW().Return(dummyMsgRW).MaxTimes(1)
			} else {
				mockVH.EXPECT().DoForInbound(mockCtx).Return(nil, errors.New("version hs failed")).MaxTimes(1)
				mockVH.EXPECT().GetMsgRW().Return(nil).MaxTimes(1)
			}

			h := NewInboundHSHandler(mockPM, mockActor, mockVM, logger, sampleChainID, samplePeerID).(*InboundWireHandshaker)
			got, err := h.handleInboundPeer(mockCtx, dummyReader)
			if (err != nil) != tt.wantErr {
				t.Errorf("InboundWireHandshaker.handleInboundPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !bytes.Equal(wbuf.Bytes(), tt.wantW) {
				t.Errorf("InboundWireHandshaker.handleInboundPeer() send resp %v, want %v", wbuf.Bytes(), tt.wantW)
			}
			if !tt.wantErr {
				if got == nil {
					t.Errorf("InboundWireHandshaker.handleInboundPeer() got msgrw nil, want not")
				}
			}
		})
	}
}

func TestOutboundWireHandshaker_handleOutboundPeer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleChainID := &types.ChainID{}
	sampleResult := &p2pcommon.HandshakeResult{}
	logger := log.NewLogger("p2p.test")
	// This bytes is actually hard-coded in source handshake_v2.go.
	outBytes := p2pcommon.HSHeadReq{Magic: p2pcommon.MAGICMain, Versions: []p2pcommon.P2PVersion{p2pcommon.P2PVersion200, p2pcommon.P2PVersion033, p2pcommon.P2PVersion032, p2pcommon.P2PVersion031}}.Marshal()

	tests := []struct {
		name string

		remoteRespVer  p2pcommon.P2PVersion
		ctxCancel      int    // 0 is not , 1 is during write, 2 is during read
		versionHSerror bool   // whether version handshaker return failed or not
		remoteResponse []byte // emulate response from remote peer

		wantErr bool
	}{
		// remote listening peer accept my best p2p version
		{"TCurrentVersion", p2pcommon.P2PVersion200, 0, false, p2pcommon.HSHeadResp{Magic: p2pcommon.MAGICMain, RespCode: p2pcommon.P2PVersion200.Uint32()}.Marshal(), false},
		// remote listening peer can connect, but old p2p version
		{"TOldVersion", p2pcommon.P2PVersion032, 0, false, p2pcommon.HSHeadResp{Magic: p2pcommon.MAGICMain, RespCode: p2pcommon.P2PVersion032.Uint32()}.Marshal(), false},
		{"TOlderVersion", p2pcommon.P2PVersion031, 0, false, p2pcommon.HSHeadResp{Magic: p2pcommon.MAGICMain, RespCode: p2pcommon.P2PVersion031.Uint32()}.Marshal(), false},
		{"TOldestVersion", p2pcommon.P2PVersion030, 0, false, p2pcommon.HSHeadResp{Magic: p2pcommon.MAGICMain, RespCode: p2pcommon.P2PVersion030.Uint32()}.Marshal(), false},
		// wrong io read
		{"TWrongResp", p2pcommon.P2PVersion032, 0, false, outBytes[:6], true},
		// {"TWrongWrite", sampleEmptyHSReq.Marshal()[:7], sampleEmptyHSResp.Marshal(), true },
		// wrong magic
		{"TWrongMagic", p2pcommon.P2PVersion032, 0, false, p2pcommon.HSHeadResp{Magic: p2pcommon.HSError, RespCode: p2pcommon.HSCodeWrongHSReq}.Marshal(), true},
		// not supported version (or wrong version)
		{"TNoVersion", p2pcommon.P2PVersionUnknown, 0, false, p2pcommon.HSHeadResp{Magic: p2pcommon.HSError, RespCode: p2pcommon.HSCodeNoMatchedVersion}.Marshal(), true},
		// protocol handshake failed
		{"TVersionHSFailed", p2pcommon.P2PVersion032, 0, true, p2pcommon.HSHeadResp{Magic: p2pcommon.MAGICMain, RespCode: p2pcommon.P2PVersion032.Uint32()}.Marshal(), true},

		// timeout while read, no reply to remote
		{"TTimeoutRead", p2pcommon.P2PVersion031, 1, false, []byte{}, true},
		// timeout while writing, sent but remote not receiving fast
		{"TTimeoutWrite", p2pcommon.P2PVersion032, 2, false, p2pcommon.HSHeadResp{Magic: p2pcommon.MAGICMain, RespCode: p2pcommon.P2PVersion032.Uint32()}.Marshal(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockVM := p2pmock.NewMockVersionedManager(ctrl)
			mockVH := p2pmock.NewMockVersionedHandshaker(ctrl)

			mockCtx := NewContextTestDouble(tt.ctxCancel)
			wbuf := bytes.NewBuffer(nil)
			dummyRWC := &RWCWrapper{bytes.NewBuffer(tt.remoteResponse), wbuf, nil}
			dummyMsgRW := p2pmock.NewMockMsgReadWriter(ctrl)

			mockVM.EXPECT().GetVersionedHandshaker(tt.remoteRespVer, gomock.Any(), gomock.Any()).Return(mockVH, nil).MaxTimes(1)
			if tt.versionHSerror {
				mockVH.EXPECT().DoForOutbound(mockCtx).Return(nil, errors.New("version hs failed")).MaxTimes(1)
				mockVH.EXPECT().GetMsgRW().Return(nil).MaxTimes(1)
			} else {
				mockVH.EXPECT().DoForOutbound(mockCtx).Return(sampleResult, nil).MaxTimes(1)
				mockVH.EXPECT().GetMsgRW().Return(dummyMsgRW).MaxTimes(1)
			}

			h := NewOutboundHSHandler(mockPM, mockActor, mockVM, logger, sampleChainID, samplePeerID).(*OutboundWireHandshaker)
			got, err := h.handleOutboundPeer(mockCtx, dummyRWC)
			if (err != nil) != tt.wantErr {
				t.Errorf("OutboundWireHandshaker.handleOutboundPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !bytes.Equal(wbuf.Bytes(), outBytes) {
				t.Errorf("OutboundWireHandshaker.handleOutboundPeer() send resp %v, want %v", wbuf.Bytes(), outBytes)
			}
			if !tt.wantErr {
				if got == nil {
					t.Errorf("OutboundWireHandshaker.handleOutboundPeer() got msgrw nil, want not")
				}
			}
		})
	}
}

type RWCWrapper struct {
	r io.Reader
	w io.Writer
	c io.Closer
}

func (rwc RWCWrapper) Read(p []byte) (n int, err error) {
	return rwc.r.Read(p)
}

func (rwc RWCWrapper) Write(p []byte) (n int, err error) {
	return rwc.w.Write(p)
}

func (rwc RWCWrapper) Close() error {
	return rwc.c.Close()
}

type ContextTestDouble struct {
	doneChannel chan struct{}
	expire      uint32
	callCnt     uint32
}

var _ context.Context = (*ContextTestDouble)(nil)

func NewContextTestDouble(expire int) *ContextTestDouble {
	if expire <= 0 {
		expire = 9999999
	}
	return &ContextTestDouble{expire: uint32(expire), doneChannel: make(chan struct{}, 1)}
}

func (*ContextTestDouble) Deadline() (deadline time.Time, ok bool) {
	panic("implement me")
}

func (c *ContextTestDouble) Done() <-chan struct{} {
	current := atomic.AddUint32(&c.callCnt, 1)
	if current >= c.expire {
		c.doneChannel <- struct{}{}
	}
	return c.doneChannel
}

func (c *ContextTestDouble) Err() error {
	if atomic.LoadUint32(&c.callCnt) >= c.expire {
		return errors.New("timeout")
	} else {
		return nil
	}
}

func (*ContextTestDouble) Value(key interface{}) interface{} {
	panic("implement me")
}
