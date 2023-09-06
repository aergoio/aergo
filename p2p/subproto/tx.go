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

type txRequestHandler struct {
	BaseMsgHandler
	asyncHelper
	msgHelper message.Helper
}

var _ p2pcommon.MessageHandler = (*txRequestHandler)(nil)

type txResponseHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*txResponseHandler)(nil)

type newTxNoticeHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*newTxNoticeHandler)(nil)

// newTxReqHandler creates handler for GetTransactionsRequest
func NewTxReqHandler(pm p2pcommon.PeerManager, sm p2pcommon.SyncManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *txRequestHandler {
	th := &txRequestHandler{
		BaseMsgHandler{protocol: p2pcommon.GetTXsRequest, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger},
		newAsyncHelper(), message.GetHelper()}
	return th
}

func (th *txRequestHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetTransactionsRequest{})
}

func (th *txRequestHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := th.peer
	body := msgBody.(*types.GetTransactionsRequest)
	p2putil.DebugLogReceive(th.logger, th.protocol, msg.ID().String(), remotePeer, body)

	if err := th.sm.HandleGetTxReq(remotePeer, msg.ID(), body); err != nil {
		th.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Str(p2putil.LogMsgID, msg.ID().String()).Err(err).Msg("return err for concurrent get tx request")
		resp := &types.GetTransactionsResponse{
			Status: types.ResultStatus_RESOURCE_EXHAUSTED,
			Hashes: nil,
			Txs:    nil, HasNext: false}
		remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.GetTXsResponse, resp))
	}
}

// newTxRespHandler creates handler for GetTransactionsResponse
func NewTxRespHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *txResponseHandler {
	th := &txResponseHandler{BaseMsgHandler{protocol: p2pcommon.GetTXsResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return th
}

func (th *txResponseHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GetTransactionsResponse{})
}

func (th *txResponseHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := th.peer
	data := msgBody.(*types.GetTransactionsResponse)
	p2putil.DebugLogReceiveResponse(th.logger, th.protocol, msg.ID().String(), msg.OriginalID().String(), th.peer, data)

	if !remotePeer.GetReceiver(msg.OriginalID())(msg, data) {
		th.logger.Warn().Str(p2putil.LogMsgID, msg.ID().String()).Msg("unknown getTX response")
		remotePeer.ConsumeRequest(msg.OriginalID())
	}
}

// newNewTxNoticeHandler creates handler for GetTransactionsResponse
func NewNewTxNoticeHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService, sm p2pcommon.SyncManager) *newTxNoticeHandler {
	th := &newTxNoticeHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.NewTxNotice, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger}}
	return th
}

func (th *newTxNoticeHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.NewTransactionsNotice{})
}

func (th *newTxNoticeHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := th.peer
	data := msgBody.(*types.NewTransactionsNotice)
	// remove to verbose log
	if th.logger.IsDebugEnabled() {
		p2putil.DebugLogReceive(th.logger, th.protocol, msg.ID().String(), remotePeer, data)
	}

	if len(data.TxHashes) == 0 {
		return
	}
	// lru cache can accept hashable key
	hashes := make([]types.TxID, len(data.TxHashes))
	for i, hash := range data.TxHashes {
		if tid, err := types.ParseToTxID(hash); err != nil {
			th.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Str("hash", enc.ToString(hash)).Msg("malformed txhash found")
			// TODO Add penalty score and break
			break
		} else {
			hashes[i] = tid
		}
	}
	added := th.peer.UpdateTxCache(hashes)
	if len(added) > 0 {
		th.sm.HandleNewTxNotice(th.peer, added, data)
	}
}
