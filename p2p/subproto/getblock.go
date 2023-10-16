/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/protobuf/proto"
)

type blockRequestHandler struct {
	BaseMsgHandler
	asyncHelper
}

var _ p2pcommon.MessageHandler = (*blockRequestHandler)(nil)

type blockResponseHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*blockResponseHandler)(nil)

// newBlockReqHandler creates handler for GetBlockRequest
func NewBlockReqHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *blockRequestHandler {
	bh := &blockRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.GetBlocksRequest, pm: pm, peer: peer, actor: actor, logger: logger}, asyncHelper: newAsyncHelper()}

	return bh
}

func (bh *blockRequestHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetBlockRequest{})
}

const (
	EmptyGetBlockResponseSize = 12 // roughly estimated maximum size if element is full
)

func (bh *blockRequestHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockRequest)
	p2putil.DebugLogReceive(bh.logger, bh.protocol, msg.ID().String(), remotePeer, data)
	if bh.issue() {
		go bh.handleBlkReq(msg, data)
	} else {
		bh.logger.Info().Stringer(p2putil.LogProtoID, bh.protocol).Stringer(p2putil.LogMsgID, msg.ID()).Str(p2putil.LogPeerName, remotePeer.Name()).Msg("return error for busy")
		resp := &types.GetBlockResponse{
			Status: types.ResultStatus_RESOURCE_EXHAUSTED,
			Blocks: nil, HasNext: false}

		remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetBlocksResponse, resp))
	}
}

func (bh *blockRequestHandler) handleBlkReq(msg p2pcommon.Message, data *types.GetBlockRequest) {
	defer bh.release()
	remotePeer := bh.peer

	requestID := msg.ID()
	sliceCap := p2pcommon.MaxBlockResponseCount
	if len(data.Hashes) < sliceCap {
		sliceCap = len(data.Hashes)
	}

	defaultMsgTimeout := time.Second * 30
	// find block info from chainservice
	idx := 0
	msgSentCount := 0
	status := types.ResultStatus_OK
	blockInfos := make([]*types.Block, 0, sliceCap)
	payloadSize := EmptyGetBlockResponseSize
	var blockSize, fieldSize int
	for _, hash := range data.Hashes {
		foundBlock, err := bh.actor.GetChainAccessor().GetBlock(hash)
		if err != nil {
			// the block hash from request must exists. this error is fatal.
			bh.logger.Warn().Err(err).Str(p2putil.LogBlkHash, enc.ToString(hash)).Str(p2putil.LogOrgReqID, requestID.String()).Msg("failed to get block while processing getBlock")
			status = types.ResultStatus_INTERNAL
			break
		}
		if foundBlock == nil {
			// the remote peer request not existing block
			bh.logger.Debug().Str(p2putil.LogBlkHash, enc.ToString(hash)).Str(p2putil.LogOrgReqID, requestID.String()).Msg("requested block hash is missing")
			status = types.ResultStatus_NOT_FOUND
			break

		}
		blockSize = proto.Size(foundBlock)
		fieldSize = blockSize + p2putil.CalculateFieldDescSize(blockSize)
		if len(blockInfos) >= sliceCap || uint32(payloadSize+fieldSize) > p2pcommon.MaxPayloadLength {
			msgSentCount++
			// send partial list
			resp := &types.GetBlockResponse{
				Status:  status,
				Blocks:  blockInfos,
				HasNext: true,
				//HasNext:msgSentCount<MaxResponseSplitCount, // always have nextItem ( see foundBlock) but msg count limit will affect
			}
			bh.logger.Debug().Uint64("first_blk_number", blockInfos[0].Header.GetBlockNo()).Int(p2putil.LogBlkCount, len(blockInfos)).Str(p2putil.LogOrgReqID, requestID.String()).Msg("Sending partial getBlock response")
			err := remotePeer.SendAndWaitMessage(remotePeer.MF().NewMsgResponseOrder(requestID, p2pcommon.GetBlocksResponse, resp), defaultMsgTimeout)
			if err != nil {
				bh.logger.Info().Uint64("first_blk_number", blockInfos[0].Header.GetBlockNo()).Err(err).Int(p2putil.LogBlkCount, len(blockInfos)).Str(p2putil.LogOrgReqID, requestID.String()).Msg("Sending failed")
				return
			}
			blockInfos = make([]*types.Block, 0, sliceCap)
			payloadSize = EmptyGetBlockResponseSize
		}
		blockInfos = append(blockInfos, foundBlock)
		payloadSize += fieldSize
		idx++
	}

	if 0 == idx {
		status = types.ResultStatus_NOT_FOUND
	}
	// Failed response does not need incomplete blocks information
	if status != types.ResultStatus_OK {
		blockInfos = blockInfos[:0]
	}
	// generate response message
	resp := &types.GetBlockResponse{
		Status: status,
		Blocks: blockInfos, HasNext: false}

	// ???: have to check arguments
	bh.logger.Debug().Int(p2putil.LogBlkCount, len(blockInfos)).Str(p2putil.LogOrgReqID, requestID.String()).Msg("Sending last part of getBlock response")
	err := remotePeer.SendAndWaitMessage(remotePeer.MF().NewMsgResponseOrder(requestID, p2pcommon.GetBlocksResponse, resp), defaultMsgTimeout)
	if err != nil {
		bh.logger.Info().Int(p2putil.LogBlkCount, len(data.Hashes)).Err(err).Str(p2putil.LogOrgReqID, requestID.String()).Msg("Sending failed")
		return
	}
}

// NewBlockRespHandler creates handler for GetBlockResponse
func NewBlockRespHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService, sm p2pcommon.SyncManager) *blockResponseHandler {
	bh := &blockResponseHandler{BaseMsgHandler{protocol: p2pcommon.GetBlocksResponse, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *blockResponseHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetBlockResponse{})
}

func (bh *blockResponseHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockResponse)
	if bh.logger.IsDebugEnabled() {
		p2putil.DebugLogReceiveResponse(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), remotePeer, data)
	}

	// locate request data and remove it if found
	if !remotePeer.GetReceiver(msg.OriginalID())(msg, data) {
		remotePeer.ConsumeRequest(msg.OriginalID())
		// TODO temporary code and will be deleted after newer syncer is made.
		if data.Status != types.ResultStatus_OK || len(data.Blocks) == 0 {
			return
		}
		bh.sm.HandleGetBlockResponse(remotePeer, msg, data)
	}
}
