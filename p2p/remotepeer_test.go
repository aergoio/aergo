package p2p

import (
	"fmt"
	"testing"
	"time"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testDuration = time.Second * 2

var samplePeerID peer.ID
var sampleErr error

var logger *log.Logger

func init() {
	logger = log.NewLogger(log.TEST)
	logger.SetLevel("fatal")
	samplePeerID, _ = peer.IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
	sampleErr = fmt.Errorf("err in unittest")
}

func TestAergoPeer_RunPeer(t *testing.T) {
	mockActorServ := new(MockActorService)
	dummyP2PServ := new(MockP2PService)

	target := newRemotePeer(PeerMeta{ID: peer.ID("ddddd")}, dummyP2PServ, mockActorServ,
		logger)
	target.pingDuration = time.Second * 10
	dummyBestBlock := types.Block{Hash: []byte("testHash"), Header: &types.BlockHeader{BlockNo: 1234}}
	mockActorServ.On("requestSync", mock.Anything, mock.AnythingOfType("message.GetBlockRsp")).Return(dummyBestBlock, true)
	target.log.SetLevel("DEBUG")
	go target.runPeer()

	time.Sleep(testDuration)
	target.stop()
}

func TestAergoPeer_writeToPeer(t *testing.T) {
	sampleMeta := PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}

	type args struct {
		StreamResult error
		needSign     bool
		signErr      error
		needResponse bool
		sendErr      error
	}
	type wants struct {
		streamCall int
		sendCnt    int
		expReqCnt  int
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{"TNReq1", args{}, wants{1, 1, 0}},
		{"TNReq2", args{needSign: true}, wants{1, 1, 0}},
		{"TNReqWResp1", args{needResponse: true}, wants{1, 1, 1}},
		{"TNReqWResp2", args{needSign: true, needResponse: true}, wants{1, 1, 1}},

		// no sign no error
		{"TFSign1", args{needSign: false, signErr: sampleErr, needResponse: true}, wants{1, 1, 1}},
		// error while signing
		{"TFSign2", args{needSign: true, signErr: sampleErr, needResponse: true}, wants{0, 0, 0}},
		// error while get stream
		{"TFStream", args{StreamResult: sampleErr, needResponse: true}, wants{1, 0, 0}},
		// {"TFReqWResp2", args{needSign: true, needResponse: true, expReqCnt: 1}},

		{"TFSend1", args{needSign: true, needResponse: true, sendErr: sampleErr}, wants{1, 1, 0}},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := new(MockActorService)
			mockPeerManager := new(MockP2PService)
			mockStream := new(Stream)
			mockOrder := new(MockMsgOrder)
			mockPeerManager.On("NewStream", mock.Anything, mock.AnythingOfType("peer.ID"),
				mock.AnythingOfType("protocol.ID")).Return(mockStream, tt.args.StreamResult)
			if tt.args.StreamResult != nil {
				mockPeerManager.On("RemovePeer", mock.AnythingOfType("peer.ID"))
			}
			mockOrder.On("IsNeedSign").Return(tt.args.needSign)
			if tt.args.needSign {
				mockOrder.On("SignWith", mockPeerManager).Return(tt.args.signErr)
			}
			mockOrder.On("IsRequest", mockPeerManager).Return(true)
			mockOrder.On("SendOver", mockStream).Return(tt.args.sendErr)
			mockOrder.On("GetProtocolID").Return(protocol.ID("dummy"))
			mockOrder.On("GetRequestID").Return("test_req")
			mockOrder.On("ResponseExpected").Return(tt.args.needResponse)

			p := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger)
			go p.runWrite()

			p.writeToPeer(mockOrder)

			// FIXME wait in more relaiable way
			time.Sleep(100 * time.Millisecond)
			p.closeWrite <- struct{}{}
			mockPeerManager.AssertNumberOfCalls(t, "NewStream", tt.wants.streamCall)
			mockOrder.AssertNumberOfCalls(t, "SendOver", tt.wants.sendCnt)
			mockOrder.AssertNumberOfCalls(t, "SendOver", tt.wants.sendCnt)
			assert.Equal(t, tt.wants.expReqCnt, len(p.requests))
		})
	}
}

func TestRemotePeer_sendPing(t *testing.T) {
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
			mockPeerManager := new(MockP2PService)

			mockActorServ.On("CallRequest", message.ChainSvc, mock.AnythingOfType("*message.GetBestBlock")).Return(dummyBestBlockRsp, tt.getBlockErr)
			mockPeerManager.On("SelfMeta").Return(sampleSelf)

			p := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger)
			go p.sendPing()

			time.Sleep(200 * time.Millisecond)

			actualWrite := false
			select {
			case msg := <-p.write:
				assert.Equal(t, protocol.ID(pingRequest), msg.GetProtocolID())
				actualWrite = true
			default:
			}
			assert.Equal(t, tt.wants.wantWrite, actualWrite)
			mockPeerManager.AssertNotCalled(t, "SelfMeta")
			mockActorServ.AssertNumberOfCalls(t, "CallRequest", 1)
		})
	}
}

func TestRemotePeer_sendStatus(t *testing.T) {
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
			mockPeerManager := new(MockP2PService)

			mockActorServ.On("CallRequest", message.ChainSvc, mock.AnythingOfType("*message.GetBestBlock")).Return(dummyBestBlockRsp, tt.getBlockErr)
			mockPeerManager.On("SelfMeta").Return(sampleSelf)

			p := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger)
			go p.sendStatus()

			time.Sleep(200 * time.Millisecond)

			actualWrite := false
			select {
			case msg := <-p.write:
				assert.Equal(t, protocol.ID(statusRequest), msg.GetProtocolID())
				actualWrite = true
			default:
			}
			assert.Equal(t, tt.wants.wantWrite, actualWrite)
			mockActorServ.AssertNumberOfCalls(t, "CallRequest", 1)
		})
	}
}
