/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"sync"

	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"

	inet "github.com/libp2p/go-libp2p-net"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/blockchain"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/multiformats/go-multicodec/protobuf"
)

// pattern: /protocol-name/request-or-response-message/version
const (
	getTXsRequest      protocol.ID = "/tx/getreq/0.1"
	getTxsResponse     protocol.ID = "/tx/getresp/0.1"
	notifyNewTxRequest protocol.ID = "/blk/newtxreq/0.1"
)

// TxProtocol handle tx messages.
// Relaying is not implemented yet.
type TxProtocol struct {
	iserv ActorService

	ps       PeerManager
	reqMutex sync.Mutex

	log *log.Logger
}

// NewTxProtocol creates transaction subprotocol
func NewTxProtocol(logger *log.Logger, chainsvc *blockchain.ChainService) *TxProtocol {
	p := &TxProtocol{reqMutex: sync.Mutex{},
		log: logger,
	}
	return p
}

func (p *TxProtocol) initWith(p2pservice PeerManager) {
	p.ps = p2pservice
	p.ps.SetStreamHandler(getTXsRequest, p.onGetTXsRequest)
	p.ps.SetStreamHandler(getTxsResponse, p.onGetTXsResponse)
	p.ps.SetStreamHandler(notifyNewTxRequest, p.onNotifynewTXs)
}

// remote peer requests handler
func (p *TxProtocol) onGetTXsRequest(s inet.Stream) {
	defer s.Close()

	peerID := s.Conn().RemotePeer()
	remotePeer, ok := p.ps.GetPeer(peerID)
	if !ok {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()
	perr := remotePeer.checkState()
	if perr != nil {
		p.log.Info().Msgf("%s: Invalid peer state to handle request %s : %s", peerID.Pretty(), s.Protocol(), perr.Error())
		return
	}

	// get request data
	data := &types.GetTransactionsRequest{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Info().Err(err).Msg("fail to decode")
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, len(data.Hashes))

	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Info().Msg("Failed to authenticate message")
		return
	}

	// find transactions from chainservice
	idx := 0
	hashesMap := make(map[string][]byte, len(data.Hashes))
	for _, hash := range data.Hashes {
		hashesMap[base64.StdEncoding.EncodeToString(hash)] = hash
	}
	hashes := make([][]byte, 0, len(data.Hashes))
	txInfos := make([]*types.Tx, 0, len(data.Hashes))
	// FIXME: chain에 들어간 트랜잭션을 볼 방법이 없다. 멤풀도 검색이 안 되서 전체를 다 본 다음에 그중에 매칭이 되는 것을 추출하는 방식으로 처리한다.
	txs, ok := extractTXsFromRequest(p.iserv.CallRequest(message.MemPoolSvc,
		&message.MemPoolGet{}))
	for _, tx := range txs {
		hash, found := hashesMap[base64.StdEncoding.EncodeToString(tx.Hash)]
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
func (p *TxProtocol) onGetTXsResponse(s inet.Stream) {
	defer s.Close()

	peerID := s.Conn().RemotePeer()
	remotePeer, ok := p.ps.GetPeer(peerID)
	if !ok {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()
	perr := remotePeer.checkState()
	if perr != nil {
		p.log.Info().Msgf("%s: Invalid peer state to handle request %s : %s", peerID.Pretty(), s.Protocol(), perr.Error())
		return
	}

	data := &types.GetTransactionsResponse{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, len(data.Txs))
	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Info().Msg("Failed to authenticate message")
		return
	}

	// TODO: Is there any better solution than passing everything to mempool service?
	if len(data.Txs) > 0 {
		p.log.Debug().Msgf("Request mempool to add %d txs", len(data.Txs))
		p.iserv.SendRequest(message.MemPoolSvc, &message.MemPoolPut{Txs: data.Txs})
	}
}

var emptyArr = make([]byte, 0)

// GetTXs send request message to peer and
func (p *TxProtocol) GetTXs(peerID peer.ID, txHashes []message.TXHash) bool {
	remotePeer, ok := p.ps.GetPeer(peerID)
	if !ok {
		p.log.Warn().Str("peer_id", peerID.Pretty()).Msg("Invalid peer. check for bug")
		return false
	}
	p.log.Debug().Msgf("%s: Sending GetTransactions request to: %s...(%d txs)", p.ps.ID(), peerID, len(txHashes))
	if len(txHashes) == 0 {
		p.log.Warn().Msg("empty hash list")
		return false
	}

	hashes := make([][]byte, len(txHashes))
	for i, hash := range txHashes {
		if len(hash) == 0 {
			p.log.Warn().Msg("empty hash value requested.")
			return false
		}
		hashes[i] = ([]byte)(hash)
	}
	// create message data
	req := &types.GetTransactionsRequest{MessageData: &types.MessageData{},
		Hashes: hashes}

	remotePeer.sendMessage(newPbMsgRequestOrder(true, true, getTXsRequest, req))
	return true
}

// NotifyNewTX notice tx(s) id created
func (p *TxProtocol) NotifyNewTX(newTXs message.NotifyNewTransactions) bool {
	p.log.Debug().Msgf("%s: Notifying new transactions ", p.ps.ID())

	hashes := make([][]byte, len(newTXs.Txs))
	for i, tx := range newTXs.Txs {
		hashes[i] = tx.Hash
	}
	p.log.Debug().Msgf("Notifying newTXs to %d peers, txHashes: %s",
		len(p.ps.GetPeers()), bytesArrToString(hashes))
	// send to peers
	for _, peer := range p.ps.GetPeers() {
		// create message data
		req := &types.NewTransactionsNotice{MessageData: &types.MessageData{},
			TxHashes: hashes,
		}
		peer.sendMessage(newPbMsgBroadcastOrder(false, notifyNewTxRequest, req))
	}

	return true
}

// remote NotifynewTXs response handler
func (p *TxProtocol) onNotifynewTXs(s inet.Stream) {
	defer s.Close()

	peerID := s.Conn().RemotePeer()
	remotePeer, ok := p.ps.GetPeer(peerID)
	if !ok {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()
	perr := remotePeer.checkState()
	if perr != nil {
		p.log.Info().Msgf("%s: Invalid peer state to handle request %s : %s", peerID.Pretty(), s.Protocol(), perr.Error())
		return
	}

	data := &types.NewTransactionsNotice{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID,
		log.DoLazyEval(func() string { return bytesArrToString(data.TxHashes) }))
	// TODO: check myself and request txs which this node don't have.
	toGet := make([]message.TXHash, len(data.TxHashes))
	// 임시조치로 일단 다 가져온다.
	for i, hashByte := range data.TxHashes {
		toGet[i] = message.TXHash(hashByte)
	}
	// create message data
	p.iserv.SendRequest(message.P2PSvc, &message.GetTransactions{ToWhom: peerID, Hashes: toGet})
	p.log.Debug().Str("peer_id", peerID.Pretty()).Msg("Request GetTransactions")
}

func bytesArrToString(bbarray [][]byte) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for _, hash := range bbarray {
		buf.WriteByte('"')
		buf.WriteString(blockchain.EncodeB64(hash))
		buf.WriteByte('"')
		buf.WriteByte(',')
	}
	buf.WriteByte(']')
	return buf.String()
}
