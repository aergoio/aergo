/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package v030

import (
	"context"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
)

const (
	sampleKeyFile = "../test/sample/sample.key"
)

var (
	myChainID, newVerChainID, theirChainID          *types.ChainID
	myChainBytes, newVerChainBytes, theirChainBytes []byte
	samplePeerID, _                                        = types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	dummyBlockHash, _                                      = hex.DecodeString("4f461d85e869ade8a0544f8313987c33a9c06534e50c4ad941498299579bd7ac")
	dummyBlockHeight                                uint64 = 100215
)

func init() {
	myChainID = types.NewChainID()
	myChainID.Magic = "itSmain1"
	myChainBytes, _ = myChainID.Bytes()

	newVerChainID = types.NewChainID()
	newVerChainID.Read(myChainBytes)
	newVerChainID.Version = 3
	newVerChainBytes, _ = newVerChainID.Bytes()

	theirChainID = types.NewChainID()
	theirChainID.Read(myChainBytes)
	theirChainID.Magic = "itsdiff2"
	theirChainBytes, _ = theirChainID.Bytes()

	baseCfg := &config.BaseConfig{AuthDir: "test"}
	p2pCfg := &config.P2PConfig{NPKey: sampleKeyFile}
	p2pkey.InitNodeInfo(baseCfg, p2pCfg, "0.0.1-test", log.NewLogger("v030.test"))
}

func TestDeepEqual(t *testing.T) {
	b1, _ := myChainID.Bytes()
	b2 := make([]byte, len(b1), len(b1)<<1)
	copy(b2, b1)

	s1 := &types.Status{ChainID: b1}
	s2 := &types.Status{ChainID: b2}

	if !reflect.DeepEqual(s1, s2) {
		t.Errorf("byte slice cant do DeepEqual! %v, %v", b1, b2)
	}

}

func TestV030StatusHS_doForOutbound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test")
	mockActor := p2pmock.NewMockActorService(ctrl)
	mockCA := p2pmock.NewMockChainAccessor(ctrl)
	mockPM := p2pmock.NewMockPeerManager(ctrl)

	dummyMeta := p2pcommon.NewMetaWith1Addr(samplePeerID, "dummy.aergo.io", 7846, "v2.0.0")
	dummyAddr := dummyMeta.ToPeerAddress()
	mockPM.EXPECT().SelfMeta().Return(dummyMeta).AnyTimes()
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	mockActor.EXPECT().GetChainAccessor().Return(mockCA).AnyTimes()
	mockCA.EXPECT().GetBestBlock().Return(dummyBlock, nil).AnyTimes()

	dummyStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr, Version: dummyAddr.Version}
	succResult := &p2pcommon.HandshakeResult{Meta: dummyMeta}
	nilSenderStatusMsg := &types.Status{ChainID: myChainBytes, Sender: nil}
	diffStatusMsg := &types.Status{ChainID: theirChainBytes, Sender: &dummyAddr}
	tests := []struct {
		name       string
		readReturn *types.Status
		readError  error
		writeError error
		want       *p2pcommon.HandshakeResult
		wantErr    bool
		wantGoAway bool
	}{
		{"TSuccess", dummyStatusMsg, nil, nil, succResult, false, false},
		{"TUnexpMsg", nil, nil, nil, nil, true, true},
		{"TRFail", dummyStatusMsg, fmt.Errorf("failed"), nil, nil, true, true},
		{"TRNoSender", nilSenderStatusMsg, nil, nil, nil, true, true},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true, false},
		{"TDiffChain", diffStatusMsg, nil, nil, nil, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyReader := p2pmock.NewMockReadWriteCloser(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)

			var containerMsg *p2pcommon.MessageValue
			if tt.readReturn != nil {
				containerMsg = p2pcommon.NewSimpleMsgVal(p2pcommon.StatusRequest, p2pcommon.NewMsgID())
				statusBytes, _ := p2putil.MarshalMessageBody(tt.readReturn)
				containerMsg.SetPayload(statusBytes)
			} else {
				containerMsg = p2pcommon.NewSimpleMsgVal(p2pcommon.AddressesRequest, p2pcommon.NewMsgID())
			}
			mockRW.EXPECT().ReadMsg().Return(containerMsg, tt.readError).AnyTimes()
			if tt.wantGoAway {
				mockRW.EXPECT().WriteMsg(&MsgMatcher{p2pcommon.GoAway}).Return(tt.writeError)
			}
			mockRW.EXPECT().WriteMsg(&MsgMatcher{p2pcommon.StatusRequest}).Return(tt.writeError).MaxTimes(1)

			h := NewV030VersionedHS(mockPM, mockActor, logger, myChainID, samplePeerID, dummyReader)
			h.msgRW = mockRW
			got, err := h.DoForOutbound(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.want != nil {
				if !got.Meta.Equals(tt.want.Meta) {
					t.Errorf("PeerHandshaker.handshakeOutboundPeer() peerID = %v, want %v", got.Meta, tt.want.Meta)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestV030StatusHS_handshakeInboundPeer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// t.SkipNow()
	logger := log.NewLogger("test")
	mockActor := p2pmock.NewMockActorService(ctrl)
	mockCA := p2pmock.NewMockChainAccessor(ctrl)
	mockPM := p2pmock.NewMockPeerManager(ctrl)

	dummyMeta := p2pcommon.NewMetaWith1Addr(samplePeerID, "dummy.aergo.io", 7846, "v2.0.0")
	dummyAddr := dummyMeta.ToPeerAddress()
	mockPM.EXPECT().SelfMeta().Return(dummyMeta).AnyTimes()
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	//dummyBlkRsp := message.GetBestBlockRsp{Block: dummyBlock}
	mockActor.EXPECT().GetChainAccessor().Return(mockCA).AnyTimes()
	mockCA.EXPECT().GetBestBlock().Return(dummyBlock, nil).AnyTimes()

	dummyStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr, Version: dummyAddr.Version}
	succResult := &p2pcommon.HandshakeResult{Meta: dummyMeta}
	nilSenderStatusMsg := &types.Status{ChainID: myChainBytes, Sender: nil}
	diffStatusMsg := &types.Status{ChainID: theirChainBytes, Sender: &dummyAddr}
	tests := []struct {
		name       string
		readReturn *types.Status
		readError  error
		writeError error
		want       *p2pcommon.HandshakeResult
		wantErr    bool
		wantGoAway bool
	}{
		{"TSuccess", dummyStatusMsg, nil, nil, succResult, false, false},
		{"TUnexpMsg", nil, nil, nil, nil, true, true},
		{"TRFail", dummyStatusMsg, fmt.Errorf("failed"), nil, nil, true, true},
		{"TRNoSender", nilSenderStatusMsg, nil, nil, nil, true, true},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true, false},
		{"TDiffChain", diffStatusMsg, nil, nil, nil, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyReader := p2pmock.NewMockReadWriteCloser(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)

			containerMsg := &p2pcommon.MessageValue{}
			if tt.readReturn != nil {
				containerMsg = p2pcommon.NewSimpleMsgVal(p2pcommon.StatusRequest, p2pcommon.NewMsgID())
				statusBytes, _ := p2putil.MarshalMessageBody(tt.readReturn)
				containerMsg.SetPayload(statusBytes)
			} else {
				containerMsg = p2pcommon.NewSimpleMsgVal(p2pcommon.AddressesRequest, p2pcommon.NewMsgID())
			}

			mockRW.EXPECT().ReadMsg().Return(containerMsg, tt.readError).AnyTimes()
			if tt.wantGoAway {
				mockRW.EXPECT().WriteMsg(&MsgMatcher{p2pcommon.GoAway}).Return(tt.writeError)
			}
			mockRW.EXPECT().WriteMsg(&MsgMatcher{p2pcommon.StatusRequest}).Return(tt.writeError).MaxTimes(1)

			h := NewV030VersionedHS(mockPM, mockActor, logger, myChainID, samplePeerID, dummyReader)
			h.msgRW = mockRW
			got, err := h.DoForInbound(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.handshakeInboundPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.want != nil {
				if !got.Meta.Equals(tt.want.Meta) {
					t.Errorf("PeerHandshaker.handshakeOutboundPeer() peerID = %v, want %v", got.Meta, tt.want.Meta)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.handshakeInboundPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MsgMatcher struct {
	sub p2pcommon.SubProtocol
}

func (m MsgMatcher) Matches(x interface{}) bool {
	return x.(p2pcommon.Message).Subprotocol() == m.sub
}

func (m MsgMatcher) String() string {
	return "matcher " + m.sub.String()
}

func Test_createMessage(t *testing.T) {
	type args struct {
		protocolID p2pcommon.SubProtocol
		msgBody    p2pcommon.MessageBody
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{"TStatus", args{protocolID: p2pcommon.StatusRequest, msgBody: &types.Status{Version: "11"}}, false},
		{"TGOAway", args{protocolID: p2pcommon.GoAway, msgBody: &types.GoAwayNotice{Message: "test"}}, false},
		{"TNil", args{protocolID: p2pcommon.StatusRequest, msgBody: nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createMessage(tt.args.protocolID, p2pcommon.NewMsgID(), tt.args.msgBody)
			if (got == nil) != tt.wantNil {
				t.Errorf("createMessage() = %v, want nil %v", got, tt.wantNil)
			}
			if got != nil && got.Subprotocol() != tt.args.protocolID {
				t.Errorf("status.ProtocolID = %v, want %v", got.Subprotocol(), tt.args.protocolID)
			}
		})
	}
}
