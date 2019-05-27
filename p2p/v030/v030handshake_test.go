/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package v030

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/p2p/p2pkey"
	"reflect"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/p2p/subproto"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
)

var (
	myChainID, theirChainID       *types.ChainID
	myChainBytes, theirChainBytes []byte
	samplePeerID, _                      = types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	dummyBlockHash, _                    = hex.DecodeString("4f461d85e869ade8a0544f8313987c33a9c06534e50c4ad941498299579bd7ac")
	dummyBlockHeight              uint64 = 100215
)

func init() {
	myChainID = types.NewChainID()
	myChainID.Magic = "itSmain1"
	myChainBytes, _ = myChainID.Bytes()

	theirChainID = types.NewChainID()
	theirChainID.Read(myChainBytes)
	theirChainID.Magic = "itsdiff2"
	theirChainBytes, _ = theirChainID.Bytes()

	sampleKeyFile := "../../test/sample.key"
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

	dummyMeta := p2pcommon.PeerMeta{ID: samplePeerID, IPAddress: "dummy.aergo.io"}
	dummyAddr := dummyMeta.ToPeerAddress()
	mockPM.EXPECT().SelfMeta().Return(dummyMeta).AnyTimes()
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	mockActor.EXPECT().GetChainAccessor().Return(mockCA).AnyTimes()
	mockCA.EXPECT().GetBestBlock().Return(dummyBlock, nil).AnyTimes()

	dummyStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr}
	nilSenderStatusMsg := &types.Status{ChainID: myChainBytes, Sender: nil}
	diffStatusMsg := &types.Status{ChainID: theirChainBytes, Sender: &dummyAddr}
	tests := []struct {
		name       string
		readReturn *types.Status
		readError  error
		writeError error
		want       *types.Status
		wantErr    bool
	}{
		{"TSuccess", dummyStatusMsg, nil, nil, dummyStatusMsg, false},
		{"TUnexpMsg", nil, nil, nil, nil, true},
		{"TRFail", dummyStatusMsg, fmt.Errorf("failed"), nil, nil, true},
		{"TRNoSender", nilSenderStatusMsg, nil, nil, nil, true},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true},
		{"TDiffChain", diffStatusMsg, nil, nil, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyReader := p2pmock.NewMockReader(ctrl)
			dummyWriter := p2pmock.NewMockWriter(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)

			var containerMsg *p2pcommon.MessageValue
			if tt.readReturn != nil {
				containerMsg = p2pcommon.NewSimpleMsgVal(subproto.StatusRequest, p2pcommon.NewMsgID())
				statusBytes, _ := p2putil.MarshalMessageBody(tt.readReturn)
				containerMsg.SetPayload(statusBytes)
			} else {
				containerMsg = p2pcommon.NewSimpleMsgVal(subproto.AddressesRequest, p2pcommon.NewMsgID())
			}
			mockRW.EXPECT().ReadMsg().Return(containerMsg, tt.readError).AnyTimes()
			mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeError).AnyTimes()

			h := NewV030StateHS(mockPM, mockActor, logger, myChainID, samplePeerID, dummyReader, dummyWriter)
			h.msgRW = mockRW
			got, err := h.DoForOutbound(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.want != nil {
				if !reflect.DeepEqual(got.ChainID, tt.want.ChainID) {
					fmt.Printf("got:(%d) %s \n", len(got.ChainID), hex.EncodeToString(got.ChainID))
					fmt.Printf("got:(%d) %s \n", len(tt.want.ChainID), hex.EncodeToString(tt.want.ChainID))
					t.Errorf("PeerHandshaker.handshakeOutboundPeer() = %v, want %v", got.ChainID, tt.want.ChainID)
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

	dummyMeta := p2pcommon.PeerMeta{ID: samplePeerID, IPAddress: "dummy.aergo.io"}
	dummyAddr := dummyMeta.ToPeerAddress()
	mockPM.EXPECT().SelfMeta().Return(dummyMeta).AnyTimes()
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	//dummyBlkRsp := message.GetBestBlockRsp{Block: dummyBlock}
	mockActor.EXPECT().GetChainAccessor().Return(mockCA).AnyTimes()
	mockCA.EXPECT().GetBestBlock().Return(dummyBlock, nil).AnyTimes()

	dummyStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr}
	nilSenderStatusMsg := &types.Status{ChainID: myChainBytes, Sender: nil}
	diffStatusMsg := &types.Status{ChainID: theirChainBytes, Sender: &dummyAddr}
	tests := []struct {
		name       string
		readReturn *types.Status
		readError  error
		writeError error
		want       *types.Status
		wantErr    bool
	}{
		{"TSuccess", dummyStatusMsg, nil, nil, dummyStatusMsg, false},
		{"TUnexpMsg", nil, nil, nil, nil, true},
		{"TRFail", dummyStatusMsg, fmt.Errorf("failed"), nil, nil, true},
		{"TRNoSender", nilSenderStatusMsg, nil, nil, nil, true},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true},
		{"TDiffChain", diffStatusMsg, nil, nil, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyReader := p2pmock.NewMockReader(ctrl)
			dummyWriter := p2pmock.NewMockWriter(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)

			containerMsg := &p2pcommon.MessageValue{}
			if tt.readReturn != nil {
				containerMsg = p2pcommon.NewSimpleMsgVal(subproto.StatusRequest, p2pcommon.NewMsgID())
				statusBytes, _ := p2putil.MarshalMessageBody(tt.readReturn)
				containerMsg.SetPayload(statusBytes)
			} else {
				containerMsg = p2pcommon.NewSimpleMsgVal(subproto.AddressesRequest, p2pcommon.NewMsgID())
			}

			mockRW.EXPECT().ReadMsg().Return(containerMsg, tt.readError).AnyTimes()
			mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeError).AnyTimes()

			h := NewV030StateHS(mockPM, mockActor, logger, myChainID, samplePeerID, dummyReader, dummyWriter)
			h.msgRW = mockRW
			got, err := h.DoForInbound(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.handshakeInboundPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.want != nil {
				if !reflect.DeepEqual(got.ChainID, tt.want.ChainID) {
					fmt.Printf("got:(%d) %s \n", len(got.ChainID), hex.EncodeToString(got.ChainID))
					fmt.Printf("got:(%d) %s \n", len(tt.want.ChainID), hex.EncodeToString(tt.want.ChainID))
					t.Errorf("PeerHandshaker.handshakeOutboundPeer() = %v, want %v", got.ChainID, tt.want.ChainID)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.handshakeInboundPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}
