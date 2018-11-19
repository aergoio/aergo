/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

type getHashRequestHandler struct {
	BaseMsgHandler
}

type getHashResponseHandler struct {
	BaseMsgHandler
}


// newBlockReqHandler creates handler for GetBlockRequest
func newGetHashesReqHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *getHashRequestHandler {
	bh := &getHashRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetHashesRequest, pm: pm, peer: peer, actor: actor, logger: logger}}

	return bh
}

func (bh *getHashRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetHashesRequest{})
}

func (bh *getHashRequestHandler) handle(msg Message, msgBody proto.Message) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer
	data := msgBody.(*types.GetHashesRequest)
	debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, data)
	chainAccessor := bh.actor.GetChainAccessor()

	// check if requested too many hashes
	if data.Size > MaxBlockResponseCount {
		resp := &types.GetHashesResponse{Status:types.ResultStatus_INVALID_ARGUMENT}
		remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetHashesResponse, resp))
		return
	}
	// check if remote peer has valid chain,
	// TODO also check if found prevBlock is on main chain or side chain, assume in main chain for now.
	prevHash, err := chainAccessor.GetHashByNo(data.PrevNumber)
	if err != nil || !bytes.Equal(prevHash, data.PrevHash) {
		resp := &types.GetHashesResponse{Status:types.ResultStatus_INVALID_ARGUMENT}
		remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetHashesResponse, resp))
		return
	}
	// decide total fetched size
	bestBlock, err := bh.actor.GetChainAccessor().GetBestBlock()
	if err != nil {
		resp := &types.GetHashesResponse{Status:types.ResultStatus_INTERNAL}
		remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetHashesResponse, resp))
		return
	}
	startNumber, endNumber, fetchSize := determineFetchSize(data.PrevNumber, bestBlock.Header.BlockNo, int(data.Size))
	if fetchSize <= 0 {
		resp := &types.GetHashesResponse{Status:types.ResultStatus_INTERNAL}
		remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetHashesResponse, resp))
		return
	}
	hashes := make([][]byte, fetchSize)
	cursorNo := endNumber
	for i:=fetchSize-1; i >= 0; i-- {
		hash, err := bh.actor.GetChainAccessor().GetHashByNo(cursorNo)
		if err != nil {
			resp := &types.GetHashesResponse{Status: types.ResultStatus_INTERNAL}
			remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetHashesResponse, resp))
			return
		}
		hashes[i] = hash
		cursorNo--
	}
	// check again if data is changed during fetch
	// check if reorg (or such like it) occured and mainchain is changed during
	endHash, err := chainAccessor.GetHashByNo(endNumber)
	if err != nil || !bytes.Equal(endHash, hashes[fetchSize-1]) {
		resp := &types.GetHashesResponse{Status:types.ResultStatus_INTERNAL}
		remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetHashesResponse, resp))
		return
	}
	startBlock, err := chainAccessor.GetBlock(hashes[0])
	if err != nil || !bytes.Equal(startBlock.Header.PrevBlockHash, prevHash) || startBlock.Header.BlockNo != startNumber {
		resp := &types.GetHashesResponse{Status:types.ResultStatus_INTERNAL}
		remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetHashesResponse, resp))
		return
	}

	// generate response message
	resp := &types.GetHashesResponse{
		Hashes: hashes,
		Status: types.ResultStatus_OK,
		HasNext:false,
	}
	remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetHashesResponse, resp))
}

func determineFetchSize(prevNum, currentLast types.BlockNo, maxSize int) (types.BlockNo, types.BlockNo, int) {
	if currentLast <= prevNum {
		return 0,0,-1
	}
	fetchSize := int(currentLast - prevNum)
	if fetchSize > maxSize {
		fetchSize = maxSize
	}

	return prevNum+1, prevNum+types.BlockNo(fetchSize), fetchSize
}

// newBlockReqHandler creates handler for GetBlockRequest
func newGetHashesRespHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *getHashResponseHandler {
	bh := &getHashResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetHashesResponse, pm: pm, peer: peer, actor: actor, logger: logger}}

	return bh
}

func (bh *getHashResponseHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetHashesResponse{})
}


func (bh *getHashResponseHandler) handle(msg Message, msgBody proto.Message) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer
	data := msgBody.(*types.GetHashesResponse)
	debugLogReceiveResponseMsg(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), peerID, fmt.Sprintf("blk_cnt=%d,hasNext=%t",len(data.Hashes),data.HasNext) )

	// locate request data and remove it if found
	remotePeer.GetReceiver(msg.OriginalID())(msg, data)
}
