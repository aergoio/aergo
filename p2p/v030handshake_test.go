/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/p2p/subproto"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
)

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

	logger = log.NewLogger("test")
	mockActor := p2pmock.NewMockActorService(ctrl)
	mockCA := p2pmock.NewMockChainAccessor(ctrl)
	mockPM := p2pmock.NewMockPeerManager(ctrl)

	dummyMeta := p2pcommon.PeerMeta{ID: dummyPeerID, IPAddress: "dummy.aergo.io"}
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

			containerMsg := &V030Message{}
			if tt.readReturn != nil {
				containerMsg.subProtocol = subproto.StatusRequest
				statusBytes, _ := p2putil.MarshalMessage(tt.readReturn)
				containerMsg.payload = statusBytes
			} else {
				containerMsg.subProtocol = subproto.AddressesRequest
			}
			mockRW.EXPECT().ReadMsg().Return(containerMsg, tt.readError).AnyTimes()
			mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeError).AnyTimes()

			h := newV030StateHS(mockPM, mockActor, logger, myChainID, samplePeerID, dummyReader, dummyWriter)
			h.msgRW = mockRW
			got, err := h.doForOutbound()
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
	logger = log.NewLogger("test")
	mockActor := p2pmock.NewMockActorService(ctrl)
	mockCA := p2pmock.NewMockChainAccessor(ctrl)
	mockPM := p2pmock.NewMockPeerManager(ctrl)

	dummyMeta := p2pcommon.PeerMeta{ID: dummyPeerID, IPAddress: "dummy.aergo.io"}
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

			containerMsg := &V030Message{}
			if tt.readReturn != nil {
				containerMsg.subProtocol = subproto.StatusRequest
				statusBytes, _ := p2putil.MarshalMessage(tt.readReturn)
				containerMsg.payload = statusBytes
			} else {
				containerMsg.subProtocol = subproto.AddressesRequest
			}

			mockRW.EXPECT().ReadMsg().Return(containerMsg, tt.readError).AnyTimes()
			mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeError).AnyTimes()

			h := newV030StateHS(mockPM, mockActor, logger, myChainID, samplePeerID, dummyReader, dummyWriter)
			h.msgRW = mockRW
			got, err := h.doForInbound()
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
