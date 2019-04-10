/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"fmt"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

type listBlockHeadersRequestHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*listBlockHeadersRequestHandler)(nil)

type listBlockHeadersResponseHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*listBlockHeadersResponseHandler)(nil)

type newBlockNoticeHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*newBlockNoticeHandler)(nil)

type getAncestorRequestHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*getAncestorRequestHandler)(nil)

type getAncestorResponseHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*getAncestorResponseHandler)(nil)

// newListBlockHeadersReqHandler creates handler for GetBlockHeadersRequest
func NewListBlockHeadersReqHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *listBlockHeadersRequestHandler {
	bh := &listBlockHeadersRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlockHeadersRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *listBlockHeadersRequestHandler) ParsePayload(rawbytes []byte) (proto.Message, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetBlockHeadersRequest{})
}

func (bh *listBlockHeadersRequestHandler) Handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockHeadersRequest)
	p2putil.DebugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), remotePeer, data)

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
	remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), GetBlockHeadersResponse, resp))
}

func getBlockHeader(blk *types.Block) *types.BlockHeader {
	return blk.Header
}

// newListBlockRespHandler creates handler for GetBlockHeadersResponse
func NewListBlockRespHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *listBlockHeadersResponseHandler {
	bh := &listBlockHeadersResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlockHeadersResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *listBlockHeadersResponseHandler) ParsePayload(rawbytes []byte) (proto.Message, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetBlockHeadersResponse{})
}

func (bh *listBlockHeadersResponseHandler) Handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockHeadersResponse)
	p2putil.DebugLogReceiveResponseMsg(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), bh.peer, len(data.Hashes))

	// send block headers to blockchain service
	remotePeer.ConsumeRequest(msg.OriginalID())

	// TODO: it's not used yet, but used in RPC and can be used in future performance tuning
}

// newNewBlockNoticeHandler creates handler for NewBlockNotice
func NewNewBlockNoticeHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService, sm p2pcommon.SyncManager) *newBlockNoticeHandler {
	bh := &newBlockNoticeHandler{BaseMsgHandler: BaseMsgHandler{protocol: NewBlockNotice, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *newBlockNoticeHandler) ParsePayload(rawbytes []byte) (proto.Message, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.NewBlockNotice{})
}

func (bh *newBlockNoticeHandler) Handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.NewBlockNotice)
	// remove to verbose log
	// debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), peerID, log.DoLazyEval(func() string { return enc.ToString(data.BlkHash) }))

	if _, err := types.ParseToBlockID(data.BlockHash); err != nil {
		// TODO Add penelty score and break
		bh.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Str("hash", enc.ToString(data.BlockHash)).Msg("malformed blockHash")
		return
	}
	// lru cache can accept hashable key
	if !remotePeer.UpdateBlkCache(data.BlockHash, data.BlockNo) {
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
func NewGetAncestorReqHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *getAncestorRequestHandler {
	bh := &getAncestorRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetAncestorRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *getAncestorRequestHandler) ParsePayload(rawbytes []byte) (proto.Message, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetAncestorRequest{})
}

func (bh *getAncestorRequestHandler) Handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetAncestorRequest)
	status := types.ResultStatus_OK
	if bh.logger.IsDebugEnabled() {
		p2putil.DebugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), remotePeer, p2putil.BytesArrToString(data.Hashes))
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
	remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), GetAncestorResponse, resp))
}

// newBlockRespHandler creates handler for GetAncestorResponse
func NewGetAncestorRespHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *getAncestorResponseHandler {
	bh := &getAncestorResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetAncestorResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *getAncestorResponseHandler) ParsePayload(rawbytes []byte) (proto.Message, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetAncestorResponse{})
}

func (bh *getAncestorResponseHandler) Handle(msg p2pcommon.Message, msgBody proto.Message) {
	data := msgBody.(*types.GetAncestorResponse)
	p2putil.DebugLogReceiveResponseMsg(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), bh.peer, fmt.Sprintf("status=%d, ancestor hash=%s,no=%d", data.Status, enc.ToString(data.AncestorHash), data.AncestorNo))

	// locate request data and remove it if found
	bh.peer.GetReceiver(msg.OriginalID())(msg, data)
}
