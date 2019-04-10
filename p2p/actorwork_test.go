/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/p2p/subproto"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/golang/mock/gomock"
)

func TestP2P_GetBlocksChunk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleMsg := &message.GetBlockChunks{GetBlockInfos: message.GetBlockInfos{ToWhom: samplePeerID}, TTL: time.Minute}

	// fail: cancel create receiver and return fail instantly
	mockPM := p2pmock.NewMockPeerManager(ctrl)
	mockPM.EXPECT().GetPeer(gomock.Any()).Return(nil, false)
	mockCtx := p2pmock.NewMockContext(ctrl)
	mockCtx.EXPECT().Respond(gomock.Any()).Times(1)
	ps := &P2P{}
	ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, ps, log.NewLogger("p2p"))
	ps.pm = mockPM

	ps.GetBlocksChunk(mockCtx, sampleMsg)

	// success case
	mockPM = p2pmock.NewMockPeerManager(ctrl)
	mockCtx = p2pmock.NewMockContext(ctrl)
	mockPeer := p2pmock.NewMockRemotePeer(ctrl)
	mockMF := p2pmock.NewMockMoFactory(ctrl)

	mockPM.EXPECT().GetPeer(gomock.Any()).Return(mockPeer, true)
	mockCtx.EXPECT().Respond(gomock.Any()).Times(0)
	mockPeer.EXPECT().MF().Return(mockMF)
	mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)

	dummyMo = createDummyMo(ctrl)
	mockMF.EXPECT().NewMsgBlockRequestOrder(gomock.Any(), subproto.GetBlocksRequest, gomock.Any()).
		Return(dummyMo).Times(1)

	ps = &P2P{}
	ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, ps, log.NewLogger("p2p"))
	ps.pm = mockPM
	ps.GetBlocksChunk(mockCtx, sampleMsg)

	//mockCtx.AssertNotCalled(t, "Respond", mock.Anything)
	// verify that receiver start working.
	//mockMF.AssertNumberOfCalls(t, "newMsgBlockRequestOrder", 1)
	//mockPeer.AssertNumberOfCalls(t, "sendMessage", 1)
}

func TestP2P_GetBlockHashByNo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleMsg := &message.GetHashByNo{ToWhom: samplePeerID, BlockNo: uint64(111111)}

	// fail: cancel create receiver and return fail instantly
	mockPM := p2pmock.NewMockPeerManager(ctrl)
	mockCtx := p2pmock.NewMockContext(ctrl)
	mockPM.EXPECT().GetPeer(gomock.Any()).Return(nil, false)
	mockCtx.EXPECT().Respond(gomock.Any()).Times(1)
	ps := &P2P{}
	ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, ps, log.NewLogger("p2p"))
	ps.pm = mockPM

	ps.GetBlockHashByNo(mockCtx, sampleMsg)

	// success case
	mockPM = p2pmock.NewMockPeerManager(ctrl)
	mockCtx = p2pmock.NewMockContext(ctrl)
	mockPeer := p2pmock.NewMockRemotePeer(ctrl)
	mockMF := p2pmock.NewMockMoFactory(ctrl)

	mockCtx.EXPECT().Respond(gomock.Any()).Times(0)
	mockPM.EXPECT().GetPeer(gomock.Any()).Return(mockPeer, true)
	mockPeer.EXPECT().MF().Return(mockMF)
	mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)

	dummyMo = createDummyMo(ctrl)
	mockMF.EXPECT().NewMsgBlockRequestOrder(gomock.Any(), subproto.GetHashByNoRequest, gomock.Any()).
		Return(dummyMo).Times(1)

	ps = &P2P{}
	ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, ps, log.NewLogger("p2p"))
	ps.pm = mockPM
	ps.GetBlockHashByNo(mockCtx, sampleMsg)
}
