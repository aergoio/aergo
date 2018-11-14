/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestP2P_GetBlocksChunk(t *testing.T) {
	sampleMsg := &message.GetBlockChunks{GetBlockInfos:message.GetBlockInfos{ToWhom:samplePeerID},TTL:time.Minute}

	// fail: cancel create receiver and return fail instantly
	mockPM := new(MockPeerManager)
	mockCtx := new(mockContext)
	mockPM.On("GetPeer", mock.AnythingOfType("peer.ID")).Return(nil, false)
	mockCtx.On("Respond", mock.AnythingOfType("*message.GetBlockChunksRsp"))
	ps := &P2P{}
	ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, ps, log.NewLogger("p2p"))
	ps.pm = mockPM

	ps.GetBlocksChunk(mockCtx, sampleMsg)

	mockCtx.AssertNumberOfCalls(t, "Respond", 1)


	// success case
	mockPM = new(MockPeerManager)
	mockCtx = new(mockContext)
	mockPeer := new(MockRemotePeer)
	mockMF := new(MockMoFactory)
	mockPM.On("GetPeer", mock.AnythingOfType("peer.ID")).Return(mockPeer, true)
	mockPeer.On("MF").Return(mockMF)
	mockPeer.On("sendMessage",mock.Anything)
	mockMF.On("newMsgBlockRequestOrder",mock.Anything, GetBlocksRequest, mock.AnythingOfType("*types.GetBlockRequest")).Return(dummyMo)

	ps = &P2P{}
	ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, ps, log.NewLogger("p2p"))
	ps.pm = mockPM
	ps.GetBlocksChunk(mockCtx, sampleMsg)

	mockCtx.AssertNotCalled(t, "Respond", mock.Anything)
	// verify that receiver start working.
	mockMF.AssertNumberOfCalls(t,"newMsgBlockRequestOrder",1)
	mockPeer.AssertNumberOfCalls(t, "sendMessage", 1)
}
