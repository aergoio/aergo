/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/internal/enc/proto"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/crypto"
)

func TestNewBlockProducedNoticeHandlerOfBP(t *testing.T) {
	logger := log.NewLogger("test.subproto")
	agentID := types.RandomPeerID()
	type args struct {
		remoteID types.PeerID
		role     types.PeerRole
	}
	tests := []struct {
		name string
		args args

		wantMine bool
	}{
		{"TMyAg", args{agentID, types.PeerRole_Agent}, true},
		{"TMyAgNoCert", args{agentID, types.PeerRole_Watcher}, true},
		{"TOtherAg", args{types.RandomPeerID(), types.PeerRole_Agent}, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIS := p2pmock.NewMockInternalService(ctrl)
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockIS.EXPECT().LocalSettings().Return(p2pcommon.LocalSettings{AgentID: agentID}).AnyTimes()
			mockPeer.EXPECT().ID().Return(tt.args.remoteID).AnyTimes()
			mockPeer.EXPECT().Name().Return("16..aadecf@1").AnyTimes()
			mockPeer.EXPECT().AcceptedRole().Return(tt.args.role).AnyTimes()
			mockCA := p2pmock.NewMockChainAccessor(ctrl)
			mockActor.EXPECT().GetChainAccessor().Return(mockCA).MaxTimes(1)
			mockSM := p2pmock.NewMockSyncManager(ctrl)

			got := NewBlockProducedNoticeHandler(mockIS, mockPM, mockPeer, logger, mockActor, mockSM)
			if got.myAgent != tt.wantMine {
				t.Errorf("NewBlockProducedNoticeHandler() myAgent = %v, want %v", got.myAgent, tt.wantMine)
			}
		})
	}
}

func Test_blockProducedNoticeHandler_handle_FromBP(t *testing.T) {
	logger := log.NewLogger("test.subproto")
	dummyBlockHash, _ := base58.Decode("v6zbuQ4aVSdbTwQhaiZGp5pcL5uL55X3kt2wfxor5W6")
	dummyBlockID := types.MustParseBlockID(dummyBlockHash)
	bpKey, bpPub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	bpID, _ := types.IDFromPrivateKey(bpKey)
	pubKeyBytes, _ := crypto.MarshalPublicKey(bpPub)

	dummyBlock := &types.Block{Hash: dummyBlockHash,
		Header: &types.BlockHeader{PubKey: pubKeyBytes}, Body: &types.BlockBody{}}
	wrongBlock := &types.Block{Hash: nil,
		Header: &types.BlockHeader{}, Body: &types.BlockBody{}}
	type args struct {
		msg     p2pcommon.Message
		msgBody proto.Message
	}
	tests := []struct {
		name       string
		peerID     types.PeerID
		cached     bool
		payloadBlk *types.Block

		syncmanagerCallCnt int
	}{
		// 1. normal case.
		{"TSucc", bpID, false, dummyBlock, 1},
		// 2. wrong notice (block data is missing)
		{"TW1", bpID, false, nil, 0},
		// 2. wrong notice1 (invalid block data)
		{"TW2", bpID, false, wrongBlock, 0},
		{"TWrongBP", types.RandomPeerID(), false, dummyBlock, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIS := p2pmock.NewMockInternalService(ctrl)
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockIS.EXPECT().LocalSettings().Return(p2pcommon.LocalSettings{}).AnyTimes()
			mockPeer.EXPECT().ID().Return(tt.peerID).AnyTimes()
			mockPeer.EXPECT().Name().Return("16..aadecf@1").AnyTimes()
			mockPeer.EXPECT().UpdateLastNotice(dummyBlockID, gomock.Any()).Times(tt.syncmanagerCallCnt)
			mockPeer.EXPECT().AcceptedRole().Return(types.PeerRole_Producer).AnyTimes()
			mockCA := p2pmock.NewMockChainAccessor(ctrl)
			mockActor.EXPECT().GetChainAccessor().Return(mockCA).MaxTimes(1)

			mockSM := p2pmock.NewMockSyncManager(ctrl)
			mockSM.EXPECT().HandleBlockProducedNotice(gomock.Any(), gomock.AssignableToTypeOf(&types.Block{})).Times(tt.syncmanagerCallCnt)

			dummyMsg := &testMessage{id: p2pcommon.NewMsgID(), subProtocol: p2pcommon.BlockProducedNotice}
			body := &types.BlockProducedNotice{Block: tt.payloadBlk}
			h := NewBlockProducedNoticeHandler(mockIS, mockPM, mockPeer, logger, mockActor, mockSM)
			h.Handle(dummyMsg, body)
		})
	}
}

func Test_checkBPNoticeSender(t *testing.T) {
	sampleVer := "v2.0.0"
	a1, a2 := "192.168.0.3", "172.21.0.3"
	addrs := []string{a1, a2}
	rp, ra, rw := types.PeerRole_Producer, types.PeerRole_Agent, types.PeerRole_Watcher
	rl := types.PeerRole_LegacyVersion

	bpSize := 5
	agentID := types.RandomPeerID()
	bpIds := make([]types.PeerID, bpSize)
	bpKeys := make([]crypto.PrivKey, bpSize)
	certs := make([]*p2pcommon.AgentCertificateV1, bpSize)
	for i := 0; i < bpSize; i++ {
		pk, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		bpIds[i], _ = types.IDFromPrivateKey(pk)
		bpKeys[i] = pk
		certs[i], _ = p2putil.NewAgentCertV1(bpIds[i], agentID, p2putil.ConvertPKToBTCEC(pk), addrs, time.Hour)
	}

	type args struct {
		bpID   types.PeerID
		peerID types.PeerID
		addr   string
		role   types.PeerRole
		certs  []*p2pcommon.AgentCertificateV1
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"TBP", args{bpIds[1], bpIds[1], a1, rp, nil}, true},
		{"TAgent", args{bpIds[1], agentID, a1, ra, certs}, true},
		{"TLegacy", args{bpIds[3], bpIds[3], a1, rl, nil}, true},

		{"TWatcher", args{bpIds[1], types.RandomPeerID(), a1, rw, nil}, false},
		{"TWatcher2", args{bpIds[1], bpIds[1], a1, rw, nil}, true},
		{"TDiffBP", args{bpIds[1], bpIds[2], a1, rp, nil}, false},
		{"TDiffLegacy", args{bpIds[3], bpIds[2], a1, rl, nil}, false},
		{"TMissingCert", args{bpIds[3], agentID, a1, ra, certs[:3]}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			inMeta := p2pcommon.NewMetaWith1Addr(tt.args.peerID, tt.args.addr, 7846, sampleVer)
			inRI := p2pcommon.RemoteInfo{Meta: inMeta, AcceptedRole: tt.args.role, Certificates: tt.args.certs}
			peer := p2pmock.NewMockRemotePeer(ctrl)
			peer.EXPECT().ID().Return(tt.args.peerID).AnyTimes()
			peer.EXPECT().RemoteInfo().Return(inRI).AnyTimes()
			peer.EXPECT().AcceptedRole().Return(tt.args.role).AnyTimes()

			if got := checkBPNoticeSender(tt.args.bpID, peer); got != tt.want {
				t.Errorf("checkBPNoticeSender() = %v, want %v", got, tt.want)
			}
		})
	}
}
