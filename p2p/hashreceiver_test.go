/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/funkygao/golib/rand"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBlockHashesReceiver_StartGet(t *testing.T) {
	sampleBlk := &types.BlockInfo{Hash: dummyBlockHash, No: 10000}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inputHashes := make([]message.BlockHash, len(sampleBlks))
	for i, hash := range sampleBlks {
		inputHashes[i] = hash
	}
	tests := []struct {
		name  string
		input *message.GetHashes
		ttl   time.Duration
	}{
		{"TSimple", &message.GetHashes{Seq: 100, ToWhom: dummyPeerID, PrevInfo: sampleBlk, Count: 100}, time.Millisecond * 10},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//mockContext := new(mockContext)
			mockActor := p2pmock.NewMockActorService(ctrl)
			//mockActor.On("SendRequest", message.P2PSvc, mock.AnythingOfType("*types.GetBlock"))
			//mockActor.On("TellRequest", message.SyncerSvc, mock.AnythingOfType("*types.GetBlock"))
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMo := createDummyMo(ctrl)
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().MF().Return(mockMF)
			mockPeer.EXPECT().SendMessage(mockMo).Times(1)

			expire := time.Now().Add(test.ttl)
			br := NewBlockHashesReceiver(mockActor, mockPeer, test.input.Seq, test.input, test.ttl)

			br.StartGet()

			assert.False(t, expire.After(br.timeout))
		})
	}
}

func TestBlockHashesReceiver_ReceiveResp(t *testing.T) {
	//t.Skip("make check by status. and make another test to check handleInWaiting method")
	sampleBlk := &types.BlockInfo{Hash: dummyBlockHash, No: 10000}
	limit := uint64(10)
	chain.Init(1<<20, "", false, 1, 1)

	totalInCnt := 10
	seqNo := uint64(8723)
	inputHashes := make([][]byte, totalInCnt)
	for i := 0; i < totalInCnt; i++ {
		inputHashes[i] = rand.RandomByteSlice(types.HashIDLength)
	}
	wrongHashes := [][]byte{rand.RandomByteSlice(types.HashIDLength - 3)}
	emptyHashes := [][]byte{{}, {}}
	tests := []struct {
		name         string
		input        *message.GetHashes
		ttl          time.Duration
		hashInterval time.Duration
		hashInput    [][][]byte

		// to verify
		instantFinish int
		sentResp      int
		respError     bool
	}{
		{"TSingleResp", &message.GetHashes{Seq: seqNo, ToWhom: dummyPeerID, PrevInfo: sampleBlk, Count: limit}, time.Minute, 0, [][][]byte{inputHashes}, 1, 1, false},
		{"TMultiResp", &message.GetHashes{Seq: seqNo, ToWhom: dummyPeerID, PrevInfo: sampleBlk, Count: limit}, time.Minute, 0, [][][]byte{inputHashes[:1], inputHashes[1:3], inputHashes[3:]}, 1, 1, false},
		{"TFewerHashes", &message.GetHashes{Seq: seqNo, ToWhom: dummyPeerID, PrevInfo: sampleBlk, Count: limit}, time.Minute, 0, [][][]byte{inputHashes[:3], inputHashes[3:7]}, 1, 1, false},
		// Fail1 remote err
		{"TWrongHash1", &message.GetHashes{Seq: seqNo, ToWhom: dummyPeerID, PrevInfo: sampleBlk, Count: limit}, time.Second >> 1, 0, [][][]byte{emptyHashes, inputHashes[:2], inputHashes[3:7]}, 0, 1, true},
		{"TWrongHash2", &message.GetHashes{Seq: seqNo, ToWhom: dummyPeerID, PrevInfo: sampleBlk, Count: limit}, time.Minute, 0, [][][]byte{inputHashes[:2], wrongHashes}, 1, 1, true},
		{"TRemoteFail", &message.GetHashes{Seq: seqNo, ToWhom: dummyPeerID, PrevInfo: sampleBlk, Count: limit}, time.Minute, 0, [][][]byte{inputHashes[:0]}, 1, 1, true},
		{"TTooManyBlks", &message.GetHashes{Seq: seqNo, ToWhom: dummyPeerID, PrevInfo: sampleBlk, Count: limit - 2}, time.Minute * 4, 0, [][][]byte{inputHashes[:1], inputHashes[1:3], inputHashes[3:]}, 1, 1, true},
		// Fail4 response sent after timeout
		{"TTimeout", &message.GetHashes{Seq: seqNo, ToWhom: dummyPeerID, PrevInfo: sampleBlk, Count: limit}, time.Millisecond * 10, time.Millisecond * 20, [][][]byte{inputHashes[:1], inputHashes[1:3], inputHashes[3:]}, 1, 0, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//mockContext := new(mockContext)
			mockActor := p2pmock.NewMockActorService(ctrl)
			if test.sentResp > 0 {
				mockActor.EXPECT().TellRequest(message.SyncerSvc, gomock.AssignableToTypeOf(&message.GetHashesRsp{})).
					DoAndReturn(func(a string, arg *message.GetHashesRsp) {
						if !((arg.Err != nil) == test.respError) {
							t.Fatalf("Wrong error (have %v)", arg.Err)
						}
						if arg.Seq != seqNo {
							t.Fatalf("Wrong seqNo %d, want %d)", arg.Seq, seqNo)
						}
						if arg.Count != uint64(len(arg.Hashes)) {
							t.Fatalf("Count mismatch Count %d, real hash count %d)", arg.Count, len(arg.Hashes))
						}
						if arg.Err == nil {
							for _, h := range arg.Hashes {
								_, err := types.ParseToBlockID(h)
								if err != nil {
									t.Fatalf("Wrong block hash %s, err %v)", enc.ToString(h), err.Error())
								}
							}
						}

					}).Times(test.sentResp)
			}

			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMo := createDummyMo(ctrl)
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().MF().Return(mockMF)
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)
			mockPeer.EXPECT().ConsumeRequest(gomock.Any()).Times(1) //mock.AnythingOfType("p2pcommon.MsgID"))

			//expire := time.Now().Add(test.ttl)
			br := NewBlockHashesReceiver(mockActor, mockPeer, seqNo, test.input, test.ttl)
			br.StartGet()

			msg := p2pcommon.NewSimpleMsgVal(p2pcommon.GetHashesRequest, sampleMsgID)
			for i, hashes := range test.hashInput {
				if test.hashInterval > 0 {
					time.Sleep(test.hashInterval)
				}
				body := &types.GetHashesResponse{Hashes: hashes, HasNext: i < len(test.hashInput)-1}
				br.ReceiveResp(msg, body)
				//if br.status == receiverStatusFinished {
				//	break
				//}
			}
			// if sender is supposed to send other partial messages, receiver will wait and cleanup those messages
			if (br.senderFinished != nil) != (test.instantFinish == 0) {
				t.Fatalf("cleanup routine is %v , want %v)", (br.senderFinished != nil), (test.instantFinish == 0))
			}
			if test.instantFinish == 0 {
				time.Sleep(test.ttl * 2)
			}
		})
	}
}
