/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package p2p

import (
	"testing"

	"github.com/aergoio/aergo/pkg/log"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
)

func TestPingProtocol_onStatusRequest(t *testing.T) {
	mockP2PS := &MockP2PService{}
	mockIStream := &Stream{}
	mockConn := &MockConn{}

	samplePeerID, _ := peer.IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
	// dummyPeer := AergoPeer{}

	mockIStream.On("Conn").Return(mockConn)
	mockIStream.On("Protocol").Return(protocol.ID(statusRequest))
	mockConn.On("RemotePeer").Return(samplePeerID)
	mockP2PS.On("LookupPeer", samplePeerID).Return(nil, false)

	type fields struct {
		actorServ ActorService
		ps        PeerManager
		logger    log.ILogger
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
				logger:    log.NewLogger(log.TEST).WithCtx("test", "p2p"),
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
			p := &PingProtocol{
				actorServ: tt.fields.actorServ,
				ps:        tt.fields.ps,
				log:       tt.fields.logger,
			}
			p.onStatusRequest(tt.args.s)
			tt.expect()
		})
	}
}
