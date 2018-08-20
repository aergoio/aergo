/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"sync"

	"github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"

	inet "github.com/libp2p/go-libp2p-net"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/blockchain"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/multiformats/go-multicodec/protobuf"
)

// pattern: /protocol-name/request-or-response-message/version
const (
	getBlocksRequest        protocol.ID = "/blk/getreq/0.2"
	getBlocksResponse       protocol.ID = "/blk/getresp/0.2"
	getBlockHeadersRequest  protocol.ID = "/blk/headerreq/0.1"
	getBlockHeadersResponse protocol.ID = "/blk/headerresp/0.1"
	getMissingRequest       protocol.ID = "/blk/getmreq/0.1"
	getMissingResponse      protocol.ID = "/blk/getmresp/0.1"
	notifyNewBlockRequest   protocol.ID = "/blk/newblockreq/0.1"
)

// BlockProtocol handle block messages.
// Relaying is not implemented yet.
type BlockProtocol struct {
	iserv ActorService

	ps       PeerManager
	reqMutex sync.Mutex

	chainsvc *blockchain.ChainService
	log      *log.Logger
}

// NewBlockProtocol create block subprotocol
func NewBlockProtocol(logger *log.Logger, chainsvc *blockchain.ChainService) *BlockProtocol {
	p := &BlockProtocol{reqMutex: sync.Mutex{},
		log: logger,
	}
	p.chainsvc = chainsvc
	return p
}

func (p *BlockProtocol) initWith(p2pservice PeerManager) {
	p.ps = p2pservice
	p.ps.SetStreamHandler(getBlocksRequest, p.onGetBlockRequest)
	p.ps.SetStreamHandler(getBlocksResponse, p.onGetBlockResponse)
	p.ps.SetStreamHandler(getBlockHeadersRequest, p.onGetBlockHeadersRequest)
	p.ps.SetStreamHandler(getBlockHeadersResponse, p.onGetBlockHeadersResponse)
	p.ps.SetStreamHandler(notifyNewBlockRequest, p.onNotifyNewBlock)
	p.ps.SetStreamHandler(getMissingRequest, p.onGetMissingRequest)
}

// remote peer requests handler
func (p *BlockProtocol) onGetBlockRequest(s inet.Stream) {
	defer s.Close()

	peerID := s.Conn().RemotePeer()
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()
	perr := remotePeer.checkState()
	if perr != nil {
		p.log.Info().Msgf("%s: Invalid peer state to handle request %s : %s", peerID.Pretty(), s.Protocol(), perr.Error())
		return
	}

	// get request data
	data := &types.GetBlockRequest{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Info().Err(err).Msg("fail to decode")
		return
	}

	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, len(data.Hashes))

	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Info().Msg("Failed to authenticate message")
		return
	}

	// find block info from chainservice
	idx := 0
	blockInfos := make([]*types.Block, 0, len(data.Hashes))
	for _, hash := range data.Hashes {
		foundBlock, err := extractBlockFromRequest(p.iserv.CallRequest(message.ChainSvc,
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
func (p *BlockProtocol) onGetBlockResponse(s inet.Stream) {
	defer s.Close()

	peerID := s.Conn().RemotePeer()
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()
	perr := remotePeer.checkState()
	if perr != nil {
		p.log.Info().Msgf("%s: Invalid peer state to handle request %s : %s", peerID.Pretty(), s.Protocol(), perr.Error())
		return
	}

	data := &types.GetBlockResponse{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, len(data.Blocks))
	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Info().Msg("Failed to authenticate message")
		return
	}
	// locate request data and remove it if found
	remotePeer.consumeRequest(data.MessageData.Id)

	// got block
	p.log.Debug().Msgf("Request chainservice to add %d blocks", len(data.Blocks))
	for _, block := range data.Blocks {
		p.iserv.SendRequest(message.ChainSvc, &message.AddBlock{PeerID: peerID, Block: block})
	}

}

// GetBlocks send request message to peer and
func (p *BlockProtocol) GetBlocks(peerID peer.ID, blockHashes []message.BlockHash) bool {
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		p.log.Warn().Str("peer_id", peerID.Pretty()).Msgf("Message %s to Unknown peer, check if a bug", getBlocksRequest)
		return false
	}
	p.log.Debug().Msgf("Sending Get block request to: %s...(%d blocks)", peerID.Pretty(), len(blockHashes))

	hashes := make([][]byte, len(blockHashes))
	for i, hash := range blockHashes {
		hashes[i] = ([]byte)(hash)
	}
	// create message data
	req := &types.GetBlockRequest{MessageData: &types.MessageData{},
		Hashes: hashes}

	remotePeer.sendMessage(newPbMsgRequestOrder(true, true, getBlocksRequest, req))
	return true
}

// GetBlockHeaders send request message to peer and
func (p *BlockProtocol) GetBlockHeaders(msg *message.GetBlockHeaders) bool {
	remotePeer, exists := p.ps.GetPeer(msg.ToWhom)
	if !exists {
		p.log.Warn().Str("peer_id", msg.ToWhom.Pretty()).Msg("Request to invalid peer")
		return false
	}
	peerID := remotePeer.meta.ID

	p.log.Debug().Str("peer_id", peerID.Pretty()).Interface("msg", msg).Msg("Sending Get block Header request")
	// create message data
	reqMsg := &types.GetBlockHeadersRequest{MessageData: &types.MessageData{}, Hash: msg.Hash,
		Height: msg.Height, Offset: msg.Offset, Size: msg.MaxSize, Asc: msg.Asc,
	}
	remotePeer.sendMessage(newPbMsgRequestOrder(true, true, getBlockHeadersRequest, reqMsg))
	return true
}

// remote peer requests handler
func (p *BlockProtocol) onGetBlockHeadersRequest(s inet.Stream) {
	defer s.Close()

	peerID := s.Conn().RemotePeer()
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()
	perr := remotePeer.checkState()
	if perr != nil {
		p.log.Info().Msgf("%s: Invalid peer state to handle request %s : %s", peerID.Pretty(), s.Protocol(), perr.Error())
		return
	}

	// get request data
	data := &types.GetBlockHeadersRequest{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Info().Err(err).Msg("fail to decode")
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, data)

	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Info().Msg("Failed to authenticate message")
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
			foundBlock, err := extractBlockFromRequest(p.iserv.CallRequest(message.ChainSvc,
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
			foundBlock, err := extractBlockFromRequest(p.iserv.CallRequest(message.ChainSvc,
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
func (p *BlockProtocol) onGetBlockHeadersResponse(s inet.Stream) {
	defer s.Close()

	peerID := s.Conn().RemotePeer()
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()
	perr := remotePeer.checkState()
	if perr != nil {
		p.log.Info().Msgf("%s: Invalid peer state to handle request %s : %s", peerID.Pretty(), s.Protocol(), perr.Error())
		return
	}

	data := &types.GetBlockHeadersResponse{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, nil)
	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Info().Msg("Failed to authenticate message")
		return
	}

	// send block headers to blockchain service
	p.log.Debug().Msgf("Got blockHeaders response %v \n %v", data.Hashes, data.Headers)
	remotePeer.consumeRequest(data.MessageData.Id)
}

// NotifyNewBlock send notice message of new block to a peer
func (p *BlockProtocol) NotifyNewBlock(newBlock message.NotifyNewBlock) bool {
	// create message data
	for _, neighbor := range p.ps.GetPeers() {
		if neighbor == nil {
			continue
		}
		req := &types.NewBlockNotice{MessageData: &types.MessageData{},
			BlockHash: newBlock.Block.Hash,
			BlockNo:   newBlock.BlockNo}
		msg := newPbMsgBroadcastOrder(false, notifyNewBlockRequest, req)
		if neighbor.State() == types.RUNNING {
			p.log.Debug().Str(LogPeerID, neighbor.meta.ID.Pretty()).Msg("Notifying new block")
			// FIXME need to check if remote peer knows this hash already.
			// but can't do that in peer's write goroutine, since the context is gone in
			// protobuf serialization.
			neighbor.sendMessage(msg)
		}
	}
	return true
}

// remote NotifyNewBlock response handler
func (p *BlockProtocol) onNotifyNewBlock(s inet.Stream) {
	defer s.Close()

	peerID := s.Conn().RemotePeer()
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()
	perr := remotePeer.checkState()
	if perr != nil {
		p.log.Info().Msgf("%s: Invalid peer state to handle request %s : %s", peerID.Pretty(), s.Protocol(), perr.Error())
		return
	}

	data := &types.NewBlockNotice{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID,
		log.DoLazyEval(func() string { return blockchain.EncodeB64(data.BlockHash) }))

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
func (p *BlockProtocol) NotifyBranchBlock(peer *RemotePeer, hash message.BlockHash, blockno types.BlockNo) bool {
	p.log.Debug().Str("peer_id", peer.meta.ID.Pretty()).Msg("Notifying branch block")

	// create message data
	req := &types.NewBlockNotice{MessageData: &types.MessageData{},
		BlockHash: hash,
		BlockNo:   uint64(blockno)}

	peer.sendMessage(newPbMsgRequestOrder(false, false, notifyNewBlockRequest, req))
	return true
}

// replying chain tree
func (p *BlockProtocol) sendMissingResp(remotePeer *RemotePeer, requestID string, missing []message.BlockHash) {
	// find block info from chainservice
	blockInfos := make([]*types.Block, 0, len(missing))
	for _, hash := range missing {
		foundBlock, err := extractBlockFromRequest(p.iserv.CallRequest(message.ChainSvc,
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

	// ???: have to check arguemnts
	remotePeer.sendMessage(newPbMsgResponseOrder(requestID, true, getBlocksResponse, resp))
}

// remote peer requests handler
func (p *BlockProtocol) onGetMissingRequest(s inet.Stream) {
	defer s.Close()

	p.log.Debug().Msg("Received GetMissingRequest request")
	peerID := s.Conn().RemotePeer()
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}
	perr := remotePeer.checkState()
	if perr != nil {
		p.log.Info().Msgf("%s: Invalid peer state to handle request %s : %s", peerID.Pretty(), s.Protocol(), perr.Error())
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()

	// get request data
	data := &types.GetMissingRequest{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Info().Err(err).Msg("fail to decode")
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, log.DoLazyEval(func() string {
		return bytesArrToString(data.Hashes)
	}))
	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Info().Msg("Failed to authenticate message")
		return
	}

	// send to ChainSvc
	// find block info from chainservice
	rawResponse, err := p.iserv.CallRequest(
		message.ChainSvc, &message.GetMissing{Hashes: data.Hashes, StopHash: data.Stophash})
	if err != nil {
		p.log.Warn().Err(err).Msg("failed to get missing")

		return
	}
	v := rawResponse.(message.GetMissingRsp)
	missing := (*message.GetMissingRsp)(&v)

	// generate response message
	p.log.Debug().Msgf("Sending GetMssingRequest response to %s. Message id: %s...", peerID.Pretty(), data.MessageData.Id)

	p.sendMissingResp(remotePeer, data.MessageData.Id, missing.Hashes)
	/*
		for i := 0; i < len(missing.Hashes); i++ {
			p.NotifyBranchBlock(remotePeer, missing.Hashes[i], missing.Blocknos[i])
		}
	*/
}

// remote GetBlock response handler
/*
func (p *BlockProtocol) onGetMissingResponse(s inet.Stream) {
	defer s.Close()

	remotePeer, exists := p.ps.GetPeer(s.Conn().RemotePeer())
	if !exists {
		p.log.Warnf("Request to invalid peer %s ", s.Conn().RemotePeer().Pretty())
		return
	}
	p.log.Debugf("Received GetMissingRequest response from %s.", remotePeer.meta.ID.Pretty())
	data := &types.GetMissingResponse{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		return
	}
	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Info("Failed to authenticate message")
		return
	}

	// send back to caller
	p.log.Infof("Got Missing response ")
	p.iserv.SendRequest(message.ChainSvc, &message.GetMissingRsp{Hashes: data.Hashes, Headers: data.Headers})
	remotePeer.ConsumeRequest(data.MessageData.Id)
}
*/

// GetMissingBlocks send request message to peer about blocks which my local peer doesn't have
func (p *BlockProtocol) GetMissingBlocks(peerID peer.ID, hashes []message.BlockHash) bool {
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		p.log.Warn().Str("peer_id", peerID.Pretty()).Msg("invalid peer id")
		return false
	}
	p.log.Debug().Str("peer_id", peerID.Pretty()).Msg("Send Get Missing Blocks")

	bhashes := make([][]byte, 0)
	for _, a := range hashes {
		bhashes = append(bhashes, a)
	}
	// create message data
	req := &types.GetMissingRequest{
		MessageData: &types.MessageData{},
		Hashes:      bhashes[1:],
		Stophash:    bhashes[0]}

	remotePeer.sendMessage(newPbMsgRequestOrder(false, true, getMissingRequest, req))
	return true
}
