/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"bytes"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
)

func TestStartGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		peerCnt int
		timeout time.Duration
	}
	tests := []struct {
		name string
		args args

		wantSentCnt int  // count of sent to remote peers
		wantTimeout bool // whether receiver returns result or not (=timeout)
		wantErrResp bool // result with error or not
	}{
		{"TTimeout", args{peerCnt: 1}, 1, true, false},
		{"TNoPeers", args{peerCnt: 0}, 0, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMo := createDummyMo(ctrl)
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo).Times(tt.wantSentCnt)
			peers := make([]p2pcommon.RemotePeer, 0, tt.args.peerCnt)
			for i := 0; i < tt.args.peerCnt; i++ {
				dummyPeerID, _ := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
				peers = append(peers, createDummyPeer(ctrl, dummyPeerID, types.RUNNING))
			}
			replyChan := make(chan *message.GetClusterRsp)
			dummyReq := &message.GetCluster{ReplyC: replyChan}
			target := NewClusterInfoReceiver(mockActor, mockMF, peers, time.Millisecond, dummyReq)
			target.StartGet()

			if !tt.wantTimeout {
				timer := time.NewTimer(time.Second * 2)
				select {
				case resp := <-replyChan:
					if (resp.Err != nil) != tt.wantErrResp {
						t.Errorf("resp error %v, wantErr %v ", resp.Err, tt.wantErrResp)
					}
				case <-timer.C:
					t.Errorf("timeout occurred, want no time")
				}
			} else {
				timer := time.NewTimer(time.Millisecond * 100)
				select {
				case resp := <-replyChan:
					t.Errorf("unexpected response (%d mems, err:%v), want timeout", len(resp.Members), resp.Err)
				case <-timer.C:
					// expected timeout
				}
			}
		})
	}
}

func createDummyPeer(ctrl *gomock.Controller, pid types.PeerID, state types.PeerState) *p2pmock.MockRemotePeer {
	mockPeer := p2pmock.NewMockRemotePeer(ctrl)
	mockPeer.EXPECT().State().Return(state).AnyTimes()
	mockPeer.EXPECT().ID().Return(pid).AnyTimes()
	mockPeer.EXPECT().ConsumeRequest(gomock.Any()).AnyTimes()
	mockPeer.EXPECT().SendMessage(gomock.Any()).AnyTimes()
	return mockPeer
}

func createDummyMo(ctrl *gomock.Controller) *p2pmock.MockMsgOrder {
	dummyMo := p2pmock.NewMockMsgOrder(ctrl)
	dummyMo.EXPECT().IsNeedSign().Return(true).AnyTimes()
	dummyMo.EXPECT().IsRequest().Return(true).AnyTimes()
	dummyMo.EXPECT().GetProtocolID().Return(p2pcommon.NewTxNotice).AnyTimes()
	dummyMo.EXPECT().GetMsgID().Return(p2pcommon.NewMsgID()).AnyTimes()
	return dummyMo
}

func TestClusterInfoReceiver_trySendNextPeer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		stats []int
	}
	tests := []struct {
		name string
		args args

		wantSentCnt int
	}{
		{"TAllRunning", args{[]int{1, 1, 1, 1, 1}}, 5},
		{"TNoPeers", args{[]int{}}, 0},
		{"TNoRunning", args{[]int{0, 0, 0, 0, 0}}, 0},
		{"TMixed", args{[]int{0, 0, 1, 1, 1}}, 3},
		{"TMixed2", args{[]int{1, 1, 0, 0, 0}}, 2},
		{"TMixed3", args{[]int{1, 0, 1, 0, 0, 1}}, 3},
		{"TMixed4", args{[]int{0, 1, 0, 1, 0, 1, 0}}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMo := createDummyMo(ctrl)
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo).Times(tt.wantSentCnt)
			peers := make([]p2pcommon.RemotePeer, 0, len(tt.args.stats))
			for _, run := range tt.args.stats {
				dummyPeerID, _ := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
				stat := types.RUNNING
				if run == 0 {
					stat = types.STOPPING
				}
				peers = append(peers, createDummyPeer(ctrl, dummyPeerID, stat))
			}

			sentCnt := 0
			replyChan := make(chan *message.GetClusterRsp)
			dummyReq := &message.GetCluster{ReplyC: replyChan}
			target := NewClusterInfoReceiver(mockActor, mockMF, peers, time.Millisecond, dummyReq)
			for target.trySendNextPeer() {
				sentCnt++
			}

			if sentCnt != tt.wantSentCnt {
				t.Errorf("resp error %v, wantErr %v ", sentCnt, tt.wantSentCnt)
			}
		})
	}
}

func TestClusterInfoReceiver_ReceiveResp(t *testing.T) {
	sampleChainID := []byte("testChain")
	members := make([]*types.MemberAttr, 4)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		stats []int
	}
	tests := []struct {
		name string
		args args

		wantSentCnt int  // count of sent to remote peers
		wantTimeout bool // whether receiver returns result or not (=timeout)
		wantErrResp bool // result with error or not
	}{
		{"TAllRet", args{[]int{1, 1, 1, 1, 1}}, 1, false, false},
		{"TErrRet", args{[]int{0, 0, 0, 0, 0}}, 5, false, true},
		{"TMixed", args{[]int{0, 0, 1, 1, 1}}, 3, false, false},
		{"TTimeout", args{[]int{0, 0}}, 3, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			peers := make([]p2pcommon.RemotePeer, 0, len(tt.args.stats))

			mockActor := p2pmock.NewMockActorService(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMo := createDummyMo(ctrl)
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo).Times(tt.wantSentCnt)

			replyChan := make(chan *message.GetClusterRsp)
			dummyReq := &message.GetCluster{ReplyC: replyChan}
			target := NewClusterInfoReceiver(mockActor, mockMF, peers, time.Second, dummyReq)

			seq := int32(0)
			for i := 0; i < 5; i++ {
				dummyPeerID, _ := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
				stat := types.RUNNING
				mockPeer := p2pmock.NewMockRemotePeer(ctrl)
				mockPeer.EXPECT().State().Return(stat).AnyTimes()
				mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
				mockPeer.EXPECT().ConsumeRequest(gomock.Any()).AnyTimes()
				mockPeer.EXPECT().SendMessage(gomock.Any()).Do(func(mo p2pcommon.MsgOrder) {
					time.Sleep(time.Millisecond * 5)
					callSeq := atomic.LoadInt32(&seq)
					msg := p2pmock.NewMockMessage(ctrl)
					msg.EXPECT().ID().Return(p2pcommon.NewMsgID()).AnyTimes()
					msg.EXPECT().OriginalID().Return(p2pcommon.NewMsgID()).AnyTimes()
					msg.EXPECT().Timestamp().Return(time.Now().UnixNano()).AnyTimes()
					msg.EXPECT().Subprotocol().Return(p2pcommon.GetClusterResponse).AnyTimes()
					if callSeq < int32(len(tt.args.stats)) {
						err := ""
						if tt.args.stats[callSeq] == 0 {
							err = "getCluster failed"
						}
						body := &types.GetClusterInfoResponse{ChainID: sampleChainID, MbrAttrs: members, Error: err}
						atomic.AddInt32(&seq, 1)
						go target.ReceiveResp(msg, body)
					} else {
						atomic.AddInt32(&seq, 1)
					}
				}).MaxTimes(1)
				peers = append(peers, mockPeer)
			}
			// force inject peers
			target.peers = peers

			target.StartGet()

			if !tt.wantTimeout {
				timer := time.NewTimer(time.Second * 2)
				select {
				case resp := <-replyChan:
					if (resp.Err != nil) != tt.wantErrResp {
						t.Errorf("resp error %v, wantErr %v ", resp.Err, tt.wantErrResp)
					}
					// receiver return valid result
					if !tt.wantErrResp {
						if !bytes.Equal(resp.ChainID, sampleChainID) {
							t.Errorf("resp chainid %v, want %v ", resp.ChainID, sampleChainID)
						}
						if len(resp.Members) != len(members) {
							t.Errorf("resp members %v, want %v ", resp.Members, len(members))
						}
					}
				case <-timer.C:
					t.Errorf("timeout occurred, want no time")
				}
			} else {
				timer := time.NewTimer(time.Millisecond * 100)
				select {
				case resp := <-replyChan:
					t.Errorf("unexpected response (%d mems, err:%v), want timeout", len(resp.Members), resp.Err)
				case <-timer.C:
					// expected timeout
				}
			}

		})
	}
}
