/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package p2p

import (
	"testing"
	"time"

	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/mock"
)

const testDuration = time.Second * 2

func TestAergoPeer_RunPeer(t *testing.T) {
	mockActorServ := new(MockActorService)
	dummyP2PServ := new(MockP2PService)

	target := newRemotePeer(PeerMeta{ID: peer.ID("ddddd")}, dummyP2PServ, mockActorServ,
		log.NewLogger(log.TEST).WithCtx("test", "peer"))
	target.pingDuration = time.Second * 10
	dummyBestBlock := types.Block{Hash: []byte("testHash"), Header: &types.BlockHeader{BlockNo: 1234}}
	mockActorServ.On("requestSync", mock.Anything, mock.AnythingOfType("message.GetBlockRsp")).Return(dummyBestBlock, true)
	target.log.SetLevel("DEBUG")
	go target.runPeer()

	time.Sleep(testDuration)
	target.stop()
}

func TestAergoPeer_writeToPeer(t *testing.T) {
	type fields struct {
		log          log.ILogger
		pingDuration time.Duration
		meta         PeerMeta
		actorServ    ActorService
		ps           PeerManager
		stopChan     chan struct{}
		write        chan msgOrder
		closeWrite   chan struct{}
		requests     map[string]msgOrder
	}
	type args struct {
		m msgOrder
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RemotePeer{
				log:          tt.fields.log,
				pingDuration: tt.fields.pingDuration,
				meta:         tt.fields.meta,
				actorServ:    tt.fields.actorServ,
				ps:           tt.fields.ps,
				stopChan:     tt.fields.stopChan,
				write:        tt.fields.write,
				closeWrite:   tt.fields.closeWrite,
				requests:     tt.fields.requests,
			}

			go p.runWrite()
			p.writeToPeer(tt.args.m)

			p.closeWrite <- struct{}{}
		})
	}
}
