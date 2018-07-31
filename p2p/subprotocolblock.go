/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"reflect"
	"sync"

	"github.com/libp2p/go-libp2p-peer"

	inet "github.com/libp2p/go-libp2p-net"

	"github.com/aergoio/aergo/blockchain"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	"github.com/multiformats/go-multicodec/protobuf"
)

// pattern: /protocol-name/request-or-response-message/version
const (
	getBlocksRequest        = "/blk/getreq/0.2"
	getBlocksResponse       = "/blk/getresp/0.2"
	getBlockHeadersRequest  = "/blk/headerreq/0.1"
	getBlockHeadersResponse = "/blk/headerresp/0.1"
	getMissingRequest       = "/blk/getmreq/0.1"
	getMissingResponse      = "/blk/getmresp/0.1"
	notifyNewBlockRequest   = "/blk/newblockreq/0.1"
)

// BlockProtocol handle block messages.
// Relaying is not implemented yet.
type BlockProtocol struct {
	iserv ActorService

	ps       PeerManager
	reqMutex sync.Mutex

	chainsvc *blockchain.ChainService
	log      log.ILogger
}

// NewBlockProtocol create block subprotocol
func NewBlockProtocol(logger log.ILogger, chainsvc *blockchain.ChainService) *BlockProtocol {
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
	peerID := s.Conn().RemotePeer()
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}
	// get request data
	data := &types.GetBlockRequest{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Info(err)
		return
	}

	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, len(data.Hashes))

	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Info("Failed to authenticate message")
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
	peerID := s.Conn().RemotePeer()
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
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
		p.log.Info("Failed to authenticate message")
		return
	}
	// locate request data and remove it if found
	remotePeer.consumeRequest(data.MessageData.Id)

	// got block
	p.log.Debugf("Request chainservice to add %d blocks", len(data.Blocks))
	for _, block := range data.Blocks {
		p.iserv.SendRequest(message.ChainSvc, &message.AddBlock{PeerID: s.Conn().RemotePeer(), Block: block})
	}

}

// GetBlocks send request message to peer and
func (p *BlockProtocol) GetBlocks(peerID peer.ID, blockHashes []message.BlockHash) bool {
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		p.log.Warnf("Message %s to Unknown peer %s, check for bug.", getBlocksRequest, peerID.Pretty())
		return false
	}
	p.log.Debugf("Sending Get block request to: %s...(%d blocks)", peerID.Pretty(), len(blockHashes))

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
		p.log.Warnf("Request to invalid peer %s ", msg.ToWhom.Pretty())
		return false
	}
	peerID := remotePeer.meta.ID

	p.log.Debugf("Sending Get block Header request to: %s (%v)", peerID.Pretty(), msg)
	// create message data
	reqMsg := &types.GetBlockHeadersRequest{MessageData: &types.MessageData{}, Hash: msg.Hash,
		Height: msg.Height, Offset: msg.Offset, Size: msg.MaxSize, Asc: msg.Asc,
	}
	remotePeer.sendMessage(newPbMsgRequestOrder(true, true, getBlockHeadersRequest, reqMsg))
	return true
}

// remote peer requests handler
func (p *BlockProtocol) onGetBlockHeadersRequest(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	// get request data
	data := &types.GetBlockHeadersRequest{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Info(err)
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, data)

	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Info("Failed to authenticate message")
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
	peerID := s.Conn().RemotePeer()
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
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
		p.log.Info("Failed to authenticate message")
		return
	}

	// TODO: send back to caller
	// send block headers to blockchain service
	p.log.Debugf("Got blockHeaders response %v \n %v", data.Hashes, data.Headers)
	remotePeer.consumeRequest(data.MessageData.Id)
}

// NotifyNewBlock send notice message of new block to a peer
func (p *BlockProtocol) NotifyNewBlock(newBlock message.NotifyNewBlock) bool {
	for _, neighbor := range p.ps.GetPeers() {
		p.log.Debugf("Notifying new block to: %s", neighbor.meta.ID.Pretty())
		// create message data
		req := &types.NewBlockNotice{MessageData: &types.MessageData{},
			BlockHash: newBlock.Block.Hash,
			BlockNo:   newBlock.BlockNo}
		neighbor.sendMessage(newPbMsgBroadcastOrder(false, notifyNewBlockRequest, req))
	}
	return true
}

// remote NotifyNewBlock response handler
func (p *BlockProtocol) onNotifyNewBlock(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	_, exists := p.ps.GetPeer(peerID)
	if !exists {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
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

	// request block info if selfnode does not have block already
	rawResp, err := p.iserv.CallRequest(message.ChainSvc, &message.GetBlock{BlockHash: message.BlockHash(data.BlockHash)})
	if err != nil {
		p.log.Warnf("actor return error on getblock : %s", err.Error())
		return
	}
	resp, ok := rawResp.(message.GetBlockRsp)
	if !ok {
		p.log.Warnf("chainservice return unexpected type : %v", reflect.TypeOf(rawResp))
		return
	}
	if resp.Err != nil {
		p.log.Debugf("chainservice responded that block %s not found. so request back to notifier: %s", blockchain.EncodeB64(data.BlockHash), peerID.Pretty())
		p.iserv.SendRequest(message.P2PSvc, &message.GetBlockInfos{ToWhom: peerID,
			Hashes: []message.BlockHash{message.BlockHash(data.BlockHash)}})
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
func (p *BlockProtocol) NotifyBranchBlock(peer *RemotePeer, hash message.BlockHash, blockno types.BlockNo) bool {
	p.log.Debugf("Notifying branch block to: %s", peer.meta.ID.Pretty())

	// create message data
	req := &types.NewBlockNotice{MessageData: &types.MessageData{},
		BlockHash: hash,
		BlockNo:   uint64(blockno)}

	peer.sendMessage(newPbMsgRequestOrder(false, false, notifyNewBlockRequest, req))
	return true
}

// remote peer requests handler
func (p *BlockProtocol) onGetMissingRequest(s inet.Stream) {
	p.log.Debugf("Received GetMissingRequest request")
	peerID := s.Conn().RemotePeer()
	remotePeer, exists := p.ps.GetPeer(peerID)
	if !exists {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	// get request data
	data := &types.GetMissingRequest{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Info(err)
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, data.Hashes)
	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Info("Failed to authenticate message")
		return
	}

	// send to ChainSvc
	// find block info from chainservice
	rawResponse, err := p.iserv.CallRequest(
		message.ChainSvc, &message.GetMissing{Hashes: data.Hashes, StopHash: data.Stophash})
	if err != nil {
		p.log.Warnf("failed to get missing : %s ", err.Error())
		return
	}
	v := rawResponse.(message.GetMissingRsp)
	missing := (*message.GetMissingRsp)(&v)

	// generate response message
	p.log.Debugf("Sending GetMssingRequest response to %s. Message id: %s...", peerID.Pretty(), data.MessageData.Id)

	for i := 0; i < len(missing.Hashes); i++ {
		p.NotifyBranchBlock(remotePeer, missing.Hashes[i], missing.Blocknos[i])
	}
}

// remote GetBlock response handler
/*
func (p *BlockProtocol) onGetMissingResponse(s inet.Stream) {
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
		p.log.Warnf("invalid peer id %s ", peerID.Pretty())
		return false
	}
	p.log.Debugf("Send Get Missing Blocks to: %s", peerID.Pretty())

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
