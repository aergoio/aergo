/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

type txRequestHandler struct {
	BaseMsgHandler
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
	th := &txRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: getTXsRequest, pm: pm, peer: peer, actor: peer.actorServ, logger: logger, signer: signer}}
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
	idx := 0
	var keyArray [txhashLen]byte
	hashesMap := make(map[[txhashLen]byte][]byte, len(data.Hashes))
	for _, hash := range data.Hashes {
		if len(hash) != txhashLen {
			// TODO ignore just single hash or return invalid request
			continue
		}
		copy(keyArray[:], hash)
		hashesMap[keyArray] = hash
	}
	hashes := make([][]byte, 0, len(data.Hashes))
	txInfos := make([]*types.Tx, 0, len(data.Hashes))
	// FIXME: chain에 들어간 트랜잭션을 볼 방법이 없다. 멤풀도 검색이 안 되서 전체를 다 본 다음에 그중에 매칭이 되는 것을 추출하는 방식으로 처리한다.
	txs, _ := extractTXsFromRequest(th.actor.CallRequest(message.MemPoolSvc,
		&message.MemPoolGet{}))
	for _, tx := range txs {
		copy(keyArray[:], tx.Hash)
		hash, found := hashesMap[keyArray]
		if !found {
			continue
		}
		hashes = append(hashes, hash)
		txInfos = append(txInfos, tx)
		idx++
	}
	status := types.ResultStatus_OK

	// generate response message
	resp := &types.GetTransactionsResponse{
		Status: status,
		Hashes: hashes,
		Txs:    txInfos}

	remotePeer.sendMessage(newPbMsgResponseOrder(msgHeader.GetId(), getTxsResponse, resp, th.signer))
}

// newTxRespHandler creates handler for GetTransactionsResponse
func newTxRespHandler(pm PeerManager, peer *RemotePeer, logger *log.Logger, signer msgSigner) *txResponseHandler {
	th := &txResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: getTxsResponse, pm: pm, peer: peer, actor: peer.actorServ, logger: logger, signer: signer}}
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
	th := &newTxNoticeHandler{BaseMsgHandler: BaseMsgHandler{protocol: newTxNotice, pm: pm, peer: peer, actor: peer.actorServ, logger: logger, signer: signer}}
	return th
}

func (th *newTxNoticeHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.NewTransactionsNotice{})
}

func (th *newTxNoticeHandler) handle(msgHeader *types.MsgHeader, msgBody proto.Message) {
	peerID := th.peer.ID()
	data := msgBody.(*types.NewTransactionsNotice)
	debugLogReceiveMsg(th.logger, th.protocol, msgHeader.GetId(), peerID, log.DoLazyEval(func() string { return bytesArrToString(data.TxHashes) }))

	// TODO: check myself and request txs which this node don't have.
	toGet := make([]message.TXHash, len(data.TxHashes))
	// 임시조치로 일단 다 가져온다.
	for i, hashByte := range data.TxHashes {
		toGet[i] = message.TXHash(hashByte)
	}
	// create message data
	th.actor.SendRequest(message.P2PSvc, &message.GetTransactions{ToWhom: peerID, Hashes: toGet})
	th.logger.Debug().Str(LogPeerID, peerID.Pretty()).Msg("Request GetTransactions")
}

func bytesArrToString(bbarray [][]byte) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for _, hash := range bbarray {
		buf.WriteByte('"')
		buf.WriteString(enc.ToString(hash))
		buf.WriteByte('"')
		buf.WriteByte(',')
	}
	buf.WriteByte(']')
	return buf.String()
}
