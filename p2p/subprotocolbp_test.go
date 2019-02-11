/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/golang/protobuf/proto"
)

func Test_blockProducedNoticeHandler_handle(t *testing.T) {
	dummyBlock := &types.Block{Hash:dummyBlockHash,
		Header:&types.BlockHeader{}, Body:&types.BlockBody{}}
	wrongBlock := &types.Block{Hash:nil,
		Header:&types.BlockHeader{}, Body:&types.BlockBody{}}
	type args struct {
		msg     p2pcommon.Message
		msgBody proto.Message
	}
	tests := []struct {
		name   string
		cached bool
		payloadBlk *types.Block

		smCalled bool
	}{
		// 1. normal case
		{"TSucc", false, dummyBlock, true},
		// 2. wrong notice (block data is missing)
		{"TW1", false, nil , false},
		// 2. wrong notice1 (invalid block data)
		{"TW2", false, wrongBlock, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockPeer := new(MockRemotePeer)
			mockPeer.On("ID").Return(dummyPeerID)
			mockPeer.On("Name").Return("16..aadecf@1")
			mockPeer.On("updateLastNotice",mock.Anything, mock.AnythingOfType("uint64")).Return(false)
			mockActor := new(MockActorService)
			mockSM := new(MockSyncManager)
			mockSM.On("HandleBlockProducedNotice",mock.Anything, mock.AnythingOfType("*types.Block"))

			msg := &V030Message{subProtocol:BlockProducedNotice, id: sampleMsgID}
			body := &types.BlockProducedNotice{Block:tt.payloadBlk}
			h := newBlockProducedNoticeHandler(mockPM, mockPeer, logger, mockActor, mockSM)
			h.handle(msg, body)
			if  tt.smCalled {
				mockSM.AssertNumberOfCalls(t, "HandleBlockProducedNotice", 1)
			} else {
				mockSM.AssertNumberOfCalls(t, "HandleBlockProducedNotice", 0)
			}
		})
	}
}
