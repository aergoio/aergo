package subproto

import (
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func Test_certRenewedNoticeHandler_Handle(t *testing.T) {
	logger := log.NewLogger("test.subproto")
	selfID := types.RandomPeerID()
	bpLPK, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	bpID, _ := peer.IDFromPrivateKey(bpLPK)
	bpPK := p2putil.ConvertPKToBTCEC(bpLPK)

	RAG, RWT := types.PeerRole_Agent, types.PeerRole_Watcher
	otherID := types.RandomPeerID()
	now := time.Now()
	dayAfter := now.Add(time.Hour * 24)

	type args struct {
		agentID      types.PeerID
		cTime, eTime time.Time
	}
	tests := []struct {
		name          string
		inClaimedRole types.PeerRole
		inAccRole     types.PeerRole
		hasBP         bool

		args args

		wantAddCert    bool
		wantUpdateRole bool
	}{
		{"TAlreadyA", RAG, RAG, true, args{selfID, now, dayAfter}, true, false},
		{"TWatcherToA", RAG, RWT, true, args{selfID, now, dayAfter}, true, true},
		{"TWrongAgID", RAG, RWT, true, args{otherID, now, dayAfter}, false, false},
		{"TNotByBP", RAG, RWT, false, args{otherID, now, dayAfter}, false, false},
		{"TNotAg", RWT, RWT, false, args{otherID, now, dayAfter}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			pm := p2pmock.NewMockPeerManager(ctrl)
			cm := p2pmock.NewMockCertificateManager(ctrl)
			peer := p2pmock.NewMockRemotePeer(ctrl)
			actor := p2pmock.NewMockActorService(ctrl)

			selfMeta := p2pcommon.NewMetaWith1Addr(selfID, "192.168.0.2", 7846, "v2.0.0")
			selfMeta.Role = tt.inClaimedRole
			selfMeta.ProducerIDs = []types.PeerID{types.RandomPeerID(), types.RandomPeerID()}
			if tt.hasBP {
				selfMeta.ProducerIDs = append(selfMeta.ProducerIDs, bpID)
			}
			ri := p2pcommon.RemoteInfo{Meta: selfMeta, AcceptedRole: tt.inAccRole}

			peer.EXPECT().ID().Return(selfID).AnyTimes()
			peer.EXPECT().Meta().Return(selfMeta).AnyTimes()
			peer.EXPECT().Name().Return("samplePeer").AnyTimes()
			peer.EXPECT().RemoteInfo().Return(ri).AnyTimes()
			peer.EXPECT().AcceptedRole().Return(tt.inAccRole).AnyTimes()

			cert, _ := p2putil.NewAgentCertV1(bpID, tt.args.agentID, bpPK, []string{"192.168.0.2"}, time.Hour*24)
			pCert, _ := p2putil.ConvertCertToProto(cert)

			dummyMsg := &testMessage{id: p2pcommon.NewMsgID(), subProtocol: p2pcommon.CertificateRenewedNotice}
			body := &types.CertificateRenewedNotice{Certificate: pCert}

			if tt.wantAddCert {
				peer.EXPECT().AddCertificate(gomock.Any())
			}
			h := NewCertRenewedNoticeHandler(pm, cm, peer, logger, actor)
			h.Handle(dummyMsg, body)
		})
	}
}
