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
	p := &BlockHandler{BaseMsgHandler: BaseMsgHandler{protocol: pingRequest, pm: pm, peer: peer, actor: peer.actorServ, logger: logger}}
	return p
}
func (p *BlockHandler) setPeerManager(pm PeerManager) {
	p.pm = pm
}

func (p *BlockHandler) startHandling() {
	// p.pm.SetStreamHandler(getBlocksRequest, p.onGetBlockRequest)
	// p.pm.SetStreamHandler(getBlocksResponse, p.onGetBlockResponse)
	// p.pm.SetStreamHandler(getBlockHeadersRequest, p.onGetBlockHeadersRequest)
	// p.pm.SetStreamHandler(getBlockHeadersResponse, p.onGetBlockHeadersResponse)
	// p.pm.SetStreamHandler(notifyNewBlockRequest, p.onNotifyNewBlock)
	// p.pm.SetStreamHandler(getMissingRequest, p.onGetMissingRequest)
}

// remote peer requests handler
func (p *BlockHandler) handleBlockRequest(msg *types.P2PMessage) {
	peerID := p.peer.ID()
	remotePeer := p.peer

	// get request data
	data := &types.GetBlockRequest{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		p.logger.Info().Err(err).Msg("fail to decode")
		return
	}

	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, len(data.Hashes))

	valid := p.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.logger.Info().Msg("Failed to authenticate message")
		return
	}

	// find block info from chainservice
	idx := 0
	blockInfos := make([]*types.Block, 0, len(data.Hashes))
	for _, hash := range data.Hashes {
		foundBlock, err := extractBlockFromRequest(p.actor.CallRequest(message.ChainSvc,
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
func (p *BlockHandler) handleGetBlockResponse(msg *types.P2PMessage) {
	peerID := p.peer.ID()
	remotePeer := p.peer

	data := &types.GetBlockResponse{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, len(data.Blocks))
	valid := p.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.logger.Info().Msg("Failed to authenticate message")
		return
	}
	// locate request data and remove it if found
	remotePeer.consumeRequest(data.MessageData.Id)

	// got block
	p.logger.Debug().Int("block_cnt", len(data.Blocks)).Msg("Request chainservice to add blocks")
	for _, block := range data.Blocks {
		p.actor.SendRequest(message.ChainSvc, &message.AddBlock{PeerID: peerID, Block: block})
	}

}

// remote peer requests handler
func (p *BlockHandler) handleGetBlockHeadersRequest(msg *types.P2PMessage) {
	peerID := p.peer.ID()
	remotePeer := p.peer

	// get request data
	data := &types.GetBlockHeadersRequest{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		p.logger.Info().Err(err).Msg("fail to decode")
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, data)

	valid := p.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.logger.Info().Msg("Failed to authenticate message")
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
			foundBlock, err := extractBlockFromRequest(p.actor.CallRequest(message.ChainSvc,
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
			foundBlock, err := extractBlockFromRequest(p.actor.CallRequest(message.ChainSvc,
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
func (p *BlockHandler) handleGetBlockHeadersResponse(msg *types.P2PMessage) {
	peerID := p.peer.ID()
	remotePeer := p.peer

	data := &types.GetBlockHeadersResponse{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, nil)
	valid := p.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.logger.Info().Msg("Failed to authenticate message")
		return
	}

	// send block headers to blockchain service
	remotePeer.consumeRequest(data.MessageData.Id)
}

// remote NotifyNewBlock response handler
func (p *BlockHandler) handleNewBlockNotice(msg *types.P2PMessage) {
	peerID := p.peer.ID()
	remotePeer := p.peer

	data := &types.NewBlockNotice{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID,
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
func (p *BlockHandler) NotifyBranchBlock(peer *RemotePeer, hash message.BlockHash, blockno types.BlockNo) bool {
	p.logger.Debug().Str(LogPeerID, peer.meta.ID.Pretty()).Msg("Notifying branch block")

	// create message data
	req := &types.NewBlockNotice{MessageData: &types.MessageData{},
		BlockHash: hash,
		BlockNo:   uint64(blockno)}

	peer.sendMessage(newPbMsgRequestOrder(false, false, newBlockNotice, req))
	return true
}

// replying chain tree
func (p *BlockHandler) sendMissingResp(remotePeer *RemotePeer, requestID string, missing []message.BlockHash) {
	// find block info from chainservice
	blockInfos := make([]*types.Block, 0, len(missing))
	for _, hash := range missing {
		foundBlock, err := extractBlockFromRequest(p.actor.CallRequest(message.ChainSvc,
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
func (p *BlockHandler) handleGetMissingRequest(msg *types.P2PMessage) {
	peerID := p.peer.ID()
	remotePeer := p.peer

	// get request data
	data := &types.GetMissingRequest{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		p.logger.Info().Err(err).Msg("fail to decode")
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, log.DoLazyEval(func() string {
		return bytesArrToString(data.Hashes)
	}))
	valid := p.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.logger.Info().Msg("Failed to authenticate message")
		return
	}

	// send to ChainSvc
	// find block info from chainservice
	rawResponse, err := p.actor.CallRequest(
		message.ChainSvc, &message.GetMissing{Hashes: data.Hashes, StopHash: data.Stophash})
	if err != nil {
		p.logger.Warn().Err(err).Msg("failed to get missing")

		return
	}
	v := rawResponse.(message.GetMissingRsp)
	missing := (*message.GetMissingRsp)(&v)

	// generate response message
	p.logger.Debug().Str(LogPeerID, peerID.Pretty()).Str(LogMsgID, data.MessageData.Id).Msg("Sending GetMssingRequest response")

	p.sendMissingResp(remotePeer, data.MessageData.Id, missing.Hashes)
	/*
		for i := 0; i < len(missing.Hashes); i++ {
			p.NotifyBranchBlock(remotePeer, missing.Hashes[i], missing.Blocknos[i])
		}
	*/
}

// remote GetBlock response handler
/*
func (p *BlockHandler) onGetMissingResponse(s inet.Stream) {
	defer s.Close()

	remotePeer, exists := p.pm.GetPeer(s.Conn().RemotePeer())
	if !exists {
		p.logger.Warnf("Request to invalid peer %s ", s.Conn().RemotePeer().Pretty())
		return
	}
	p.logger.Debugf("Received GetMissingRequest response from %s.", remotePeer.meta.ID.Pretty())
	data := &types.GetMissingResponse{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		return
	}
	valid := p.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.logger.Info("Failed to authenticate message")
		return
	}

	// send back to caller
	p.logger.Infof("Got Missing response ")
	p.iserv.SendRequest(message.ChainSvc, &message.GetMissingRsp{Hashes: data.Hashes, Headers: data.Headers})
	remotePeer.ConsumeRequest(data.MessageData.Id)
}
*/
