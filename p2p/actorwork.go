/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
)

// GetAddresses send getAddress request to other peer
func (p2ps *P2P) GetAddresses(peerID peer.ID, size uint32) bool {
	remotePeer, ok := p2ps.pm.GetPeer(peerID)
	if !ok {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Msg("Message addressRequest to Unknown peer, check if a bug")

		return false
	}
	senderAddr := p2ps.pm.SelfMeta().ToPeerAddress()
	// create message data
	req := &types.AddressesRequest{MessageData: &types.MessageData{},
		Sender: &senderAddr, MaxSize: 50}
	remotePeer.sendMessage(newPbMsgRequestOrder(true, false, addressesRequest, req))
	return true
}

// GetBlockHeaders send request message to peer and
func (p2ps *P2P) GetBlockHeaders(msg *message.GetBlockHeaders) bool {
	remotePeer, exists := p2ps.pm.GetPeer(msg.ToWhom)
	if !exists {
		p2ps.Warn().Str(LogPeerID, msg.ToWhom.Pretty()).Msg("Request to invalid peer")
		return false
	}
	peerID := remotePeer.meta.ID

	p2ps.Debug().Str(LogPeerID, peerID.Pretty()).Interface("msg", msg).Msg("Sending Get block Header request")
	// create message data
	reqMsg := &types.GetBlockHeadersRequest{Hash: msg.Hash,
		Height: msg.Height, Offset: msg.Offset, Size: msg.MaxSize, Asc: msg.Asc,
	}
	remotePeer.sendMessage(newPbMsgRequestOrder(true, true, getBlockHeadersRequest, reqMsg))
	return true
}

// GetBlocks send request message to peer and
func (p2ps *P2P) GetBlocks(peerID peer.ID, blockHashes []message.BlockHash) bool {
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Str(LogProtoID, string(getBlocksRequest)).Msg("Message to Unknown peer, check if a bug")
		return false
	}
	p2ps.Debug().Str(LogPeerID, peerID.Pretty()).Int("block_cnt", len(blockHashes)).Msg("Sending Get block request")

	hashes := make([][]byte, len(blockHashes))
	for i, hash := range blockHashes {
		hashes[i] = ([]byte)(hash)
	}
	// create message data
	req := &types.GetBlockRequest{Hashes: hashes}

	remotePeer.sendMessage(newPbMsgRequestOrder(true, true, getBlocksRequest, req))
	return true
}

// NotifyNewBlock send notice message of new block to a peer
func (p2ps *P2P) NotifyNewBlock(newBlock message.NotifyNewBlock) bool {
	// create message data
	for _, neighbor := range p2ps.pm.GetPeers() {
		if neighbor == nil {
			continue
		}
		req := &types.NewBlockNotice{
			BlockHash: newBlock.Block.Hash,
			BlockNo:   newBlock.BlockNo}
		msg := newPbMsgBroadcastOrder(false, newBlockNotice, req)
		if neighbor.State() == types.RUNNING {
			p2ps.Debug().Str(LogPeerID, neighbor.meta.ID.Pretty()).Str("hash", enc.ToString(newBlock.Block.Hash)).Msg("Notifying new block")
			// FIXME need to check if remote peer knows this hash already.
			// but can't do that in peer's write goroutine, since the context is gone in
			// protobuf serialization.
			neighbor.sendMessage(msg)
		}
	}
	return true
}

// GetMissingBlocks send request message to peer about blocks which my local peer doesn't have
func (p2ps *P2P) GetMissingBlocks(peerID peer.ID, hashes []message.BlockHash) bool {
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Msg("invalid peer id")
		return false
	}
	p2ps.Debug().Str(LogPeerID, peerID.Pretty()).Msg("Send Get Missing Blocks")

	bhashes := make([][]byte, 0)
	for _, a := range hashes {
		bhashes = append(bhashes, a)
	}
	// create message data
	req := &types.GetMissingRequest{
		Hashes:   bhashes[1:],
		Stophash: bhashes[0]}

	remotePeer.sendMessage(newPbMsgRequestOrder(false, true, getMissingRequest, req))
	return true
}

// GetTXs send request message to peer and
func (p2ps *P2P) GetTXs(peerID peer.ID, txHashes []message.TXHash) bool {
	remotePeer, ok := p2ps.pm.GetPeer(peerID)
	if !ok {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Msg("Invalid peer. check for bug")
		return false
	}
	p2ps.Debug().Str(LogPeerID, peerID.Pretty()).Int("tx_cnt", len(txHashes)).Msg("Sending GetTransactions request")
	if len(txHashes) == 0 {
		p2ps.Warn().Msg("empty hash list")
		return false
	}

	hashes := make([][]byte, len(txHashes))
	for i, hash := range txHashes {
		if len(hash) == 0 {
			p2ps.Warn().Msg("empty hash value requested.")
			return false
		}
		hashes[i] = ([]byte)(hash)
	}
	// create message data
	req := &types.GetTransactionsRequest{Hashes: hashes}

	remotePeer.sendMessage(newPbMsgRequestOrder(true, true, getTXsRequest, req))
	return true
}

// NotifyNewTX notice tx(s) id created
func (p2ps *P2P) NotifyNewTX(newTXs message.NotifyNewTransactions) bool {
	hashes := make([][]byte, len(newTXs.Txs))
	for i, tx := range newTXs.Txs {
		hashes[i] = tx.Hash
	}
	p2ps.Debug().Int("peer_cnt", len(p2ps.pm.GetPeers())).Str("hashes", bytesArrToString(hashes)).Msg("Notifying newTXs to peers")
	// send to peers
	for _, peer := range p2ps.pm.GetPeers() {
		// create message data
		req := &types.NewTransactionsNotice{TxHashes: hashes}
		peer.sendMessage(newPbMsgBroadcastOrder(false, newTxNotice, req))
	}

	return true
}
