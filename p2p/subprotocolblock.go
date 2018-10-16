/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

type blockRequestHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*blockRequestHandler)(nil)

type blockResponseHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*blockResponseHandler)(nil)

type listBlockHeadersRequestHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*listBlockHeadersRequestHandler)(nil)

type listBlockHeadersResponseHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*listBlockHeadersResponseHandler)(nil)

type newBlockNoticeHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*newBlockNoticeHandler)(nil)

type getMissingRequestHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*getMissingRequestHandler)(nil)

// newBlockReqHandler creates handler for GetBlockRequest
func newBlockReqHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *blockRequestHandler {
	bh := &blockRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlocksRequest, pm: pm, peer: peer, actor: actor, logger: logger}}

	return bh
}

func (bh *blockRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetBlockRequest{})
}

func (bh *blockRequestHandler) handle(msg Message, msgBody proto.Message) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockRequest)
	debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, len(data.Hashes))

	// find block info from chainservice
	idx := 0
	blockInfos := make([]*types.Block, 0, len(data.Hashes))
	for _, hash := range data.Hashes {
		foundBlock, err := extractBlockFromRequest(bh.actor.CallRequest(message.ChainSvc,
			&message.GetBlock{BlockHash: hash}))
		if err != nil || foundBlock == nil {
			continue
		}
		blockInfos = append(blockInfos, foundBlock)
		idx++
	}
	status := types.ResultStatus_OK
	if 0 == len(blockInfos) {
		status = types.ResultStatus_NOT_FOUND
	}

	// generate response message
	resp := &types.GetBlockResponse{
		Status: status,
		Blocks: blockInfos}

	remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetBlocksResponse, resp))
}

// newBlockRespHandler creates handler for GetBlockResponse
func newBlockRespHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *blockResponseHandler {
	bh := &blockResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlocksResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *blockResponseHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetBlockResponse{})
}

func (bh *blockResponseHandler) handle(msg Message, msgBody proto.Message) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockResponse)
	debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, len(data.Blocks))

	// locate request data and remove it if found
	remotePeer.consumeRequest(msg.ID())

	// got block
	bh.logger.Debug().Int("block_cnt", len(data.Blocks)).Msg("Request chainservice to add blocks")
	for _, block := range data.Blocks {
		bh.actor.SendRequest(message.ChainSvc, &message.AddBlock{PeerID: peerID, Block: block, Bstate: nil})
	}

}

// newListBlockReqHandler creates handler for GetBlockHeadersRequest
func newListBlockReqHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *listBlockHeadersRequestHandler {
	bh := &listBlockHeadersRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlockHeadersRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *listBlockHeadersRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetBlockHeadersRequest{})
}

func (bh *listBlockHeadersRequestHandler) handle(msg Message, msgBody proto.Message) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockHeadersRequest)
	debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, data)

	// find block info from chainservice
	maxFetchSize := min(1000, data.Size)
	idx := uint32(0)
	hashes := make([][]byte, 0, data.Size)
	headers := make([]*types.BlockHeader, 0, data.Size)
	if len(data.Hash) > 0 {
		hash := data.Hash
		for idx < maxFetchSize {
			foundBlock, err := extractBlockFromRequest(bh.actor.CallRequest(message.ChainSvc,
				&message.GetBlock{BlockHash: hash}))
			if err != nil || foundBlock == nil {
				break
			}
			hashes = append(hashes, foundBlock.BlockHash())
			headers = append(headers, getBlockHeader(foundBlock))
			idx++
			hash = foundBlock.Header.PrevBlockHash
			if len(hash) == 0 {
				break
			}
		}
	} else {
		end := types.BlockNo(0)
		if types.BlockNo(data.Height) >= types.BlockNo(maxFetchSize) {
			end = types.BlockNo(data.Height - uint64(maxFetchSize-1))
		}
		for i := types.BlockNo(data.Height); i >= end; i-- {
			foundBlock, err := extractBlockFromRequest(bh.actor.CallRequest(message.ChainSvc,
				&message.GetBlockByNo{BlockNo: i}))
			if err != nil || foundBlock == nil {
				break
			}
			hashes = append(hashes, foundBlock.BlockHash())
			headers = append(headers, getBlockHeader(foundBlock))
			idx++
		}
	}
	// generate response message
	resp := &types.GetBlockHeadersResponse{
		Hashes: hashes, Headers: headers,
		Status: types.ResultStatus_OK,
	}
	remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetBlockHeadersResponse, resp))
}

func getBlockHeader(blk *types.Block) *types.BlockHeader {
	return blk.Header
}

// newListBlockRespHandler creates handler for GetBlockHeadersResponse
func newListBlockRespHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *listBlockHeadersResponseHandler {
	bh := &listBlockHeadersResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlockHeadersResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *listBlockHeadersResponseHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetBlockHeadersResponse{})
}

func (bh *listBlockHeadersResponseHandler) handle(msg Message, msgBody proto.Message) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockHeadersResponse)
	debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, len(data.Hashes))

	// send block headers to blockchain service
	remotePeer.consumeRequest(msg.ID())

	// TODO: it's not used yet, but used in RPC and can be used in future performance tuning
}

// newNewBlockNoticeHandler creates handler for NewBlockNotice
func newNewBlockNoticeHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService, sm SyncManager) *newBlockNoticeHandler {
	bh := &newBlockNoticeHandler{BaseMsgHandler: BaseMsgHandler{protocol: NewBlockNotice, pm: pm, sm:sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *newBlockNoticeHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.NewBlockNotice{})
}

func (bh *newBlockNoticeHandler) handle(msg Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.NewBlockNotice)
	// remove to verbose log
	// debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, log.DoLazyEval(func() string { return enc.ToString(data.BlockHash) }))

	// lru cache can accept hashable key
	var hash BlockHash
	copy(hash[:], data.BlockHash)
	if ! remotePeer.updateBlkCache(hash) {
		bh.sm.HandleNewBlockNotice(remotePeer, hash, data)
	}
}

func max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}
func min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// TODO need to add comment
func (bh *getMissingRequestHandler) notifyBranchBlock(peer RemotePeer, hash message.BlockHash, blockno types.BlockNo) bool {
	bh.logger.Debug().Str(LogPeerID, peer.ID().Pretty()).Msg("Notifying branch block")

	// create message data
	req := &types.NewBlockNotice{
		BlockHash: hash,
		BlockNo:   uint64(blockno)}

	peer.sendMessage(peer.MF().newMsgRequestOrder(false, NewBlockNotice, req))
	return true
}

// newGetMissingReqHandler creates handler for GetMissingRequest
func newGetMissingReqHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *getMissingRequestHandler {
	bh := &getMissingRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetMissingRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *getMissingRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetMissingRequest{})
}

func (bh *getMissingRequestHandler) handle(msg Message, msgBody proto.Message) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer
	data := msgBody.(*types.GetMissingRequest)
	if bh.logger.IsDebugEnabled() {
		debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, bytesArrToString(data.Hashes))
	}

	// send to ChainSvc
	// find block info from chainservice
	rawResponse, err := bh.actor.CallRequest(
		message.ChainSvc, &message.GetMissing{Hashes: data.Hashes, StopHash: data.Stophash})
	if err != nil {
		bh.logger.Warn().Err(err).Msg("failed to get missing")

		return
	}
	v := rawResponse.(message.GetMissingRsp)
	missing := (*message.GetMissingRsp)(&v)

	// generate response message
	bh.logger.Debug().Str(LogPeerID, peerID.Pretty()).Str(LogMsgID, msg.ID().String()).Msg("Sending GetMssingRequest response")

	bh.sendMissingResp(remotePeer, msg.ID(), missing.Hashes)
	/*
		for i := 0; i < len(missing.Hashes); i++ {
			bh.notifyBranchBlock(remotePeer, missing.Hashes[i], missing.Blocknos[i])
		}
	*/
}

// replying chain tree
func (bh *getMissingRequestHandler) sendMissingResp(remotePeer RemotePeer, requestID MsgID, missing []message.BlockHash) {
	// find block info from chainservice
	blockInfos := make([]*types.Block, 0, len(missing))
	for _, hash := range missing {
		foundBlock, err := extractBlockFromRequest(bh.actor.CallRequest(message.ChainSvc,
			&message.GetBlock{BlockHash: hash}))
		if err != nil || foundBlock == nil {
			continue
		}
		blockInfos = append(blockInfos, foundBlock)
	}
	status := types.ResultStatus_OK
	if 0 == len(blockInfos) {
		status = types.ResultStatus_NOT_FOUND
	}

	// generate response message
	resp := &types.GetBlockResponse{
		Status: status,
		Blocks: blockInfos}

	// ???: have to check arguments
	remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(requestID, GetBlocksResponse, resp))
}
