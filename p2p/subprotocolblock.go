/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
)

func (sp SubProtocol) Uint32() uint32 {
	return uint32(sp)
}

// BlockHandler handle block messages.
// Relaying is not implemented yet.
type BlockHandler struct {
	BaseMsgHandler
}

func NewBlockHandler(pm PeerManager, peer *RemotePeer, logger *log.Logger) *BlockHandler {
	bh := &BlockHandler{BaseMsgHandler: BaseMsgHandler{protocol: pingRequest, pm: pm, peer: peer, actor: peer.actorServ, logger: logger}}
	return bh
}
func (bh *BlockHandler) setPeerManager(pm PeerManager) {
	bh.pm = pm
}

func (bh *BlockHandler) startHandling() {
	// bh.pm.SetStreamHandler(getBlocksRequest, bh.onGetBlockRequest)
	// bh.pm.SetStreamHandler(getBlocksResponse, bh.onGetBlockResponse)
	// bh.pm.SetStreamHandler(getBlockHeadersRequest, bh.onGetBlockHeadersRequest)
	// bh.pm.SetStreamHandler(getBlockHeadersResponse, bh.onGetBlockHeadersResponse)
	// bh.pm.SetStreamHandler(notifyNewBlockRequest, bh.onNotifyNewBlock)
	// bh.pm.SetStreamHandler(getMissingRequest, bh.onGetMissingRequest)
}

// remote peer requests handler
func (bh *BlockHandler) handleBlockRequest(msg *types.P2PMessage) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer

	// get request data
	data := &types.GetBlockRequest{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		bh.logger.Info().Err(err).Msg("fail to decode")
		return
	}

	debugLogReceiveMsg(bh.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, len(data.Hashes))

	valid := bh.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		bh.logger.Info().Msg("Failed to authenticate message")
		return
	}

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
	resp := &types.GetBlockResponse{MessageData: &types.MessageData{},
		Status: status,
		Blocks: blockInfos}

	remotePeer.sendMessage(newPbMsgResponseOrder(data.MessageData.Id, true, getBlocksResponse, resp))
}

// remote GetBlock response handler
func (bh *BlockHandler) handleGetBlockResponse(msg *types.P2PMessage) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer

	data := &types.GetBlockResponse{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(bh.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, len(data.Blocks))
	valid := bh.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		bh.logger.Info().Msg("Failed to authenticate message")
		return
	}
	// locate request data and remove it if found
	remotePeer.consumeRequest(data.MessageData.Id)

	// got block
	bh.logger.Debug().Int("block_cnt", len(data.Blocks)).Msg("Request chainservice to add blocks")
	for _, block := range data.Blocks {
		bh.actor.SendRequest(message.ChainSvc, &message.AddBlock{PeerID: peerID, Block: block, Bstate: nil})
	}

}

// remote peer requests handler
func (bh *BlockHandler) handleGetBlockHeadersRequest(msg *types.P2PMessage) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer

	// get request data
	data := &types.GetBlockHeadersRequest{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		bh.logger.Info().Err(err).Msg("fail to decode")
		return
	}
	debugLogReceiveMsg(bh.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, data)

	valid := bh.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		bh.logger.Info().Msg("Failed to authenticate message")
		return
	}

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
			hashes = append(hashes, foundBlock.Hash)
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
			hashes = append(hashes, foundBlock.Hash)
			headers = append(headers, getBlockHeader(foundBlock))
			idx++
		}
	}
	// generate response message
	resp := &types.GetBlockHeadersResponse{MessageData: &types.MessageData{},
		Hashes: hashes, Headers: headers,
		Status: types.ResultStatus_OK,
	}
	remotePeer.sendMessage(newPbMsgResponseOrder(data.MessageData.Id, true, getBlockHeadersResponse, resp))
}

func getBlockHeader(blk *types.Block) *types.BlockHeader {
	return blk.Header
}

// remote GetBlock response handler
func (bh *BlockHandler) handleGetBlockHeadersResponse(msg *types.P2PMessage) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer

	data := &types.GetBlockHeadersResponse{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(bh.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, nil)
	valid := bh.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		bh.logger.Info().Msg("Failed to authenticate message")
		return
	}

	// send block headers to blockchain service
	remotePeer.consumeRequest(data.MessageData.Id)
}

// remote NotifyNewBlock response handler
func (bh *BlockHandler) handleNewBlockNotice(msg *types.P2PMessage) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer

	data := &types.NewBlockNotice{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(bh.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID,
		log.DoLazyEval(func() string { return enc.ToString(data.BlockHash) }))

	remotePeer.handleNewBlockNotice(data)

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
func (bh *BlockHandler) NotifyBranchBlock(peer *RemotePeer, hash message.BlockHash, blockno types.BlockNo) bool {
	bh.logger.Debug().Str(LogPeerID, peer.meta.ID.Pretty()).Msg("Notifying branch block")

	// create message data
	req := &types.NewBlockNotice{MessageData: &types.MessageData{},
		BlockHash: hash,
		BlockNo:   uint64(blockno)}

	peer.sendMessage(newPbMsgRequestOrder(false, false, newBlockNotice, req))
	return true
}

// replying chain tree
func (bh *BlockHandler) sendMissingResp(remotePeer *RemotePeer, requestID string, missing []message.BlockHash) {
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
	resp := &types.GetBlockResponse{MessageData: &types.MessageData{},
		Status: status,
		Blocks: blockInfos}

	// ???: have to check arguments
	remotePeer.sendMessage(newPbMsgResponseOrder(requestID, true, getBlocksResponse, resp))
}

// remote peer requests handler
func (bh *BlockHandler) handleGetMissingRequest(msg *types.P2PMessage) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer

	// get request data
	data := &types.GetMissingRequest{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		bh.logger.Info().Err(err).Msg("fail to decode")
		return
	}
	debugLogReceiveMsg(bh.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, log.DoLazyEval(func() string {
		return bytesArrToString(data.Hashes)
	}))
	valid := bh.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		bh.logger.Info().Msg("Failed to authenticate message")
		return
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
	bh.logger.Debug().Str(LogPeerID, peerID.Pretty()).Str(LogMsgID, data.MessageData.Id).Msg("Sending GetMssingRequest response")

	bh.sendMissingResp(remotePeer, data.MessageData.Id, missing.Hashes)
	/*
		for i := 0; i < len(missing.Hashes); i++ {
			bh.NotifyBranchBlock(remotePeer, missing.Hashes[i], missing.Blocknos[i])
		}
	*/
}

// remote GetBlock response handler
/*
func (bh *BlockHandler) onGetMissingResponse(s inet.Stream) {
	defer s.Close()

	remotePeer, exists := bh.pm.GetPeer(s.Conn().RemotePeer())
	if !exists {
		bh.logger.Warnf("Request to invalid peer %s ", s.Conn().RemotePeer().Pretty())
		return
	}
	bh.logger.Debugf("Received GetMissingRequest response from %s.", remotePeer.meta.ID.Pretty())
	data := &types.GetMissingResponse{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		return
	}
	valid := bh.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		bh.logger.Info("Failed to authenticate message")
		return
	}

	// send back to caller
	bh.logger.Infof("Got Missing response ")
	bh.iserv.SendRequest(message.ChainSvc, &message.GetMissingRsp{Hashes: data.Hashes, Headers: data.Headers})
	remotePeer.ConsumeRequest(data.MessageData.Id)
}
*/
