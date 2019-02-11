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
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

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


type getAncestorRequestHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*getAncestorRequestHandler)(nil)

type getAncestorResponseHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*getAncestorResponseHandler)(nil)


// newListBlockHeadersReqHandler creates handler for GetBlockHeadersRequest
func newListBlockHeadersReqHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *listBlockHeadersRequestHandler {
	bh := &listBlockHeadersRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlockHeadersRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *listBlockHeadersRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetBlockHeadersRequest{})
}

func (bh *listBlockHeadersRequestHandler) handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockHeadersRequest)
	debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), remotePeer, data)

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

func (bh *listBlockHeadersResponseHandler) handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockHeadersResponse)
	debugLogReceiveResponseMsg(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), bh.peer, len(data.Hashes))

	// send block headers to blockchain service
	remotePeer.consumeRequest(msg.OriginalID())

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

func (bh *newBlockNoticeHandler) handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.NewBlockNotice)
	// remove to verbose log
	// debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, log.DoLazyEval(func() string { return enc.ToString(data.BlkHash) }))

	if _, err := types.ParseToBlockID(data.BlockHash) ; err != nil {
		// TODO Add penelty score and break
		bh.logger.Info().Str(LogPeerName, remotePeer.Name()).Str("hash", enc.ToString(data.BlockHash)).Msg("malformed blockHash")
		return
	}
	// lru cache can accept hashable key
	if !remotePeer.updateBlkCache(data.BlockHash, data.BlockNo) {
		bh.sm.HandleNewBlockNotice(remotePeer, data)
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

// newGetAncestorReqHandler creates handler for GetAncestorRequest
func newGetAncestorReqHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *getAncestorRequestHandler {
	bh := &getAncestorRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetAncestorRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *getAncestorRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetAncestorRequest{})
}

func (bh *getAncestorRequestHandler) handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetAncestorRequest)
	status := types.ResultStatus_OK
	if bh.logger.IsDebugEnabled() {
		debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), remotePeer, bytesArrToString(data.Hashes))
	}

	// send to ChainSvc
	// find ancestor from chainservice
	ancestor := &types.BlockInfo{}

	//TODO split rsp handler
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

func (bh *getAncestorResponseHandler) handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetAncestorResponse)
	debugLogReceiveResponseMsg(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), bh.peer, fmt.Sprintf("status=%d, ancestor hash=%s,no=%d", data.Status, enc.ToString(data.AncestorHash), data.AncestorNo))

	// locate request data and remove it if found
	remotePeer.consumeRequest(msg.OriginalID())

	var ancestor *types.BlockInfo
	if data.Status == types.ResultStatus_OK {
		ancestor = &types.BlockInfo{Hash: data.AncestorHash, No: data.AncestorNo}
	}
	// send GetSyncAncestorRsp to syncer
	// if error, ancestor is nil
	bh.actor.TellRequest(message.SyncerSvc, &message.GetSyncAncestorRsp{Ancestor: ancestor})
}
