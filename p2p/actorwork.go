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
	req := &types.AddressesRequest{Sender: &senderAddr, MaxSize: 50}
	remotePeer.sendMessage(newPbMsgRequestOrder(true, AddressesRequest, req, p2ps.signer))
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
	remotePeer.sendMessage(newPbMsgRequestOrder(true, GetBlockHeadersRequest, reqMsg, p2ps.signer))
	return true
}

// GetBlocks send request message to peer and
func (p2ps *P2P) GetBlocks(peerID peer.ID, blockHashes []message.BlockHash) bool {
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Str(LogProtoID, string(GetBlocksRequest)).Msg("Message to Unknown peer, check if a bug")
		return false
	}
	p2ps.Debug().Str(LogPeerID, peerID.Pretty()).Int("block_cnt", len(blockHashes)).Msg("Sending Get block request")

	hashes := make([][]byte, len(blockHashes))
	for i, hash := range blockHashes {
		hashes[i] = ([]byte)(hash)
	}
	// create message data
	req := &types.GetBlockRequest{Hashes: hashes}

	remotePeer.sendMessage(newPbMsgRequestOrder(true, GetBlocksRequest, req, p2ps.signer))
	return true
}

// NotifyNewBlock send notice message of new block to a peer
func (p2ps *P2P) NotifyNewBlock(newBlock message.NotifyNewBlock) bool {
	req := &types.NewBlockNotice{
		BlockHash: newBlock.Block.BlockHash(),
		BlockNo:   newBlock.BlockNo}
	msg := newPbMsgBlkBroadcastOrder(req, p2ps.signer)

	skipped, sent := 0, 0
	// create message data
	for _, neighbor := range p2ps.pm.GetPeers() {
		if neighbor != nil && neighbor.State() == types.RUNNING {
			sent++
			neighbor.sendMessage(msg)
		} else {
			skipped++
		}
	}
	p2ps.Debug().Int("skippeer_cnt", skipped).Int("sendpeer_cnt", sent).Str("hash", enc.ToString(newBlock.Block.BlockHash())).Msg("Notifying new block")
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

	remotePeer.sendMessage(newPbMsgRequestOrder(false, GetMissingRequest, req, p2ps.signer))
	return true
}

// GetTXs send request message to peer and
func (p2ps *P2P) GetTXs(peerID peer.ID, txHashes []message.TXHash) bool {
	remotePeer, ok := p2ps.pm.GetPeer(peerID)
	if !ok {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Msg("Invalid peer. check for bug")
		return false
	}
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
	p2ps.Debug().Str(LogPeerID, peerID.Pretty()).Str(LogTxHash, bytesArrToString(hashes)).Msg("Sending GetTransactions request")
	// create message data
	req := &types.GetTransactionsRequest{Hashes: hashes}

	remotePeer.sendMessage(newPbMsgRequestOrder(true, GetTXsRequest, req, p2ps.signer))
	return true
}

// NotifyNewTX notice tx(s) id created
func (p2ps *P2P) NotifyNewTX(newTXs message.NotifyNewTransactions) bool {
	hashes := make([][]byte, len(newTXs.Txs))
	for i, tx := range newTXs.Txs {
		hashes[i] = tx.Hash
	}
	// create message data
	req := &types.NewTransactionsNotice{TxHashes: hashes}
	msg := newPbMsgTxBroadcastOrder(req, p2ps.signer)
	skipped, sent := 0, 0
	// send to peers
	for _, peer := range p2ps.pm.GetPeers() {
		if peer != nil && peer.State() == types.RUNNING {
			sent++
			peer.sendMessage(msg)
		} else {
			skipped++
		}
	}
	p2ps.Debug().Int("skippeer_cnt", skipped).Int("sendpeer_cnt", sent).Str("hashes", bytesArrToString(hashes)).Msg("Notifying newTXs to peers")

	return true
}
