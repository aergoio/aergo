/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/blockchain"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
)

// TxHandler handle tx messages.
// Relaying is not implemented yet.
type TxHandler struct {
	BaseMsgHandler
}

// NewTxProtocol creates transaction subprotocol
func NewTxProtocol(logger *log.Logger, chainsvc *blockchain.ChainService) *TxHandler {
	p := &TxHandler{}
	return p
}

func NewTxHandler(pm PeerManager, peer *RemotePeer, logger *log.Logger) *TxHandler {
	h := &TxHandler{BaseMsgHandler: BaseMsgHandler{protocol: pingRequest, pm: pm, peer: peer, actor: peer.actorServ, logger: logger}}
	return h
}
func (p *TxHandler) setPeerManager(pm PeerManager) {
	p.pm = pm
}

func (p *TxHandler) startHandling() {
	// p.pm.SetStreamHandler(getTXsRequest, p.onGetTXsRequest)
	// p.pm.SetStreamHandler(getTxsResponse, p.onGetTXsResponse)
	// p.pm.SetStreamHandler(notifyNewTxRequest, p.onNotifynewTXs)
}

// remote peer requests handler
func (p *TxHandler) handleGetTXsRequest(msg *types.P2PMessage) {
	peerID := p.peer.ID()
	remotePeer := p.peer

	// get request data
	data := &types.GetTransactionsRequest{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		p.logger.Info().Err(err).Msg("fail to decode")
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, len(data.Hashes))

	valid := p.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.logger.Info().Msg("Failed to authenticate message")
		return
	}

	// find transactions from chainservice
	idx := 0
	hashesMap := make(map[string][]byte, len(data.Hashes))
	for _, hash := range data.Hashes {
		hashesMap[enc.ToString(hash)] = hash
	}
	hashes := make([][]byte, 0, len(data.Hashes))
	txInfos := make([]*types.Tx, 0, len(data.Hashes))
	// FIXME: chain에 들어간 트랜잭션을 볼 방법이 없다. 멤풀도 검색이 안 되서 전체를 다 본 다음에 그중에 매칭이 되는 것을 추출하는 방식으로 처리한다.
	txs, _ := extractTXsFromRequest(p.actor.CallRequest(message.MemPoolSvc,
		&message.MemPoolGet{}))
	for _, tx := range txs {
		hash, found := hashesMap[enc.ToString(tx.Hash)]
		if !found {
			continue
		}
		hashes = append(hashes, hash)
		txInfos = append(txInfos, tx)
		idx++
	}
	status := types.ResultStatus_OK

	// generate response message
	resp := &types.GetTransactionsResponse{MessageData: &types.MessageData{},
		Status: status,
		Hashes: hashes,
		Txs:    txInfos}

	remotePeer.sendMessage(newPbMsgResponseOrder(data.MessageData.Id, true, getTxsResponse, resp))
}

// remote GetTransactions response handler
func (p *TxHandler) handleGetTXsResponse(msg *types.P2PMessage) {
	peerID := p.peer.ID()

	data := &types.GetTransactionsResponse{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID, len(data.Txs))
	valid := p.pm.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.logger.Info().Msg("Failed to authenticate message")
		return
	}

	// TODO: Is there any better solution than passing everything to mempool service?
	if len(data.Txs) > 0 {
		p.logger.Debug().Int("tx_cnt", len(data.Txs)).Msg("Request mempool to add txs")
		p.actor.SendRequest(message.MemPoolSvc, &message.MemPoolPut{Txs: data.Txs})
	}
}

var emptyArr = make([]byte, 0)

// remote NotifynewTXs response handler
func (p *TxHandler) handleNewTXsNotice(msg *types.P2PMessage) {
	peerID := p.peer.ID()

	data := &types.NewTransactionsNotice{}
	err := unmarshalMessage(msg.Data, data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), data.MessageData.Id, peerID,
		log.DoLazyEval(func() string { return bytesArrToString(data.TxHashes) }))
	// TODO: check myself and request txs which this node don't have.
	toGet := make([]message.TXHash, len(data.TxHashes))
	// 임시조치로 일단 다 가져온다.
	for i, hashByte := range data.TxHashes {
		toGet[i] = message.TXHash(hashByte)
	}
	// create message data
	p.actor.SendRequest(message.P2PSvc, &message.GetTransactions{ToWhom: peerID, Hashes: toGet})
	p.logger.Debug().Str(LogPeerID, peerID.Pretty()).Msg("Request GetTransactions")
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
