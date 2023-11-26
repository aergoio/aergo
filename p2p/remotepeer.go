/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/metric"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

var TimeoutError = errors.New("timeout")
var CancelError = errors.New("canceled")

type requestInfo struct {
	cTime    time.Time
	reqMO    p2pcommon.MsgOrder
	receiver p2pcommon.ResponseReceiver
}

type queryMsg struct {
	handler p2pcommon.MessageHandler
	msg     p2pcommon.Message
	msgBody p2pcommon.MessageBody
}

// remotePeerImpl represent remote peer to which is connected
type remotePeerImpl struct {
	logger       *log.Logger
	pingDuration time.Duration

	manageNum  uint32
	remoteInfo p2pcommon.RemoteInfo
	name       string
	state      types.PeerState
	actor      p2pcommon.ActorService
	pm         p2pcommon.PeerManager
	mf         p2pcommon.MoFactory
	signer     p2pcommon.MsgSigner
	metric     *metric.PeerMetric

	certChan chan *p2pcommon.AgentCertificateV1
	stopChan chan struct{}

	// direct write channel
	writeBuf    chan p2pcommon.MsgOrder
	writeDirect chan p2pcommon.MsgOrder
	closeWrite  chan struct{}

	// used to access request data from response handlers
	requests map[p2pcommon.MsgID]*requestInfo
	reqMutex *sync.Mutex

	handlers map[p2pcommon.SubProtocol]p2pcommon.MessageHandler

	blkHashCache *lru.Cache
	txHashCache  *lru.Cache
	lastStatus   *types.LastBlockStatus
	// lastBlkNoticeTime is time that local peer sent NewBlockNotice to this remote peer
	lastBlkNoticeTime time.Time
	skipCnt           int32

	txQueueLock         *sync.Mutex
	txNoticeQueue       *p2putil.PressableQueue
	maxTxNoticeHashSize int

	rw p2pcommon.MsgReadWriter

	taskChannel chan p2pcommon.PeerTask

	// lastTxQuery indicate last message for querying tx
	blkQuerySlot int64
}

var _ p2pcommon.RemotePeer = (*remotePeerImpl)(nil)

// newRemotePeer create an object which represent a remote peer.
func newRemotePeer(remote p2pcommon.RemoteInfo, manageNum uint32, pm p2pcommon.PeerManager, actor p2pcommon.ActorService, log *log.Logger, mf p2pcommon.MoFactory, signer p2pcommon.MsgSigner, rw p2pcommon.MsgReadWriter) *remotePeerImpl {
	rPeer := &remotePeerImpl{
		remoteInfo: remote, manageNum: manageNum, pm: pm,
		name:  fmt.Sprintf("%s#%d", p2putil.ShortForm(remote.Meta.ID), manageNum),
		actor: actor, logger: log, mf: mf, signer: signer, rw: rw,
		pingDuration: defaultPingInterval,
		state:        types.STARTING,

		lastStatus: &types.LastBlockStatus{},
		stopChan:   make(chan struct{}, 1),
		closeWrite: make(chan struct{}),
		certChan:   make(chan *p2pcommon.AgentCertificateV1),
		requests:   make(map[p2pcommon.MsgID]*requestInfo),
		reqMutex:   &sync.Mutex{},

		handlers: make(map[p2pcommon.SubProtocol]p2pcommon.MessageHandler),

		txQueueLock:         &sync.Mutex{},
		txNoticeQueue:       p2putil.NewPressableQueue(DefaultPeerTxQueueSize),
		maxTxNoticeHashSize: DefaultPeerTxQueueSize,
		taskChannel:         make(chan p2pcommon.PeerTask, 1),
	}
	rPeer.writeBuf = make(chan p2pcommon.MsgOrder, writeMsgBufferSize)
	rPeer.writeDirect = make(chan p2pcommon.MsgOrder)

	var err error
	rPeer.blkHashCache, err = lru.New(DefaultPeerBlockCacheSize)
	if err != nil {
		panic("Failed to create remote peer " + err.Error())
	}
	rPeer.txHashCache, err = lru.New(DefaultPeerTxCacheSize)
	if err != nil {
		panic("Failed to create remote peer " + err.Error())
	}

	return rPeer
}

// ID return id of peer, same as peer.remoteInfo.ID
func (p *remotePeerImpl) ID() types.PeerID {
	return p.remoteInfo.Meta.ID
}

func (p *remotePeerImpl) RemoteInfo() p2pcommon.RemoteInfo {
	return p.remoteInfo
}

func (p *remotePeerImpl) Meta() p2pcommon.PeerMeta {
	return p.remoteInfo.Meta
}

func (p *remotePeerImpl) ManageNumber() uint32 {
	return p.manageNum
}

func (p *remotePeerImpl) Name() string {
	return p.name
}

func (p *remotePeerImpl) Version() string {
	return p.remoteInfo.Meta.Version
}

func (p *remotePeerImpl) AcceptedRole() types.PeerRole {
	return p.remoteInfo.AcceptedRole
}
func (p *remotePeerImpl) ChangeRole(role types.PeerRole) {
	p.remoteInfo.AcceptedRole = role
}

func (p *remotePeerImpl) AddMessageHandler(subProtocol p2pcommon.SubProtocol, handler p2pcommon.MessageHandler) {
	p.handlers[subProtocol] = handler
}

func (p *remotePeerImpl) MF() p2pcommon.MoFactory {
	return p.mf
}

// State returns current state of peer
func (p *remotePeerImpl) State() types.PeerState {
	return p.state.Get()
}

func (p *remotePeerImpl) LastStatus() *types.LastBlockStatus {
	return p.lastStatus
}

// runPeer should be called by go routine
func (p *remotePeerImpl) RunPeer() {
	p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Msg("Starting peer")
	pingTicker := time.NewTicker(p.pingDuration)

	go p.runWrite()
	go p.runRead()

	txNoticeTicker := time.NewTicker(txNoticeInterval)
	certCleanupTicker := time.NewTicker(p2pcommon.RemoteCertCheckInterval)

	// peer state is changed to RUNNING after all sub goroutine is ready, and to STOPPED before fll sub goroutine is stopped.
	p.state.SetAndGet(types.RUNNING)
READNOPLOOP:
	for {
		select {
		case <-pingTicker.C:
			p.sendPing()
			// no operation for now
		case <-txNoticeTicker.C:
			p.trySendTxNotices()
		case <-certCleanupTicker.C:
			p.cleanupCerts()
		case c := <-p.certChan:
			p.addCert(c)
		case task := <-p.taskChannel:
			p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Msg("Executing task for peer")
			task(p)
		case <-p.stopChan:
			break READNOPLOOP
		}
	}

	p.logger.Info().Str(p2putil.LogPeerName, p.Name()).Msg("Finishing peer")
	txNoticeTicker.Stop()
	pingTicker.Stop()
	// finish goroutine write. read goroutine will be closed automatically when disconnect
	close(p.closeWrite)
	close(p.stopChan)
	p.state.SetAndGet(types.STOPPED)

	p.pm.RemovePeer(p)
}

func (p *remotePeerImpl) runWrite() {
	cleanupTicker := time.NewTicker(cleanRequestInterval)
	defer func() {
		if r := recover(); r != nil {
			p.logger.Panic().Str("callStack", string(debug.Stack())).Str(p2putil.LogPeerName, p.Name()).Str("recover", fmt.Sprint(r)).Msg("There were panic in runWrite ")
		}
	}()

WRITELOOP:
	for {
		select {
		case m := <-p.writeBuf:
			p.writeToPeer(m)
		case m := <-p.writeDirect:
			p.writeToPeer(m)
		case <-cleanupTicker.C:
			p.pruneRequests()
		case <-p.closeWrite:
			p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Msg("Quitting runWrite")
			break WRITELOOP
		}
	}
	cleanupTicker.Stop()
	p.cleanupWrite()
	p.rw.Close()

	// closing channel is up to golang runtime
	// close(p.write)
	// close(p.consumeChan)
}

func (p *remotePeerImpl) cleanupWrite() {
	// 1. cleaning receive handlers. TODO add code

	// 2. canceling not sent orders
	for {
		select {
		case m := <-p.writeBuf:
			m.CancelSend(p)
		default:
			return
		}
	}
}

func (p *remotePeerImpl) runRead() {
	for {
		msg, err := p.rw.ReadMsg()
		if err != nil {
			// TODO set different log level by case (i.e. it can be expected if peer is disconnecting )
			p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Err(err).Msg("Failed to read message")
			p.Stop()
			return
		}
		if err = p.handleMsg(msg); err != nil {
			// TODO set different log level by case (i.e. it can be expected if peer is disconnecting )
			p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Err(err).Msg("Failed to handle message")
			p.Stop()
			return
		}
	}
}

func (p *remotePeerImpl) handleMsg(msg p2pcommon.Message) (err error) {
	subProto := msg.Subprotocol()
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error().Stringer(p2putil.LogProtoID, subProto).Str("callStack", string(debug.Stack())).Interface("panic", r).Msg("There were panic in handler.")
			err = fmt.Errorf("internal error")
		}
	}()

	if p.State() > types.RUNNING {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogMsgID, msg.ID()).Stringer(p2putil.LogProtoID, subProto).Str("current_state", p.State().String()).Msg("peer is not running. silently drop input message")
		return nil
	}

	handler, found := p.handlers[subProto]
	if !found {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogMsgID, msg.ID()).Stringer(p2putil.LogProtoID, subProto).Msg("invalid protocol")
		return fmt.Errorf("invalid protocol %s", subProto)
	}

	handler.PreHandle()

	payload, err := handler.ParsePayload(msg.Payload())
	if err != nil {
		p.logger.Warn().Err(err).Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogMsgID, msg.ID()).Stringer(p2putil.LogProtoID, subProto).Msg("invalid message data")
		return fmt.Errorf("invalid message data")
	}
	//err = p.signer.verifyMsg(msg, p.remoteInfo.ID)
	//if err != nil {
	//	p.logger.Warn().Err(err).Str(LogPeerName, p.Name()).Str(LogMsgID, msg.ID().String()).Str(LogProtoID, subProto.String()).Msg("Failed to check signature")
	//	return fmt.Errorf("Failed to check signature")
	//}
	err = handler.CheckAuth(msg, payload)
	if err != nil {
		p.logger.Warn().Err(err).Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogMsgID, msg.ID()).Stringer(p2putil.LogProtoID, subProto).Msg("Failed to authenticate message")
		return fmt.Errorf("Failed to authenticate message.")
	}

	handler.Handle(msg, payload)

	handler.PostHandle(msg, payload)
	return nil
}

// Stop stops aPeer works
func (p *remotePeerImpl) Stop() {
	prevState := p.state.SetAndGet(types.STOPPING)
	if prevState <= types.RUNNING {
		p.stopChan <- struct{}{}
	}

}

func (p *remotePeerImpl) SendMessage(msg p2pcommon.MsgOrder) {
	if p.State() > types.RUNNING {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, msg.GetProtocolID()).
			Str(p2putil.LogMsgID, msg.GetMsgID().String()).Str("current_state", p.State().String()).Msg("Cancel sending message, since peer is not running state")
		return
	}
	select {
	case p.writeBuf <- msg:
		// it's OK
	default:
		p.logger.Info().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, msg.GetProtocolID()).
			Str(p2putil.LogMsgID, msg.GetMsgID().String()).Msg("Remote peer is busy or down")
		// TODO find more elegant way to handled flooding queue. in lots of cases, pending for dropped tx notice or newBlock notice (not blockProduced notice) are not critical in lots of cases.
		p.Stop()
	}
}

func (p *remotePeerImpl) TrySendMessage(msg p2pcommon.MsgOrder) bool {
	if p.State() > types.RUNNING {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, msg.GetProtocolID()).
			Str(p2putil.LogMsgID, msg.GetMsgID().String()).Str("current_state", p.State().String()).Msg("Cancel sending message, since peer is not running state")
		return false
	}
	select {
	case p.writeBuf <- msg:
		// succeed to send message
		return true
	default:
		return false
	}
}

func (p *remotePeerImpl) SendAndWaitMessage(msg p2pcommon.MsgOrder, timeout time.Duration) error {
	if p.State() > types.RUNNING {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, msg.GetProtocolID()).
			Str(p2putil.LogMsgID, msg.GetMsgID().String()).Str("current_state", p.State().String()).Msg("Cancel sending message, since peer is not running state")
		return fmt.Errorf("not running")
	}
	select {
	case p.writeBuf <- msg:
		return nil
	case <-time.NewTimer(timeout).C:
		p.logger.Info().Str(p2putil.LogPeerName, p.Name()).Stringer(p2putil.LogProtoID, msg.GetProtocolID()).
			Str(p2putil.LogMsgID, msg.GetMsgID().String()).Msg("Remote peer is busy or down")
		// TODO find more elegant way to handled flooding queue. in lots of cases, pending for dropped tx notice or newBlock notice (not blockProduced notice) are not critical in lots of cases.
		p.Stop()
		return TimeoutError
	}
}

func (p *remotePeerImpl) PushTxsNotice(txHashes []types.TxID) {
	p.txQueueLock.Lock()
	defer p.txQueueLock.Unlock()
	for _, hash := range txHashes {
		if !p.txNoticeQueue.Offer(hash) {
			p.sendTxNotices()
			// this Offer is always succeeded by invariant
			p.txNoticeQueue.Offer(hash)
		}
	}
}

// ConsumeRequest remove request from request history.
func (p *remotePeerImpl) ConsumeRequest(msgID p2pcommon.MsgID) p2pcommon.MsgOrder {
	p.reqMutex.Lock()
	defer p.reqMutex.Unlock()
	if r, ok := p.requests[msgID]; ok {
		delete(p.requests, msgID)
		return r.reqMO
	}
	return nil
}

// requestIDNotFoundReceiver is to handle response msg which the original message is not identified
func (p *remotePeerImpl) requestIDNotFoundReceiver(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) bool {
	return true
}

// passThroughReceiver is bypass message to legacy handler.
func (p *remotePeerImpl) passThroughReceiver(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) bool {
	return false
}

func (p *remotePeerImpl) GetReceiver(originalID p2pcommon.MsgID) p2pcommon.ResponseReceiver {
	p.reqMutex.Lock()
	defer p.reqMutex.Unlock()
	req, found := p.requests[originalID]
	if !found {
		return p.requestIDNotFoundReceiver
	} else if req.receiver == nil {
		return p.passThroughReceiver
	} else {
		return req.receiver
	}
}

func (p *remotePeerImpl) writeToPeer(m p2pcommon.MsgOrder) {
	if err := m.SendTo(p); err != nil {
		// write fail
		p.Stop()
	}
}

func (p *remotePeerImpl) trySendTxNotices() {
	p.txQueueLock.Lock()
	defer p.txQueueLock.Unlock()
	p.sendTxNotices()
}

// sendTxNotices must be called in txQueueLock
func (p *remotePeerImpl) sendTxNotices() {
	// make create multiple message if queue is too many hashes.
	for {
		// no need to send if queue is empty
		if p.txNoticeQueue.Size() == 0 {
			return
		}
		hashes := make([][]byte, 0, p.txNoticeQueue.Size())
		skippedTxIDs := make([]types.TxID, 0)
		for element := p.txNoticeQueue.Poll(); element != nil; element = p.txNoticeQueue.Poll() {
			hash := element.(types.TxID)
			if p.txHashCache.Contains(hash) {
				skippedTxIDs = append(skippedTxIDs, hash)
				continue
			}
			hashes = append(hashes, hash[:])
			p.txHashCache.Add(hash, cachePlaceHolder)
		}
		if len(hashes) > 0 {
			mo := p.mf.NewMsgTxBroadcastOrder(&types.NewTransactionsNotice{TxHashes: hashes})
			p.SendMessage(mo)
		}
	}
}

// this method MUST be called in same go routine as AergoPeer.RunPeer()
func (p *remotePeerImpl) sendPing() {
	// find my best block
	bestBlock, err := p.actor.GetChainAccessor().GetBestBlock()
	if err != nil {
		p.logger.Warn().Err(err).Msg("cancel ping. failed to get best block")
		return
	}
	// create message data
	pingMsg := &types.Ping{
		BestBlockHash: bestBlock.BlockHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	p.SendMessage(p.mf.NewMsgRequestOrder(true, p2pcommon.PingRequest, pingMsg))
}

// send notice message and then disconnect. this routine should only run in RunPeer go routine
func (p *remotePeerImpl) goAwayMsg(msg string) {
	p.logger.Info().Str(p2putil.LogPeerName, p.Name()).Str("msg", msg).Msg("Peer is closing")
	p.SendAndWaitMessage(p.mf.NewMsgRequestOrder(false, p2pcommon.GoAway, &types.GoAwayNotice{Message: msg}), time.Second)
	p.Stop()
}

func (p *remotePeerImpl) pruneRequests() {
	debugLog := p.logger.IsDebugEnabled()
	deletedCnt := 0
	var deletedReqs []string
	expireTime := time.Now().Add(-1 * time.Hour)
	p.reqMutex.Lock()
	defer p.reqMutex.Unlock()
	for key, m := range p.requests {
		if m.cTime.Before(expireTime) {
			delete(p.requests, key)
			if debugLog {
				deletedReqs = append(deletedReqs, m.reqMO.GetProtocolID().String()+"/"+key.String()+m.cTime.String())
			}
			deletedCnt++
		}
	}
	p.logger.Info().Int("count", deletedCnt).Str(p2putil.LogPeerName, p.Name()).
		Time("until", expireTime).Msg("Pruned requests which response was not came")
	//.Msg("Pruned %d requests but no response to peer %s until %v", deletedCnt, p.remoteInfo.ID.String(), time.Unix(expireTime, 0))
	if debugLog {
		p.logger.Debug().Strs("reqs", deletedReqs).Msg("Pruned")
	}
}

func (p *remotePeerImpl) UpdateBlkCache(blkHash types.BlockID, blkNumber types.BlockNo) bool {
	p.UpdateLastNotice(blkHash, blkNumber)
	// lru cache can't accept byte slice key
	found, _ := p.blkHashCache.ContainsOrAdd(blkHash, true)
	return found
}

func (p *remotePeerImpl) UpdateTxCache(hashes []types.TxID) []types.TxID {
	// lru cache can't accept byte slice key
	added := make([]types.TxID, 0, len(hashes))
	for _, hash := range hashes {
		if found, _ := p.txHashCache.ContainsOrAdd(hash, true); !found {
			added = append(added, hash)
		}
	}
	return added
}

func (p *remotePeerImpl) UpdateLastNotice(blkHash types.BlockID, blkNumber types.BlockNo) {
	p.lastStatus = &types.LastBlockStatus{CheckTime: time.Now(), BlockHash: blkHash[:], BlockNumber: blkNumber}
}

func (p *remotePeerImpl) sendGoAway(msg string) {
	// TODO: send goaway message and close connection
}

func (p *remotePeerImpl) addCert(cert *p2pcommon.AgentCertificateV1) {
	newCerts := make([]*p2pcommon.AgentCertificateV1, 0, len(p.remoteInfo.Certificates)+1)
	for _, oldCert := range p.remoteInfo.Certificates {
		if !types.IsSamePeerID(oldCert.BPID, cert.BPID) {
			// replace old certificate if it already exists.
			newCerts = append(newCerts, oldCert)
		}
	}
	newCerts = append(newCerts, cert)
	p.logger.Info().Str(p2putil.LogPeerName, p.Name()).Stringer("bpID", types.LogPeerShort(cert.BPID)).Time("cTime", cert.CreateTime).Time("eTime", cert.ExpireTime).Msg("agent certificate is added to certificate list of remote peer")
	p.remoteInfo.Certificates = newCerts
	if len(newCerts) > 0 && p.AcceptedRole() == types.PeerRole_Watcher {
		p.logger.Info().Str(p2putil.LogPeerName, p.Name()).Msg("peer has certificates now. peer is promoted to Agent")
		p.pm.UpdatePeerRole([]p2pcommon.RoleModifier{{ID: p.ID(), Role: types.PeerRole_Agent}})
	}
}

func (p *remotePeerImpl) cleanupCerts() {
	now := time.Now()
	certs2 := p.remoteInfo.Certificates[:0]
	for _, cert := range p.remoteInfo.Certificates {
		if cert.IsValidInTime(now, p2pcommon.TimeErrorTolerance) {
			certs2 = append(certs2, cert)
		} else {
			p.logger.Info().Str(p2putil.LogPeerName, p.Name()).Stringer("issuer", types.LogPeerShort(cert.BPID)).Msg("Certificate is expired")
		}
	}
	p.remoteInfo.Certificates = certs2
	if len(certs2) == 0 && p.AcceptedRole() == types.PeerRole_Agent {
		p.logger.Info().Str(p2putil.LogPeerName, p.Name()).Msg("All Certificates are expired. peer is demoted to Watcher")
		p.pm.UpdatePeerRole([]p2pcommon.RoleModifier{{ID: p.ID(), Role: types.PeerRole_Watcher}})
	}
}

func (p *remotePeerImpl) AddCertificate(cert *p2pcommon.AgentCertificateV1) {
	p.certChan <- cert
}

func (p *remotePeerImpl) DoTask(task p2pcommon.PeerTask) bool {
	select {
	case p.taskChannel <- task:
		return true
	default:
		// peer is busy
		return false
	}
}
