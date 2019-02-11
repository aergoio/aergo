/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"crypto/sha256"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestGetHashRequestHandler_handle(t *testing.T) {
	baseHeight := uint64(110000)
	sampleSize := 21
	mainChainHashes := make([][]byte, sampleSize)
	sideChainHashes := make([][]byte, sampleSize)
	digest := sha256.New()
	for i:=0;i<sampleSize; i++ {
		digest.Write(uuid.Must(uuid.NewV4()).Bytes())
		mainChainHashes[i] = digest.Sum(nil)
		digest.Write(uuid.Must(uuid.NewV4()).Bytes())
		sideChainHashes[i] = digest.Sum(nil)
	}
	tests := []struct {
		name string
		inNum uint64
		inHash []byte
		inSize uint64

		firstChain [][]byte
		reorgIdx int
		lastChain [][]byte

		expectedStatus types.ResultStatus
		expectedHashCnt int
	}{
		// 1. success (exact prev and enough chaining)
		{"Tsucc", baseHeight, mainChainHashes[0], 20,
			mainChainHashes, 99999, mainChainHashes, types.ResultStatus_OK, 20},
		// 2. exact prev but smaller chaining
		{"TShorter", baseHeight, mainChainHashes[0], 20,
			mainChainHashes[:16], 99999, mainChainHashes[:16], types.ResultStatus_OK, 15},
		// 3. wrong prev
		{"TWrongPrev", baseHeight, sampleBlks[0], 20,
			mainChainHashes, 99999, mainChainHashes, types.ResultStatus_INVALID_ARGUMENT, 0},
		// 4. missing prev (smaller best block than prev)
		{"TMissingPrev", baseHeight+30, mainChainHashes[0], 20,
			mainChainHashes, 99999, mainChainHashes, types.ResultStatus_INVALID_ARGUMENT, 0},
		// 5. exact prev , but reorg from middle before first fetch
		{"TReorgBefore", baseHeight, mainChainHashes[0], 20,
			mainChainHashes, 0, append(append(make([][]byte,0,sampleSize),mainChainHashes[:5]...),sideChainHashes[5:]...), types.ResultStatus_OK, 20},
		// 6. exact prev , but changed (by such as reorg) during fetch
		{"TReorgMid", baseHeight, mainChainHashes[0], 20,
			mainChainHashes, 10, append(append(make([][]byte,0,sampleSize),mainChainHashes[:5]...),sideChainHashes[5:]...), types.ResultStatus_INTERNAL, 0},
		// 7. exact prev at first, but changed prev (and decent blocks also) before first fetch
		{"TReorgWhole", baseHeight, mainChainHashes[0], 20,
			mainChainHashes, 0, sideChainHashes, types.ResultStatus_INTERNAL, 0},
		// 7. exact prev at first, but changed prev (and decent blocks also) during fetch
		{"TReorgWhole2", baseHeight, mainChainHashes[0], 20,
			mainChainHashes, 10, sideChainHashes, types.ResultStatus_INTERNAL, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockPeer := new(MockRemotePeer)
			mockActor := new(MockActorService)
			dummyMF := new(testDoubleHashRespFactory)
			mockPeer.On("ID").Return(dummyPeerID)
			mockPeer.On("Name").Return("16..aadecf@1")
			mockPeer.On("MF").Return(dummyMF)
			mockPeer.On("sendMessage", mock.Anything)

			mockAcc := &testDoubleChainAccessor{firstChain:test.firstChain, lastChain:test.lastChain, baseHeight:baseHeight, reorgTiming:test.reorgIdx}
			mockActor.On("GetChainAccessor").Return(mockAcc)

			msg := &V030Message{subProtocol:GetHashesRequest, id: sampleMsgID}
			body := &types.GetHashesRequest{PrevNumber:test.inNum, PrevHash:test.inHash, Size:test.inSize}

			h := newGetHashesReqHandler(mockPM, mockPeer, logger, mockActor)
			h.handle(msg, body)

			// verify
			assert.Equal(t, test.expectedStatus.String(), dummyMF.lastStatus.String())
			if test.expectedStatus == types.ResultStatus_OK {
				assert.Equal(t, test.expectedHashCnt, len(dummyMF.lastResp.Hashes) )
			}
			// send only one response whether success or not
			mockPeer.AssertNumberOfCalls(t, "sendMessage",1)
		})
	}
}

type testDoubleHashRespFactory struct {
	v030MOFactory
	lastResp *types.GetHashesResponse
	lastStatus types.ResultStatus
}
func (f *testDoubleHashRespFactory) newMsgResponseOrder(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message pbMessage) msgOrder {
	f.lastResp = message.(*types.GetHashesResponse)
	f.lastStatus = f.lastResp.Status
	return f.v030MOFactory.newMsgResponseOrder(reqID, protocolID, message)
}


func TestGetHashByNoRequestHandler_handle(t *testing.T) {
	baseHeight := uint64(110000)
	wrongHeight := uint64(21531535)
	tests := []struct {
		name string
		inNum uint64

		expectedStatus types.ResultStatus
	}{
		// 1. success (exact prev and enough chaining)
		{"Tsucc", baseHeight, types.ResultStatus_OK},
		// 2. exact prev but smaller chaining
		{"TMissing", wrongHeight, types.ResultStatus_NOT_FOUND},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockPeer := new(MockRemotePeer)
			mockActor := new(MockActorService)
			dummyMF := new(testDoubleHashByNoRespFactory)
			mockPeer.On("ID").Return(dummyPeerID)
			mockPeer.On("MF").Return(dummyMF)
			mockPeer.On("Name").Return("16..aadecf@1")
			mockPeer.On("sendMessage", mock.Anything)

			mockAcc := new(MockChainAccessor)
			mockActor.On("GetChainAccessor").Return(mockAcc)
			mockAcc.On("GetHashByNo", baseHeight).Return(sampleBlks[0], nil)
			mockAcc.On("GetHashByNo", wrongHeight).Return(nil, &chain.ErrNoBlock{})
			msg := &V030Message{subProtocol:GetHashByNoRequest, id: sampleMsgID}
			body := &types.GetHashByNo{BlockNo:test.inNum}

			h := newGetHashByNoReqHandler(mockPM, mockPeer, logger, mockActor)
			h.handle(msg, body)

			// verify
			assert.Equal(t, test.expectedStatus.String(), dummyMF.lastStatus.String())
			// send only one response whether success or not
			mockPeer.AssertNumberOfCalls(t, "sendMessage",1)
			mockAcc.AssertNumberOfCalls(t, "GetHashByNo", 1)
		})
	}
}

type testDoubleHashByNoRespFactory struct {
	v030MOFactory
	lastResp *types.GetHashByNoResponse
	lastStatus types.ResultStatus
}
func (f *testDoubleHashByNoRespFactory) newMsgResponseOrder(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message pbMessage) msgOrder {
	f.lastResp = message.(*types.GetHashByNoResponse)
	f.lastStatus = f.lastResp.Status
	return f.v030MOFactory.newMsgResponseOrder(reqID, protocolID, message)
}


// emulate db
type testDoubleChainAccessor struct {
	MockChainAccessor
	baseHeight types.BlockNo

	callCount int
	reorgTiming int
	firstChain [][]byte
	lastChain [][]byte
}

func (a *testDoubleChainAccessor) getChain() [][]byte {
	if a.callCount <= a.reorgTiming {
		return a.firstChain
	} else {
		return a.lastChain
	}
}

func (a *testDoubleChainAccessor) GetBestBlock() (*types.Block, error) {
	mychain := a.getChain()
	idx := len(mychain)-1
	return &types.Block{Hash:mychain[idx],Header:&types.BlockHeader{BlockNo:a.baseHeight+types.BlockNo(idx)}}, nil
}
// GetBlock return block of blockHash. It return nil and error if not found block of that hash or there is a problem in db store
func (a *testDoubleChainAccessor) GetBlock(blockHash []byte) (*types.Block, error) {
	mychain := a.getChain()
	for i, hash := range mychain {
		if bytes.Equal(hash, blockHash) {
			prevHash := []byte(nil)
			if i>0 {
				prevHash = mychain[i-1]
			}
			number := a.baseHeight + types.BlockNo(i)
			return &types.Block{Hash:hash,Header:&types.BlockHeader{BlockNo:number, PrevBlockHash:prevHash}}, nil
		}
	}
	return nil, chain.ErrNoBlock{}
}
// GetHashByNo returns hash of block. It return nil and error if not found block of that number or there is a problem in db store
func (a *testDoubleChainAccessor) GetHashByNo(blockNo types.BlockNo) ([]byte, error) {
	mychain := a.getChain()
	a.callCount++
	idx := blockNo - a.baseHeight
	if idx < 0 || idx >= types.BlockNo(len(mychain)) {
		return nil, chain.ErrNoBlock{}
	}
	return mychain[int(idx)], nil
}

