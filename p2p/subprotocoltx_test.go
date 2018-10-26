/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */
package p2p

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/message/mocks"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"

	"github.com/aergoio/aergo/internal/enc"

	"github.com/satori/go.uuid"
)

var sampleMsgID MsgID
var sampleHeader Message

func init() {
	sampleMsgID = NewMsgID()
	sampleHeader = &V030Message{id:sampleMsgID}
}

func TestTxRequestHandler_handle(t *testing.T) {
	logger := log.NewLogger("test.p2p")
	dummyMeta := PeerMeta{ID:dummyPeerID,IPAddress:"192.168.1.2",Port:4321}
	mockMo := new(MockMsgOrder)
	mockMo.On("GetProtocolID").Return(GetTxsResponse)
	mockMo.On("GetMsgID").Return(sampleMsgID)
	//mockSigner := new(mockMsgSigner)
	//mockSigner.On("signMsg",mock.Anything).Return(nil)
	tests := []struct {
		name string
		setup func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockMF *MockMoFactory, mockRW *MockMsgReadWriter) (Message,*types.GetTransactionsRequest)
		verify func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockMF *MockMoFactory, mockRW *MockMsgReadWriter)
	}{
		// 1. success case (single tx)
		{"TSucc1",func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockMF *MockMoFactory, mockRW *MockMsgReadWriter) (Message,*types.GetTransactionsRequest){
			dummyTx := &types.Tx{Hash:nil}
			actor.On("CallRequestDefaultTimeout",message.MemPoolSvc, mock.AnythingOfType("*message.MemPoolExist")).Return(&message.MemPoolExistRsp{Tx:dummyTx}, nil)
			msgHelper.On("ExtractTxFromResponseAndError", mock.AnythingOfType("*message.MemPoolExistRsp"), nil).Return(dummyTx, nil)
			hashes := sampleTxs[:1]
			mockMF.On("newMsgResponseOrder",sampleMsgID,GetTxsResponse, mock.AnythingOfType("*types.GetTransactionsResponse")).Run(func(args mock.Arguments) {
				resp := args[2].(*types.GetTransactionsResponse)
				assert.Equal(tt, types.ResultStatus_OK, resp.Status)
				assert.Equal(tt, 1, len(resp.Hashes))
				assert.Equal(tt, sampleTxs[0], resp.Hashes[0])
			}).Return(mockMo)
			return sampleHeader, &types.GetTransactionsRequest{Hashes:hashes}
		}, func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockMF *MockMoFactory, mockRW *MockMsgReadWriter) {
			actor.AssertNumberOfCalls(tt,"CallRequestDefaultTimeout",1)
			msgHelper.AssertNumberOfCalls(tt,"ExtractTxFromResponseAndError",1)
			mockMF.AssertNumberOfCalls(tt, "newMsgResponseOrder", 1)
		}},
		// 1-1 success case2 (multiple tx)
		{"TSucc2",func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockMF *MockMoFactory, mockRW *MockMsgReadWriter) (Message,*types.GetTransactionsRequest){
			dummyTx := &types.Tx{Hash:nil}
			actor.On("CallRequestDefaultTimeout",message.MemPoolSvc, mock.AnythingOfType("*message.MemPoolExist")).Return(&message.MemPoolExistRsp{Tx:dummyTx}, nil)
			msgHelper.On("ExtractTxFromResponseAndError", mock.AnythingOfType("*message.MemPoolExistRsp"), nil).Return(dummyTx, nil)
			hashes := sampleTxs
			mockMF.On("newMsgResponseOrder",sampleMsgID,GetTxsResponse, mock.AnythingOfType("*types.GetTransactionsResponse")).Run(func(args mock.Arguments) {
				resp := args[2].(*types.GetTransactionsResponse)
				assert.Equal(tt, types.ResultStatus_OK, resp.Status)
				assert.Equal(tt, len(sampleTxs), len(resp.Hashes))

			}).Return(mockMo)
			return sampleHeader, &types.GetTransactionsRequest{Hashes:hashes}
		}, func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockMF *MockMoFactory, mockRW *MockMsgReadWriter) {
			actor.AssertNumberOfCalls(tt,"CallRequestDefaultTimeout",len(sampleTxs))
			msgHelper.AssertNumberOfCalls(tt,"ExtractTxFromResponseAndError",len(sampleTxs))
			mockMF.AssertNumberOfCalls(tt, "newMsgResponseOrder", 1)
		}},
		// 2. hash not found (partial)
		{"TPartialExist",func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockMF *MockMoFactory, mockRW *MockMsgReadWriter) (Message,*types.GetTransactionsRequest){
			dummyTx := &types.Tx{Hash:nil}
			// emulate second tx is not exists.
			actor.On("CallRequestDefaultTimeout",message.MemPoolSvc, mock.MatchedBy(func(m *message.MemPoolExist) bool {
				if bytes.Equal(m.Hash,sampleTxs[1]) {
					return false
				}
				return true
			})).Return(&message.MemPoolExistRsp{Tx:dummyTx}, nil)
			actor.On("CallRequestDefaultTimeout",message.MemPoolSvc, mock.MatchedBy(func(m *message.MemPoolExist) bool {
				if bytes.Equal(m.Hash,sampleTxs[1]) {
					return true
				}
				return false
			})).Return(&message.MemPoolExistRsp{Tx:nil}, nil)
			msgHelper.On("ExtractTxFromResponseAndError", mock.MatchedBy(func(m *message.MemPoolExistRsp) bool {
				if m.Tx == nil {
					return false
				}
				return true
			}), nil).Return(dummyTx, nil)
			msgHelper.On("ExtractTxFromResponseAndError", mock.MatchedBy(func(m *message.MemPoolExistRsp) bool {
				if m.Tx == nil {
					return true
				}
				return false
				}), nil).Return(nil, nil)
			hashes := sampleTxs
			mockMF.On("newMsgResponseOrder",sampleMsgID,GetTxsResponse, mock.AnythingOfType("*types.GetTransactionsResponse")).Run(func(args mock.Arguments) {
					resp := args[2].(*types.GetTransactionsResponse)
					assert.Equal(tt, types.ResultStatus_OK, resp.Status)
					assert.Equal(tt, len(sampleTxs)-1, len(resp.Hashes))
				}).Return(mockMo)
			return sampleHeader, &types.GetTransactionsRequest{Hashes:hashes}
		}, func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockMF *MockMoFactory, mockRW *MockMsgReadWriter) {
			actor.AssertNumberOfCalls(tt,"CallRequestDefaultTimeout",len(sampleTxs))
			msgHelper.AssertNumberOfCalls(tt,"ExtractTxFromResponseAndError",len(sampleTxs))
			mockMF.AssertNumberOfCalls(tt, "newMsgResponseOrder", 1)
		}},
		// 3. hash not found (all)
		{"TNoExist",func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockMF *MockMoFactory, mockRW *MockMsgReadWriter) (Message,*types.GetTransactionsRequest){
			dummyTx := &types.Tx{Hash:nil}
			// emulate second tx is not exists.
			actor.On("CallRequestDefaultTimeout",message.MemPoolSvc, mock.AnythingOfType("*message.MemPoolExist")).Return(&message.MemPoolExistRsp{}, nil)
			msgHelper.On("ExtractTxFromResponseAndError", mock.MatchedBy(func(m *message.MemPoolExistRsp) bool {
				if m.Tx == nil {
					return false
				}
				return true
			}), nil).Return(dummyTx, nil)
			msgHelper.On("ExtractTxFromResponseAndError", mock.MatchedBy(func(m *message.MemPoolExistRsp) bool {
				if m.Tx == nil {
					return true
				}
				return false
			}), nil).Return(nil, nil)
			hashes := sampleTxs
			mockMF.On("newMsgResponseOrder",sampleMsgID,GetTxsResponse, mock.AnythingOfType("*types.GetTransactionsResponse")).Run(func(args mock.Arguments) {
				resp := args[2].(*types.GetTransactionsResponse)
				assert.Equal(tt, types.ResultStatus_NOT_FOUND, resp.Status)
				assert.Equal(tt, 0, len(resp.Hashes))
			}).Return(mockMo)
			return sampleHeader, &types.GetTransactionsRequest{Hashes:hashes}
		}, func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockMF *MockMoFactory, mockRW *MockMsgReadWriter) {
			actor.AssertNumberOfCalls(tt,"CallRequestDefaultTimeout",len(sampleTxs))
			msgHelper.AssertNumberOfCalls(tt,"ExtractTxFromResponseAndError",len(sampleTxs))
			mockMF.AssertNumberOfCalls(tt, "newMsgResponseOrder", 1)
		}},
		// 4. actor failure
		{"TActorError",func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockMF *MockMoFactory, mockRW *MockMsgReadWriter) (Message,*types.GetTransactionsRequest){
			//dummyTx := &types.Tx{Hash:nil}
			actor.On("CallRequestDefaultTimeout",message.MemPoolSvc, mock.AnythingOfType("*message.MemPoolExist")).Return(nil, fmt.Errorf("timeout"))
			//msgHelper.On("ExtractTxFromResponseAndError", nil, mock.AnythingOfType("error")).Return(nil, fmt.Errorf("error"))
			msgHelper.On("ExtractTxFromResponseAndError",  nil, mock.MatchedBy(func(err error) bool {
				if err != nil {
					return true
				}
				return false})).Return(nil, fmt.Errorf("error"))
			//msgHelper.On("ExtractTxFromResponseAndError", mock.AnythingOfType("*message.MemPoolExistRsp"), nil).Return(dummyTx, nil)
			hashes := sampleTxs
			mockMF.On("newMsgResponseOrder",sampleMsgID,GetTxsResponse, mock.AnythingOfType("*types.GetTransactionsResponse")).Run(func(args mock.Arguments) {
				resp := args[2].(*types.GetTransactionsResponse)
				assert.Equal(tt, types.ResultStatus_INTERNAL, resp.Status)
			}).Return(mockMo)
			return sampleHeader, &types.GetTransactionsRequest{Hashes:hashes}
		}, func(tt *testing.T, pm *MockPeerManager, actor *MockActorService, msgHelper *mocks.Helper, mockMF *MockMoFactory, mockRW *MockMsgReadWriter) {
			// break at first eval
			actor.AssertNumberOfCalls(tt,"CallRequestDefaultTimeout",1)
			msgHelper.AssertNumberOfCalls(tt,"ExtractTxFromResponseAndError",1)
			// TODO need check that error response was sent
		}},

		// 5. invalid parameter (no input hash, or etc.)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockActor := new(MockActorService)
			mockRW := new(MockMsgReadWriter)
			mockMF := new(MockMoFactory)
			dummyPeer := newRemotePeer(dummyMeta, mockPM, mockActor, logger, mockMF, &dummySigner{}, mockRW)
			mockMsgHelper := new(mocks.Helper)

			header, body := test.setup(t, mockPM, mockActor, mockMsgHelper, mockMF, mockRW)
			target := newTxReqHandler(mockPM, dummyPeer, logger, mockActor)
			target.msgHelper = mockMsgHelper

			target.handle(header, body)

			test.verify(t, mockPM, mockActor, mockMsgHelper, mockMF, mockRW)
		})
	}
}


func TestTxRequestHandler_handleBySize(t *testing.T) {
	bigTxBytes := make([]byte,2*1024*1024)
	logger := log.NewLogger("test")
	//validSmallBlockRsp := &message.GetBlockRsp{Block:&types.Block{Hash:make([]byte,40)},Err:nil}
	validBigMempoolRsp := &message.MemPoolExistRsp{Tx: &types.Tx{Hash: dummyTxHash,Body:&types.TxBody{Payload: bigTxBytes}}}
	notExistMempoolRsp := &message.MemPoolExistRsp{Tx:nil}
	//dummyMO := new(MockMsgOrder)
	tests := []struct {
		name string
		hashCnt int
		validCallCount int
		expectedSendCount int
	}{
		{"TSingle", 1, 1, 1},
		{"TNotFounds", 100, 0, 1},
		{"TFound10", 10, 10, 4},
		{"TFoundAll", 20, 100, 7},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockMF := &v030MOFactory{}
			mockPeer := new(MockRemotePeer)
			mockActor := new(MockActorService)
			mockPeer.On("MF").Return(mockMF)
			mockPeer.On("ID").Return(dummyPeerID)
			mockPeer.On("sendMessage", mock.Anything)
			callReqCount :=0
			mockActor.On("CallRequestDefaultTimeout",message.MemPoolSvc, mock.MatchedBy(func(arg *message.MemPoolExist) bool{
				callReqCount++
				if callReqCount <= test.validCallCount {
					return true
				}
				return false
			})).Return(validBigMempoolRsp, nil)
			mockActor.On("CallRequestDefaultTimeout",message.MemPoolSvc, mock.MatchedBy(func(arg *message.MemPoolExist) bool{
				callReqCount++
				if callReqCount <= test.validCallCount {
					return false
				}
				return true
			})).Return(notExistMempoolRsp, nil)

			h:=newTxReqHandler(mockPM, mockPeer, logger, mockActor)
			dummyMsg := &V030Message{id:NewMsgID()}
			msgBody := &types.GetTransactionsRequest{Hashes:make([][]byte, test.hashCnt)}
			h.handle(dummyMsg, msgBody)

			mockPeer.AssertNumberOfCalls(t, "sendMessage", test.expectedSendCount)
		})
	}
}


func TestNewTxNoticeHandler_handle(t *testing.T) {
	logger := log.NewLogger("test.p2p")
	sampleMeta := PeerMeta{ID:dummyPeerID,IPAddress:"192.168.1.2",Port:4321}
	var filledArrs []TxHash = make([]TxHash,1)
	copy(filledArrs[0][:],dummyTxHash)
	var emptyArrs []TxHash =make([]TxHash,0)

	tests := []struct {
		name string
		//hashes [][]byte
		//calledUpdataCache bool
		//passedToSM bool
		setup func(tt *testing.T, pm *MockPeerManager,mockPeer *MockRemotePeer, mockSM *MockSyncManager) (Message,*types.NewTransactionsNotice)
		verify func(tt *testing.T, pm *MockPeerManager,mockPeer *MockRemotePeer, mockSM *MockSyncManager)
	}{
		// 1. success case (single tx)
		{"TSuccSingle",func(tt *testing.T, pm *MockPeerManager,mockPeer *MockRemotePeer, mockSM *MockSyncManager) (Message,*types.NewTransactionsNotice){
			hashes := sampleTxs[:1]
			mockPeer.On("updateTxCache",mock.Anything).Return(filledArrs)
			mockSM.On("HandleNewTxNotice",mock.Anything,mock.Anything, mock.AnythingOfType("*types.NewTransactionsNotice"))
			return sampleHeader, &types.NewTransactionsNotice{TxHashes:hashes}
		}, func(tt *testing.T, pm *MockPeerManager,mockPeer *MockRemotePeer, mockSM *MockSyncManager) {
			mockPeer.AssertCalled(t, "updateTxCache", mock.MatchedBy(func(arg []TxHash) bool {
				return len(arg) == 1
			}))
			mockSM.AssertCalled(t, "HandleNewTxNotice", mockPeer, filledArrs, mock.Anything)
		}},
		// 1-1 success case2 (multiple tx)
		{"TSuccMultiHash",func(tt *testing.T, pm *MockPeerManager,mockPeer *MockRemotePeer, mockSM *MockSyncManager) (Message,*types.NewTransactionsNotice){
			hashes := sampleTxs
			mockPeer.On("updateTxCache",mock.Anything).Return(filledArrs)
			mockSM.On("HandleNewTxNotice",mock.Anything,mock.Anything, mock.AnythingOfType("*types.NewTransactionsNotice"))
			return sampleHeader, &types.NewTransactionsNotice{TxHashes:hashes}
		}, func(tt *testing.T, pm *MockPeerManager,mockPeer *MockRemotePeer, mockSM *MockSyncManager) {
			mockPeer.AssertCalled(t, "updateTxCache", mock.MatchedBy(func(arg []TxHash) bool {
				return len(arg) == len(sampleTxs)
			}))
			mockSM.AssertCalled(t, "HandleNewTxNotice", mockPeer, filledArrs, mock.Anything)
		}},
		//// 2. All hashes already exist
		{"TSuccMultiHash",func(tt *testing.T, pm *MockPeerManager,mockPeer *MockRemotePeer, mockSM *MockSyncManager) (Message,*types.NewTransactionsNotice){
			hashes := sampleTxs
			mockPeer.On("updateTxCache",mock.Anything).Return(emptyArrs)
			mockSM.On("HandleNewTxNotice",mock.Anything,mock.Anything, mock.AnythingOfType("*types.NewTransactionsNotice"))
			return sampleHeader, &types.NewTransactionsNotice{TxHashes:hashes}
		}, func(tt *testing.T, pm *MockPeerManager,mockPeer *MockRemotePeer, mockSM *MockSyncManager) {
			mockSM.AssertNotCalled(t, "HandleNewTxNotice", mockPeer, filledArrs, mock.Anything)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockActor := new(MockActorService)
			mockPeer := new(MockRemotePeer)
			mockPeer.On("Meta").Return(sampleMeta)
			mockPeer.On("ID").Return(sampleMeta.ID)
			mockSM := new(MockSyncManager)

			header, body := test.setup(t, mockPM, mockPeer, mockSM)

			target := newNewTxNoticeHandler(mockPM, mockPeer, logger, mockActor, mockSM)
			target.handle(header, body)

			test.verify(t, mockPM, mockPeer, mockSM)
		})
	}
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
