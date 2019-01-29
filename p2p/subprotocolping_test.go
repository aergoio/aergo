package p2p

import (
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
)

func TestPingProtocol_onStatusRequest(t *testing.T) {
	//// TODO this test should be moved to handshake later.
	//t.SkipNow()
	mockP2PS := &MockPeerManager{}
	mockIStream := &MockStream{}
	mockConn := &MockConn{}

	samplePeerID, _ := peer.IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
	// dummyPeer := AergoPeer{}

	mockIStream.On("Conn").Return(mockConn)
	mockIStream.On("Protocol").Return(protocol.ID(StatusRequest))
	mockIStream.On("Close").Return(nil)
	mockConn.On("MessageImpl").Return(samplePeerID)
	mockP2PS.On("LookupPeer", samplePeerID).Return(nil, false)

	type fields struct {
		actorServ ActorService
		ps        PeerManager
		logger    *log.Logger
	}
	type args struct {
		s inet.Stream
	}
	tests := []struct {
		name   string
		fields *fields
		args   args
		expect func()
	}{
		{
			"normal",
			&fields{
				actorServ: &MockActorService{},
				logger:    log.NewLogger("test.p2p"),
				ps:        mockP2PS,
			},
			args{s: mockIStream},
			func() {
				// dummy
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 	p := &PingProtocol{
			// 		actorService: tt.fields.actorService,
			// 		ps:        tt.fields.ps,
			// 		log:       tt.fields.logger,
			// 	}

			tt.expect()
		})
	}
}

func Test_pingRequestHandler_handle(t *testing.T) {
	type args struct {
		hash    []byte
		height 	uint64
	}
	tests := []struct {
		name   string
		args   args

		wantCallUpdate bool
	}{
		{"TSucc", args{dummyBlockHash, 10}, true},
		{"TWrongHash", args{[]byte{}, 10},false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockPeer := new(MockRemotePeer)
			mockActor := new(MockActorService)
			dummyMF := &v030MOFactory{}

			mockPeer.On("MF").Return(dummyMF)
			mockPeer.On("ID").Return(dummyPeerID)
			mockPeer.On("updateLastNotice", tt.args.hash, tt.args.height)
			mockPeer.On("sendMessage", mock.Anything)

			mockCA := new(MockChainAccessor)
			mockActor.On("GetChainAccessor").Return(mockCA)

			msg := &V030Message{subProtocol:PingRequest, id: sampleMsgID}
			body := &types.Ping{BestBlockHash:tt.args.hash, BestHeight: tt.args.height}

			ph := newPingReqHandler(mockPM, mockPeer, logger, mockActor)

			ph.handle(msg, body)

			if !tt.wantCallUpdate {
				mockPeer.AssertNotCalled(t, "updateLastNotice",mock.Anything,mock.Anything)
			}
			// other call verifications are already checked by On() method
		})
	}
}
