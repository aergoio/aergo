/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package p2p

import (
	"testing"

	"github.com/aergoio/aergo-lib/log"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
)

func TestPingProtocol_onStatusRequest(t *testing.T) {
	// TODO this test should be moved to handshake later.
	t.SkipNow()
	mockP2PS := &MockPeerManager{}
	mockIStream := &MockStream{}
	mockConn := &MockConn{}

	samplePeerID, _ := peer.IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
	// dummyPeer := AergoPeer{}

	mockIStream.On("Conn").Return(mockConn)
	mockIStream.On("Protocol").Return(protocol.ID(StatusRequest))
	mockIStream.On("Close").Return(nil)
	mockConn.On("RemotePeer").Return(samplePeerID)
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
			// 		actorServ: tt.fields.actorServ,
			// 		ps:        tt.fields.ps,
			// 		log:       tt.fields.logger,
			// 	}

			tt.expect()
		})
	}
}
