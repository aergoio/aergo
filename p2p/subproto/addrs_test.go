/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/crypto"
	"testing"

	"github.com/aergoio/aergo/p2p/p2pcommon"

	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

var samplePeers []p2pcommon.RemotePeer

func MakeSenderSlice(slis ...[]p2pcommon.RemotePeer) []p2pcommon.RemotePeer {
	result := make([]p2pcommon.RemotePeer, 0, 10)
	for _, sli := range slis {
		result = append(result, sli...)
	}
	return result
}

func Test_addressesRequestHandler_handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.subproto")

	var samplePeers = make([]p2pcommon.RemotePeer, 20)
	for i := 0; i < 20; i++ {
		_, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		peerid, _ := types.IDFromPublicKey(pub)
		// first 10 are visible, others are hidden
		meta := p2pcommon.PeerMeta{ID: peerid, Hidden: i >= 10}
		samplePeer := p2pmock.NewMockRemotePeer(ctrl)
		samplePeer.EXPECT().ID().Return(meta.ID).AnyTimes()
		samplePeer.EXPECT().Meta().Return(meta).AnyTimes()
		samplePeers[i] = samplePeer
	}
	var dummyPeerID, _ = types.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")
	var dummyPeerID2, _ = types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")

	dummySender := &types.PeerAddress{PeerID: []byte(dummyPeerID), Port: 7845}
	senderPeer := p2pmock.NewMockRemotePeer(ctrl)
	senderPeer.EXPECT().ID().Return(dummyPeerID2).AnyTimes()
	senderPeer.EXPECT().Meta().Return(p2pcommon.PeerMeta{ID:dummyPeerID2}).AnyTimes()

	tests := []struct {
		name     string
		gotPeers []p2pcommon.RemotePeer
		wantSize int
	}{
		{"TVisible", samplePeers[:10], 10},
		{"THidden", samplePeers[10:], 0},
		{"TMix", samplePeers[5:15], 5},
		// get peers contains sender peer itself. it will be skipped in response
		{"TVisibleWithSender", MakeSenderSlice(samplePeers[:10], []p2pcommon.RemotePeer{senderPeer}), 10},
		{"THiddenWithSender", MakeSenderSlice(samplePeers[10:], []p2pcommon.RemotePeer{senderPeer}), 0},
		{"TMixWithSender", MakeSenderSlice(samplePeers[5:15], []p2pcommon.RemotePeer{senderPeer}), 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPM.EXPECT().SelfNodeID().Return(dummyPeerID2).AnyTimes()
			mockPM.EXPECT().GetPeers().Return(tt.gotPeers).AnyTimes()
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer.EXPECT().ID().Return(dummyPeerID2).AnyTimes()
			mockPeer.EXPECT().Name().Return("16..aadecf@1").AnyTimes()
			mockPeer.EXPECT().MF().Return(mockMF).MinTimes(1)
			mockPeer.EXPECT().Meta().Return(p2pcommon.PeerMeta{ID:"16..aadecf@1"}).AnyTimes()
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)
			dummyMo := &testMo{}
			mockMF.EXPECT().NewMsgResponseOrder(gomock.Any(), AddressesResponse, &addrRespSizeMatcher{tt.wantSize}).Return(dummyMo)

			ph := NewAddressesReqHandler(mockPM, mockPeer, logger, mockActor)
			dummyMsg :=&testMessage{id:p2pcommon.NewMsgID()}
			msgBody := &types.AddressesRequest{Sender: dummySender, MaxSize: 50}
			ph.Handle(dummyMsg, msgBody)

		})
	}
}

type addrRespSizeMatcher struct {
	wantSize int
}

func (rsm addrRespSizeMatcher) Matches(x interface{}) bool {
	m, ok := x.(*types.AddressesResponse)
	if !ok {
		return false
	}

	return rsm.wantSize == len(m.Peers)
}

func (rsm addrRespSizeMatcher) String() string {
	return fmt.Sprintf("len(Peers) = %d",rsm.wantSize)
}

func Test_addressesResponseHandler_checkAndAddPeerAddresses(t *testing.T) {
	type args struct {
		peers []*types.PeerAddress
	}
	tests := []struct {
		name string
		args args
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
			ph.Handle(tt.args.msg, tt.args.msgBody)
		})
	}
}
