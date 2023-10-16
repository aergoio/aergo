/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/stretchr/testify/assert"
)

//const testDuration = time.Second >> 1

// TODO refactor rw and modify this test
/*
func TestAergoPeer_RunPeer(t *testing.T) {
	//t.SkipNow()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//mockActorServ := new(p2pmock.MockActorService)
	mockActorServ := p2pmock.NewMockActorService(ctrl)
	dummyP2PServ := new(p2pmock.MockPeerManager)
	mockMF := new(p2pmock.MockMoFactory)
	dummyRW := new(p2pmock.MockMsgReadWriter)
	target := newRemotePeer(p2pcommon.PeerMeta{ID: types.PeerID("ddddd")}, 0, dummyP2PServ, mockActorServ, logger, mockMF, nil, nil, dummyRW)

	target.pingDuration = time.Second * 10
	dummyBestBlock := types.Block{Hash: []byte("testHash"), Header: &types.BlockHeader{BlockNo: 1234}}

	mockActorServ.On("requestSync", mock.Anything, mock.AnythingOfType("message.GetBlockRsp")).Return(dummyBestBlock, true)

	mockActorServ.EXPECT().
	go target.RunPeer()

	time.Sleep(testDuration)
	target.Stop()
}

func TestRemotePeer_sendPing(t *testing.T) {
	t.Skip("Send ping is not used for now, and will be reused after")
	selfPeerID, _ := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	sampleSelf := p2pcommon.PeerMeta{ID: selfPeerID, IPAddress: "192.168.1.1", Port: 6845}

	dummyBestBlockRsp := message.GetBestBlockRsp{Block: &types.Block{Header: &types.BlockHeader{}}}
	type wants struct {
		wantWrite bool
	}
	tests := []struct {
		name        string
		getBlockErr error
		wants       wants
	}{
		{"TSucc", nil, wants{wantWrite: true}},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := new(p2pmock.MockActorService)
			mockPeerManager := new(p2pmock.MockPeerManager)
			mockMF := new(p2pmock.MockMoFactory)

			mockActorServ.On("CallRequest", message.ChainSvc, mock.AnythingOfType("*message.GetBestBlock")).Return(dummyBestBlockRsp, tt.getBlockErr)
			mockPeerManager.On("SelfMeta").Return(sampleSelf)
			mockMF.On("signMsg", mock.AnythingOfType("*types.P2PMessage")).Return(nil)
			p := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, mockMF, nil, nil, nil)
			p.state.SetAndGet(types.RUNNING)

			go p.sendPing()

			time.Sleep(200 * time.Millisecond)

			actualWrite := false
			select {
			case msg := <-p.writeBuf:
				assert.Equal(t, PingRequest, msg.(msgOrder).GetProtocolID())
				actualWrite = true
			default:
			}
			assert.Equal(t, tt.wants.wantWrite, actualWrite)
			mockPeerManager.AssertNotCalled(t, "SelfMeta")
			// ping request does not contain best block information anymore.
			mockActorServ.AssertNotCalled(t, "CallRequest")
		})
	}
}

func TestRemotePeer_pruneRequests(t *testing.T) {
	tests := []struct {
		name     string
		loglevel string
	}{
		{"T1", "info"},
		//		{"T2", "debug"},
		// TODO: Add test cases.
	}
	// prevLevel := logger.Level()
	// defer logger.SetLevel(prevLevel)
	for _, tt := range tests {
		// logger.SetLevel(tt.loglevel)
		mockActorServ := new(p2pmock.MockActorService)
		mockPeerManager := new(p2pmock.MockPeerManager)
		mockStream := new(p2pmock.MockStream)
		mockStream.On("Close").Return(nil)

		p := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, nil, nil, mockStream, nil)
		t.Run(tt.name, func(t *testing.T) {
			mid1, mid2, midn := p2pcommon.NewMsgID(), p2pcommon.NewMsgID(), p2pcommon.NewMsgID()
			p.requests[mid1] = &requestInfo{cTime: time.Now().Add(time.Minute * -61), reqMO: &pbRequestOrder{pbMessageOrder{message: &MessageValue{id: mid1}}, nil}}
			p.requests[mid2] = &requestInfo{cTime: time.Now().Add(time.Minute * -60).Add(time.Second * -1), reqMO: &pbRequestOrder{pbMessageOrder{message: &MessageValue{id: mid2}}, nil}}
			p.requests[midn] = &requestInfo{cTime: time.Now().Add(time.Minute * -59), reqMO: &pbRequestOrder{pbMessageOrder{message: &MessageValue{id: midn}}, nil}}
			p.pruneRequests()

			assert.Equal(t, 1, len(p.requests))
		})
	}
}

func TestRemotePeer_sendMessage(t *testing.T) {

	type args struct {
		msgID    p2pcommon.MsgID
		protocol protocol.ID
		timeout  time.Duration
	}
	tests := []struct {
		name    string
		args    args
		timeout bool
	}{
		{"TSucc", args{p2pcommon.NewMsgID(), "p1", time.Millisecond * 100}, false},
		{"TTimeout", args{p2pcommon.NewMsgID(), "p1", time.Millisecond * 100}, true},
	}
	for _, tt := range tests {
		mockActorServ := new(p2pmock.MockActorService)
		mockPeerManager := new(p2pmock.MockPeerManager)
		mockMsg := new(p2pmock.MockMsgOrder)
		mockMsg.On("GetMsgID").Return(tt.args.msgID)
		mockMsg.On("GetProtocolID").Return(NewBlockNotice)

		writeCnt := int32(0)
		t.Run(tt.name, func(t *testing.T) {
			finishTest := make(chan interface{}, 1)
			wg := &sync.WaitGroup{}
			wg.Add(1)
			wg2 := &sync.WaitGroup{}
			wg2.Add(1)
			p := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, nil, nil, nil, nil)
			p.state.SetAndGet(types.RUNNING)

			if !tt.timeout {
				go func() {
					wg.Wait()
					for {
						select {
						case mo := <-p.writeBuf:
							p.logger.Info().Msgf("Got order from chan %v", mo)
							msg := mo.(msgOrder)
							p.logger.Info().Str(LogMsgID, msg.GetMsgID().String()).Msg("Got order")
							atomic.AddInt32(&writeCnt, 1)
							wg2.Done()
							continue
						case <-finishTest:
							return
						}
					}
				}()
			} else {
				wg2.Done()
			}
			wg.Done()
			p.sendMessage(mockMsg)
			wg2.Wait()
			if !tt.timeout {
				assert.Equal(t, int32(1), atomic.LoadInt32(&writeCnt))
			}
			finishTest <- struct{}{}
		})
	}
}

func TestRemotePeer_handleMsg(t *testing.T) {
	sampleMsgID := p2pcommon.NewMsgID()
	mockMO := new(p2pmock.MockMsgOrder)
	mockMO.On("GetMsgID").Return(sampleMsgID)
	mockMO.On("Subprotocol").Return(PingRequest)

	type args struct {
		nohandler bool
		parerr    error
		autherr   error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"TSucc", args{false, nil, nil}, false},
		{"Tnopro", args{true, nil, nil}, true},
		{"Tparcefail", args{false, fmt.Errorf("not proto"), nil}, true},
		{"Tauthfail", args{false, nil, fmt.Errorf("no permission")}, true},

		// TODO: make later use
		//{"TTimeout", args{false, nil, fmt.Errorf("no permission")}, true},
	}
	for _, tt := range tests {
		mockActorServ := new(p2pmock.MockActorService)
		mockPeerManager := new(p2pmock.MockPeerManager)
		mockMsgHandler := new(p2pmock.MockMessageHandler)
		mockSigner := new(p2pmock.MockMsgSigner)
		mockMF := new(p2pmock.MockMoFactory)
		t.Run(tt.name, func(t *testing.T) {
			msg := new(p2pmock.MockMessage)
			if tt.args.nohandler {
				msg.On("Subprotocol").Return(p2pcommon.SubProtocol(3999999999))
			} else {
				msg.On("Subprotocol").Return(PingRequest)
			}
			bodyStub := &types.Ping{}
			bytes, _ := proto.Marshal(bodyStub)
			msg.On("ID").Return(sampleMsgID)
			msg.On("Payload").Return(bytes)
			mockMsgHandler.On("parsePayload", mock.AnythingOfType("[]uint8")).Return(bodyStub, tt.args.parerr)
			mockMsgHandler.On("checkAuth", mock.Anything, mock.Anything).Return(tt.args.autherr)
			mockMsgHandler.On("handle", mock.Anything, mock.Anything)
			mockSigner.On("verifyMsg", mock.Anything, mock.Anything).Return(nil)

			target := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, mockMF, mockSigner, nil, nil)
			target.handlers[PingRequest] = mockMsgHandler

			if err := target.handleMsg(msg); (err != nil) != tt.wantErr {
				t.Errorf("remotePeerImpl.handleMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.nohandler {
				mockMsgHandler.AssertNotCalled(t, "parsePayload", mock.AnythingOfType("[]uint8"))
			} else {
				mockMsgHandler.AssertCalled(t, "parsePayload", mock.AnythingOfType("[]uint8"))
			}
			if tt.args.nohandler || tt.args.parerr != nil {
				mockMsgHandler.AssertNotCalled(t, "checkAuth", mock.Anything, mock.Anything)
			} else {
				mockMsgHandler.AssertCalled(t, "checkAuth", msg, bodyStub)
			}
			if tt.args.nohandler || tt.args.parerr != nil || tt.args.autherr != nil {
				mockMsgHandler.AssertNotCalled(t, "handle", mock.Anything, mock.Anything)
			} else {
				mockMsgHandler.AssertCalled(t, "handle", msg, bodyStub)
			}
		})
	}
}

func TestRemotePeer_sendTxNotices(t *testing.T) {
	t.Skip("meaningless after 20181030 refactoring")
	sampleSize := DefaultPeerTxQueueSize << 1
	sampleHashes := make([]types.TxID, sampleSize)
	maxTxHashSize := 100
	for i := 0; i < sampleSize; i++ {
		sampleHashes[i] = generateHash(uint64(i))
	}
	tests := []struct {
		name    string
		initCnt int
		moCnt   int
	}{
		{"TZero", 0, 0},
		{"TSingle", 1, 1},
		{"TSmall", 100, 1},
		{"TBig", 1001, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockActorServ := new(p2pmock.MockActorService)
			mockPeerManager := new(p2pmock.MockPeerManager)
			mockSigner := new(p2pmock.MockMsgSigner)
			mockMF := new(p2pmock.MockMoFactory)
			mockOrder := new(p2pmock.MockMsgOrder)
			mockOrder.On("IsNeedSign").Return(true)
			mockOrder.On("GetProtocolID").Return(NewTxNotice)
			mockOrder.On("GetMsgID").Return(p2pcommon.NewMsgID())

			mockMF.On("newMsgTxBroadcastOrder", mock.AnythingOfType("*types.NewTransactionsNotice")).Return(mockOrder)

			target := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, mockMF, mockSigner, nil, nil)
			target.maxTxNoticeHashSize = maxTxHashSize

			for i := 0; i < test.initCnt; i++ {
				target.txNoticeQueue.Press(sampleHashes[i])
			}
			target.sendTxNotices()
			mockMF.AssertNumberOfCalls(t, "newMsgTxBroadcastOrder", test.moCnt)
		})
	}
}

*/

func generateHash(i uint64) types.TxID {
	bs := types.TxID{}
	binary.LittleEndian.PutUint64(bs[:], i)
	return bs
}

func TestRemotePeerImpl_UpdateBlkCache(t *testing.T) {

	tests := []struct {
		name        string
		hash        types.BlockID
		inCache     []types.BlockID
		prevLastBlk types.BlockID
		expected    bool
	}{
		{"TAllNew", sampleBlksIDs[0], sampleBlksIDs[2:], sampleBlksIDs[2], false},
		{"TAllExist", sampleBlksIDs[0], sampleBlksIDs, sampleBlksIDs[1], true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockActor := new(p2pmock.MockActorService)
			mockPeerManager := new(p2pmock.MockPeerManager)
			mockSigner := new(p2pmock.MockMsgSigner)
			mockMF := new(p2pmock.MockMoFactory)
			sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP(sampleMeta.PrimaryAddress()), Port: sampleMeta.PrimaryPort()}
			sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn}

			target := newRemotePeer(sampleRemote, 0, mockPeerManager, mockActor, logger, mockMF, mockSigner, nil)
			for _, hash := range test.inCache {
				target.blkHashCache.Add(hash, true)
			}
			target.lastStatus = &types.LastBlockStatus{BlockHash: test.prevLastBlk[:], BlockNumber: 0, CheckTime: time.Now()}
			actual := target.UpdateBlkCache(test.hash, 0)
			assert.Equal(t, test.expected, actual)
			assert.True(t, bytes.Equal(test.hash[:], target.LastStatus().BlockHash))
		})
	}
}

func TestRemotePeerImpl_UpdateTxCache(t *testing.T) {
	tests := []struct {
		name     string
		hashes   []types.TxID
		inCache  []types.TxID
		expected []types.TxID
	}{
		{"TAllNew", sampleTxIDs, sampleTxIDs[:0], sampleTxIDs},
		{"TPartial", sampleTxIDs, sampleTxIDs[2:], sampleTxIDs[:2]},
		{"TAllExist", sampleTxIDs, sampleTxIDs, make([]types.TxID, 0)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockActor := new(p2pmock.MockActorService)
			mockPeerManager := new(p2pmock.MockPeerManager)
			mockSigner := new(p2pmock.MockMsgSigner)
			mockMF := new(p2pmock.MockMoFactory)
			sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP(sampleMeta.PrimaryAddress()), Port: sampleMeta.PrimaryPort()}
			sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn}

			target := newRemotePeer(sampleRemote, 0, mockPeerManager, mockActor, logger, mockMF, mockSigner, nil)
			for _, hash := range test.inCache {
				target.txHashCache.Add(hash, true)
			}
			actual := target.UpdateTxCache(test.hashes)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestRemotePeerImpl_GetReceiver(t *testing.T) {
	idSize := 10
	idList := make([]p2pcommon.MsgID, idSize)
	recvList := make(map[p2pcommon.MsgID]p2pcommon.ResponseReceiver)
	// first 5 have receiver, but lattes don't
	for i := 0; i < idSize; i++ {
		idList[i] = p2pcommon.NewMsgID()
		if i < 5 {
			recvList[idList[i]] = func(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) bool {
				logger.Debug().Int("seq", i).Msg("receiver called")
				return true
			}
		}
	}
	// GetReceiver should not return nil and consumeRequest must be thread-safe
	tests := []struct {
		name  string
		toAdd []p2pcommon.MsgID
		inID  p2pcommon.MsgID

		receiverReturn bool
	}{
		// 1. not anything
		{"TEmpty", idList[1:10], idList[0], true},
		// 2. have request history but no receiver
		{"TNOReceiver", idList[1:10], idList[5], false},
		// 3. have request history with receiver
		{"THave", idList[1:10], idList[1], true},
		// 4. have multiple receivers
		// TODO maybe to make separate test
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockActor := new(p2pmock.MockActorService)
			mockPeerManager := new(p2pmock.MockPeerManager)
			mockSigner := new(p2pmock.MockMsgSigner)
			mockMF := new(p2pmock.MockMoFactory)
			sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP(sampleMeta.PrimaryAddress()), Port: sampleMeta.PrimaryPort()}
			sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn}
			mockMo := p2pmock.NewMockMsgOrder(ctrl)
			mockMo.EXPECT().GetProtocolID().Return(p2pcommon.GetBlocksRequest).AnyTimes()

			p := newRemotePeer(sampleRemote, 0, mockPeerManager, mockActor, logger, mockMF, mockSigner, nil)
			for _, add := range test.toAdd {
				p.requests[add] = &requestInfo{receiver: recvList[add], reqMO: mockMo}
			}
			actual := p.GetReceiver(test.inID)
			assert.NotNil(t, actual)
			dummyMsg := p2pcommon.NewSimpleRespMsgVal(p2pcommon.AddressesResponse, p2pcommon.NewMsgID(), test.inID)
			assert.Equal(t, test.receiverReturn, actual(dummyMsg, nil))

			// after consuming request, GetReceiver always return requestIDNotFoundReceiver, which always return true
			p.ConsumeRequest(test.inID)
			actual2 := p.GetReceiver(test.inID)
			assert.NotNil(t, actual2)
			assert.Equal(t, true, actual2(dummyMsg, nil))
		})
	}
}

func TestRemotePeerImpl_pushTxsNotice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sampleSize := 100
	sampleHashes := make([]types.TxID, sampleSize)
	maxTxHashSize := 10
	for i := 0; i < sampleSize; i++ {
		sampleHashes[i] = generateHash(uint64(i))
	}
	tests := []struct {
		name       string
		in         []types.TxID
		expectSend int
	}{
		// 1. single tx
		{"TSingle", sampleHashes[:1], 0},
		// 2, multiple tx less than capacity
		{"TSmall", sampleHashes[:maxTxHashSize], 0},
		// 3. multiple tx more than capacity. last one is not sent but just queued.
		{"TLarge", sampleHashes[:maxTxHashSize*3+1], 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockMO := p2pmock.NewMockMsgOrder(ctrl)
			mockPeerManager := p2pmock.NewMockPeerManager(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockSigner := new(p2pmock.MockMsgSigner)

			mockMO.EXPECT().GetMsgID().Return(p2pcommon.NewMsgID()).AnyTimes()
			mockMF.EXPECT().NewMsgTxBroadcastOrder(gomock.Any()).Return(mockMO).
				Times(test.expectSend)

			sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP(sampleMeta.PrimaryAddress()), Port: sampleMeta.PrimaryPort()}
			sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn}

			p := newRemotePeer(sampleRemote, 0, mockPeerManager, nil, logger, mockMF, mockSigner, nil)
			p.txNoticeQueue = p2putil.NewPressableQueue(maxTxHashSize)
			p.maxTxNoticeHashSize = maxTxHashSize

			p.PushTxsNotice(test.in)
		})
	}
}
func TestRemotePeer_writeToPeer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rand := uuid.Must(uuid.NewV4())
	var sampleMsgID p2pcommon.MsgID
	copy(sampleMsgID[:], rand[:])
	type args struct {
		StreamResult error
		signErr      error
		needResponse bool
		sendErr      error
	}
	type wants struct {
		sendCnt   int
		expReqCnt int
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{"TNReq1", args{}, wants{1, 0}},
		// {"TNReqWResp1", args{needResponse: true}, wants{1, 1}},

		// error while signing
		// error while get stream
		// TODO this case is evaluated in pbMsgOrder tests
		// {"TFSend1", args{needResponse: true, sendErr: sampleErr}, wants{1, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPeerManager := p2pmock.NewMockPeerManager(ctrl)
			mockMO := p2pmock.NewMockMsgOrder(ctrl)
			mockStream := p2pmock.NewMockStream(ctrl)
			dummyRW := p2pmock.NewMockMsgReadWriter(ctrl)
			dummyRW.EXPECT().Close().AnyTimes()

			mockStream.EXPECT().Close().Return(nil).AnyTimes()
			mockMO.EXPECT().IsNeedSign().Return(true).AnyTimes()
			mockMO.EXPECT().SendTo(gomock.Any()).Return(tt.args.sendErr)
			mockMO.EXPECT().GetProtocolID().Return(p2pcommon.PingRequest).AnyTimes()
			mockMO.EXPECT().GetMsgID().Return(sampleMsgID).AnyTimes()

			sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP(sampleMeta.PrimaryAddress()), Port: sampleMeta.PrimaryPort()}
			sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn}

			p := newRemotePeer(sampleRemote, 0, mockPeerManager, nil, logger, nil, nil, dummyRW)
			p.state.SetAndGet(types.RUNNING)
			go p.runWrite()
			p.state.SetAndGet(types.RUNNING)

			p.writeToPeer(mockMO)

			// FIXME wait in more reliable way
			time.Sleep(50 * time.Millisecond)
			close(p.closeWrite)
			//mockOrder.AssertNumberOfCalls(t, "SendTo", tt.wants.sendCnt)
			//assert.Equal(t, tt.wants.expReqCnt, len(p.requests))
		})
	}
}

func Test_remotePeerImpl_handleMsg_InPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msg := p2pmock.NewMockMessage(ctrl)
	msg.EXPECT().Subprotocol().Return(p2pcommon.PingRequest).MaxTimes(1)
	msg.EXPECT().ID().Return(p2pcommon.NewMsgID()).AnyTimes()
	msg.EXPECT().Payload().DoAndReturn(func() []byte {
		panic("for test")
	})
	mockHandler := p2pmock.NewMockMessageHandler(ctrl)
	mockHandler.EXPECT().PreHandle().AnyTimes()

	p := &remotePeerImpl{
		logger:   logger,
		handlers: make(map[p2pcommon.SubProtocol]p2pcommon.MessageHandler),
	}
	p.handlers[p2pcommon.PingRequest] = mockHandler

	if err := p.handleMsg(msg); err == nil {
		t.Errorf("remotePeerImpl.handleMsg() no error, err by panic")
	} else {
		t.Logf("expected error %v", err)
	}
}

func Test_remotePeerImpl_addCert(t *testing.T) {
	logger := log.NewLogger("p2p.test")
	addrs := []string{"192.168.1.2"}
	peerID := types.RandomPeerID()

	bpSize := 4
	bpKeys := make([]crypto.PrivKey, bpSize)
	bpIds := make([]types.PeerID, bpSize)
	bpCerts := make([]*p2pcommon.AgentCertificateV1, bpSize)
	for i := 0; i < bpSize; i++ {
		bpKeys[i], _, _ = crypto.GenerateKeyPair(crypto.Secp256k1, 11)
		bpIds[i], _ = types.IDFromPrivateKey(bpKeys[i])
		bpCerts[i], _ = p2putil.NewAgentCertV1(bpIds[i], peerID, p2putil.ConvertPKToBTCEC(bpKeys[i]), addrs, time.Hour)
	}
	sampleMeta := p2pcommon.NewMetaWith1Addr(peerID, addrs[0], 7846, "v2.0.0")
	sampleMeta.Role = types.PeerRole_Agent
	sampleMeta.ProducerIDs = bpIds
	sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP(sampleMeta.PrimaryAddress()), Port: sampleMeta.PrimaryPort()}

	type fields struct {
		prevCert []*p2pcommon.AgentCertificateV1
	}

	tests := []struct {
		name           string
		fields         fields
		args           *p2pcommon.AgentCertificateV1
		wantAdd        bool
		wantChangeRole bool
	}{
		{"TNew1", fields{bpCerts[:0]}, bpCerts[1], true, true},
		{"TNew2", fields{bpCerts[:3]}, bpCerts[3], true, false},
		{"TExist", fields{bpCerts}, bpCerts[1], false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			prevRole := types.PeerRole_Watcher
			preAdded := make([]*p2pcommon.AgentCertificateV1, len(tt.fields.prevCert))
			if len(tt.fields.prevCert) > 0 {
				copy(preAdded, tt.fields.prevCert)
				prevRole = types.PeerRole_Agent
			}
			sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn, Certificates: preAdded, AcceptedRole: prevRole}
			mockPM := p2pmock.NewMockPeerManager(ctrl)

			expectSize := len(sampleRemote.Certificates)
			if tt.wantAdd {
				expectSize++
			}
			if tt.wantChangeRole {
				mockPM.EXPECT().UpdatePeerRole(gomock.Any())
			}
			p := &remotePeerImpl{
				logger:     logger,
				remoteInfo: sampleRemote,
				pm:         mockPM,
			}
			p.addCert(tt.args)
			if len(p.remoteInfo.Certificates) != expectSize {
				t.Errorf("addCert() result size %v, want %v", len(p.remoteInfo.Certificates), expectSize)
			}
		})
	}
}

func Test_remotePeerImpl_cleanupCerts(t *testing.T) {
	logger := log.NewLogger("p2p.test")
	addrs := []string{"192.168.1.2"}
	peerID := types.RandomPeerID()

	bpSize := 4
	oldSize := 2
	yesterday := time.Now().Add(-time.Hour * 24)
	bpKeys := make([]crypto.PrivKey, bpSize)
	bpIds := make([]types.PeerID, bpSize)
	bpCerts := make([]*p2pcommon.AgentCertificateV1, bpSize)
	for i := 0; i < bpSize; i++ {
		bpKeys[i], _, _ = crypto.GenerateKeyPair(crypto.Secp256k1, 11)
		bpIds[i], _ = types.IDFromPrivateKey(bpKeys[i])
		bpCerts[i], _ = p2putil.NewAgentCertV1(bpIds[i], peerID, p2putil.ConvertPKToBTCEC(bpKeys[i]), addrs, time.Hour)
		if i < oldSize {
			bpCerts[i].CreateTime = yesterday
			bpCerts[i].ExpireTime = yesterday.Add(time.Hour)
		}
	}
	sampleMeta := p2pcommon.NewMetaWith1Addr(peerID, addrs[0], 7846, "v2.0.0")
	sampleMeta.Role = types.PeerRole_Agent
	sampleMeta.ProducerIDs = bpIds
	sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP(sampleMeta.PrimaryAddress()), Port: sampleMeta.PrimaryPort()}

	type fields struct {
		prevCert []*p2pcommon.AgentCertificateV1
	}

	tests := []struct {
		name           string
		fields         fields
		wantSize       int
		wantChangeRole bool
	}{
		{"TNothing", fields{bpCerts[:0]}, 0, false},
		{"TAllOld", fields{bpCerts[:2]}, 0, true},
		{"TPartial", fields{bpCerts[:4]}, 2, false},
		{"TAllNew", fields{bpCerts[2:]}, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			prevRole := types.PeerRole_Watcher
			preAdded := make([]*p2pcommon.AgentCertificateV1, len(tt.fields.prevCert))
			if len(tt.fields.prevCert) > 0 {
				copy(preAdded, tt.fields.prevCert)
				prevRole = types.PeerRole_Agent
			}
			sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn, Certificates: preAdded, AcceptedRole: prevRole}
			mockPM := p2pmock.NewMockPeerManager(ctrl)

			if tt.wantChangeRole {
				mockPM.EXPECT().UpdatePeerRole(gomock.Any())
			}
			p := &remotePeerImpl{
				logger:     logger,
				remoteInfo: sampleRemote,
				pm:         mockPM,
			}
			p.cleanupCerts()
			if len(p.remoteInfo.Certificates) != tt.wantSize {
				t.Errorf("addCert() result size %v, want %v", len(p.remoteInfo.Certificates), tt.wantSize)
			}
		})
	}
}
