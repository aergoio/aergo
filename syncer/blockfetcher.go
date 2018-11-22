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
	"sync"
	"sync/atomic"
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
	retryQueue   TaskQueue

	responseCh chan interface{} //BlockResponse, AddBlockResponse message
	peers      *PeerSet

	blockProcessor *BlockProcessor

	name string

	maxFetchSize   int
	maxFetchTasks  int
	maxPendingConn int

	debug bool

	stat BlockFetcherStat

	waitGroup *sync.WaitGroup
	isRunning bool
}

type BlockFetcherStat struct {
	maxRspBlock  atomic.Value
	lastAddBlock atomic.Value
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
	count   int
	hashes  []message.BlockHash
	startNo types.BlockNo

	syncPeer *SyncPeer

	started time.Time
	retry   int
}

type PeerSet struct {
	total int
	free  int
	bad   int

	freePeers *list.List
	badPeers  *list.List
}

var (
	schedTick          = time.Millisecond * 100
	fetchTimeOut       = time.Second * 30
	DfltBlockFetchSize = 20
	//DfltBlockFetchSize   = 5
	MaxPeerFailCount     = 3
	DfltBlockFetchTasks  = 5
	MaxBlockPendingTasks = 10
)

var (
	ErrAllPeerBad   = errors.New("BlockFetcher: error no avaliable peers")
	ErrNotFoundTask = errors.New("not found next task")
)

func newBlockFetcher(ctx *types.SyncContext, hub component.ICompRequester, maxFetchSize int, maxRunningFetchTasks int,
	maxPendingConnTasks int) *BlockFetcher {
	bf := &BlockFetcher{ctx: ctx, hub: hub, name: NameBlockFetcher}

	bf.quitCh = make(chan interface{})
	bf.hfCh = make(chan *HashSet)
	bf.responseCh = make(chan interface{}, maxRunningFetchTasks*2) //for safety. In normal situdation, it should use only one

	bf.peers = newPeerSet()
	bf.maxFetchSize = maxFetchSize
	bf.maxFetchTasks = maxRunningFetchTasks
	bf.maxPendingConn = maxPendingConnTasks

	bf.blockProcessor = NewBlockProcessor(hub, bf, ctx.CommonAncestor, ctx.TargetNo)

	bf.blockProcessor.connQueue = make([]*ConnectTask, 0, 16)

	bf.runningQueue.Init()
	bf.pendingQueue.Init()
	bf.retryQueue.Init()

	return bf
}

func (bf *BlockFetcher) Start() {
	bf.waitGroup = &sync.WaitGroup{}
	bf.waitGroup.Add(1)

	schedTicker := time.NewTicker(schedTick)

	bf.isRunning = true

	run := func() {
		defer bf.waitGroup.Done()

		logger.Debug().Msg("start block fetcher")

		if err := bf.init(); err != nil {
			stopSyncer(bf.hub, bf.name, err)
			return
		}

		logger.Debug().Msg("block fetcher loop start")

		for {
			select {
			case <-schedTicker.C:
				if err := bf.checkTaskTimeout(); err != nil {
					logger.Error().Err(err).Msg("failed checkTaskTimeout")
					stopSyncer(bf.hub, bf.name, err)
					return
				}

			case msg, ok := <-bf.responseCh:
				if !ok {
					logger.Info().Msg("BlockFetcher responseCh is closed. Syncer is maybe stopping.")
					return
				}

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

		msg := result.(*message.GetPeersRsp)

		for i, peerElem := range msg.Peers {
			state := msg.States[i]
			if state.Get() == types.RUNNING {
				bf.peers.addNew(peer.ID(peerElem.PeerID))
			}
		}

		if bf.peers.freePeers.Len() != bf.peers.free {
			panic(fmt.Sprintf("free peer len mismatch %d,%d", bf.peers.freePeers.Len(), bf.peers.free))
		}

		return nil
	}

	logger.Debug().Msg("block fetcher init")

	if err := setPeers(); err != nil {
		logger.Error().Err(err).Msg("failed to set peers")
		return err
	}

	return nil
}

func (bf *BlockFetcher) schedule() error {
	//check no free peer
	if bf.peers.free == 0 {
		return nil
	}

	//check max concurrent runing task count
	curRunning := bf.runningQueue.Len()
	if curRunning >= bf.maxFetchTasks {
		//logger.Debug().Int("runnig", curRunning).Int("pending", bf.pendingQueue.Len()).Msg("max running")
		return nil
	}

	//check no task
	candTask, err := bf.searchCandidateTask()
	if err != nil {
		logger.Error().Err(err).Msg("failed to search candidate task")
		return err
	}
	if candTask == nil {
		return nil
	}

	//check max pending connect
	//	if next task is retry task, must run. it can be next block to connect
	curPendingConn := len(bf.blockProcessor.connQueue)
	if curPendingConn >= bf.maxPendingConn && candTask.retry <= 0 {
		return nil
	}

	freePeer, err := bf.popFreePeer()
	if err != nil {
		logger.Error().Err(err).Msg("error to get free peer")
		return err
	}
	if freePeer == nil {
		panic("free peer can't be nil")
	}

	bf.popNextTask(candTask)
	if candTask == nil {
		panic("task can't be nil")
	}

	logger.Debug().Int("pendingConn", curPendingConn).Int("running", curRunning).Msg("schedule")
	bf.runTask(candTask, freePeer)

	return nil
}

func (bf *BlockFetcher) checkTaskTimeout() error {
	now := time.Now()
	var next *list.Element

	for e := bf.runningQueue.Front(); e != nil; e = next {
		// do something with e.Value
		task := e.Value.(*FetchTask)
		next = e.Next()

		if !task.isTimeOut(now) {
			continue
		}

		bf.runningQueue.Remove(e)

		if err := bf.processFailedTask(task, false); err != nil {
			return err
		}

		logger.Error().Uint64("StartNo", task.startNo).Str("start", enc.ToString(task.hashes[0])).Int("cout", task.count).Int("runqueue", bf.runningQueue.Len()).Int("pendingqueue", bf.pendingQueue.Len()).
			Msg("timeouted task pushed to pending queue")

		//time.Sleep(10000*time.Second)
	}

	return nil
}

func (bf *BlockFetcher) processFailedTask(task *FetchTask, isErr bool) error {
	logger.Error().Int("peerno", task.syncPeer.No).Uint64("StartNo", task.startNo).Str("start", enc.ToString(task.hashes[0])).Msg("task fail, move to retry queue")

	failPeer := task.syncPeer
	bf.peers.processPeerFail(failPeer, isErr)

	task.retry++
	task.syncPeer = nil

	bf.retryQueue.PushBack(task)

	if bf.peers.isAllBad() {
		return ErrAllPeerBad
	}

	return nil
}

func (bf *BlockFetcher) popNextTask(task *FetchTask) {
	logger.Debug().Int("retry", task.retry).Uint64("StartNo", task.startNo).Str("start", enc.ToString(task.hashes[0])).Str("end", enc.ToString(task.hashes[task.count-1])).
		Int("tasks retry", bf.retryQueue.Len()).Int("tasks pending", bf.pendingQueue.Len()).Msg("next fetchtask")

	var poppedTask *FetchTask
	if task.retry > 0 {
		poppedTask = bf.retryQueue.Pop()
	} else {
		poppedTask = bf.pendingQueue.Pop()
	}

	if poppedTask != task {
		logger.Panic().Uint64("next task", task.startNo).Uint64("popped task", poppedTask.startNo).
			Int("retry", task.retry).Msg("peeked task is not popped task")
		panic("peeked task is not popped task")
	}
}

func (bf *BlockFetcher) searchCandidateTask() (*FetchTask, error) {
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
				//logger.Debug().Msg("BlockFetcher has no input HashSet")
				return nil
			}
		}
	}

	addNewFetchTasks := func(hashSet *HashSet) {
		start, end := 0, 0
		count := hashSet.Count

		logger.Debug().Uint64("startno", hashSet.StartNo).Str("start", enc.ToString(hashSet.Hashes[0])).Int("count", hashSet.Count).Msg("add new fetchtasks from HashSet")

		for start < count {
			end = start + bf.maxFetchSize
			if end > count {
				end = count
			}

			task := &FetchTask{count: end - start, hashes: hashSet.Hashes[start:end], startNo: hashSet.StartNo + uint64(start), retry: 0}

			logger.Debug().Uint64("StartNo", task.startNo).Int("count", task.count).Msg("add fetchtask")

			bf.pendingQueue.PushBack(task)

			start = end
		}
		logger.Debug().Int("pendingqueue", bf.pendingQueue.Len()).Msg("addNewTasks end")
	}

	var newTask *FetchTask

	if bf.retryQueue.Len() > 0 {
		newTask = bf.retryQueue.Peek()
	} else {
		if bf.pendingQueue.Len() == 0 {
			logger.Debug().Msg("pendingqueue is empty")

			hashSet := getNewHashSet()
			if hashSet == nil {
				logger.Debug().Msg("BlockFetcher no hashSet")
				return nil, nil
			}

			logger.Debug().Uint64("startno", hashSet.StartNo).Str("start", enc.ToString(hashSet.Hashes[0])).Int("count", hashSet.Count).Msg("BlockFetcher got hashset")

			bf.curHashSet = hashSet
			addNewFetchTasks(hashSet)
		}

		newTask = bf.pendingQueue.Peek()
	}

	//logger.Debug().Int("retry", newTask.retry).Uint64("StartNo", newTask.startNo).Str("start", enc.ToString(newTask.hashes[0])).Str("end", enc.ToString(newTask.hashes[newTask.count-1])).
	//	Int("tasks retry", bf.retryQueue.Len()).Int("tasks pending", bf.pendingQueue.Len()).Msg("candidate fetchtask")

	return newTask, nil
}

func (bf *BlockFetcher) popFreePeer() (*SyncPeer, error) {
	freePeer, err := bf.peers.popFree()
	if err != nil {
		logger.Error().Err(err).Msg("pop free peer failed")
		return nil, err
	}

	if freePeer != nil {
		logger.Debug().Int("peerno", freePeer.No).Int("free", bf.peers.free).Int("total", bf.peers.total).Int("bad", bf.peers.bad).Msg("popped free peer")
	} else {
		logger.Debug().Int("free", bf.peers.free).Int("total", bf.peers.total).Int("bad", bf.peers.bad).Msg("not exist free peer")
	}

	return freePeer, nil
}

func (bf *BlockFetcher) pushFreePeer(syncPeer *SyncPeer) {
	bf.peers.pushFree(syncPeer)

	logger.Debug().Int("peerno", syncPeer.No).Int("free", bf.peers.free).Msg("pushed free peer")
}

func (bf *BlockFetcher) runTask(task *FetchTask, peer *SyncPeer) {
	task.started = time.Now()
	task.syncPeer = peer
	bf.runningQueue.PushBack(task)

	logger.Debug().Int("peerno", task.syncPeer.No).Int("count", task.count).Uint64("StartNo", task.startNo).Str("start", enc.ToString(task.hashes[0])).Int("runqueue", bf.runningQueue.Len()).Msg("send block fetch request")

	bf.hub.Tell(message.P2PSvc, &message.GetBlockChunks{GetBlockInfos: message.GetBlockInfos{ToWhom: peer.ID, Hashes: task.hashes}, TTL: fetchTimeOut})
}

//TODO refactoring matchFunc
func (bf *BlockFetcher) findFinished(msg *message.GetBlockChunksRsp, peerMatch bool) (*FetchTask, error) {
	count := len(msg.Blocks)

	var next *list.Element
	for e := bf.runningQueue.Front(); e != nil; e = next {
		// do something with e.Value
		task := e.Value.(*FetchTask)
		next = e.Next()

		//find failed peer
		if peerMatch {
			if task.isPeerMatched(msg.ToWhom) {
				bf.runningQueue.Remove(e)

				logger.Debug().Str("peer", msg.ToWhom.Pretty()).Err(msg.Err).Str("start", enc.ToString(task.hashes[0])).Int("count", task.count).Int("runqueue", bf.runningQueue.Len()).Msg("task finished with error")
				return task, nil
			}
		} else {
			//find finished peer
			if task.isMatched(msg.ToWhom, msg.Blocks, count) {
				bf.runningQueue.Remove(e)

				logger.Debug().Uint64("StartNo", task.startNo).Str("start", enc.ToString(task.hashes[0])).Int("count", task.count).Int("runqueue", bf.runningQueue.Len()).
					Msg("task finished")

				return task, nil
			}
		}
	}

	return nil, &ErrSyncMsg{msg: msg}
}

func (bf *BlockFetcher) handleBlockRsp(msg interface{}) error {
	bf.responseCh <- msg
	return nil
}

func (bf *BlockFetcher) stop() {
	if bf == nil {
		return
	}

	//logger.Info().Bool("isrunning", bf.isRunning).Bool("isnil", bf.quitCh== nil).Msg("BlockFetcher stop")

	if bf.isRunning {
		logger.Info().Msg("BlockFetcher stop#1")

		close(bf.quitCh)
		bf.quitCh = nil

		close(bf.hfCh)
		bf.hfCh = nil

		bf.waitGroup.Wait()
		bf.isRunning = false
	}
	logger.Info().Msg("BlockFetcher stopped")
}

func (stat *BlockFetcherStat) setMaxChunkRsp(lastBlock *types.Block) {
	curMaxRspBlock := stat.getMaxChunkRsp()

	if curMaxRspBlock == nil || curMaxRspBlock.GetHeader().BlockNo < lastBlock.GetHeader().BlockNo {
		stat.maxRspBlock.Store(lastBlock)
		logger.Debug().Uint64("no", lastBlock.GetHeader().BlockNo).Msg("last block chunk response")
	}
}

func (stat *BlockFetcherStat) setLastAddBlock(block *types.Block) {
	stat.lastAddBlock.Store(block)
	logger.Debug().Uint64("no", block.GetHeader().BlockNo).Msg("last block add response")
}

func (stat *BlockFetcherStat) getMaxChunkRsp() *types.Block {
	aopv := stat.maxRspBlock.Load()
	if aopv != nil {
		return aopv.(*types.Block)
	}

	return nil
}

func (stat *BlockFetcherStat) getLastAddBlock() *types.Block {
	aopv := stat.lastAddBlock.Load()
	if aopv != nil {
		return aopv.(*types.Block)
	}

	return nil
}

func newPeerSet() *PeerSet {
	ps := &PeerSet{}

	ps.freePeers = list.New()
	ps.badPeers = list.New()

	return ps
}

func (ps *PeerSet) isAllBad() bool {
	if ps.total == ps.bad {
		return true
	}

	return false
}

func (ps *PeerSet) addNew(peerID peer.ID) {
	peerno := ps.total
	ps.pushFree(&SyncPeer{No: peerno, ID: peerID})
	ps.total++

	logger.Info().Str("peer", peerID.Pretty()).Int("peerno", peerno).Int("no", ps.total).Msg("new peer added")
}

/*
func (ps *PeerSet) print() {

}*/
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
	logger.Debug().Int("peerno", freePeer.No).Int("no", freePeer.No).Msg("free peer poped")
	return freePeer, nil
}

func (ps *PeerSet) processPeerFail(failPeer *SyncPeer, isErr bool) {
	//TODO handle connection closed
	failPeer.FailCnt++
	failPeer.IsErr = isErr

	logger.Error().Int("peerno", failPeer.No).Int("failcnt", failPeer.FailCnt).Int("maxfailcnt", MaxPeerFailCount).Bool("iserr", failPeer.IsErr).Msg("peer failed")

	if isErr || failPeer.FailCnt >= MaxPeerFailCount {
		ps.badPeers.PushBack(failPeer)
		ps.bad++

		if ps.badPeers.Len() != ps.bad {
			panic(fmt.Sprintf("bad peer len mismatch %d,%d", ps.badPeers.Len(), ps.bad))
		}

		logger.Error().Int("peerno", failPeer.No).Int("total", ps.total).Int("free", ps.free).Int("bad", ps.bad).Msg("peer move to bad")
	} else {
		ps.freePeers.PushBack(failPeer)
		ps.free++

		logger.Error().Int("peerno", failPeer.No).Int("total", ps.total).Int("free", ps.free).Int("bad", ps.bad).Msg("peer move to free")
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

func (tq *TaskQueue) Peek() *FetchTask {
	elem := tq.Front()
	if elem == nil {
		return nil
	}

	return elem.Value.(*FetchTask)
}

func (task *FetchTask) isTimeOut(now time.Time) bool {
	if now.Sub(task.started) > fetchTimeOut {
		logger.Info().Int("peerno", task.syncPeer.No).Uint64("startno", task.startNo).Str("start", enc.ToString(task.hashes[0])).Int("cout", task.count).Msg("FetchTask peer timeouted")
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
			logger.Info().Int("peerno", task.syncPeer.No).Str("hash", enc.ToString(task.hashes[0])).Int("idx", i).Msg("task hash mismatch")
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
