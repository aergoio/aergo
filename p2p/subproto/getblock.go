/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"fmt"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

type blockRequestHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*blockRequestHandler)(nil)

type blockResponseHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*blockResponseHandler)(nil)

// newBlockReqHandler creates handler for GetBlockRequest
func NewBlockReqHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *blockRequestHandler {
	bh := &blockRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlocksRequest, pm: pm, peer: peer, actor: actor, logger: logger}}

	return bh
}

func (bh *blockRequestHandler) ParsePayload(rawbytes []byte) (proto.Message, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetBlockRequest{})
}

const (
	EmptyGetBlockResponseSize = 12 // roughly estimated maximum size if element is full
)

func (bh *blockRequestHandler) Handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockRequest)
	p2putil.DebugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), remotePeer, len(data.Hashes))

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
			bh.logger.Warn().Err(err).Str(p2putil.LogBlkHash, enc.ToString(hash)).Str("req_id", requestID.String()).Msg("failed to get block while processing getBlock")
			status = types.ResultStatus_INTERNAL
			break
		}
		if foundBlock == nil {
			// the remote peer request not existing block
			bh.logger.Debug().Str(p2putil.LogBlkHash, enc.ToString(hash)).Str("req_id", requestID.String()).Msg("requested block hash is missing")
			status = types.ResultStatus_NOT_FOUND
			break

		}
		blockSize = proto.Size(foundBlock)
		fieldSize = blockSize + p2putil.CalculateFieldDescSize(blockSize)
		if len(blockInfos) >= sliceCap || (payloadSize+fieldSize) > p2pcommon.MaxPayloadLength {
			msgSentCount++
			// send partial list
			resp := &types.GetBlockResponse{
				Status:  status,
				Blocks:  blockInfos,
				HasNext: true,
				//HasNext:msgSentCount<MaxResponseSplitCount, // always have nextItem ( see foundBlock) but msg count limit will affect
			}
			bh.logger.Debug().Uint64("first_blk_number", blockInfos[0].Header.GetBlockNo()).Int(p2putil.LogBlkCount, len(blockInfos)).Str("req_id", requestID.String()).Msg("Sending partial getBlock response")
			err := remotePeer.SendAndWaitMessage(remotePeer.MF().NewMsgResponseOrder(requestID, GetBlocksResponse, resp), defaultMsgTimeout)
			if err != nil {
				bh.logger.Info().Uint64("first_blk_number", blockInfos[0].Header.GetBlockNo()).Err(err).Int(p2putil.LogBlkCount, len(blockInfos)).Str("req_id", requestID.String()).Msg("Sending failed")
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
	bh.logger.Debug().Int(p2putil.LogBlkCount, len(blockInfos)).Str("req_id", requestID.String()).Msg("Sending last part of getBlock response")
	err := remotePeer.SendAndWaitMessage(remotePeer.MF().NewMsgResponseOrder(requestID, GetBlocksResponse, resp), defaultMsgTimeout)
	if err != nil {
		bh.logger.Info().Int(p2putil.LogBlkCount, len(data.Hashes)).Err(err).Str("req_id", requestID.String()).Msg("Sending failed")
		return
	}
}

// newBlockRespHandler creates handler for GetBlockResponse
func NewBlockRespHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService, sm p2pcommon.SyncManager) *blockResponseHandler {
	bh := &blockResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlocksResponse, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *blockResponseHandler) ParsePayload(rawbytes []byte) (proto.Message, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetBlockResponse{})
}

func (bh *blockResponseHandler) Handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockResponse)
	if bh.logger.IsDebugEnabled() {
		additional := fmt.Sprintf("hashNext=%t,%s", data.HasNext, p2putil.PrintHashList(data.Blocks))
		p2putil.DebugLogReceiveResponseMsg(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), remotePeer, additional)
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
