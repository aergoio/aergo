/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"reflect"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/etcd/raft/raftpb"
)

const (
	// fetchTimeOut was copied from syncer package. it can be problem if these value become different
	fetchTimeOut = time.Second * 100
)

// GetAddresses send getAddress request to other peer
func (p2ps *P2P) GetAddresses(peerID types.PeerID, size uint32) bool {
	remotePeer, ok := p2ps.pm.GetPeer(peerID)
	if !ok {
		p2ps.Warn().Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Msg("Message addressRequest to Unknown peer, check if a bug")

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
		p2ps.Warn().Stringer(p2putil.LogPeerID, types.LogPeerShort(msg.ToWhom)).Msg("Request to invalid peer")
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
		p2ps.Warn().Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Stringer(p2putil.LogProtoID, p2pcommon.GetBlocksRequest).Msg("Message to Unknown peer, check if a bug")
		return false
	}
	if len(blockHashes) == 0 {
		p2ps.Warn().Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Stringer(p2putil.LogProtoID, p2pcommon.GetBlocksRequest).Msg("meaningless GetBlocks request with zero hash")
		return false
	}
	p2ps.Debug().Str(p2putil.LogPeerName, remotePeer.Name()).Int(p2putil.LogBlkCount, len(blockHashes)).Stringer("first_hash", types.LogBase58(blockHashes[0])).Msg("Sending Get block request")

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
		p2ps.Warn().Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Stringer(p2putil.LogProtoID, p2pcommon.GetBlocksRequest).Msg("Message to Unknown peer, check if a bug")
		context.Respond(&message.GetBlockChunksRsp{Seq: msg.Seq, ToWhom: peerID, Err: fmt.Errorf("invalid peer")})
		return
	}
	receiver := NewBlockReceiver(p2ps, remotePeer, msg.Seq, blockHashes, msg.TTL)
	receiver.StartGet()
}

// GetBlockHashes send request message to peer and make response message for block hashes
func (p2ps *P2P) GetBlockHashes(context actor.Context, msg *message.GetHashes) {
	peerID := msg.ToWhom

	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Stringer(p2putil.LogProtoID, p2pcommon.GetHashesRequest).Msg("Invalid peerID")
		context.Respond(&message.GetHashesRsp{Seq: msg.Seq, Hashes: nil, PrevInfo: msg.PrevInfo, Count: 0, Err: message.PeerNotFoundError})
		return
	}
	receiver := NewBlockHashesReceiver(p2ps, remotePeer, msg.Seq, msg, fetchTimeOut)
	receiver.StartGet()
}

// GetBlockHashes send request message to peer and make response message for block hashes
func (p2ps *P2P) GetBlockHashByNo(context actor.Context, msg *message.GetHashByNo) {
	peerID := msg.ToWhom

	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Stringer(p2putil.LogProtoID, p2pcommon.GetHashByNoRequest).Msg("Invalid peerID")
		context.Respond(&message.GetHashByNoRsp{Seq: msg.Seq, Err: message.PeerNotFoundError})
		return
	}
	receiver := NewBlockHashByNoReceiver(p2ps, remotePeer, msg.Seq, msg.BlockNo, fetchTimeOut)
	receiver.StartGet()
}

// NotifyNewBlock send notice message of new block to a peer
func (p2ps *P2P) NotifyNewBlock(blockNotice message.NotifyNewBlock) bool {
	req := &types.NewBlockNotice{
		BlockHash: blockNotice.Block.BlockHash(),
		BlockNo:   blockNotice.BlockNo}
	mo := p2ps.mf.NewMsgBlkBroadcastOrder(req)

	// sending new block notice (relay inv message is not need to every nodes)
	peers := p2ps.prm.FilterNewBlockNoticeReceiver(blockNotice.Block, p2ps.pm)
	sent, skipped := 0, 0
	for _, neighbor := range peers {
		if neighbor != nil && neighbor.State() == types.RUNNING {
			sent++
			neighbor.SendMessage(mo)
		} else {
			skipped++
		}
	}

	p2ps.Debug().Int("skipped_cnt", skipped).Int("sent_cnt", sent).Str("hash", enc.ToString(blockNotice.Block.BlockHash())).Msg("Notifying new block")
	return true
}

// NotifyBlockProduced send notice message of new block to a peer
func (p2ps *P2P) NotifyBlockProduced(blockNotice message.NotifyNewBlock) bool {
	// TODO fill producerID, but actually there is no way go find producer, for now.
	req := &types.BlockProducedNotice{ProducerID: nil, BlockNo: blockNotice.BlockNo, Block: blockNotice.Block}
	mo := p2ps.mf.NewMsgBPBroadcastOrder(req)

	peers := p2ps.pm.GetPeers()
	sent, skipped := 0, 0
	for _, neighbor := range peers {
		if neighbor.State() == types.RUNNING {
			sent++
			neighbor.SendMessage(mo)
		} else {
			skipped++
		}
	}

	p2ps.Debug().Int("skipped_cnt", skipped).Int("sent_cnt", sent).Str("hash", enc.ToString(blockNotice.Block.BlockHash())).Uint64("block_no", req.BlockNo).Msg("Notifying block produced")
	return true
}

// GetTXs send request message to peer and
func (p2ps *P2P) GetTXs(peerID types.PeerID, txHashes []message.TXHash) bool {
	remotePeer, ok := p2ps.pm.GetPeer(peerID)
	if !ok {
		p2ps.Warn().Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Msg("Invalid peer. check for bug")
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
func (p2ps *P2P) NotifyNewTX(msg *message.NotifyNewTransactions) bool {
	hashes := make([]types.TxID, len(msg.Txs))
	for i, tx := range msg.Txs {
		hashes[i] = types.ToTxID(tx.Hash)
	}
	// create message data
	skipped, sent := 0, 0
	// send to peers
	peers := p2ps.pm.GetPeers()
	p2ps.sm.RegisterTxNotice(msg.Txs)
	for _, rPeer := range peers {
		if rPeer != nil && rPeer.State() == types.RUNNING {
			sent++
			rPeer.PushTxsNotice(hashes)
		} else {
			skipped++
		}
	}
	//p2ps.Debug().Int("skippeer_cnt", skipped).Int("sendpeer_cnt", sent).Int("hash_cnt", len(hashes)).Msg("Notifying newTXs to peers")

	return true
}

// GetSyncAncestor request remote peer to find ancestor
func (p2ps *P2P) GetSyncAncestor(context actor.Context, msg *message.GetSyncAncestor) {
	peerID := msg.ToWhom
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		p2ps.Warn().Stringer(p2putil.LogPeerID, types.LogPeerShort(peerID)).Msg("invalid peer id")
		context.Respond(&message.GetSyncAncestorRsp{Seq: msg.Seq, Ancestor: nil})
		return
	}
	if len(msg.Hashes) == 0 {
		p2ps.Warn().Str(p2putil.LogPeerName, remotePeer.Name()).Msg("empty hash list received")
		context.Respond(&message.GetSyncAncestorRsp{Seq: msg.Seq, Ancestor: nil})
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
		p2ps.Error().Str("actual", reflect.TypeOf(msg.Body).String()).Msg("body is not raftpb.Message")
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

func (p2ps *P2P) SendIssueCertMessage(context actor.Context, msg message.IssueAgentCertificate) {
	peerID := msg.ProducerID
	remotePeer, exists := p2ps.pm.GetPeer(peerID)
	if !exists {
		return
	}
	if remotePeer.AcceptedRole() != types.PeerRole_Producer {
		p2ps.Debug().Str(p2putil.LogPeerName, remotePeer.Name()).Msg("to peer")
		return
	}

	body := &types.IssueCertificateRequest{}
	remotePeer.SendMessage(p2ps.mf.NewMsgRequestOrder(true, p2pcommon.IssueCertificateRequest, body))
}

func (p2ps *P2P) NotifyCertRenewed(context actor.Context, renewed message.NotifyCertRenewed) {
	body := &types.CertificateRenewedNotice{Certificate: renewed.Cert}
	msg := p2ps.mf.NewMsgRequestOrder(false, p2pcommon.CertificateRenewedNotice, body)

	skipped, sent := 0, 0
	for _, neighbor := range p2ps.pm.GetPeers() {
		if neighbor != nil && neighbor.State() == types.RUNNING {
			sent++
			neighbor.SendMessage(msg)
		} else {
			skipped++
		}
	}
	p2ps.Debug().Int("skipped_cnt", skipped).Int("sent_cnt", sent).Str("cert", renewed.Cert.String()).Msg("Notifying certificate renewed")

}

func (p2ps *P2P) TossBPNotice(msg message.TossBPNotice) bool {
	orgMsg := msg.OriginalMsg.(p2pcommon.Message)
	mo := p2ps.mf.NewTossMsgOrder(orgMsg)

	targetZone := p2pcommon.PeerZone(msg.TossIn)
	peers := p2ps.prm.FilterBPNoticeReceiver(msg.Block, p2ps.pm, targetZone)
	skipped, sent := 0, 0
	for _, neighbor := range peers {
		if neighbor != nil && neighbor.State() == types.RUNNING {
			sent++
			neighbor.SendMessage(mo)
		} else {
			skipped++
		}
	}

	p2ps.Debug().Int("skipCnt", skipped).Int("sendCnt", sent).Str("Target", targetZone.String()).Str(p2putil.LogMsgID, orgMsg.ID().String()).Msg("Tossing block produced notice")
	return true
}
