/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"bytes"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
)

func TestConcurrentClusterInfoReceiver_StartGet(t *testing.T) {
	logger := log.NewLogger("raft.support.test")
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
		{"TTimeout", args{peerCnt: 3}, 3, true, true},
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
			target := NewConcClusterInfoReceiver(mockActor, mockMF, peers, time.Millisecond*20, dummyReq, logger)
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
				timer := time.NewTimer(time.Millisecond * 200)
				select {
				case resp := <-replyChan:
					if (resp.Err != nil) != tt.wantErrResp {
						t.Errorf("resp error %v, wantErr %v ", resp.Err, tt.wantErrResp)
					}
				case <-timer.C:
					// expected timeout
				}
			}
		})
	}
}

func TestConcurrentClusterInfoReceiver_trySendAllPeers(t *testing.T) {
	logger := log.NewLogger("raft.support.test")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		stats []int
	}
	tests := []struct {
		name string
		args args

		wantSentCnt int
		wantSucc    bool
	}{
		{"TAllRunning", args{[]int{1, 1, 1, 1, 1}}, 5, true},
		{"TNoPeers", args{[]int{}}, 0, false},
		{"TNoRunning", args{[]int{0, 0, 0, 0, 0}}, 0, false},
		{"TMixed", args{[]int{0, 0, 1, 1, 1}}, 3, true},
		{"TMixed2", args{[]int{1, 1, 0, 0, 0}}, 2, false},
		{"TMixed3", args{[]int{1, 0, 1, 0, 1, 1}}, 4, true},
		{"TMixed3Fail", args{[]int{1, 0, 1, 0, 0, 1}}, 3, false},
		{"TMixed4", args{[]int{0, 1, 0, 1, 0, 1, 0}}, 3, false},
		{"TSingle", args{[]int{1}}, 1, true},
		{"TSingleFail", args{[]int{0}}, 0, false},
		{"TTwoFail", args{[]int{0, 1}}, 1, false},
		{"TTwoFail2", args{[]int{0, 0}}, 0, false},
		{"TThree", args{[]int{1, 0, 1}}, 2, true},
		{"TThreeFail", args{[]int{1, 0, 0}}, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(respReceiver p2pcommon.ResponseReceiver, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) p2pcommon.MsgOrder {
				return createDummyMo(ctrl)
			}).Times(tt.wantSentCnt)
			peers := make([]p2pcommon.RemotePeer, 0, len(tt.args.stats))
			for _, run := range tt.args.stats {
				dummyPeerID := types.RandomPeerID()
				stat := types.RUNNING
				if run == 0 {
					stat = types.STOPPING
				}
				peers = append(peers, createDummyPeer(ctrl, dummyPeerID, stat))
			}

			replyChan := make(chan *message.GetClusterRsp)
			dummyReq := &message.GetCluster{ReplyC: replyChan}
			target := NewConcClusterInfoReceiver(mockActor, mockMF, peers, time.Millisecond, dummyReq, logger)
			ret := target.trySendAllPeers()

			if len(target.sent) != tt.wantSentCnt {
				t.Errorf("trySendAllPeers sentCount %v, want %v ", len(target.sent), tt.wantSentCnt)
			}
			if ret != tt.wantSucc {
				t.Errorf("trySendAllPeers() ret %v, want %v ", ret, tt.wantSucc)
			}
		})
	}
}

type retStat int

const (
	ERR = -1
	NOR = -2
)

func TestConcurrentClusterInfoReceiver_ReceiveResp(t *testing.T) {
	logger := log.NewLogger("raft.support.test")

	sampleChainID := []byte("testChain")
	members := make([]*types.MemberAttr, 4)

	type args struct {
		retStats []retStat
	}
	tests := []struct {
		name string
		args args

		wantBestNo  int  // count of sent to remote peers
		wantErrResp bool // result with error or not
	}{
		{"TAllSame", args{[]retStat{10, 10, 10, 10, 10}}, 10, false},
		{"TErrRet", args{[]retStat{ERR, ERR, ERR, ERR, ERR}}, 10, true},
		{"TMixed", args{[]retStat{100, ERR, 99, 98, ERR}}, 100, false},
		{"TMixed2", args{[]retStat{100, ERR, NOR, ERR, 100}}, 100, true},
		{"TTimeSucc", args{[]retStat{NOR, 99, NOR, 98, 99}}, 99, false},
		{"TTime1", args{[]retStat{NOR, ERR, NOR, 100, 100}}, 100, true},
		{"TTime2", args{[]retStat{NOR, NOR, NOR, 100, 100}}, 100, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			peers := make([]p2pcommon.RemotePeer, len(tt.args.retStats))

			mockActor := p2pmock.NewMockActorService(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMO := p2pmock.NewMockMsgOrder(ctrl)

			replyChan := make(chan *message.GetClusterRsp)
			dummyReq := &message.GetCluster{ReplyC: replyChan}

			sMap := make(map[p2pcommon.MsgID]p2pcommon.RemotePeer)
			rHeads := make([]p2pcommon.Message, 0, len(tt.args.retStats))
			rBodies := make([]*types.GetClusterInfoResponse, 0, len(tt.args.retStats))

			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), p2pcommon.GetClusterRequest, gomock.Any()).
				Times(len(tt.args.retStats)).Return(mockMO)
			sentTrigger := int32(0)
			prevCall := (*gomock.Call)(nil)
			for i, st := range tt.args.retStats {
				dummyPeerID := types.RandomPeerID()
				msgID := p2pcommon.NewMsgID()
				stat := types.RUNNING
				mockPeer := p2pmock.NewMockRemotePeer(ctrl)
				mockPeer.EXPECT().State().Return(stat).AnyTimes()
				mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
				mockPeer.EXPECT().Name().Return("peer" + p2putil.ShortForm(dummyPeerID)).AnyTimes()
				mockPeer.EXPECT().ConsumeRequest(gomock.Any()).AnyTimes()
				mockPeer.EXPECT().SendMessage(mockMO).Do(func(arg interface{}) {
					atomic.StoreInt32(&sentTrigger, 1)
				})
				peers[i] = mockPeer
				sMap[msgID] = mockPeer

				// mockMO
				if prevCall == nil {
					prevCall = mockMO.EXPECT().GetMsgID().Return(msgID)
				} else {
					prevCall = mockMO.EXPECT().GetMsgID().Return(msgID).After(prevCall)
				}

				head := p2pcommon.NewLiteMessageValue(p2pcommon.RaftWrapperMessage, p2pcommon.NewMsgID(), msgID, time.Now().UnixNano())
				body := &types.GetClusterInfoResponse{ChainID: sampleChainID, MbrAttrs: members, Error: ""}
				switch st {
				case NOR:
					continue
				case ERR:
					body.Error = "getCluster failed"
					fallthrough
				default:
					rHeads = append(rHeads, head)
					rBodies = append(rBodies, body)
				}
			}
			ttl := time.Second >> 4
			target := NewConcClusterInfoReceiver(mockActor, mockMF, peers, ttl, dummyReq, logger)
			target.StartGet()

			wg := sync.WaitGroup{}
			wg.Add(len(rHeads))
			bar := sync.WaitGroup{}
			bar.Add(1)
			for i, head := range rHeads {
				go func(msg p2pcommon.Message, body *types.GetClusterInfoResponse) {
					wg.Done()
					bar.Wait()
					target.ReceiveResp(msg, body)
				}(head, rBodies[i])
			}
			for atomic.LoadInt32(&sentTrigger) == 0 {
				time.Sleep(time.Millisecond)
			}
			wg.Wait()
			bar.Done()
			timer := time.NewTimer(ttl << 1)
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
						t.Errorf("resp members %v, want %v ", len(resp.Members), len(members))
					}
				}
			case <-timer.C:
				t.Fatalf("timeout occurred, want not")
			}

			ctrl.Finish()
		})
	}
}
