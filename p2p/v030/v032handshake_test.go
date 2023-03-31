/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package v030

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
)

func TestV032VersionedHS_DoForOutbound(t *testing.T) {
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
	dummyGenHash := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	diffGenesis := []byte{0xff, 0xfe, 0xfd, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	dummyStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr, Genesis: dummyGenHash, Version: dummyAddr.Version}
	succResult := &p2pcommon.HandshakeResult{Meta: dummyMeta}
	diffGenesisStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr, Genesis: diffGenesis, Version: dummyAddr.Version}
	nilGenesisStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr, Genesis: nil, Version: dummyAddr.Version}
	nilSenderStatusMsg := &types.Status{ChainID: myChainBytes, Sender: nil, Genesis: dummyGenHash, Version: dummyAddr.Version}
	diffStatusMsg := &types.Status{ChainID: theirChainBytes, Sender: &dummyAddr, Genesis: dummyGenHash, Version: dummyAddr.Version}
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
		{"TNilGenesis", nilGenesisStatusMsg, nil, nil, nil, true, true},
		{"TDiffGenesis", diffGenesisStatusMsg, nil, nil, nil, true, true},
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
			mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeError).AnyTimes()

			h := NewV032VersionedHS(mockPM, mockActor, logger, myChainID, samplePeerID, dummyReader, dummyGenHash)
			h.msgRW = mockRW
			got, err := h.DoForOutbound(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.DoForOutbound() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.want != nil {
				if !got.Meta.Equals(tt.want.Meta) {
					t.Errorf("PeerHandshaker.handshakeOutboundPeer() peerID = %v, want %v", got.Meta, tt.want.Meta)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.DoForOutbound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestV032VersionedHS_DoForInbound(t *testing.T) {
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

	dummyGenHash := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	diffGenHash := []byte{0xff, 0xfe, 0xfd, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	dummyStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr, Genesis: dummyGenHash, Version: dummyAddr.Version}
	succResult := &p2pcommon.HandshakeResult{Meta: dummyMeta}
	diffGenesisStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr, Genesis: diffGenHash, Version: dummyAddr.Version}
	nilGenesisStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr, Genesis: nil, Version: dummyAddr.Version}
	nilSenderStatusMsg := &types.Status{ChainID: myChainBytes, Sender: nil, Genesis: dummyGenHash, Version: dummyAddr.Version}
	diffStatusMsg := &types.Status{ChainID: theirChainBytes, Sender: &dummyAddr, Genesis: dummyGenHash, Version: dummyAddr.Version}
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
		{"TNilGenesis", nilGenesisStatusMsg, nil, nil, nil, true, true},
		{"TDiffGenesis", diffGenesisStatusMsg, nil, nil, nil, true, true},
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
			mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeError).AnyTimes()

			h := NewV032VersionedHS(mockPM, mockActor, logger, myChainID, samplePeerID, dummyReader, dummyGenHash)
			h.msgRW = mockRW
			got, err := h.DoForInbound(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.DoForInbound() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.want != nil {
				if !got.Meta.Equals(tt.want.Meta) {
					t.Errorf("PeerHandshaker.handshakeOutboundPeer() peerID = %v, want %v", got.Meta, tt.want.Meta)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.DoForInbound() = %v, want %v", got, tt.want)
			}
		})
	}
}
