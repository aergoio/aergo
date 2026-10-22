/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"bytes"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

type getHashRequestHandler struct {
	BaseMsgHandler
}

type getHashResponseHandler struct {
	BaseMsgHandler
}

// newBlockReqHandler creates handler for GetBlockRequest
func NewGetHashesReqHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *getHashRequestHandler {
	bh := &getHashRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.GetHashesRequest, pm: pm, peer: peer, actor: actor, logger: logger}}

	return bh
}

func (bh *getHashRequestHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetHashesRequest{})
}

func (bh *getHashRequestHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetHashesRequest)
	p2putil.DebugLogReceive(bh.logger, bh.protocol, msg.ID().String(), remotePeer, data)
	chainAccessor := bh.actor.GetChainAccessor()

	// check if requested too many hashes
	if data.Size > p2pcommon.MaxBlockResponseCount {
		resp := &types.GetHashesResponse{Status: types.ResultStatus_INVALID_ARGUMENT}
		remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetHashesResponse, resp))
		return
	}
	// check if remote peer has valid chain,
	// TODO also check if found prevBlock is on main chain or side chain, assume in main chain for now.
	prevHash, err := chainAccessor.GetHashByNo(data.PrevNumber)
	if err != nil || !bytes.Equal(prevHash, data.PrevHash) {
		resp := &types.GetHashesResponse{Status: types.ResultStatus_INVALID_ARGUMENT}
		remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetHashesResponse, resp))
		return
	}
	// decide total fetched size
	bestBlock, err := bh.actor.GetChainAccessor().GetBestBlock()
	if err != nil {
		resp := &types.GetHashesResponse{Status: types.ResultStatus_INTERNAL}
		remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetHashesResponse, resp))
		return
	}
	startNumber, endNumber, fetchSize := determineFetchSize(data.PrevNumber, bestBlock.Header.BlockNo, int(data.Size))
	if fetchSize <= 0 {
		resp := &types.GetHashesResponse{Status: types.ResultStatus_INTERNAL}
		remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetHashesResponse, resp))
		return
	}
	hashes := make([][]byte, fetchSize)
	cursorNo := endNumber
	for i := fetchSize - 1; i >= 0; i-- {
		hash, err := bh.actor.GetChainAccessor().GetHashByNo(cursorNo)
		if err != nil {
			resp := &types.GetHashesResponse{Status: types.ResultStatus_INTERNAL}
			remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetHashesResponse, resp))
			return
		}
		hashes[i] = hash
		cursorNo--
	}
	// check again if data is changed during fetch
	// check if reorg (or such like it) occurred and mainchain is changed during
	endHash, err := chainAccessor.GetHashByNo(endNumber)
	if err != nil || !bytes.Equal(endHash, hashes[fetchSize-1]) {
		resp := &types.GetHashesResponse{Status: types.ResultStatus_INTERNAL}
		remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetHashesResponse, resp))
		return
	}
	startBlock, err := chainAccessor.GetBlock(hashes[0])
	if err != nil || !bytes.Equal(startBlock.Header.PrevBlockHash, prevHash) || startBlock.Header.BlockNo != startNumber {
		resp := &types.GetHashesResponse{Status: types.ResultStatus_INTERNAL}
		remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetHashesResponse, resp))
		return
	}

	// generate response message
	resp := &types.GetHashesResponse{
		Hashes:  hashes,
		Status:  types.ResultStatus_OK,
		HasNext: false,
	}
	remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetHashesResponse, resp))
}

func determineFetchSize(prevNum, currentLast types.BlockNo, maxSize int) (types.BlockNo, types.BlockNo, int) {
	if currentLast <= prevNum {
		return 0, 0, -1
	}
	fetchSize := int(currentLast - prevNum)
	if fetchSize > maxSize {
		fetchSize = maxSize
	}

	return prevNum + 1, prevNum + types.BlockNo(fetchSize), fetchSize
}

// newBlockReqHandler creates handler for GetBlockRequest
func NewGetHashesRespHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *getHashResponseHandler {
	bh := &getHashResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.GetHashesResponse, pm: pm, peer: peer, actor: actor, logger: logger}}

	return bh
}

func (bh *getHashResponseHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetHashesResponse{})
}

func (bh *getHashResponseHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetHashesResponse)
	p2putil.DebugLogReceiveResponse(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), bh.peer, data)

	// locate request data and remove it if found
	remotePeer.GetReceiver(msg.OriginalID())(msg, data)
}

type getHashByNoRequestHandler struct {
	BaseMsgHandler
}

type getHashByNoResponseHandler struct {
	BaseMsgHandler
}

// newBlockReqHandler creates handler for GetBlockRequest
func NewGetHashByNoReqHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *getHashByNoRequestHandler {
	bh := &getHashByNoRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.GetHashByNoRequest, pm: pm, peer: peer, actor: actor, logger: logger}}

	return bh
}

func (bh *getHashByNoRequestHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetHashByNo{})
}

func (bh *getHashByNoRequestHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetHashByNo)
	p2putil.DebugLogReceive(bh.logger, bh.protocol, msg.ID().String(), remotePeer, data)
	chainAccessor := bh.actor.GetChainAccessor()

	// check if remote peer has valid chain,
	// TODO also check if found prevBlock is on main chain or side chain, assume in main chain for now.
	targetHash, err := chainAccessor.GetHashByNo(data.BlockNo)
	if err != nil {
		resp := &types.GetHashByNoResponse{Status: types.ResultStatus_NOT_FOUND}
		remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetHashByNoResponse, resp))
		return
	}

	// generate response message
	resp := &types.GetHashByNoResponse{
		Status:    types.ResultStatus_OK,
		BlockHash: targetHash,
	}
	remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetHashByNoResponse, resp))
}

// newBlockReqHandler creates handler for GetBlockRequest
func NewGetHashByNoRespHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *getHashByNoResponseHandler {
	bh := &getHashByNoResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.GetHashByNoResponse, pm: pm, peer: peer, actor: actor, logger: logger}}

	return bh
}

func (bh *getHashByNoResponseHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetHashByNoResponse{})
}

func (bh *getHashByNoResponseHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	data := msgBody.(*types.GetHashByNoResponse)
	p2putil.DebugLogReceiveResponse(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), bh.peer, data)

	// locate request data and remove it if found
	bh.peer.GetReceiver(msg.OriginalID())(msg, data)
}
