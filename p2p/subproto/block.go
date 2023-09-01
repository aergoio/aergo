/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

type getBlockHeadersRequestHandler struct {
	BaseMsgHandler
	asyncHelper
}

var _ p2pcommon.MessageHandler = (*getBlockHeadersRequestHandler)(nil)

type getBlockHeadersResponseHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*getBlockHeadersResponseHandler)(nil)

type newBlockNoticeHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*newBlockNoticeHandler)(nil)

type getAncestorRequestHandler struct {
	BaseMsgHandler
	asyncHelper
}

var _ p2pcommon.MessageHandler = (*getAncestorRequestHandler)(nil)

type getAncestorResponseHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*getAncestorResponseHandler)(nil)

// newListBlockHeadersReqHandler creates handler for GetBlockHeadersRequest
func NewGetBlockHeadersReqHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *getBlockHeadersRequestHandler {
	bh := &getBlockHeadersRequestHandler{BaseMsgHandler{protocol: p2pcommon.GetBlockHeadersRequest, pm: pm, peer: peer, actor: actor, logger: logger}, newAsyncHelper()}
	return bh
}

func (bh *getBlockHeadersRequestHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetBlockHeadersRequest{})
}

func (bh *getBlockHeadersRequestHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockHeadersRequest)
	p2putil.DebugLogReceive(bh.logger, bh.protocol, msg.ID().String(), remotePeer, data)
	if bh.issue() {
		go bh.handleGetBlockHeaders(msg, data)
	} else {
		resp := &types.GetBlockHeadersResponse{
			Hashes: nil, Headers: nil,
			Status: types.ResultStatus_RESOURCE_EXHAUSTED,
		}
		remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetBlockHeadersResponse, resp))
	}
}

func (bh *getBlockHeadersRequestHandler) handleGetBlockHeaders(msg p2pcommon.Message, data *types.GetBlockHeadersRequest) {
	defer bh.release()
	remotePeer := bh.peer

	// find block info from chainservice
	maxFetchSize := min(p2pcommon.MaxBlockHeaderResponseCount, data.Size)
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
			foundBlock, err := p2putil.ExtractBlockFromRequest(bh.actor.CallRequestDefaultTimeout(message.ChainSvc,
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
	remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetBlockHeadersResponse, resp))
}

func getBlockHeader(blk *types.Block) *types.BlockHeader {
	return blk.Header
}

// newListBlockRespHandler creates handler for GetBlockHeadersResponse
func NewGetBlockHeaderRespHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *getBlockHeadersResponseHandler {
	bh := &getBlockHeadersResponseHandler{BaseMsgHandler{protocol: p2pcommon.GetBlockHeadersResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *getBlockHeadersResponseHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetBlockHeadersResponse{})
}

func (bh *getBlockHeadersResponseHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockHeadersResponse)
	p2putil.DebugLogReceiveResponse(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), bh.peer, data)

	// send block headers to blockchain service
	remotePeer.ConsumeRequest(msg.OriginalID())

	// TODO: it's not used yet, but used in RPC and can be used in future performance tuning
}

// newNewBlockNoticeHandler creates handler for NewBlockNotice
func NewNewBlockNoticeHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService, sm p2pcommon.SyncManager) *newBlockNoticeHandler {
	bh := &newBlockNoticeHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.NewBlockNotice, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *newBlockNoticeHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.NewBlockNotice{})
}

func (bh *newBlockNoticeHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := bh.peer
	data := msgBody.(*types.NewBlockNotice)
	// remove to verbose log
	// debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, log.DoLazyEval(func() string { return enc.ToString(data.BlkHash) }))

	if blockID, err := types.ParseToBlockID(data.BlockHash); err != nil {
		// TODO Add penalty score and break
		bh.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Str("hash", enc.ToString(data.BlockHash)).Msg("malformed blockHash")
		return
	} else {
		// lru cache can't accept byte slice key
		if !remotePeer.UpdateBlkCache(blockID, data.BlockNo) {
			bh.sm.HandleNewBlockNotice(remotePeer, data)
		}
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
func NewGetAncestorReqHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *getAncestorRequestHandler {
	bh := &getAncestorRequestHandler{BaseMsgHandler{protocol: p2pcommon.GetAncestorRequest, pm: pm, peer: peer, actor: actor, logger: logger}, newAsyncHelper()}
	return bh
}

func (bh *getAncestorRequestHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetAncestorRequest{})
}

func (bh *getAncestorRequestHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	data := msgBody.(*types.GetAncestorRequest)

	if bh.issue() {
		go bh.handleGetAncestorReq(msg, data)
	} else {
		resp := &types.GetAncestorResponse{
			Status: types.ResultStatus_RESOURCE_EXHAUSTED,
		}
		bh.peer.SendMessage(bh.peer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetAncestorResponse, resp))
	}
}

func (bh *getAncestorRequestHandler) handleGetAncestorReq(msg p2pcommon.Message, data *types.GetAncestorRequest) {
	defer bh.release()
	remotePeer := bh.peer
	status := types.ResultStatus_OK
	if bh.logger.IsDebugEnabled() {
		p2putil.DebugLogReceive(bh.logger, bh.protocol, msg.ID().String(), remotePeer, data)
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
	remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetAncestorResponse, resp))
}

// newBlockRespHandler creates handler for GetAncestorResponse
func NewGetAncestorRespHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *getAncestorResponseHandler {
	bh := &getAncestorResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.GetAncestorResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *getAncestorResponseHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetAncestorResponse{})
}

func (bh *getAncestorResponseHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	data := msgBody.(*types.GetAncestorResponse)
	p2putil.DebugLogReceiveResponse(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), bh.peer, data)

	// locate request data and remove it if found
	bh.peer.GetReceiver(msg.OriginalID())(msg, data)
}
