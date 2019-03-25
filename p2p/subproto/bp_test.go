/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-peer"
	"testing"
)

func Test_blockProducedNoticeHandler_handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.subproto")
	dummyBlockHash, _ := enc.ToBytes("v6zbuQ4aVSdbTwQhaiZGp5pcL5uL55X3kt2wfxor5W6")
	var dummyPeerID, _ = peer.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")

	dummyBlock := &types.Block{Hash: dummyBlockHash,
		Header: &types.BlockHeader{}, Body: &types.BlockBody{}}
	wrongBlock := &types.Block{Hash: nil,
		Header: &types.BlockHeader{}, Body: &types.BlockBody{}}
	type args struct {
		msg     p2pcommon.Message
		msgBody proto.Message
	}
	tests := []struct {
		name       string
		cached     bool
		payloadBlk *types.Block

		syncmanagerCallCnt int
	}{
		// 1. normal case.
		{"TSucc", false, dummyBlock, 1},
		// 2. wrong notice (block data is missing)
		{"TW1", false, nil, 0},
		// 2. wrong notice1 (invalid block data)
		{"TW2", false, wrongBlock, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().Name().Return("16..aadecf@1").AnyTimes()
			mockPeer.EXPECT().UpdateLastNotice(dummyBlockHash, gomock.Any()).Times(tt.syncmanagerCallCnt)
			mockCA := p2pmock.NewMockChainAccessor(ctrl)
			mockActor.EXPECT().GetChainAccessor().Return(mockCA).MaxTimes(1)

			mockSM := p2pmock.NewMockSyncManager(ctrl)
			mockSM.EXPECT().HandleBlockProducedNotice(gomock.Any(), gomock.AssignableToTypeOf(&types.Block{})).Times(tt.syncmanagerCallCnt)

			dummyMsg :=&testMessage{id:p2pcommon.NewMsgID(), subProtocol:BlockProducedNotice}
			body := &types.BlockProducedNotice{Block: tt.payloadBlk}
			h := NewBlockProducedNoticeHandler(mockPM, mockPeer, logger, mockActor, mockSM)
			h.Handle(dummyMsg, body)
		})
	}
}
