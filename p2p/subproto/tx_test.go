/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */
package subproto

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/message/messagemock"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2putil"

	"github.com/gofrs/uuid"
)

func TestTxRequestHandler_handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//var dummyPeerID, _ = peer.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")
	var sampleMsgID = p2pcommon.NewMsgID()
	var sampleHeader = p2pmock.NewMockMessage(ctrl)
	sampleHeader.EXPECT().ID().Return(sampleMsgID).AnyTimes()
	sampleHeader.EXPECT().Subprotocol().Return(GetTXsResponse).AnyTimes()

	var sampleTxsB58 = []string{
		"4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45",
		"6xfk39kuyDST7NwCu8tx3wqwFZ5dwKPDjxUS14tU7NZb8",
		"E8dbBGe9Hnuhk35cJoekPjL3VoL4xAxtnRuP47UoxzHd",
		"HB7Hg5GUbHuxwe8Lp5PcYUoAaQ7EZjRNG6RuvS6DnDRf",
		"BxKmDg9VbWHxrWnStEeTzJ2Ze7RF7YK4rpyjcsWSsnxs",
		"DwmGqFU4WgADpYN36FXKsYxMjeppvh9Najg4KxJ8gtX3",
	}
	var sampleTxs = make([][]byte, len(sampleTxsB58))
	var sampleTxHashes = make([]types.TxID, len(sampleTxsB58))
	for i, hashb58 := range sampleTxsB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleTxs[i] = hash
		copy(sampleTxHashes[i][:], hash)
	}

	//dummyPeerID2, _ = peer.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	//dummyPeerID3, _ = peer.IDB58Decode("16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD")

	logger := log.NewLogger("test.subproto")
	//dummyMeta := p2pcommon.PeerMeta{ID: dummyPeerID, IPAddress: "192.168.1.2", Port: 4321}
	mockMo := p2pmock.NewMockMsgOrder(ctrl)
	mockMo.EXPECT().GetProtocolID().Return(GetTXsResponse).AnyTimes()
	mockMo.EXPECT().GetMsgID().Return(sampleMsgID).AnyTimes()
	//mockSigner := p2pmock.NewmockMsgSigner(ctrl)
	//mockSigner.EXPECT().signMsg",gomock.Any()).Return(nil)
	tests := []struct {
		name   string
		setup  func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) (p2pcommon.Message, *types.GetTransactionsRequest)
		verify func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter)
	}{
		// 1. success case (single tx)
		{"TSucc1", func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) (p2pcommon.Message, *types.GetTransactionsRequest) {
			// receive request for one tx , query to mempool get send response to remote peer
			dummyTxs := make([]*types.Tx, 1)
			dummyTxs[0] = &types.Tx{Hash: sampleTxs[0]}
			actor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(&message.MemPoolExistExRsp{Txs: dummyTxs}, nil).Times(1)
			msgHelper.EXPECT().ExtractTxsFromResponseAndError(gomock.AssignableToTypeOf(&message.MemPoolExistExRsp{}), nil).Return(dummyTxs, nil).Times(1)
			hashes := sampleTxs[:1]
			mockMF.EXPECT().NewMsgResponseOrder(sampleMsgID, GetTXsResponse, gomock.AssignableToTypeOf(&types.GetTransactionsResponse{})).Do(func(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) {
				resp := message.(*types.GetTransactionsResponse)
				assert.Equal(tt, types.ResultStatus_OK, resp.Status)
				assert.Equal(tt, 1, len(resp.Hashes))
				assert.Equal(tt, sampleTxs[0], resp.Hashes[0])
			}).Return(mockMo).Times(1)
			return sampleHeader, &types.GetTransactionsRequest{Hashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) {
			// verification is defined in setup
		}},
		// 1-1 success case2 (multiple tx)
		{"TSucc2", func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) (p2pcommon.Message, *types.GetTransactionsRequest) {
			dummyTxs := make([]*types.Tx, len(sampleTxs))
			for i, txHash := range sampleTxs {
				dummyTxs[i] = &types.Tx{Hash: txHash}
			}
			actor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(&message.MemPoolExistExRsp{Txs: dummyTxs}, nil).Times(1)
			msgHelper.EXPECT().ExtractTxsFromResponseAndError(gomock.AssignableToTypeOf(&message.MemPoolExistExRsp{}), nil).Return(dummyTxs, nil).Times(1)
			hashes := sampleTxs
			mockMF.EXPECT().NewMsgResponseOrder(sampleMsgID, GetTXsResponse, gomock.AssignableToTypeOf(&types.GetTransactionsResponse{})).Do(func(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) {
				resp := message.(*types.GetTransactionsResponse)
				assert.Equal(tt, types.ResultStatus_OK, resp.Status)
				assert.Equal(tt, len(sampleTxs), len(resp.Hashes))
			}).Return(mockMo).Times(1)
			return sampleHeader, &types.GetTransactionsRequest{Hashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) {
		}},
		// 2. hash not found (partial)
		{"TPartialExist", func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) (p2pcommon.Message, *types.GetTransactionsRequest) {
			dummyTxs := make([]*types.Tx, 0, len(sampleTxs))
			hashes := make([][]byte, 0, len(sampleTxs))
			for i, txHash := range sampleTxs {
				if i%2 == 0 {
					dummyTxs = append(dummyTxs, &types.Tx{Hash: txHash})
					hashes = append(hashes, txHash)
				}
			}
			actor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(&message.MemPoolExistExRsp{Txs: dummyTxs}, nil).Times(1)
			msgHelper.EXPECT().ExtractTxsFromResponseAndError(gomock.AssignableToTypeOf(&message.MemPoolExistExRsp{}), nil).Return(dummyTxs, nil).Times(1)
			mockMF.EXPECT().NewMsgResponseOrder(sampleMsgID, GetTXsResponse, gomock.AssignableToTypeOf(&types.GetTransactionsResponse{})).Do(func(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) {
				resp := message.(*types.GetTransactionsResponse)
				assert.Equal(tt, types.ResultStatus_OK, resp.Status)
				assert.Equal(tt, len(dummyTxs), len(resp.Hashes))
			}).Return(mockMo).Times(1)
			return sampleHeader, &types.GetTransactionsRequest{Hashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) {
		}},
		// 3. hash not found (all)
		{"TNoExist", func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) (p2pcommon.Message, *types.GetTransactionsRequest) {
			//dummyTx := &types.Tx{Hash:nil}
			// emulate second tx is not exists.
			actor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(&message.MemPoolExistExRsp{}, nil).Times(1)
			//msgHelper.EXPECT().ExtractTxsFromResponseAndError", mock.MatchedBy(func(m *message.MemPoolExistExRsp) bool {
			//	if len(m.Txs) == 0 {
			//		return false
			//	}
			//	return true
			//}), nil).Return(dummyTx, nil)
			msgHelper.EXPECT().ExtractTxsFromResponseAndError(&MempoolRspTxCountMatcher{0}, nil).Return(nil, nil).Times(1)
			hashes := sampleTxs
			mockMF.EXPECT().NewMsgResponseOrder(sampleMsgID, GetTXsResponse, gomock.AssignableToTypeOf(&types.GetTransactionsResponse{})).Do(func(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) {
				resp := message.(*types.GetTransactionsResponse)
				assert.Equal(tt, types.ResultStatus_NOT_FOUND, resp.Status)
				assert.Equal(tt, 0, len(resp.Hashes))
			}).Return(mockMo).Times(1)
			return sampleHeader, &types.GetTransactionsRequest{Hashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) {
		}},
		// 4. actor failure
		{"TActorError", func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) (p2pcommon.Message, *types.GetTransactionsRequest) {
			//dummyTx := &types.Tx{Hash:nil}
			actor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(nil, fmt.Errorf("timeout")).Times(1)
			//msgHelper.EXPECT().ExtractTxsFromResponseAndError", nil, gomock.AssignableToTypeOf("error")).Return(nil, fmt.Errorf("error"))
			msgHelper.EXPECT().ExtractTxsFromResponseAndError(nil, &WantErrMatcher{true}).Return(nil, fmt.Errorf("error")).Times(0)
			hashes := sampleTxs
			mockMF.EXPECT().NewMsgResponseOrder(sampleMsgID, GetTXsResponse, gomock.AssignableToTypeOf(&types.GetTransactionsResponse{})).Do(func(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) {
				resp := message.(*types.GetTransactionsResponse)
				// TODO check if the changed behavior is fair or not.
				assert.Equal(tt, types.ResultStatus_NOT_FOUND, resp.Status)
			}).Return(mockMo).Times(1)
			return sampleHeader, &types.GetTransactionsRequest{Hashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) {
			// break at first eval
			// TODO need check that error response was sent
		}},

		// 5. invalid parameter (no input hash, or etc.)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMsgHelper := messagemock.NewHelper(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().Name().Return("mockPeer").AnyTimes()
			mockPeer.EXPECT().MF().Return(mockMF).AnyTimes()
			mockPeer.EXPECT().SendMessage(mockMo)

			header, body := test.setup(t, mockPM, mockActor, mockMsgHelper, mockMF, mockRW)
			target := NewTxReqHandler(mockPM, mockPeer, logger, mockActor)
			target.msgHelper = mockMsgHelper

			target.Handle(header, body)

			test.verify(t, mockPM, mockActor, mockMsgHelper, mockMF, mockRW)
		})
	}
}

func TestTxRequestHandler_handleBySize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.subproto")
	var dummyPeerID, _ = peer.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")
	var dummyTxHash, _ = enc.ToBytes("4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45")

	bigTxBytes := make([]byte, 2*1024*1024)
	//dummyMO := p2pmock.NewMockMsgOrder(ctrl)
	tests := []struct {
		name              string
		hashCnt           int
		validCallCount    int
		expectedSendCount int
	}{
		{"TSingle", 1, 1, 1},
		{"TNotFounds", 100, 0, 1},
		{"TFound10", 10, 10, 4},
		{"TFoundAll", 20, 100, 7},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockMF := &testDoubleMOFactory{}
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer.EXPECT().MF().Return(mockMF).AnyTimes()
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().Name().Return("16..aadecf@1").AnyTimes()
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(test.expectedSendCount)

			validBigMempoolRsp := &message.MemPoolExistExRsp{}
			txs := make([]*types.Tx, 0, test.hashCnt)
			for i := 0; i < test.hashCnt; i++ {
				if i >= test.validCallCount {
					break
				}
				txs = append(txs, &types.Tx{Hash: dummyTxHash, Body: &types.TxBody{Payload: bigTxBytes}})
			}
			validBigMempoolRsp.Txs = txs

			mockActor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(validBigMempoolRsp, nil)

			h := NewTxReqHandler(mockPM, mockPeer, logger, mockActor)
			dummyMsg := &testMessage{subProtocol:GetTXsRequest, id:p2pcommon.NewMsgID()}
			msgBody := &types.GetTransactionsRequest{Hashes: make([][]byte, test.hashCnt)}
			h.Handle(dummyMsg, msgBody)

		})
	}
}

func TestNewTxNoticeHandler_handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.subproto")
	var dummyPeerID, _ = peer.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")
	var dummyTxHash, _ = enc.ToBytes("4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45")
	sampleMeta := p2pcommon.PeerMeta{ID: dummyPeerID, IPAddress: "192.168.1.2", Port: 4321}
	sampleHeader := &testMessage{id:p2pcommon.NewMsgID()}

	var filledArrs = make([]types.TxID, 1)
	filledArrs[0] = types.ToTxID(dummyTxHash)
	var emptyArrs = make([]types.TxID, 0)

	var sampleTxsB58 = []string{
		"4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45",
		"6xfk39kuyDST7NwCu8tx3wqwFZ5dwKPDjxUS14tU7NZb8",
		"E8dbBGe9Hnuhk35cJoekPjL3VoL4xAxtnRuP47UoxzHd",
		"HB7Hg5GUbHuxwe8Lp5PcYUoAaQ7EZjRNG6RuvS6DnDRf",
		"BxKmDg9VbWHxrWnStEeTzJ2Ze7RF7YK4rpyjcsWSsnxs",
		"DwmGqFU4WgADpYN36FXKsYxMjeppvh9Najg4KxJ8gtX3",
	}
	sampleTxs := make([][]byte, len(sampleTxsB58))
	sampleTxHashes := make([]types.TxID, len(sampleTxsB58))
	for i, hashb58 := range sampleTxsB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleTxs[i] = hash
		copy(sampleTxHashes[i][:], hash)
	}

	tests := []struct {
		name string
		//hashes [][]byte
		//calledUpdataCache bool
		//passedToSM bool
		setup  func(tt *testing.T, pm *p2pmock.MockPeerManager, mockPeer *p2pmock.MockRemotePeer, mockSM *p2pmock.MockSyncManager) (p2pcommon.Message, *types.NewTransactionsNotice)
		verify func(tt *testing.T, pm *p2pmock.MockPeerManager, mockPeer *p2pmock.MockRemotePeer, mockSM *p2pmock.MockSyncManager)
	}{
		// 1. success case (single tx)
		{"TSuccSingle", func(tt *testing.T, pm *p2pmock.MockPeerManager, mockPeer *p2pmock.MockRemotePeer, mockSM *p2pmock.MockSyncManager) (p2pcommon.Message, *types.NewTransactionsNotice) {
			hashes := sampleTxs[:1]
			mockPeer.EXPECT().UpdateTxCache(&TxIDCntMatcher{1}).Return(filledArrs).MinTimes(1)
			//mockPeer.EXPECT().UpdateTxCache(gomock.Any()).Return(filledArrs).AnyTimes()
			mockSM.EXPECT().HandleNewTxNotice(mockPeer, filledArrs, gomock.AssignableToTypeOf(&types.NewTransactionsNotice{})).MinTimes(1)
			return sampleHeader, &types.NewTransactionsNotice{TxHashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, mockPeer *p2pmock.MockRemotePeer, mockSM *p2pmock.MockSyncManager) {
		}},
		// 1-1 success case2 (multiple tx)
		{"TSuccMultiHash", func(tt *testing.T, pm *p2pmock.MockPeerManager, mockPeer *p2pmock.MockRemotePeer, mockSM *p2pmock.MockSyncManager) (p2pcommon.Message, *types.NewTransactionsNotice) {
			hashes := sampleTxs
			mockPeer.EXPECT().UpdateTxCache(&TxIDCntMatcher{len(sampleTxs)}).Return(filledArrs).MinTimes(1)
			//mockPeer.EXPECT().UpdateTxCache(gomock.Any()).Return(filledArrs)
			mockSM.EXPECT().HandleNewTxNotice(gomock.Any(), filledArrs, gomock.AssignableToTypeOf(&types.NewTransactionsNotice{})).MinTimes(1)
			return sampleHeader, &types.NewTransactionsNotice{TxHashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, mockPeer *p2pmock.MockRemotePeer, mockSM *p2pmock.MockSyncManager) {
		}},
		//// 2. All hashes already exist
		{"TSuccAlreadyExists", func(tt *testing.T, pm *p2pmock.MockPeerManager, mockPeer *p2pmock.MockRemotePeer, mockSM *p2pmock.MockSyncManager) (p2pcommon.Message, *types.NewTransactionsNotice) {
			hashes := sampleTxs
			mockPeer.EXPECT().UpdateTxCache(gomock.Any()).Return(emptyArrs)
			mockSM.EXPECT().HandleNewTxNotice(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			return sampleHeader, &types.NewTransactionsNotice{TxHashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, mockPeer *p2pmock.MockRemotePeer, mockSM *p2pmock.MockSyncManager) {
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().Meta().Return(sampleMeta).AnyTimes()
			mockPeer.EXPECT().ID().Return(sampleMeta.ID).AnyTimes()
			mockPeer.EXPECT().Name().Return(p2putil.ShortForm(sampleMeta.ID) + "#1").AnyTimes()
			mockSM := p2pmock.NewMockSyncManager(ctrl)

			header, body := test.setup(t, mockPM, mockPeer, mockSM)

			target := NewNewTxNoticeHandler(mockPM, mockPeer, logger, mockActor, mockSM)
			target.Handle(header, body)

			test.verify(t, mockPM, mockPeer, mockSM)
		})
	}
}


func BenchmarkArrayKey(b *testing.B) {
	size := 100000
	const hashSize = 32
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
			if got := p2putil.BytesArrToString(tt.args.bbarray); got != tt.want {
				t.Errorf("BytesArrToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MempoolRspTxCountMatcher struct {
	matchCnt int
}

func (tcm MempoolRspTxCountMatcher) Matches(x interface{}) bool {
	m, ok := x.(*message.MemPoolExistExRsp)
	if !ok {
		return false
	}
	return tcm.matchCnt == len(m.Txs)
}

func (tcm MempoolRspTxCountMatcher) String() string {
	return fmt.Sprintf("tx count = %d",tcm.matchCnt)
}

type TxIDCntMatcher struct {
	matchCnt int
}

func (scm TxIDCntMatcher) Matches(x interface{}) bool {
	m, ok := x.([]types.TxID)
	if !ok {
		return false
	}
	return scm.matchCnt == len(m)
}

func (scm TxIDCntMatcher) String() string {
	return fmt.Sprintf("len(slice) = %d",scm.matchCnt)
}

type WantErrMatcher struct {
	wantErr bool
}

func (tcm WantErrMatcher) Matches(x interface{}) bool {
	m, ok := x.(*error)
	if !ok {
		return false
	}
	return tcm.wantErr == (m != nil)
}

func (tcm WantErrMatcher) String() string {
	return fmt.Sprintf("want error = %v",tcm.wantErr)
}

