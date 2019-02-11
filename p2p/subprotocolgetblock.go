/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2pcommon"
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

func (bh *blockRequestHandler) handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockRequest)
	debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), remotePeer, len(data.Hashes))

	requestID := msg.ID()
	sliceCap := MaxBlockResponseCount
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
	for  _, hash := range data.Hashes {
		foundBlock, err := bh.actor.GetChainAccessor().GetBlock(hash)
		if err != nil {
			// the block hash from request must exists. this error is fatal.
			bh.logger.Warn().Err(err).Str(LogBlkHash, enc.ToString(hash)).Str("req_id", requestID.String()).Msg("failed to get block while processing getBlock")
			status = types.ResultStatus_INTERNAL
			break
		}
		if foundBlock == nil {
			// the remote peer request not existing block
			bh.logger.Debug().Str(LogBlkHash, enc.ToString(hash)).Str("req_id", requestID.String()).Msg("requested block hash is missing")
			status = types.ResultStatus_NOT_FOUND
			break

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
			bh.logger.Debug().Uint64("first_blk_number", blockInfos[0].Header.GetBlockNo()).Int(LogBlkCount, len(blockInfos)).Str("req_id",requestID.String()).Msg("Sending partial getBlock response")
			err := remotePeer.sendAndWaitMessage(remotePeer.MF().newMsgResponseOrder(requestID, GetBlocksResponse, resp), defaultMsgTimeout)
			if err != nil {
				bh.logger.Info().Uint64("first_blk_number", blockInfos[0].Header.GetBlockNo()).Err(err).Int(LogBlkCount, len(blockInfos)).Str("req_id",requestID.String()).Msg("Sending failed")
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
	bh.logger.Debug().Int(LogBlkCount, len(blockInfos)).Str("req_id",requestID.String()).Msg("Sending last part of getBlock response")
	err := remotePeer.sendAndWaitMessage(remotePeer.MF().newMsgResponseOrder(requestID, GetBlocksResponse, resp), defaultMsgTimeout)
	if err != nil {
		bh.logger.Info().Int(LogBlkCount, len(data.Hashes)).Err(err).Str("req_id",requestID.String()).Msg("Sending failed")
		return
	}
}


// newBlockRespHandler creates handler for GetBlockResponse
func newBlockRespHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService, sm SyncManager) *blockResponseHandler {
	bh := &blockResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlocksResponse, pm: pm, sm:sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *blockResponseHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetBlockResponse{})
}

func (bh *blockResponseHandler) handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockResponse)
	if bh.logger.IsDebugEnabled() {
		additional := fmt.Sprintf("hashNext=%t,%s", data.HasNext,PrintHashList(data.Blocks))
	debugLogReceiveResponseMsg(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), remotePeer, additional )
}

	// locate request data and remove it if found
	if !remotePeer.GetReceiver(msg.OriginalID())(msg, data) {
		remotePeer.consumeRequest(msg.OriginalID())
		// TODO temporary code and will be deleted after newer syncer is made.
		if data.Status != types.ResultStatus_OK || len(data.Blocks) == 0 {
			return
		}
		bh.sm.HandleGetBlockResponse(remotePeer, msg, data)
	}
}

