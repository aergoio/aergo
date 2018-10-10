/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/golang-lru"
)

type txRequestHandler struct {
	BaseMsgHandler
	msgHelper message.Helper
}

var _ MessageHandler = (*txRequestHandler)(nil)

type txResponseHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*txResponseHandler)(nil)

type newTxNoticeHandler struct {
	BaseMsgHandler
	txHashCache  *lru.Cache
}

var _ MessageHandler = (*newTxNoticeHandler)(nil)

// newTxReqHandler creates handler for GetTransactionsRequest
func newTxReqHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *txRequestHandler {
	th := &txRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetTXsRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	th.msgHelper = message.GetHelper()
	return th
}

func (th *txRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetTransactionsRequest{})
}

func (th *txRequestHandler) handle(msgHeader *types.MsgHeader, msgBody proto.Message) {
	peerID := th.peer.ID()
	remotePeer := th.peer
	data := msgBody.(*types.GetTransactionsRequest)
	debugLogReceiveMsg(th.logger, th.protocol, msgHeader.GetId(), peerID, len(data.Hashes))

	// find transactions from chainservice
	status := types.ResultStatus_OK
	idx := 0
	hashes := make([][]byte, 0, len(data.Hashes))
	txInfos := make([]*types.Tx, 0, len(data.Hashes))
	for _, hash := range data.Hashes {
		tx, err := th.msgHelper.ExtractTxFromResponseAndError(th.actor.CallRequest(message.MemPoolSvc,
			&message.MemPoolExist{Hash:hash}))
		if err != nil {
			// response error to peer
			status = types.ResultStatus_INTERNAL
			break
		}
		if tx == nil {
			// ignore not existing hash
			continue
		}
		hashes = append(hashes, hash)
		txInfos = append(txInfos, tx)
		idx++
	}

	// generate response message
	resp := &types.GetTransactionsResponse{
		Status: status,
		Hashes: hashes,
		Txs:    txInfos}

	remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msgHeader.GetId(), GetTxsResponse, resp))
}

// newTxRespHandler creates handler for GetTransactionsResponse
func newTxRespHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *txResponseHandler {
	th := &txResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetTxsResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return th
}

func (th *txResponseHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetTransactionsResponse{})
}

func (th *txResponseHandler) handle(msgHeader *types.MsgHeader, msgBody proto.Message) {
	peerID := th.peer.ID()
	data := msgBody.(*types.GetTransactionsResponse)
	debugLogReceiveMsg(th.logger, th.protocol, msgHeader.GetId(), peerID, len(data.Txs))

	// TODO: Is there any better solution than passing everything to mempool service?
	if len(data.Txs) > 0 {
		th.logger.Debug().Int("tx_cnt", len(data.Txs)).Msg("Request mempool to add txs")
		th.actor.SendRequest(message.MemPoolSvc, &message.MemPoolPut{Txs: data.Txs})
	}
}

// newNewTxNoticeHandler creates handler for GetTransactionsResponse
func newNewTxNoticeHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService, sm SyncManager) *newTxNoticeHandler {
	th := &newTxNoticeHandler{BaseMsgHandler: BaseMsgHandler{protocol: NewTxNotice, pm: pm, sm:sm, peer: peer, actor: actor, logger: logger}}
	var err error
	th.txHashCache, err = lru.New(DefaultPeerInvCacheSize)
	if err != nil {
		panic("Failed to create newTxNoticeHandler " + err.Error())
	}
	return th
}

func (th *newTxNoticeHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.NewTransactionsNotice{})
}

func (th *newTxNoticeHandler) handle(msgHeader *types.MsgHeader, msgBody proto.Message) {
	peerID := th.peer.ID()
	data := msgBody.(*types.NewTransactionsNotice)
	// remove to verbose log
	debugLogReceiveMsg(th.logger, th.protocol, msgHeader.GetId(), peerID, len(data.TxHashes))

	if len(data.TxHashes) == 0 {
		return
	}
	// lru cache can accept hashable key
	hashes := make([]TxHash, len(data.TxHashes))
	for i, hash := range data.TxHashes {
		copy(hashes[i][:], hash)
	}
	added := th.peer.updateTxCache(hashes)
	if len(added) > 0 {
		th.sm.HandleNewTxNotice(th.peer, added, data)
	}
}