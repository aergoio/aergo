/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

type blockProducedNoticeHandler struct {
	BaseMsgHandler
	settings p2pcommon.LocalSettings
	myAgent  bool
}

var _ p2pcommon.MessageHandler = (*blockProducedNoticeHandler)(nil)

// newNewBlockNoticeHandler creates handler for NewBlockNotice
func NewBlockProducedNoticeHandler(is p2pcommon.InternalService, pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService, sm p2pcommon.SyncManager) *blockProducedNoticeHandler {
	bh := &blockProducedNoticeHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.BlockProducedNotice, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger}, settings: is.LocalSettings()}
	// FIXME refactor later
	bh.myAgent = types.IsSamePeerID(bh.settings.AgentID, peer.ID())
	return bh
}

func (h *blockProducedNoticeHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.BlockProducedNotice{})
}

func (h *blockProducedNoticeHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := h.peer
	data := msgBody.(*types.BlockProducedNotice)
	if data.Block == nil || len(data.Block.Hash) == 0 {
		h.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Msg("invalid blockProduced notice. block is null")
		return
	}
	// remove to verbose log
	p2putil.DebugLogReceive(h.logger, h.protocol, msg.ID().String(), remotePeer, data)

	// lru cache can accept hashable key
	block := data.Block
	if blockID, err := types.ParseToBlockID(data.GetBlock().GetHash()); err != nil {
		// TODO add penalty score
		h.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Str("hash", enc.ToString(data.GetBlock().GetHash())).Msg("malformed blockHash")
		return
	} else {
		bpID, err := block.BPID()
		if err != nil {
			h.logger.Debug().Err(err).Str("blockID", block.BlockID().String()).Msg("invalid block publick key")
			return
		}
		if !h.checkSender(bpID) {
			h.logger.Debug().Err(err).Str("blockID", block.BlockID().String()).Msg("peer is not access right to send bp notice")
			return
		}
		// block by blockProduced notice must be new fresh block
		remotePeer.UpdateLastNotice(blockID, data.BlockNo)
		h.sm.HandleBlockProducedNotice(h.peer, block)
	}
}

func (h *blockProducedNoticeHandler) checkSender(bpID types.PeerID) bool {
	if h.myAgent {
		return true
	} else {
		return checkBPNoticeSender(bpID, h.peer)
	}
}

// toAgentBPNoticeHandler handle blockProducedNotice to agent node from any other peer
type toAgentBPNoticeHandler struct {
	BaseMsgHandler
	cm p2pcommon.CertificateManager
}

var _ p2pcommon.MessageHandler = (*toAgentBPNoticeHandler)(nil)

// newNewBlockNoticeHandler creates handler for NewBlockNotice
func NewAgentBlockProducedNoticeHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService, sm p2pcommon.SyncManager, cm p2pcommon.CertificateManager) *toAgentBPNoticeHandler {
	bh := &toAgentBPNoticeHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.BlockProducedNotice, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger}, cm: cm}
	return bh
}

func (h *toAgentBPNoticeHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.BlockProducedNotice{})
}

// TODO redundant code with blockProducedNoticeHandler
func (h *toAgentBPNoticeHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := h.peer
	data := msgBody.(*types.BlockProducedNotice)
	if data.Block == nil || len(data.Block.Hash) == 0 {
		h.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Msg("invalid blockProduced notice. block is null")
		return
	}
	// remove to verbose log
	p2putil.DebugLogReceive(h.logger, h.protocol, msg.ID().String(), remotePeer, data)

	// lru cache cannot accept slice as key
	block := data.Block
	if blockID, err := types.ParseToBlockID(data.GetBlock().GetHash()); err != nil {
		// TODO add penalty score
		h.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Str("hash", enc.ToString(data.GetBlock().GetHash())).Msg("malformed blockHash")
		return
	} else {
		bpID, err := block.BPID()
		if err != nil {
			h.logger.Debug().Err(err).Str("blockID", block.BlockID().String()).Msg("invalid block publick key")
			return
		}

		if !checkBPNoticeSender(bpID, remotePeer) {
			h.logger.Debug().Err(err).Str(p2putil.LogPeerName, remotePeer.Name()).Stringer("bpID", types.LogPeerShort(bpID)).Str("blockID", block.BlockID().String()).Msg("peer is not access right to send bp notice")
			return
		}

		switch h.isToToss(bpID, block) {
		case tossOut:
			h.actor.TellRequest(message.P2PSvc, message.TossBPNotice{Block: block, TossIn: bool(p2pcommon.ExternalZone), OriginalMsg: msg})
		case tossIn:
			h.actor.TellRequest(message.P2PSvc, message.TossBPNotice{Block: block, TossIn: bool(p2pcommon.InternalZone), OriginalMsg: msg})
		default:
			// do nothing
		}
		remotePeer.UpdateLastNotice(blockID, data.BlockNo)
		h.sm.HandleBlockProducedNotice(h.peer, block)
	}
}

func checkBPNoticeSender(bpID types.PeerID, peer p2pcommon.RemotePeer) bool {
	// accepted role can be not synced if local block height is low, so allow notice from watcher if peer id and bpid are same.
	if types.IsSamePeerID(peer.ID(), bpID) {
		return true
	}
	// is valid agent = has certificate for bp
	switch peer.AcceptedRole() {
	case types.PeerRole_Agent:
		for _, cert := range peer.RemoteInfo().Certificates {
			if types.IsSamePeerID(bpID, cert.BPID) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

type tossDirection int

const (
	noToss tossDirection = iota
	tossIn
	tossOut
)

func (h *toAgentBPNoticeHandler) isToToss(bpID types.PeerID, block *types.Block) tossDirection {
	// check if the remote peer has right to send (or toss) bp notice.
	if h.cm.CanHandle(bpID) {
		// toss notice to other agents or bp
		return tossOut
	} else if h.peer.RemoteInfo().Zone == p2pcommon.ExternalZone {
		return tossIn
	} else {
		return noToss
	}
}
