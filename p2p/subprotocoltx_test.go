/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package p2p

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/message/mocks"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"

	"github.com/aergoio/aergo/internal/enc"

	"github.com/satori/go.uuid"
)

const hashSize = 32
var sampleTxsB58 = []string{ "4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45", "6xfk39kuyDST7NwCu8tx3wqwFZ5dwKPDjxUS14tU7NZb8",
"E8dbBGe9Hnuhk35cJoekPjL3VoL4xAxtnRuP47UoxzHd",
}
var sampleTxs [][]byte
func init() {
	sampleTxs = make([][]byte, len(sampleTxsB58))
	for i, hashb58 := range sampleTxsB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleTxs[i] = hash
	}
}
func TestTxRequestHandler_handle(t *testing.T) {
	logger := log.NewLogger("test.p2p")
	dummyMeta := PeerMeta{ID:dummyPeerID,IPAddress:"192.168.1.2",Port:4321}
	mockSigner := new(mockMsgSigner)
	mockSigner.On("signMsg",mock.Anything).Return(nil)
	tests := []struct {
		name string
		setup func(pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockRW *MockMsgReadWriter) (*types.MsgHeader,*types.GetTransactionsRequest)
		verify func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockRW *MockMsgReadWriter)
	}{
		// 1. success case (single tx)
		{"TSucc1",func(pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockRW *MockMsgReadWriter) (*types.MsgHeader,*types.GetTransactionsRequest){
			dummyTx := &types.Tx{Hash:nil}
			actor.On("CallRequest",message.MemPoolSvc, mock.AnythingOfType("*message.MemPoolExist")).Return(&message.MemPoolExistRsp{}, nil)
			msgHelper.On("ExtractTxFromResponseAndError", mock.AnythingOfType("*message.MemPoolExistRsp"), nil).Return(dummyTx, nil)
			hashes := sampleTxs[:1]
			return nil, &types.GetTransactionsRequest{Hashes:hashes}
		}, func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockRW *MockMsgReadWriter) {
			actor.AssertNumberOfCalls(tt,"CallRequest",1)
			msgHelper.AssertNumberOfCalls(tt,"ExtractTxFromResponseAndError",1)
		}},
		// 1-1 success case2 (multiple tx)
		{"TSucc2",func(pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockRW *MockMsgReadWriter) (*types.MsgHeader,*types.GetTransactionsRequest){
			dummyTx := &types.Tx{Hash:nil}
			actor.On("CallRequest",message.MemPoolSvc, mock.AnythingOfType("*message.MemPoolExist")).Return(&message.MemPoolExistRsp{}, nil)
			msgHelper.On("ExtractTxFromResponseAndError", mock.AnythingOfType("*message.MemPoolExistRsp"), nil).Return(dummyTx, nil)
			hashes := sampleTxs
			return nil, &types.GetTransactionsRequest{Hashes:hashes}
		}, func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockRW *MockMsgReadWriter) {
			actor.AssertNumberOfCalls(tt,"CallRequest",len(sampleTxs))
			msgHelper.AssertNumberOfCalls(tt,"ExtractTxFromResponseAndError",len(sampleTxs))
		}},
		// 2. hash not found (partial)
		// TODO testcase 2 and 3 need refactoring to test. it should verify response parameter
		//{"TPartialFailure",func(pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockRW *MockMsgReadWriter) (*types.MsgHeader,*types.GetTransactionsRequest){
		//	dummyTx := &types.Tx{Hash:nil}
		//	actor.On("CallRequest",message.MemPoolSvc, mock.MatchedBy(func(in *message.MemPoolExist) bool {
		//		if( bytes.Equal(in.Hash, sampleTxs[1])  ) {
		//			return true
		//		} else {
		//			return false
		//		}
		//	})).Return(&message.MemPoolExistRsp{}, nil)
		//	actor.On("CallRequest",message.MemPoolSvc, mock.MatchedBy(func(in *message.MemPoolExist) bool {
		//		if( bytes.Equal(in.Hash, sampleTxs[1])  ) {
		//			return false
		//		} else {
		//			return true
		//		}
		//	})).Return(nil, fmt.Errorf("not found"))
		//
		//	msgHelper.On("ExtractTxFromResponseAndError", mock.AnythingOfType("*message.MemPoolExistRsp"), nil).Return(dummyTx, nil)
		//	hashes := sampleTxs
		//	return nil, &types.GetTransactionsRequest{Hashes:hashes}
		//}, func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockRW *MockMsgReadWriter) {
		//	actor.AssertNumberOfCalls(tt,"CallRequest",len(sampleTxs))
		//	msgHelper.AssertNumberOfCalls(tt,"ExtractTxFromResponseAndError",len(sampleTxs))
		//}},
		// 3. hash not found (all)
		// 4. actor failure
		{"TActorError",func(pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockRW *MockMsgReadWriter) (*types.MsgHeader,*types.GetTransactionsRequest){
			//dummyTx := &types.Tx{Hash:nil}
			actor.On("CallRequest",message.MemPoolSvc, mock.AnythingOfType("*message.MemPoolExist")).Return(nil, fmt.Errorf("timeout"))
			//msgHelper.On("ExtractTxFromResponseAndError", nil, mock.AnythingOfType("error")).Return(nil, fmt.Errorf("error"))
			msgHelper.On("ExtractTxFromResponseAndError",  nil, mock.MatchedBy(func(err error) bool {
				if err != nil {
					return true
				}
				return false})).Return(nil, fmt.Errorf("error"))
			//msgHelper.On("ExtractTxFromResponseAndError", mock.AnythingOfType("*message.MemPoolExistRsp"), nil).Return(dummyTx, nil)
			hashes := sampleTxs
			return nil, &types.GetTransactionsRequest{Hashes:hashes}
		}, func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockRW *MockMsgReadWriter) {
			// break at first eval
			actor.AssertNumberOfCalls(tt,"CallRequest",1)
			msgHelper.AssertNumberOfCalls(tt,"ExtractTxFromResponseAndError",1)
			// TODO need check that error response was sent
		}},

		// 5. invalid parameter (no input hash, or etc.)
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockActor := new(MockActorService)
			mockRW := new(MockMsgReadWriter)
			dummyPeer := newRemotePeer(dummyMeta, mockPM, mockActor, logger, mockSigner, mockRW)
			mockMsgHelper := new(mocks.Helper)

			header, body := test.setup(mockPM, mockActor, mockMsgHelper, mockRW)
			target := newTxReqHandler(mockPM, dummyPeer, logger, mockSigner)
			target.msgHelper = mockMsgHelper

			target.handle(header, body)

			test.verify(t, mockPM, mockActor, mockMsgHelper, mockRW)
		})
	}
}

func TestTxRequestHandler_handle1(t *testing.T) {
	logger := log.NewLogger("test.p2p")
	dummyMeta := PeerMeta{ID:dummyPeerID,IPAddress:"192.168.1.2",Port:4321}
	dummyTx := &types.Tx{Hash:nil}
	hashes := make([][]byte,0,0)
	hashes = append(hashes, dummyTxHash)
	dummyReq := &types.GetTransactionsRequest{Hashes:hashes}

	mockPM := new(MockPeerManager)
	mockActor := new(MockActorService)
	dummyRW := new(MockMsgReadWriter)
	mockSigner := new(mockMsgSigner)
	mockSigner.On("signMsg",mock.Anything).Return(nil)

	dummyPeer := newRemotePeer(dummyMeta, mockPM, mockActor, logger, mockSigner, dummyRW)
	mockMsgHelper := new(mocks.Helper)

	target := newTxReqHandler(mockPM, dummyPeer, logger, mockSigner)
	target.msgHelper = mockMsgHelper
	// 1. success case (single tx)
	mockActor.On("CallRequest",message.MemPoolSvc, mock.AnythingOfType("*message.MemPoolExist")).Return(&message.MemPoolExistRsp{}, nil)
	mockMsgHelper.On("ExtractTxFromResponseAndError", mock.AnythingOfType("*message.MemPoolExistRsp"), nil).Return(dummyTx, nil)
	target.handle(nil, dummyReq)
	mockActor.AssertNumberOfCalls(t,"CallRequest",1)
	mockMsgHelper.AssertNumberOfCalls(t,"ExtractTxFromResponseAndError",1)

	// 1-1 success case2 (multiple tx)
	// 2. hash not found (partial)
	// 3. hash not found (all)
	// 4. actor failure
	// 5. invalid parameter (no input hash, or etc.)
}

func BenchmarkArrayKey(b *testing.B) {
	size := 100000
	var samples = make([]([hashSize]byte), size)
	for i := 0; i < size; i++ {
		copy(samples[i][:], uuid.Must(uuid.NewV4()).Bytes())
		copy(samples[i][16:], uuid.Must(uuid.NewV4()).Bytes())
	}

	b.Run("BArray", func(b *testing.B) {
		var keyArr [hashSize]byte
		startTime := time.Now()
		fmt.Printf("P1 in byte array\n")
		target := make(map[[hashSize]byte]int)
		for i := 0; i < size; i++ {
			copy(keyArr[:], samples[i][:])
			target[keyArr] = i
		}
		endTime := time.Now()
		fmt.Printf("Takes %f sec \n", endTime.Sub(startTime).Seconds())
	})

	b.Run("Bbase64", func(b *testing.B) {
		startTime := time.Now()
		fmt.Printf("P2 in base64\n")
		target2 := make(map[string]int)
		for i := 0; i < size; i++ {
			target2[enc.ToString(samples[i][:])] = i
		}
		endTime := time.Now()
		fmt.Printf("Takes %f sec\n", endTime.Sub(startTime).Seconds())

	})

}

func Test_bytesArrToString(t *testing.T) {
	t.SkipNow()
	type args struct {
		bbarray [][]byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "TSucc-01", args: args{[][]byte{[]byte("abcde"), []byte("12345")}}, want: "[\"YWJjZGU=\",\"MTIzNDU=\",]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bytesArrToString(tt.args.bbarray); got != tt.want {
				t.Errorf("bytesArrToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
