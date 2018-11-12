package syncer

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"time"
)

type BlockFetcher struct {
	hub component.ICompRequester //for communicate with other service

	ctx *types.SyncContext

	quitCh chan interface{}

	hfCh chan *HashSet

	curHashSet *HashSet

	runningQueue TaskQueue
	pendingQueue TaskQueue

	responseCh chan interface{} //BlockResponse, AddBlockResponse message
	peers      *PeerSet

	nextTask *FetchTask

	blockProcessor *BlockProcessor

	name string

	debug bool
}

type SyncPeer struct {
	No      int
	ID      peer.ID
	FailCnt int
	IsErr   bool
}

type TaskQueue struct {
	list.List
}

type FetchTask struct {
	count  int
	hashes []message.BlockHash

	syncPeer *SyncPeer

	started time.Time
}

type PeerSet struct {
	total int
	free  int
	bad   int

	freePeers *list.List
	badPeers  *list.List
}

var (
	schedTick        = time.Millisecond * 100
	fetchTimeOut     = time.Second * 100
	MaxFetchTask     = 16
	MaxPeerFailCount = 3
)

var (
	ErrAllPeerBad = errors.New("BlockFetcher: error no avaliable peers")
)

func newBlockFetcher(ctx *types.SyncContext, hub component.ICompRequester) *BlockFetcher {
	bf := &BlockFetcher{ctx: ctx, hub: hub, name: NameBlockFetcher}

	bf.quitCh = make(chan interface{})
	bf.hfCh = make(chan *HashSet)

	bf.peers = newPeerSet()

	bf.blockProcessor = &BlockProcessor{
		hub:           hub,
		blockFetcher:  bf,
		prevBlock:     &types.Block{Hash: ctx.CommonAncestor.Hash},
		targetBlockNo: ctx.TargetNo,
		name:          "BlockProducer",
	}
	bf.blockProcessor.pendingConnect = make([]*ConnectRequest, 0, 16)
	return bf
}

func (bf *BlockFetcher) Start() {
	schedTicker := time.NewTicker(schedTick)

	run := func() {
		if err := bf.init(); err != nil {
			stopSyncer(bf.hub, bf.name, err)
			return
		}

		for {
			select {
			case <-schedTicker.C:
				bf.checkTaskTimeout()

			case msg := <-bf.responseCh:
				err := bf.blockProcessor.run(msg)
				if err != nil {
					logger.Error().Err(err).Msg("invalid block response message")
					stopSyncer(bf.hub, bf.name, err)
					return
				}

			case <-bf.quitCh:
				logger.Info().Msg("BlockFetcher exited")
				return
			}

			//TODO scheduler stop if all tasks have done
			if err := bf.schedule(); err != nil {
				logger.Error().Msg("BlockFetcher schedule failed & finished")
				stopSyncer(bf.hub, bf.name, err)
				return
			}
		}
	}

	go run()
}

func (bf *BlockFetcher) init() error {
	setPeers := func() error {
		result, err := bf.hub.RequestFutureResult(message.P2PSvc, &message.GetPeers{}, dfltTimeout, "BlockFetcher init")
		if err != nil {
			logger.Error().Err(err).Msg("failed to get peers information")
			return err
		}

		for i, peerElem := range result.(message.GetPeersRsp).Peers {
			state := result.(message.GetPeersRsp).States[i]
			if state.Get() == types.RUNNING {
				bf.peers.addNew(peer.ID(peerElem.PeerID))
			}
		}

		if bf.peers.freePeers.Len() != bf.peers.free {
			panic(fmt.Sprintf("free peer len mismatch %d,%d", bf.peers.freePeers.Len(), bf.peers.free))
		}

		return nil
	}

	if err := setPeers(); err != nil {
		logger.Error().Err(err).Msg("failed to set peers")
		return err
	}

	return nil
}

func (bf *BlockFetcher) schedule() error {
	task, err := bf.setNextTask()
	if err != nil {
		logger.Error().Err(err).Msg("error to get next task")
		return err
	}
	if task == nil {
		logger.Debug().Msg("no task to schedule")
		return nil
	}

	freePeer, err := bf.popFreePeer()
	if err != nil {
		logger.Error().Err(err).Msg("error to get free peer")
		return err
	}
	if freePeer == nil {
		logger.Debug().Msg("no freepeer to schedule")
		return nil
	}

	bf.runTask(task, freePeer)

	return nil
}

func (bf *BlockFetcher) checkTaskTimeout() {
	now := time.Now()
	var next *list.Element
	for e := bf.runningQueue.Front(); e != nil; e = next {
		// do something with e.Value
		task := e.Value.(*FetchTask)
		if !task.isTimeOut(now) {
			continue
		}

		next = e.Next()

		bf.runningQueue.Remove(e)

		bf.processFailedTask(task, false)

		logger.Debug().Str("start", enc.ToString(task.hashes[0])).Int("cout", task.count).Int("runqueue", bf.runningQueue.Len()).Int("pendingqueue", bf.pendingQueue.Len()).
			Msg("timeouted task pushed to pending queue")
	}
}

func (bf *BlockFetcher) processFailedTask(task *FetchTask, isErr bool) {
	logger.Error().Str("peer", string(task.syncPeer.ID)).Str("start", enc.ToString(task.hashes[0])).Msg("task fail")

	failPeer := task.syncPeer
	bf.peers.processPeerFail(failPeer, isErr)

	task.syncPeer = nil
	bf.pendingQueue.PushFront(task)
}

func (bf *BlockFetcher) setNextTask() (*FetchTask, error) {
	getNewHashSet := func() *HashSet {
		if bf.curHashSet == nil { //blocking
			logger.Info().Msg("BlockFetcher waiting first hashset")

			select {
			case hashSet := <-bf.hfCh:
				return hashSet
			case <-bf.quitCh:
				return nil
			}
		} else {
			select { //nonblocking
			case hashSet := <-bf.hfCh:
				return hashSet
			case <-bf.quitCh:
				return nil
			default:
				logger.Debug().Msg("BlockFetcher has no input HashSet")
				return nil
			}
		}
	}

	addNewFetchTasks := func(hashSet *HashSet) {
		start, end := 0, 0
		count := hashSet.Count

		logger.Debug().Str("start", enc.ToString(hashSet.Hashes[0])).Int("count", hashSet.Count).Msg("addNew fetchtasks from HashSet")

		for start < count {
			end = start + MaxFetchTask
			if end > count {
				end = count
			}

			task := &FetchTask{count: end - start, hashes: hashSet.Hashes[start:end]}

			logger.Debug().Int("startNo", start).Int("end", end).Msg("addNew fetchtask")

			bf.pendingQueue.PushBack(task)

			start = end
		}
		logger.Debug().Int("pendingqueue", bf.pendingQueue.Len()).Msg("addNewTasks end")
	}

	if bf.nextTask != nil {
		return bf.nextTask, nil
	}

	if bf.pendingQueue.Len() == 0 {
		logger.Debug().Msg("pendingqueue is empty")

		hashSet := getNewHashSet()
		if hashSet == nil {
			logger.Debug().Msg("BlockFetcher no hashSet")
			return nil, nil
		}

		logger.Debug().Str("start", enc.ToString(hashSet.Hashes[0])).Int("count", hashSet.Count).Msg("BlockFetcher got hashset")

		bf.curHashSet = hashSet
		addNewFetchTasks(hashSet)
	}

	//newTask = nil or task
	newTask := bf.pendingQueue.Pop()

	bf.nextTask = newTask

	logger.Debug().Str("start", enc.ToString(newTask.hashes[0])).Str("end", enc.ToString(newTask.hashes[newTask.count-1])).Int("task pending", bf.pendingQueue.Len()).Msg("set new fetchtask")

	return newTask, nil
}

func (bf *BlockFetcher) popFreePeer() (*SyncPeer, error) {
	freePeer, err := bf.peers.popFree()
	if err != nil {
		return nil, err
	}

	bf.nextTask.syncPeer = freePeer

	logger.Debug().Str("peer", string(freePeer.ID)).Int("free", bf.peers.free).Msg("popped free peer")
	return freePeer, nil
}

func (bf *BlockFetcher) pushFreePeer(syncPeer *SyncPeer) {
	bf.peers.pushFree(syncPeer)

	logger.Debug().Str("peer", string(syncPeer.ID)).Int("free", bf.peers.free).Msg("pushed free peer")
}

func (bf *BlockFetcher) runTask(task *FetchTask, peer *SyncPeer) {
	task.started = time.Now()
	bf.runningQueue.PushBack(task)
	bf.nextTask = nil

	logger.Debug().Str("peer", string(task.syncPeer.ID)).Int("count", task.count).Str("start", enc.ToString(task.hashes[0])).Int("runqueue", bf.runningQueue.Len()).Msg("run block fetch task")

	bf.hub.Tell(message.P2PSvc, &message.GetBlockChunks{GetBlockInfos: message.GetBlockInfos{ToWhom: peer.ID, Hashes: task.hashes}, TTL: fetchTimeOut})
}

//TODO refactoring matchFunc
func (bf *BlockFetcher) findFinished(msg *message.GetBlockChunksRsp) (*FetchTask, error) {
	count := len(msg.Blocks)

	var next *list.Element
	for e := bf.runningQueue.Front(); e != nil; e = next {
		// do something with e.Value
		task := e.Value.(*FetchTask)
		next = e.Next()

		if msg.Err != nil && task.isPeerMatched(msg.ToWhom) {
			bf.runningQueue.Remove(e)

			logger.Debug().Str("peer", string(msg.ToWhom)).Str("start", enc.ToString(task.hashes[0])).Int("count", task.count).Int("runqueue", bf.runningQueue.Len()).Msg("task finished with error")
			return task, nil
		}

		if task.isMatched(msg.ToWhom, msg.Blocks, count) {
			bf.runningQueue.Remove(e)

			logger.Debug().Str("start", enc.ToString(task.hashes[0])).Int("count", task.count).Int("runqueue", bf.runningQueue.Len()).
				Msg("task finished")

			return task, nil
		}
	}

	return nil, &ErrSyncMsg{msg: msg}
}

func (bf *BlockFetcher) handleBlockRsp(msg interface{}) error {
	bf.responseCh <- msg
	return nil
}

func (bf *BlockFetcher) stop() {
	logger.Info().Msg("BlockFetcher finished")

	if bf == nil {
		return
	}

	if bf.quitCh != nil {
		close(bf.quitCh)
		bf.quitCh = nil

		close(bf.hfCh)
		bf.hfCh = nil
	}
}

func newPeerSet() *PeerSet {
	ps := &PeerSet{}

	ps.freePeers = list.New()
	ps.badPeers = list.New()

	return ps
}

func (ps *PeerSet) isAllBad() bool {
	if ps.total == ps.badPeers.Len() {
		return true
	}

	return false
}

func (ps *PeerSet) addNew(peerID peer.ID) {
	ps.pushFree(&SyncPeer{No: ps.total, ID: peerID})
	ps.total++

	logger.Info().Str("peer", peerID.String()).Int("no", ps.total).Msg("new peer added")
}

func (ps *PeerSet) pushFree(freePeer *SyncPeer) {
	ps.freePeers.PushBack(freePeer)
	ps.free++

	logger.Info().Int("no", freePeer.No).Int("free", ps.free).Msg("free peer added")
}

func (ps *PeerSet) popFree() (*SyncPeer, error) {
	if ps.isAllBad() {
		logger.Error().Msg("all peers are bad")
		return nil, ErrAllPeerBad
	}

	elem := ps.freePeers.Front()
	if elem == nil {
		return nil, nil
	}

	ps.freePeers.Remove(elem)
	ps.free--

	if ps.freePeers.Len() != ps.free {
		panic(fmt.Sprintf("free peer len mismatch %d,%d", ps.freePeers.Len(), ps.free))
	}

	freePeer := elem.Value.(*SyncPeer)
	logger.Debug().Str("peer", freePeer.ID.String()).Int("no", freePeer.No).Msg("pop free peer")
	return freePeer, nil
}

func (ps *PeerSet) processPeerFail(failPeer *SyncPeer, isErr bool) {
	//TODO handle connection closed
	failPeer.FailCnt++
	failPeer.IsErr = isErr

	logger.Error().Str("peer", string(failPeer.ID)).Int("failcnt", failPeer.FailCnt).Bool("iserr", failPeer.IsErr).Msg("peer failed")

	if isErr || failPeer.FailCnt > MaxPeerFailCount {
		ps.badPeers.PushBack(failPeer)
		ps.bad++

		if ps.badPeers.Len() != ps.bad {
			panic(fmt.Sprintf("bad peer len mismatch %d,%d", ps.badPeers.Len(), ps.bad))
		}

		logger.Error().Str("peer", string(failPeer.ID)).Int("total", ps.total).Int("badqueue", ps.bad).Msg("peer move to bad")
	}
}

func (tq *TaskQueue) Pop() *FetchTask {
	elem := tq.Front()
	if elem == nil {
		return nil
	}

	tq.Remove(elem)
	return elem.Value.(*FetchTask)
}

func (task *FetchTask) isTimeOut(now time.Time) bool {
	if now.Sub(task.started) > fetchTimeOut {
		logger.Info().Str("peer", task.syncPeer.ID.String()).Str("start", enc.ToString(task.hashes[0])).Int("cout", task.count).Msg("FetchTask peer timeouted")
		return true
	}

	return false
}

func (task *FetchTask) isMatched(peerID peer.ID, blocks []*types.Block, count int) bool {
	startHash, endHash := blocks[0].GetHash(), blocks[len(blocks)-1].GetHash()

	if task.count != count ||
		task.syncPeer.ID != peerID ||
		bytes.Compare(task.hashes[0], startHash) != 0 ||
		bytes.Compare(task.hashes[len(task.hashes)-1], endHash) != 0 {
		return false
	}

	for i, block := range blocks {
		if bytes.Compare(task.hashes[i], block.GetHash()) != 0 {
			logger.Info().Str("peer", task.syncPeer.ID.String()).Str("hash", enc.ToString(task.hashes[0])).Int("idx", i).Msg("task hash mismatch")
			return false
		}
	}

	return true
}

func (task *FetchTask) isPeerMatched(peerID peer.ID) bool {
	if task.syncPeer.ID == peerID {
		return true
	}

	return false
}
