/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

var samplePeers []RemotePeer

func init() {
	samplePeers = make([]RemotePeer, 20)
	for i := 0; i < 20; i++ {
		_, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		peerid, _ := peer.IDFromPublicKey(pub)
		// first 10 are visible, others are hidden
		meta := p2pcommon.PeerMeta{ID: peerid, Hidden:i>=10}
		samplePeer := &remotePeerImpl{meta:meta}
		samplePeers[i] = samplePeer
	}
	fmt.Println("peers ",samplePeers)
}


func MakeSenderSlice(slis ...[]RemotePeer) []RemotePeer {
	result := make([]RemotePeer, 0, 10)
	for _, sli := range slis {
		result = append(result, sli...)
	}
	return result
}

func Test_addressesRequestHandler_handle(t *testing.T) {
	dummySender := &types.PeerAddress{PeerID:[]byte(dummyPeerID),Port:7845}
	senderPeer := &remotePeerImpl{meta: p2pcommon.PeerMeta{ID: dummyPeerID2}}
	tests := []struct {
		name   string
		gotPeers []RemotePeer
		wantSize int
	}{
		{"TVisible",samplePeers[:10], 10},
		{"THidden",samplePeers[10:], 0},
		{"TMix",samplePeers[5:15], 5},
		{"TVisibleWithSender",MakeSenderSlice(samplePeers[:10], []RemotePeer{senderPeer}), 10},
		{"THiddenWithSender",MakeSenderSlice(samplePeers[10:], []RemotePeer{senderPeer}), 0},
		{"TMixWithSender",MakeSenderSlice(samplePeers[5:15], []RemotePeer{senderPeer}), 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockPM .On("SelfNodeID").Return(dummyPeerID)
			mockPM.On("GetPeers").Return(tt.gotPeers)
			mockPeer := new(MockRemotePeer)
			mockMF := new(MockMoFactory)
			mockActor := new(MockActorService)
			mockPeer.On("ID").Return(dummyPeerID2)
			mockPeer.On("MF").Return(mockMF)
			mockPeer.On("sendMessage", mock.Anything)
			mockPeer.On("Name").Return("16..aadecf@1")
			mockMF.On("newMsgResponseOrder", mock.Anything, AddressesResponse, mock.MatchedBy(
				func(p0 *types.AddressesResponse) bool {
					return len(p0.Peers) == tt.wantSize
				})).Return(dummyMo)
			ph := newAddressesReqHandler(mockPM, mockPeer, logger, mockActor)
			dummyMsg := &V030Message{id: p2pcommon.NewMsgID()}
			msgBody := &types.AddressesRequest{Sender:dummySender, MaxSize:50}
			ph.handle(dummyMsg, msgBody)


		})
	}
}

func Test_addressesResponseHandler_checkAndAddPeerAddresses(t *testing.T) {
	type args struct {
		peers []*types.PeerAddress
	}
	tests := []struct {
		name   string
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func Test_addressesResponseHandler_handle(t *testing.T) {
	type fields struct {
		BaseMsgHandler BaseMsgHandler
	}
	type args struct {
		msg     p2pcommon.Message
		msgBody proto.Message
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
			ph := &addressesResponseHandler{
				BaseMsgHandler: tt.fields.BaseMsgHandler,
			}
			ph.handle(tt.args.msg, tt.args.msgBody)
		})
	}
}
