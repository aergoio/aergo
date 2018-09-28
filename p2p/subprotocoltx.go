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
}

var _ MessageHandler = (*newTxNoticeHandler)(nil)

// newTxReqHandler creates handler for GetTransactionsRequest
func newTxReqHandler(pm PeerManager, peer *RemotePeer, logger *log.Logger, signer msgSigner) *txRequestHandler {
	th := &txRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetTXsRequest, pm: pm, peer: peer, actor: peer.actorServ, logger: logger, signer: signer}}
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
			&message.MemPoolExist{}))
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

	remotePeer.sendMessage(newPbMsgResponseOrder(msgHeader.GetId(), GetTxsResponse, resp, th.signer))
}

// newTxRespHandler creates handler for GetTransactionsResponse
func newTxRespHandler(pm PeerManager, peer *RemotePeer, logger *log.Logger, signer msgSigner) *txResponseHandler {
	th := &txResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetTxsResponse, pm: pm, peer: peer, actor: peer.actorServ, logger: logger, signer: signer}}
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
func newNewTxNoticeHandler(pm PeerManager, peer *RemotePeer, logger *log.Logger, signer msgSigner) *newTxNoticeHandler {
	th := &newTxNoticeHandler{BaseMsgHandler: BaseMsgHandler{protocol: NewTxNotice, pm: pm, peer: peer, actor: peer.actorServ, logger: logger, signer: signer}}
	return th
}

func (th *newTxNoticeHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.NewTransactionsNotice{})
}

func (th *newTxNoticeHandler) handle(msgHeader *types.MsgHeader, msgBody proto.Message) {
	// peerID := th.peer.ID()
	data := msgBody.(*types.NewTransactionsNotice)
	// remove to verbose log
	// debugLogReceiveMsg(th.logger, th.protocol, msgHeader.GetId(), peerID, log.DoLazyEval(func() string { return bytesArrToString(data.TxHashes) }))

	th.peer.handleNewTxNotice(data)
}

