/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var NetError = fmt.Errorf("network err for unittest")

func Test_pingChecker_DoCall(t *testing.T) {
	ctrl := gomock.NewController(t)

	type args struct {
		rw        p2pcommon.MsgReadWriter
		writeWait int
		writeRet  error
		readWait  int
		respSub   p2pcommon.SubProtocol
		readRet2  error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// 1. msg writer succeeded send and succeeded read
		{"Tsucc", args{writeRet: nil, readRet2: nil, respSub: p2pcommon.PingResponse}, false},
		// 2. failed to write
		{"TFailWrite", args{writeRet: NetError, readRet2: nil, respSub: p2pcommon.PingResponse}, true},
		// 3. failed to read
		{"TFailRead", args{writeRet: nil, readRet2: NetError, respSub: p2pcommon.PingResponse}, true},
		// 4. read but not ping response
		{"TWrongResp", args{writeRet: nil, readRet2: nil, respSub: p2pcommon.AddressesResponse}, true},
		// 5. cancel signal  while writing
		{"TTimeoutWrite", args{writeRet: nil, writeWait: 3, readRet2: nil, respSub: p2pcommon.PingResponse}, true},
		// 6. cancel signal while reading
		{"TTimeoutRead", args{writeRet: nil, readWait: 3, readRet2: nil, respSub: p2pcommon.PingResponse}, true},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			ps := &peerState{temporary: true, PeerMapService: &PeerMapService{BaseComponent: &component.BaseComponent{Logger: log.NewLogger("test")}}}
			rw := p2pmock.NewMockMsgReadWriter(ctrl)
			pc := &pingChecker{
				peerState: ps,
				rw:        rw,
			}

			var reqID p2pcommon.MsgID
			rw.EXPECT().WriteMsg(gomock.Any()).Do(func(msg p2pcommon.Message) {
				reqID = msg.ID()
				if tt.args.writeWait > 0 {
					pc.Cancel()
					time.Sleep(time.Millisecond << 4)
				}
			}).Return(tt.args.writeRet)
			rw.EXPECT().ReadMsg().MaxTimes(1).DoAndReturn(func() (p2pcommon.Message, error) {
				if tt.args.readWait > 0 {
					pc.Cancel()
					time.Sleep(time.Millisecond << 4)
				}
				ret := p2pcommon.NewMessageValue(tt.args.respSub, EmptyMsgID, reqID, time.Now().UnixNano(), []byte{})
				return ret, tt.args.readRet2
			})

			done := make(chan interface{}, 1)
			pc.DoCall(done)
			result := <-done
			if tt.wantErr {
				assert.Nil(t, result.(*types.Ping))
			} else {
				assert.NotNil(t, result.(*types.Ping))
			}
		})

	}
	ctrl.Finish()
}

func Test_pingChecker_DoCallWithTimer(t *testing.T) {
	ctrl := gomock.NewController(t)

	type args struct {
		rw        p2pcommon.MsgReadWriter
		writeWait int
		writeRet  error
		readWait  int
		respSub   p2pcommon.SubProtocol
		readRet2  error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// 1. msg writer succeeded send and succeeded read
		{"Tsucc", args{writeRet: nil, readRet2: nil, respSub: p2pcommon.PingResponse}, false},
		// 2. failed to write
		{"TFailWrite", args{writeRet: NetError, readRet2: nil, respSub: p2pcommon.PingResponse}, true},
		// 3. failed to read
		{"TFailRead", args{writeRet: nil, readRet2: NetError, respSub: p2pcommon.PingResponse}, true},
		// 4. read but not ping response
		{"TWrongPayload", args{writeRet: nil, readRet2: nil, respSub: p2pcommon.StatusRequest}, true},
		// 5. cancel signal  while writing
		{"TTimeoutWrite", args{writeRet: nil, writeWait: 3, readRet2: nil, respSub: p2pcommon.PingResponse}, true},
		// 6. cancel signal while reading
		{"TTimeoutRead", args{writeRet: nil, readWait: 3, readRet2: nil, respSub: p2pcommon.PingResponse}, true},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			ps := &peerState{temporary: true, PeerMapService: &PeerMapService{BaseComponent: &component.BaseComponent{Logger: log.NewLogger("test")}}}
			rw := p2pmock.NewMockMsgReadWriter(ctrl)
			pc := &pingChecker{
				peerState: ps,
				rw:        rw,
			}

			var reqID p2pcommon.MsgID
			rw.EXPECT().WriteMsg(gomock.Any()).Do(func(msg p2pcommon.Message) {
				reqID = msg.ID()
				if tt.args.writeWait > 0 {
					time.Sleep(time.Second)
				}
			}).Return(tt.args.writeRet)
			rw.EXPECT().ReadMsg().MaxTimes(1).DoAndReturn(func() (p2pcommon.Message, error) {
				if tt.args.readWait > 0 {
					time.Sleep(time.Second)
				}
				ret := p2pcommon.NewMessageValue(tt.args.respSub, EmptyMsgID, reqID, time.Now().UnixNano(), []byte{})
				return ret, tt.args.readRet2
			})

			result, err := p2putil.InvokeWithTimer(pc, time.NewTimer(time.Millisecond<<5))
			if tt.args.readWait > 0 || tt.args.writeWait > 0 {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				if tt.wantErr {
					assert.Nil(t, result.(*types.Ping))
				} else {
					assert.NotNil(t, result.(*types.Ping))
				}
			}
		})

	}
	ctrl.Finish()
}
