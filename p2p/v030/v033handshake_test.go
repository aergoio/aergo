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
	"sort"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
)

type fakeChainID struct {
	genID types.ChainID
	versions []uint64
}
func newFC(genID types.ChainID, vers... uint64) fakeChainID {
	sort.Sort(BlkNoDesc(vers))
	return fakeChainID{genID:genID, versions:vers}
}
func (f fakeChainID) getChainID(no types.BlockNo) *types.ChainID{
	cp := f.genID
	for i:=len(f.versions)-1; i>=0 ; i-- {
		if f.versions[i] <= no {
			cp.Version = int32(i+2)
		}
	}
	return &cp
}
type BlkNoDesc []uint64

func (a BlkNoDesc) Len() int           { return len(a) }
func (a BlkNoDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BlkNoDesc) Less(i, j int) bool { return a[i] < a[j] }

func TestV033VersionedHS_DoForOutbound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test")
	mockActor := p2pmock.NewMockActorService(ctrl)
	mockCA := p2pmock.NewMockChainAccessor(ctrl)
	mockPM := p2pmock.NewMockPeerManager(ctrl)

	fc := newFC(*myChainID, 10000,20000,dummyBlockHeight+100)
	localChainID := *fc.getChainID(dummyBlockHeight)
	localChainBytes, _ := localChainID.Bytes()
	oldChainID := fc.getChainID(10000)
	oldChainBytes, _ := oldChainID.Bytes()
	newChainID := fc.getChainID(600000)
	newChainBytes, _ := newChainID.Bytes()

 	diffBlockNo := dummyBlockHeight+100000
	dummyMeta := p2pcommon.PeerMeta{ID: samplePeerID, IPAddress: "dummy.aergo.io"}
	dummyAddr := dummyMeta.ToPeerAddress()
	mockPM.EXPECT().SelfMeta().Return(dummyMeta).AnyTimes()
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	mockActor.EXPECT().GetChainAccessor().Return(mockCA).AnyTimes()
	mockCA.EXPECT().GetBestBlock().Return(dummyBlock, nil).AnyTimes()
	dummyGenHash := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	diffGenesis := []byte{0xff, 0xfe, 0xfd, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	dummyStatusMsg := &types.Status{ChainID: localChainBytes, Sender: &dummyAddr, Genesis: dummyGenHash, BestBlockHash:dummyBlockHash, BestHeight:dummyBlockHeight}
	diffGenesisStatusMsg := &types.Status{ChainID: localChainBytes, Sender: &dummyAddr, Genesis: diffGenesis}
	nilGenesisStatusMsg := &types.Status{ChainID: localChainBytes, Sender: &dummyAddr, Genesis: nil}
	nilSenderStatusMsg := &types.Status{ChainID: localChainBytes, Sender: nil, Genesis: dummyGenHash}
	diffStatusMsg := &types.Status{ChainID: theirChainBytes, Sender: &dummyAddr, Genesis: dummyGenHash, BestHeight:diffBlockNo}
	olderStatusMsg := &types.Status{ChainID: oldChainBytes, Sender: &dummyAddr, Genesis: dummyGenHash, BestHeight:10000}
	newerStatusMsg := &types.Status{ChainID: newChainBytes, Sender: &dummyAddr, Genesis: dummyGenHash, BestHeight:600000}
	diffVersionStatusMsg := &types.Status{ChainID: newVerChainBytes, Sender: &dummyAddr, Genesis: dummyGenHash}

	tests := []struct {
		name       string
		readReturn *types.Status
		readError  error
		writeError error
		wantStatus *types.Status
		wantErr    bool
		wantGoAway bool
	}{
		{"TSuccess", dummyStatusMsg, nil, nil, dummyStatusMsg, false, false},
		{"TOldChain", olderStatusMsg, nil, nil, dummyStatusMsg, false, false},
		{"TNewChain", newerStatusMsg, nil, nil, dummyStatusMsg, false, false},
		{"TUnexpMsg", nil, nil, nil, nil, true, true},
		{"TRFail", dummyStatusMsg, fmt.Errorf("failed"), nil, nil, true, true},
		{"TRNoSender", nilSenderStatusMsg, nil, nil, nil, true, true},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true, false},
		{"TDiffChain", diffStatusMsg, nil, nil, nil, true, true},
		{"TNilGenesis", nilGenesisStatusMsg, nil, nil, nil, true, true},
		{"TDiffGenesis", diffGenesisStatusMsg, nil, nil, nil, true, true},
		{"TDiffChainVersion", diffVersionStatusMsg, nil, nil, nil, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyReader := p2pmock.NewMockReadWriteCloser(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)
			mockVM := p2pmock.NewMockVersionedManager(ctrl)
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
			mockVM.EXPECT().GetBestChainID().Return(myChainID).AnyTimes()
			mockVM.EXPECT().GetChainID(gomock.Any()).DoAndReturn(fc.getChainID).AnyTimes()

			h := NewV033VersionedHS(mockPM, mockActor, logger, mockVM, samplePeerID, dummyReader, dummyGenHash)
			h.msgRW = mockRW
			got, err := h.DoForOutbound(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.DoForOutbound() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.wantStatus != nil {
				if !reflect.DeepEqual(got.ChainID, tt.wantStatus.ChainID) {
					fmt.Printf("got:(%d) %s \n", len(got.ChainID), hex.EncodeToString(got.ChainID))
					fmt.Printf("got:(%d) %s \n", len(tt.wantStatus.ChainID), hex.EncodeToString(tt.wantStatus.ChainID))
					t.Errorf("PeerHandshaker.DoForOutbound() = %v, want %v", got.ChainID, tt.wantStatus.ChainID)
				}
			} else if !reflect.DeepEqual(got, tt.wantStatus) {
				t.Errorf("PeerHandshaker.DoForOutbound() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func TestV033VersionedHS_DoForInbound(t *testing.T) {
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

	dummyGenHash := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	diffGenHash := []byte{0xff, 0xfe, 0xfd, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	dummyStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr, Genesis: dummyGenHash}
	diffGenesisStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr, Genesis: diffGenHash}
	nilGenesisStatusMsg := &types.Status{ChainID: myChainBytes, Sender: &dummyAddr, Genesis: nil}
	nilSenderStatusMsg := &types.Status{ChainID: myChainBytes, Sender: nil, Genesis: dummyGenHash}
	diffStatusMsg := &types.Status{ChainID: theirChainBytes, Sender: &dummyAddr, Genesis: dummyGenHash}
	diffVersionStatusMsg := &types.Status{ChainID: newVerChainBytes, Sender: &dummyAddr, Genesis: dummyGenHash}
	tests := []struct {
		name       string
		readReturn *types.Status
		readError  error
		writeError error
		want       *types.Status
		wantErr    bool
		wantGoAway bool
	}{
		{"TSuccess", dummyStatusMsg, nil, nil, dummyStatusMsg, false, false},
		{"TUnexpMsg", nil, nil, nil, nil, true, true},
		{"TRFail", dummyStatusMsg, fmt.Errorf("failed"), nil, nil, true, true},
		{"TRNoSender", nilSenderStatusMsg, nil, nil, nil, true, true},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true, false},
		{"TDiffChain", diffStatusMsg, nil, nil, nil, true, true},
		{"TNilGenesis", nilGenesisStatusMsg, nil, nil, nil, true, true},
		{"TDiffGenesis", diffGenesisStatusMsg, nil, nil, nil, true, true},
		{"TDiffChainVersion", diffVersionStatusMsg, nil, nil, nil, true, true},
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
				if !reflect.DeepEqual(got.ChainID, tt.want.ChainID) {
					fmt.Printf("got:(%d) %s \n", len(got.ChainID), hex.EncodeToString(got.ChainID))
					fmt.Printf("got:(%d) %s \n", len(tt.want.ChainID), hex.EncodeToString(tt.want.ChainID))
					t.Errorf("PeerHandshaker.DoForInbound() = %v, want %v", got.ChainID, tt.want.ChainID)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.DoForInbound() = %v, want %v", got, tt.want)
			}
		})
	}
}
