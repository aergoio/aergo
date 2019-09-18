/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/etcd/raft/raftpb"
	"reflect"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
)

const (
	// fetchTimeOut was copied from syncer package. it can be problem if these value become different
	fetchTimeOut = time.Second * 100
)

// GetAddresses send getAddress request to other peer
func (p2ps *P2P) GetAddresses(peerID types.PeerID, size uint32) bool {
	remotePeer, ok := p2ps.pm.GetPeer(peerID)
	if !ok {
		p2ps.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("Message addressRequest to Unknown peer, check if a bug")

		return false
	}
	senderAddr := p2ps.SelfMeta().ToPeerAddress()
	// createPolaris message data
	req := &types.AddressesRequest{Sender: &senderAddr, MaxSize: 50}
	remotePeer.SendMessage(p2ps.mf.NewMsgRequestOrder(true, p2pcommon.AddressesRequest, req))
	return true
}

// GetBlockHeaders send request message to peer and
func (p2ps *P2P) GetBlockHeaders(msg *message.GetBlockHeaders) bool {
	remotePeer, exists := p2ps.pm.GetPeer(msg.ToWhom)
	if !exists {
		p2ps.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(msg.ToWhom)).Msg("Request to invalid peer")
		return false
	}

	p2ps.Debug().Str(p2putil.LogPeerName, remotePeer.Name()).Interface("msg", msg).Msg("Sending Get block Header request")
	// create message data
	reqMsg := &types.GetBlockHeadersRequest{Hash: msg.Hash,
		Height: msg.Height, Offset: msg.Offset, Size: msg.MaxSize, Asc: msg.Asc,
	}
	remotePeer.SendMessage(p2ps.mf.NewMsgRequestOrder(true, p2pcommon.GetBlockHeadersRequest, reqMsg))
	return true
}

// GetBlocks send request message to peer and
func (p2ps *P2P) GetBlocks(peerID types.PeerID, blockHashes []message.BlockHash) bool {
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Str(p2putil.LogProtoID, string(p2pcommon.GetBlocksRequest)).Msg("Message to Unknown peer, check if a bug")
		return false
	}
	if len(blockHashes) == 0 {
		p2ps.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Str(p2putil.LogProtoID, string(p2pcommon.GetBlocksRequest)).Msg("meaningless GetBlocks request with zero hash")
		return false
	}
	p2ps.Debug().Str(p2putil.LogPeerName, remotePeer.Name()).Int(p2putil.LogBlkCount, len(blockHashes)).Str("first_hash", enc.ToString(blockHashes[0])).Msg("Sending Get block request")

	hashes := make([][]byte, len(blockHashes))
	for i, hash := range blockHashes {
		hashes[i] = ([]byte)(hash)
	}
	// create message data
	req := &types.GetBlockRequest{Hashes: hashes}

	remotePeer.SendMessage(p2ps.mf.NewMsgRequestOrder(true, p2pcommon.GetBlocksRequest, req))
	return true
}

// GetBlocksChunk send request message to peer and
func (p2ps *P2P) GetBlocksChunk(context actor.Context, msg *message.GetBlockChunks) {
	peerID := msg.ToWhom
	blockHashes := msg.Hashes
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Str(p2putil.LogProtoID, p2pcommon.GetBlocksRequest.String()).Msg("Message to Unknown peer, check if a bug")
		context.Respond(&message.GetBlockChunksRsp{Seq:msg.Seq, ToWhom: peerID, Err: fmt.Errorf("invalid peer")})
		return
	}
	receiver := NewBlockReceiver(p2ps, remotePeer, msg.Seq, blockHashes, msg.TTL)
	receiver.StartGet()
}

// GetBlockHashes send request message to peer and make response message for block hashes
func (p2ps *P2P) GetBlockHashes(context actor.Context, msg *message.GetHashes) {
	peerID := msg.ToWhom
	// TODO
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Str(p2putil.LogProtoID, p2pcommon.GetHashesRequest.String()).Msg("Invalid peerID")
		context.Respond(&message.GetHashesRsp{Seq:msg.Seq, Hashes: nil, PrevInfo: msg.PrevInfo, Count: 0, Err: message.PeerNotFoundError})
		return
	}
	receiver := NewBlockHashesReceiver(p2ps, remotePeer, msg.Seq, msg, fetchTimeOut)
	receiver.StartGet()
}

// GetBlockHashes send request message to peer and make response message for block hashes
func (p2ps *P2P) GetBlockHashByNo(context actor.Context, msg *message.GetHashByNo) {
	peerID := msg.ToWhom
	// TODO
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Str(p2putil.LogProtoID, p2pcommon.GetHashByNoRequest.String()).Msg("Invalid peerID")
		context.Respond(&message.GetHashByNoRsp{Seq:msg.Seq, Err: message.PeerNotFoundError})
		return
	}
	receiver := NewBlockHashByNoReceiver(p2ps, remotePeer, msg.Seq, msg.BlockNo, fetchTimeOut)
	receiver.StartGet()
}

// NotifyNewBlock send notice message of new block to a peer
func (p2ps *P2P) NotifyNewBlock(newBlock message.NotifyNewBlock) bool {
	req := &types.NewBlockNotice{
		BlockHash: newBlock.Block.BlockHash(),
		BlockNo:   newBlock.BlockNo}
	msg := p2ps.mf.NewMsgBlkBroadcastOrder(req)

	// sending new block notice (relay inv message is not need to every nodes)
	skipped, sent := p2ps.prm.NotifyNewBlockMsg(msg, p2ps.pm.GetPeers())

	p2ps.Debug().Int("skipped_cnt", skipped).Int("sent_cnt", sent).Str("hash", enc.ToString(newBlock.Block.BlockHash())).Msg("Notifying new block")
	return true
}

// NotifyNewBlock send notice message of new block to a peer
func (p2ps *P2P) NotifyBlockProduced(newBlock message.NotifyNewBlock) bool {
	// TODO fill producerID
	req := &types.BlockProducedNotice{ProducerID: nil, BlockNo: newBlock.BlockNo, Block: newBlock.Block}
	msg := p2ps.mf.NewMsgBPBroadcastOrder(req)

	skipped, sent := p2ps.prm.NotifyNewBlockMsg(msg, p2ps.pm.GetPeers())
	// TODO filter to only contain bp and trusted node.
	//for _, neighbor := range p2ps.pm.GetPeers() {
	//	if neighbor != nil && neighbor.State() == types.RUNNING {
	//		sent++
	//		neighbor.SendMessage(msg)
	//	} else {
	//		skipped++
	//	}
	//}
	p2ps.Debug().Int("skipped_cnt", skipped).Int("sent_cnt", sent).Str("hash", enc.ToString(newBlock.Block.BlockHash())).Uint64("block_no", req.BlockNo).Msg("Notifying block produced")
	return true
}

// GetTXs send request message to peer and
func (p2ps *P2P) GetTXs(peerID types.PeerID, txHashes []message.TXHash) bool {
	remotePeer, ok := p2ps.pm.GetPeer(peerID)
	if !ok {
		p2ps.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("Invalid peer. check for bug")
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

	remotePeer.SendMessage(p2ps.mf.NewMsgRequestOrder(true, p2pcommon.GetTXsRequest, req))
	return true
}

// NotifyNewTX notice tx(s) id created
func (p2ps *P2P) NotifyNewTX(newTXs notifyNewTXs) bool {
	hashes := newTXs.ids
	// create message data
	skipped, sent := 0, 0
	// send to peers
	peers := p2ps.pm.GetPeers()
	p2ps.tnt.RegisterTxNotice(hashes, len(peers), newTXs.alreadySent)
	for _, rPeer := range peers {
		if rPeer != nil && rPeer.State() == types.RUNNING {
			sent++
			rPeer.PushTxsNotice(hashes)
		} else {
			skipped++
		}
	}
	//p2ps.Debug().Int("skippeer_cnt", skipped).Int("sendpeer_cnt", sent).Int("hash_cnt", len(hashes)).Msg("Notifying newTXs to peers")
	if skipped > 0 {
		p2ps.tnt.ReportNotSend(hashes, skipped)
	}

	return true
}

// GetSyncAncestor request remote peer to find ancestor
func (p2ps *P2P) GetSyncAncestor(context actor.Context, msg *message.GetSyncAncestor) {
	peerID := msg.ToWhom
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("invalid peer id")
		context.Respond(&message.GetSyncAncestorRsp{Seq: msg.Seq, Ancestor:nil})
		return
	}
	if len(msg.Hashes) == 0 {
		p2ps.Warn().Str(p2putil.LogPeerName, remotePeer.Name()).Msg("empty hash list received")
		context.Respond(&message.GetSyncAncestorRsp{Seq: msg.Seq, Ancestor:nil})
		return
	}

	// create message data
	receiver := NewAncestorReceiver(p2ps, remotePeer, msg.Seq, msg.Hashes, fetchTimeOut)
	receiver.StartGet()
	return
}

func (p2ps *P2P) SendRaftMessage(context actor.Context, msg *message.SendRaft) {
	body, ok := msg.Body.(raftpb.Message)
	if !ok {
		p2ps.Error().Str("actual", reflect.TypeOf(msg.Body).String() ).Msg("body is not raftpb.Message")
		return
	}
	peerID := msg.ToWhom
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		// temporarily comment out warning log, since current http/p2p hybrid env can cause too much logs
		p2ps.consacc.RaftAccessor().ReportUnreachable(peerID)
		return
	}
	remotePeer.SendMessage(p2ps.mf.NewRaftMsgOrder(body.Type, &body))
	// return success
}