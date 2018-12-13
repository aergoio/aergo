/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
	"time"
)

const (
	// fetchTimeOut was copied from syncer package. it can be problem if these value become different
	fetchTimeOut     = time.Second * 100
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
	remotePeer.sendMessage(p2ps.mf.newMsgRequestOrder(true, AddressesRequest, req))
	return true
}

// GetBlockHeaders send request message to peer and
func (p2ps *P2P) GetBlockHeaders(msg *message.GetBlockHeaders) bool {
	remotePeer, exists := p2ps.pm.GetPeer(msg.ToWhom)
	if !exists {
		p2ps.Warn().Str(LogPeerID, msg.ToWhom.Pretty()).Msg("Request to invalid peer")
		return false
	}
	peerID := remotePeer.ID()

	p2ps.Debug().Str(LogPeerID, peerID.Pretty()).Interface("msg", msg).Msg("Sending Get block Header request")
	// create message data
	reqMsg := &types.GetBlockHeadersRequest{Hash: msg.Hash,
		Height: msg.Height, Offset: msg.Offset, Size: msg.MaxSize, Asc: msg.Asc,
	}
	remotePeer.sendMessage(p2ps.mf.newMsgRequestOrder(true, GetBlockHeadersRequest, reqMsg))
	return true
}

// GetBlocks send request message to peer and
func (p2ps *P2P) GetBlocks(peerID peer.ID, blockHashes []message.BlockHash) bool {
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Str(LogProtoID, string(GetBlocksRequest)).Msg("Message to Unknown peer, check if a bug")
		return false
	}
	if len(blockHashes) == 0 {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Str(LogProtoID, string(GetBlocksRequest)).Msg("meaningless GetBlocks request with zero hash")
		return false
	}
	p2ps.Debug().Str(LogPeerID, peerID.Pretty()).Int(LogBlkCount, len(blockHashes)).Str("first_hash", enc.ToString(blockHashes[0])).Msg("Sending Get block request")

	hashes := make([][]byte, len(blockHashes))
	for i, hash := range blockHashes {
		hashes[i] = ([]byte)(hash)
	}
	// create message data
	req := &types.GetBlockRequest{Hashes: hashes}

	remotePeer.sendMessage(p2ps.mf.newMsgRequestOrder(true, GetBlocksRequest, req))
	return true
}


// GetBlocksChunk send request message to peer and
func (p2ps *P2P) GetBlocksChunk(context actor.Context, msg *message.GetBlockChunks) {
	peerID := msg.ToWhom
	blockHashes := msg.Hashes
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Str(LogProtoID, GetBlocksRequest.String()).Msg("Message to Unknown peer, check if a bug")
		context.Respond(&message.GetBlockChunksRsp{ToWhom:peerID, Err:fmt.Errorf("invalid peer")})
		return
	}
	receiver := NewBlockReceiver(p2ps, remotePeer, blockHashes, msg.TTL)
	receiver.StartGet()
}


// GetBlockHashes send request message to peer and make response message for block hashes
func (p2ps *P2P) GetBlockHashes(context actor.Context, msg *message.GetHashes) {
	peerID := msg.ToWhom
	// TODO
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Str(LogProtoID, GetHashesRequest.String()).Msg("Invalid peerID")
		context.Respond(&message.GetHashesRsp{Hashes:nil, PrevInfo:msg.PrevInfo, Count:0, Err:message.PeerNotFoundError})
		return
	}
	receiver := NewBlockHashesReceiver(p2ps, remotePeer, msg, fetchTimeOut)
	receiver.StartGet()
}

// GetBlockHashes send request message to peer and make response message for block hashes
func (p2ps *P2P) GetBlockHashByNo(context actor.Context, msg *message.GetHashByNo) {
	peerID := msg.ToWhom
	// TODO
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Str(LogProtoID, GetHashByNoRequest.String()).Msg("Invalid peerID")
		context.Respond(&message.GetHashByNoRsp{Err:message.PeerNotFoundError})
		return
	}
	receiver := NewBlockHashByNoReceiver(p2ps, remotePeer, msg.BlockNo, fetchTimeOut)
	receiver.StartGet()
}

// NotifyNewBlock send notice message of new block to a peer
func (p2ps *P2P) NotifyNewBlock(newBlock message.NotifyNewBlock) bool {
	req := &types.NewBlockNotice{
		BlockHash: newBlock.Block.BlockHash(),
		BlockNo:   newBlock.BlockNo}
	msg := p2ps.mf.newMsgBlkBroadcastOrder(req)

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
	p2ps.Debug().Int("skipped_cnt", skipped).Int("sent_cnt", sent).Str("hash", enc.ToString(newBlock.Block.BlockHash())).Msg("Notifying new block")
	return true
}

// GetMissingBlocks send request message to peer about blocks which my local peer doesn't have
func (p2ps *P2P) GetMissingBlocks(peerID peer.ID, hashes []message.BlockHash) bool {
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Msg("invalid peer id")
		return false
	}
	//p2ps.Debug().Str(LogPeerID, peerID.Pretty()).Msg("Send Get Missing Blocks")
	if len(hashes) == 0 {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Msg("empty hash list received")
		return false
	}

	p2ps.sm.DoSync(remotePeer, hashes[1:], hashes[0])
	//// create message data
	//req := &types.GetMissingRequest{
	//	Hashes:   bhashes[1:],
	//	Stophash: bhashes[0]}
	//
	//remotePeer.sendMessage(p2ps.mf.newMsgRequestOrder(false, GetMissingRequest, req))
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
	// create message data
	req := &types.GetTransactionsRequest{Hashes: hashes}

	remotePeer.sendMessage(p2ps.mf.newMsgRequestOrder(true, GetTXsRequest, req))
	return true
}

// NotifyNewTX notice tx(s) id created
func (p2ps *P2P) NotifyNewTX(newTXs message.NotifyNewTransactions) bool {
	hashes := make([]TxHash, len(newTXs.Txs))
	for i, tx := range newTXs.Txs {
		copy(hashes[i][:], tx.Hash)
	}
	// create message data
	skipped, sent := 0, 0
	// send to peers
	for _, rPeer := range p2ps.pm.GetPeers() {
		if rPeer != nil && rPeer.State() == types.RUNNING {
			sent++
			rPeer.pushTxsNotice(hashes)
		} else {
			skipped++
		}
	}
	//p2ps.Debug().Int("skippeer_cnt", skipped).Int("sendpeer_cnt", sent).Int("hash_cnt", len(hashes)).Msg("Notifying newTXs to peers")

	return true
}

// Syncer.finder request remote peer to find ancestor
func (p2ps *P2P) GetSyncAncestor(peerID peer.ID, hashes [][]byte) bool {
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Msg("invalid peer id")
		return false
	}
	//p2ps.Debug().Str(LogPeerID, peerID.Pretty()).Msg("Send Get Missing Blocks")
	if len(hashes) == 0 {
		p2ps.Warn().Str(LogPeerID, peerID.Pretty()).Msg("empty hash list received")
		return false
	}

	// create message data
	req := &types.GetAncestorRequest{Hashes: hashes}

	remotePeer.sendMessage(p2ps.mf.newMsgRequestOrder(true, GetAncestorRequest, req))
	return true
}
