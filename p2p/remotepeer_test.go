package p2p

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testDuration = time.Second >> 1

var samplePeerID peer.ID
var sampleErr error

var logger *log.Logger

func init() {
	logger = log.NewLogger("test")

	samplePeerID, _ = peer.IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
	sampleErr = fmt.Errorf("err in unittest")
}

// TODO refactor rw and modify this test
func TestAergoPeer_RunPeer(t *testing.T) {
	t.SkipNow()
	mockActorServ := new(MockActorService)
	dummyP2PServ := new(MockPeerManager)

	dummyRW := new(MockMsgReadWriter)
	target := newRemotePeer(PeerMeta{ID: peer.ID("ddddd")}, dummyP2PServ, mockActorServ,
		logger, nil, dummyRW)

	target.pingDuration = time.Second * 10
	dummyBestBlock := types.Block{Hash: []byte("testHash"), Header: &types.BlockHeader{BlockNo: 1234}}
	mockActorServ.On("requestSync", mock.Anything, mock.AnythingOfType("message.GetBlockRsp")).Return(dummyBestBlock, true)

	go target.runPeer()

	time.Sleep(testDuration)
	target.stop()
}

func TestRemotePeer_writeToPeer(t *testing.T) {
	sampleMeta := PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}

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
		// TODO this case is evaled in pbMsgOrder tests
		// {"TFSend1", args{needResponse: true, sendErr: sampleErr}, wants{1, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := new(MockActorService)
			mockPeerManager := new(MockPeerManager)
			mockOrder := new(MockMsgOrder)
			dummyRW := new(MockMsgReadWriter)
			mockOrder.On("IsNeedSign").Return(true)
			mockOrder.On("IsRequest", mockPeerManager).Return(true)
			mockOrder.On("SendTo", mock.AnythingOfType("*p2p.RemotePeer")).Return(tt.args.sendErr == nil)
			mockOrder.On("GetProtocolID").Return(PingRequest)
			mockOrder.On("GetMsgID").Return("test_req")
			mockOrder.On("ResponseExpected").Return(tt.args.needResponse)

			p := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger, nil, dummyRW)
			p.state.SetAndGet(types.RUNNING)
			p.write.Open()
			defer p.write.Close()
			go p.runWrite()
			p.state.SetAndGet(types.RUNNING)

			p.writeToPeer(mockOrder)

			// FIXME wait in more relaiable way
			time.Sleep(50 * time.Millisecond)
			p.closeWrite <- struct{}{}
			mockOrder.AssertNumberOfCalls(t, "SendTo", tt.wants.sendCnt)
			assert.Equal(t, tt.wants.expReqCnt, len(p.requests))
		})
	}
}

func TestePeer_sendPing(t *testing.T) {
	selfPeerID, _ := peer.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	sampleSelf := PeerMeta{ID: selfPeerID, IPAddress: "192.168.1.1", Port: 6845}

	sampleMeta := PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
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
			mockActorServ := new(MockActorService)
			mockPeerManager := new(MockPeerManager)
			mockMF := new(MockMoFactory)

			mockActorServ.On("CallRequest", message.ChainSvc, mock.AnythingOfType("*message.GetBestBlock")).Return(dummyBestBlockRsp, tt.getBlockErr)
			mockPeerManager.On("SelfMeta").Return(sampleSelf)
			mockMF.On("signMsg", mock.AnythingOfType("*types.P2PMessage")).Return(nil)
			p := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger, mockMF, nil)
			p.write.Open()
			p.state.SetAndGet(types.RUNNING)
			defer p.write.Close()

			go p.sendPing()

			time.Sleep(200 * time.Millisecond)

			actualWrite := false
			select {
			case msg := <-p.write.Out():
				assert.Equal(t, PingRequest, msg.(msgOrder).GetProtocolID())
				actualWrite = true
				p.write.Done() <- msg
			default:
			}
			assert.Equal(t, tt.wants.wantWrite, actualWrite)
			mockPeerManager.AssertNotCalled(t, "SelfMeta")
			// ping request does not contain best block information anymore.
			mockActorServ.AssertNotCalled(t, "CallRequest")
		})
	}
}

// TODO sendStatus will be deleted
func IgnoreTestRemotePeer_sendStatus(t *testing.T) {
	selfPeerID, _ := peer.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	sampleSelf := PeerMeta{ID: selfPeerID, IPAddress: "192.168.1.1", Port: 6845}

	sampleMeta := PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	dummyBestBlockRsp := message.GetBestBlockRsp{Block: &types.Block{Header: &types.BlockHeader{}}}
	type wants struct {
		wantWrite bool
	}
	tests := []struct {
		name        string
		getBlockErr error
		wants       wants
	}{
		{"TN", nil, wants{wantWrite: true}},
		{"TF", sampleErr, wants{wantWrite: false}},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := new(MockActorService)
			mockPeerManager := new(MockPeerManager)

			mockActorServ.On("CallRequest", message.ChainSvc, mock.AnythingOfType("*message.GetBestBlock")).Return(dummyBestBlockRsp, tt.getBlockErr)
			mockPeerManager.On("SelfMeta").Return(sampleSelf)

			p := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger, nil, nil)
			p.write.Open()
			p.state.SetAndGet(types.RUNNING)
			defer p.write.Close()

			go p.sendStatus()

			time.Sleep(200 * time.Millisecond)

			actualWrite := false
			select {
			case msg := <-p.write.Out():
				assert.Equal(t, StatusRequest, msg.(msgOrder).GetProtocolID())
				actualWrite = true
				p.write.Done() <- msg
			default:
			}
			assert.Equal(t, tt.wants.wantWrite, actualWrite)
			mockActorServ.AssertNumberOfCalls(t, "CallRequest", 1)
		})
	}
}

func TestRemotePeer_pruneRequests(t *testing.T) {
	sampleMeta := PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
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
		mockActorServ := new(MockActorService)
		mockPeerManager := new(MockPeerManager)
		p := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger, nil, nil)
		t.Run(tt.name, func(t *testing.T) {
			p.requests["r1"] = &pbMessageOrder{message: &types.P2PMessage{Header: &types.MsgHeader{Id: "r1", Timestamp: time.Now().Add(time.Minute * -61).Unix()}}}
			p.requests["r2"] = &pbMessageOrder{message: &types.P2PMessage{Header: &types.MsgHeader{Id: "r2", Timestamp: time.Now().Add(time.Minute*-60 - time.Second).Unix()}}}
			p.requests["rn"] = &pbMessageOrder{message: &types.P2PMessage{Header: &types.MsgHeader{Id: "rn", Timestamp: time.Now().Add(time.Minute * -60).Unix()}}}
			p.pruneRequests()

			assert.Equal(t, 1, len(p.requests))
		})
	}
}

func TestRemotePeer_tryGetStream(t *testing.T) {
	sampleMeta := PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	mockStream := new(MockStream)
	type args struct {
		msgID    string
		protocol protocol.ID
		timeout  time.Duration
	}
	tests := []struct {
		name    string
		args    args
		timeout bool
		want    inet.Stream
	}{
		{"TN", args{"m1", "p1", time.Millisecond * 100}, false, mockStream},
		{"TTimeout", args{"m1", "p1", time.Millisecond * 100}, true, nil},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		mockActorServ := new(MockActorService)
		mockPeerManager := new(MockPeerManager)
		if tt.timeout {
			mockPeerManager.On("NewStream", mock.Anything, mock.Anything, mock.Anything).After(time.Second).Return(mockStream, nil)
		} else {
			mockPeerManager.On("NewStream", mock.Anything, mock.Anything, mock.Anything).Return(mockStream, nil)
		}
		t.Run(tt.name, func(t *testing.T) {
			p := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger, nil, nil)
			got := p.tryGetStream(tt.args.msgID, tt.args.protocol, tt.args.timeout)

			assert.Equal(t, got, tt.want)
		})
	}
}

func TestRemotePeer_sendMessage(t *testing.T) {
	sampleMeta := PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}

	type args struct {
		msgID    string
		protocol protocol.ID
		timeout  time.Duration
	}
	tests := []struct {
		name    string
		args    args
		timeout bool
	}{
		{"TSucc", args{"m1", "p1", time.Millisecond * 100}, false},
		{"TTimeout", args{"mtimeout", "p1", time.Millisecond * 100}, true},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		mockActorServ := new(MockActorService)
		mockPeerManager := new(MockPeerManager)
		mockMsg := new(MockMsgOrder)
		mockMsg.On("GetMsgID").Return(tt.args.msgID)
		mockMsg.On("GetProtocolID").Return(NewBlockNotice)

		writeCnt := int32(0)
		t.Run(tt.name, func(t *testing.T) {
			finishTest := make(chan interface{}, 1)
			wg := &sync.WaitGroup{}
			wg.Add(1)
			wg2 := &sync.WaitGroup{}
			wg2.Add(1)
			p := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger, nil, nil)
			p.write.Open()
			p.state.SetAndGet(types.RUNNING)
			defer p.write.Close()

			if !tt.timeout {
				go func() {
					wg.Wait()
					for {
						select {
						case mo := <-p.write.Out():
							p.logger.Info().Msgf("Got order from chan %v", mo)
							msg := mo.(msgOrder)
							p.logger.Info().Str(LogMsgID, msg.GetMsgID()).Msg("Got order")
							atomic.AddInt32(&writeCnt, 1)
							p.write.Done() <- msg
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
	sampleMeta := PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	mockMsg := new(MockMsgOrder)
	mockMsg.On("GetMsgID").Return("m1")
	mockMsg.On("GetProtocolID").Return(NewBlockNotice)

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
		mockActorServ := new(MockActorService)
		mockPeerManager := new(MockPeerManager)
		mockMsgHandler := new(MockMessageHandler)
		mockMF := new(MockMoFactory)
		t.Run(tt.name, func(t *testing.T) {
			msg := &types.P2PMessage{Header: &types.MsgHeader{Subprotocol: PingRequest.Uint32()}}
			if tt.args.nohandler {
				msg.Header.Subprotocol = 3999999999
			}
			bodyStub := &types.Ping{}
			mockMsgHandler.On("parsePayload", mock.AnythingOfType("[]uint8")).Return(bodyStub, tt.args.parerr)
			mockMsgHandler.On("checkAuth", mock.AnythingOfType("*types.P2PMessage"), mock.Anything).Return(tt.args.autherr)
			mockMsgHandler.On("handle", mock.AnythingOfType("*types.MsgHeader"), mock.Anything)
			target := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger, mockMF, nil)
			target.handlers[PingRequest] = mockMsgHandler

			if err := target.handleMsg(msg); (err != nil) != tt.wantErr {
				t.Errorf("RemotePeer.handleMsg() error = %v, wantErr %v", err, tt.wantErr)
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
				mockMsgHandler.AssertCalled(t, "handle", msg.Header, bodyStub)
			}
		})
	}
}
