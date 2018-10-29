/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	"time"
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

type getAncestorRequestHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*getAncestorRequestHandler)(nil)

type getAncestorResponseHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*getAncestorResponseHandler)(nil)

// newBlockReqHandler creates handler for GetBlockRequest
func newBlockReqHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *blockRequestHandler {
	bh := &blockRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlocksRequest, pm: pm, peer: peer, actor: actor, logger: logger}}

	return bh
}

func (bh *blockRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetBlockRequest{})
}

const (
	EmptyGetBlockResponseSize = 12 // roughly estimated maximum size if element is full
)

func (bh *blockRequestHandler) handle(msg Message, msgBody proto.Message) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockRequest)
	debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, len(data.Hashes))

	// find block info from chainservice
	idx := 0
	status := types.ResultStatus_OK
	blockInfos := make([]*types.Block, 0, 10)
	// TODO consider to make async if deadlock with remote peer can occurs
	// NOTE size estimation is tied to protobuf3 it should be changed when protobuf is changed.
	payloadSize := EmptyGetBlockResponseSize
	var blockSize, fieldSize int
	for _, hash := range data.Hashes {
		foundBlock, err := bh.actor.GetChainAccessor().GetBlock(hash)
		if err != nil || foundBlock == nil {
			continue
		}
		blockSize = proto.Size(foundBlock)
		fieldSize = blockSize + calculateFieldDescSize(blockSize)
		if (payloadSize + fieldSize) > MaxPayloadLength {
			// send partial list
			resp := &types.GetBlockResponse{
				Status: status,
				Blocks: blockInfos, HasNext:true}
			bh.logger.Debug().Int(LogBlkCount, len(blockInfos)).Str("req_id",msg.ID().String()).Msg("Sending partial response")
			remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetBlocksResponse, resp))
			// reset list
			blockInfos = make([]*types.Block, 0, 10)
			payloadSize = EmptyGetBlockResponseSize
		}
		blockInfos = append(blockInfos, foundBlock)
		payloadSize += fieldSize
		idx++
	}
	// send remained blocks
	if 0 == idx {
		status = types.ResultStatus_NOT_FOUND
	}
	// generate response message
	resp := &types.GetBlockResponse{
		Status: status,
		Blocks: blockInfos,HasNext:false}

	bh.logger.Debug().Int(LogBlkCount, len(blockInfos)).Str("req_id",msg.ID().String()).Msg("Sending last part response")
	remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetBlocksResponse, resp))
}

// newBlockRespHandler creates handler for GetBlockResponse
func newBlockRespHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService, sm SyncManager) *blockResponseHandler {
	bh := &blockResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlocksResponse, pm: pm, sm:sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *blockResponseHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetBlockResponse{})
}

func (bh *blockResponseHandler) handle(msg Message, msgBody proto.Message) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockResponse)
	debugLogReceiveResponseMsg(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), peerID, fmt.Sprintf("blk_cnt=%d,hasNext=%t",len(data.Blocks),data.HasNext) )

	// locate request data and remove it if found
	remotePeer.consumeRequest(msg.ID())
	if data.Status != types.ResultStatus_OK || len(data.Blocks) == 0 {
		return
	}
	bh.sm.HandleGetBlockResponse(remotePeer, msg, data)
}

// newListBlockHeadersReqHandler creates handler for GetBlockHeadersRequest
func newListBlockHeadersReqHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *listBlockHeadersRequestHandler {
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
	maxFetchSize := min(MaxBlockHeaderResponseCount, data.Size)
	idx := uint32(0)
	hashes := make([][]byte, 0, data.Size)
	headers := make([]*types.BlockHeader, 0, data.Size)
	if len(data.Hash) > 0 {
		hash := data.Hash
		for idx < maxFetchSize {
			foundBlock, err := bh.actor.GetChainAccessor().GetBlock(hash)
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
			foundBlock, err := extractBlockFromRequest(bh.actor.CallRequestDefaultTimeout(message.ChainSvc,
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
	bh := &newBlockNoticeHandler{BaseMsgHandler: BaseMsgHandler{protocol: NewBlockNotice, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *newBlockNoticeHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.NewBlockNotice{})
}

func (bh *newBlockNoticeHandler) handle(msg Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.NewBlockNotice)
	// remove to verbose log
	// debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, log.DoLazyEval(func() string { return enc.ToString(data.BlkHash) }))

	// lru cache can accept hashable key
	var hash BlkHash
	copy(hash[:], data.BlockHash)
	if !remotePeer.updateBlkCache(hash, data) {
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
	rawResponse, err := bh.actor.CallRequestDefaultTimeout(
		message.ChainSvc, &message.GetMissing{Hashes: data.Hashes, StopHash: data.Stophash})
	if err != nil {
		bh.logger.Warn().Err(err).Msg("failed to get missing")

		return
	}
	v := rawResponse.(message.GetMissingRsp)
	missingInfo := (*message.GetMissingRsp)(&v)

	if missingInfo.TopMatched == nil {
		// TODO process that internal error or remote is different chain, not just ignore

		return
	}
	// generate response message
	bh.logger.Debug().Str(LogPeerID, peerID.Pretty()).Str(LogMsgID, msg.ID().String()).Uint64("from_no", missingInfo.TopNumber).Uint64("to_no", missingInfo.StopNumber).Msg("Sending GetMissingRequest response")

	bh.sendMissingResp(remotePeer, msg.ID(), missingInfo)
	/*
		for i := 0; i < len(missing.Hashes); i++ {
			bh.notifyBranchBlock(remotePeer, missing.Hashes[i], missing.Blocknos[i])
		}
	*/
}

// replying chain tree
func (bh *getMissingRequestHandler) sendMissingResp(remotePeer RemotePeer, requestID MsgID, missing *message.GetMissingRsp) {
	// To get around load issues. Split message by byte size and block count. and limit 10 messages at a single missing handling
	if missing.StopNumber <= missing.TopNumber {
		return
	}
	totalCount := missing.StopNumber - missing.TopNumber

	// limit block count in single message
	sliceCap := MaxBlockResponseCount
	if totalCount < uint64(sliceCap) {
		sliceCap = int(totalCount)
	}

	defaultMsgTimeout := time.Second * 10

	// TODO very similar with blockRequestHandler.handle() consider refactoring
	// find block info from chainservice
	idx := 0
	msgSentCount := 0
	status := types.ResultStatus_OK
	blockInfos := make([]*types.Block, 0, sliceCap)
	payloadSize := EmptyGetBlockResponseSize
	var blockSize, fieldSize int
	for i := missing.TopNumber + 1; i <= missing.StopNumber; i++ {
		foundBlock, err := extractBlockFromRequest(bh.actor.CallRequestDefaultTimeout(message.ChainSvc,
			&message.GetBlockByNo{BlockNo: i}))
		if err != nil || foundBlock == nil {
			// the block get from getMissing must exists. this error is fatal.
			bh.logger.Warn().Err(err).Uint64("blk_number", i).Str("req_id", requestID.String()).Msg("failed to get block while processing getMissing")
			return
		}
		blockSize = proto.Size(foundBlock)
		fieldSize = blockSize + calculateFieldDescSize(blockSize)
		if len(blockInfos) >= sliceCap || (payloadSize+fieldSize) > MaxPayloadLength {
			msgSentCount++
			// send partial list
			resp := &types.GetBlockResponse{
				Status: status,
				Blocks: blockInfos,
				HasNext:true,
				//HasNext:msgSentCount<MaxResponseSplitCount, // always have nextItem ( see foundBlock) but msg count limit will affect
			}
			bh.logger.Debug().Uint64("first_blk_number", blockInfos[0].Header.GetBlockNo()).Int(LogBlkCount, len(blockInfos)).Str("req_id",requestID.String()).Msg("Sending partial getMissing response")
			err := remotePeer.sendAndWaitMessage(remotePeer.MF().newMsgResponseOrder(requestID, GetBlocksResponse, resp), defaultMsgTimeout)
			if err != nil {
				bh.logger.Info().Uint64("first_blk_number", blockInfos[0].Header.GetBlockNo()).Err(err).Int(LogBlkCount, len(blockInfos)).Str("req_id",requestID.String()).Msg("Sending failed")
				return
			}
			//if msgSentCount >= MaxResponseSplitCount {
			//	return
			//}
			// reset list
			blockInfos = make([]*types.Block, 0, sliceCap)
			payloadSize = EmptyGetBlockResponseSize
		}
		blockInfos = append(blockInfos, foundBlock)
		payloadSize += fieldSize
		idx++
	}

	if idx == 0 { // have nothing to send
		return
	}
	// generate response message
	resp := &types.GetBlockResponse{
		Status: status,
		Blocks: blockInfos,HasNext:false}

	// ???: have to check arguments
	bh.logger.Debug().Uint64("first_blk_number", blockInfos[0].Header.GetBlockNo()).Int(LogBlkCount, len(blockInfos)).Str("req_id",requestID.String()).Msg("Sending last part of getMissing response")
	err := remotePeer.sendAndWaitMessage(remotePeer.MF().newMsgResponseOrder(requestID, GetBlocksResponse, resp), defaultMsgTimeout)
	if err != nil {
		bh.logger.Info().Uint64("first_blk_number", blockInfos[0].Header.GetBlockNo()).Err(err).Int(LogBlkCount, len(blockInfos)).Str("req_id",requestID.String()).Msg("Sending failed")
		return
	}
}

// newGetMissingReqHandler creates handler for GetMissingRequest
func newGetAncestorReqHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *getAncestorRequestHandler {
	bh := &getAncestorRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetAncestorRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *getAncestorRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetAncestorRequest{})
}

func (bh *getAncestorRequestHandler) handle(msg Message, msgBody proto.Message) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer
	data := msgBody.(*types.GetAncestorRequest)
	status := types.ResultStatus_OK
	if bh.logger.IsDebugEnabled() {
		debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, bytesArrToString(data.Hashes))
	}

	// send to ChainSvc
	// find ancestor from chainservice
	ancestor := &types.BlockInfo{}

	rawResponse, err := bh.actor.CallRequestDefaultTimeout(
		message.ChainSvc, &message.GetAncestor{Hashes: data.Hashes})
	if err != nil {
		//TODO error handling
		status = types.ResultStatus_ABORTED
	} else {
		v := rawResponse.(message.GetAncestorRsp)
		result := (*message.GetAncestorRsp)(&v)
		if result.Err != nil {
			status = types.ResultStatus_NOT_FOUND
		} else {
			ancestor = result.Ancestor
		}
	}

	resp := &types.GetAncestorResponse{
		Status:       status,
		AncestorHash: ancestor.Hash,
		AncestorNo:   ancestor.No,
	}

	bh.logger.Debug().Uint64("ancestorno", ancestor.No).Str("ancestorhash", enc.ToString(ancestor.Hash)).Msg("Sending get ancestor response")
	remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), GetAncestorResponse, resp))
}

// newBlockRespHandler creates handler for GetAncestorResponse
func newGetAncestorRespHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *getAncestorResponseHandler {
	bh := &getAncestorResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetAncestorResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *getAncestorResponseHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetAncestorResponse{})
}

func (bh *getAncestorResponseHandler) handle(msg Message, msgBody proto.Message) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer
	data := msgBody.(*types.GetAncestorResponse)
	debugLogReceiveResponseMsg(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), peerID, fmt.Sprintf("status=%d, ancestor hash=%s,no=%d", data.Status, enc.ToString(data.AncestorHash), data.AncestorNo))

	// locate request data and remove it if found
	remotePeer.consumeRequest(msg.ID())

	var ancestor *types.BlockInfo
	if data.Status == types.ResultStatus_OK {
		ancestor = &types.BlockInfo{Hash: data.AncestorHash, No: data.AncestorNo}
	}
	// send GetSyncAncestorRsp to syncer
	// if error, ancestor is nil
	bh.actor.TellRequest(message.SyncerSvc, &message.GetSyncAncestorRsp{Ancestor: ancestor})
}
