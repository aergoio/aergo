package subproto

import (
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/golang/mock/gomock"
	inet "github.com/libp2p/go-libp2p-net"
)

func TestPingProtocol_onStatusRequest(t *testing.T) {
	//// TODO this test should be moved to handshake later.
	//t.SkipNow()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockP2PS := p2pmock.NewMockPeerManager(ctrl)
	mockIStream := p2pmock.NewMockStream(ctrl)
	//mockConn := p2pmock.NewMockConn(ctrl)

	//samplePeerID, _ := peer.IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
	// dummyPeer := AergoPeer{}

	//mockIStream.EXPECT().Conn().Return(mockConn)
	//mockIStream.EXPECT().Protocol().Return(protocol.ID(StatusRequest))
	//	mockIStream.EXPECT().Close().Return(nil)
	//mockConn.EXPECT().("MessageImpl").Return(samplePeerID)
	//mockP2PS.EXPECT().LookupPeer(, samplePeerID).Return(nil, false)

	type fields struct {
		actorServ p2pcommon.ActorService
		ps        p2pcommon.PeerManager
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
				actorServ: p2pmock.NewMockActorService(ctrl),
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.subproto")
	dummyBlockHash, _ := enc.ToBytes("v6zbuQ4aVSdbTwQhaiZGp5pcL5uL55X3kt2wfxor5W6")
	var dummyPeerID, _ = peer.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")

	type args struct {
		hash   []byte
		height uint64
	}
	tests := []struct {
		name string
		args args

		sendRespCnt int
	}{
		{"TSucc", args{dummyBlockHash, 10}, 1},
		{"TWrongHash", args{[]byte{}, 10}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			dummyMF := p2pmock.NewMockMoFactory(ctrl)
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().Name().Return("16..aadecf@1").AnyTimes()
			mockPeer.EXPECT().MF().Return(dummyMF).MinTimes(tt.sendRespCnt)
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(tt.sendRespCnt)
			mockPeer.EXPECT().UpdateLastNotice(tt.args.hash, tt.args.height).Times(tt.sendRespCnt)
			mockCA := p2pmock.NewMockChainAccessor(ctrl)
			mockActor.EXPECT().GetChainAccessor().Return(mockCA).MaxTimes(1)

			reqID := p2pcommon.NewMsgID()
			dummyMF.EXPECT().NewMsgResponseOrder(reqID, PingResponse, gomock.AssignableToTypeOf(&types.Pong{})).Return(nil).Times(tt.sendRespCnt)

			msg := p2pmock.NewMockMessage(ctrl)
			msg.EXPECT().ID().Return(reqID).AnyTimes()
			msg.EXPECT().Subprotocol().Return(PingRequest).AnyTimes()
			body := &types.Ping{BestBlockHash: tt.args.hash, BestHeight: tt.args.height}

			ph := NewPingReqHandler(mockPM, mockPeer, logger, mockActor)

			ph.Handle(msg, body)

		})
	}
}

