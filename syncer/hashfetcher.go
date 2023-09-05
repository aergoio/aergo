package syncer

import (
	"sync"
	"time"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/pkg/errors"
)

type HashFetcher struct {
	compRequester component.IComponentRequester //for communicate with other service

	ctx *types.SyncContext

	responseCh chan *message.GetHashesRsp //HashSet response channel (<- Syncer)
	resultCh   chan *HashSet              //BlockFetcher input channel (-> BlockFetcher)
	//HashFetcher can wait in resultCh
	quitCh chan interface{}

	lastBlockInfo *types.BlockInfo
	reqCount      uint64
	reqTime       time.Time
	isRequesting  bool

	maxHashReq uint64
	name       string

	timeout time.Duration
	debug   bool

	isRunning bool

	waitGroup *sync.WaitGroup
}

type HashSet struct {
	Count   int
	Hashes  []message.BlockHash
	StartNo types.BlockNo
}

type HashRequest struct {
	prevInfo *types.BlockInfo
	count    uint64
}

var (
	dfltTimeout     = time.Second * 180
	DfltHashReqSize = uint64(1000)
)

var (
	ErrQuitHashFetcher    = errors.New("Hashfetcher quit")
	ErrInvalidHashSet     = errors.New("Invalid hash set reply")
	ErrHashFetcherTimeout = errors.New("HashFetcher response timeout")
)

func newHashFetcher(ctx *types.SyncContext, compRequester component.IComponentRequester, bfCh chan *HashSet, cfg *SyncerConfig) *HashFetcher {
	hf := &HashFetcher{ctx: ctx, compRequester: compRequester, name: NameHashFetcher}

	hf.quitCh = make(chan interface{})
	hf.responseCh = make(chan *message.GetHashesRsp)

	hf.resultCh = bfCh

	hf.lastBlockInfo = &types.BlockInfo{Hash: ctx.CommonAncestor.GetHash(), No: ctx.CommonAncestor.BlockNo()}

	hf.maxHashReq = cfg.maxHashReqSize

	hf.timeout = dfltTimeout

	return hf
}

func (hf *HashFetcher) setTimeout(timeout time.Duration) {
	hf.timeout = timeout
}

func (hf *HashFetcher) recover() {

}
func (hf *HashFetcher) Start() {
	hf.waitGroup = &sync.WaitGroup{}
	hf.waitGroup.Add(1)

	hf.isRunning = true

	run := func() {
		defer RecoverSyncer(NameHashFetcher, hf.GetSeq(), hf.compRequester, func() { hf.waitGroup.Done() })

		logger.Debug().Msg("start hash fetcher")

		timer := time.NewTimer(hf.timeout)

		hf.requestHashSet()

		for {
			select {
			case msg, ok := <-hf.responseCh:
				if !ok {
					logger.Error().Msg("HashFetcher responseCh is closed. Syncer is stopping now")
					return
				}
				logger.Debug().Msg("process GetHashesRsp")

				timer.Stop()
				res, err := hf.isValidResponse(msg)
				if res {
					HashSet := &HashSet{Count: len(msg.Hashes), Hashes: msg.Hashes, StartNo: msg.PrevInfo.No + 1}

					if err := hf.processHashSet(HashSet); err != nil {
						//TODO send errmsg to syncer & stop sync
						logger.Error().Err(err).Msg("error! process hash chunk, HashFetcher exited")
						if err != ErrQuitHashFetcher {
							stopSyncer(hf.compRequester, hf.GetSeq(), hf.name, err)
						}
						return
					}

					if hf.isFinished(HashSet) {
						closeFetcher(hf.compRequester, hf.GetSeq(), hf.name)
						logger.Info().Msg("HashFetcher finished")
						return
					}
					hf.requestHashSet()
				} else if err != nil {
					stopSyncer(hf.compRequester, hf.GetSeq(), hf.name, err)
				}

				//timer restart
				timer.Reset(hf.timeout)
			case <-timer.C:
				if hf.requestTimeout() {
					logger.Error().Msg("HashFetcher response timeout.")
					stopSyncer(hf.compRequester, hf.GetSeq(), hf.name, ErrHashFetcherTimeout)
				}

			case <-hf.quitCh:
				logger.Info().Msg("HashFetcher exited")
				return
			}
		}
	}

	go run()
}

func (hf *HashFetcher) GetSeq() uint64 {
	return hf.ctx.Seq
}

func (hf *HashFetcher) isFinished(HashSet *HashSet) bool {
	return (hf.lastBlockInfo.No == hf.ctx.TargetNo)
}

func (hf *HashFetcher) requestTimeout() bool {
	return hf.isRequesting && time.Now().Sub(hf.reqTime) > hf.timeout
}

func (hf *HashFetcher) requestHashSet() {
	count := hf.maxHashReq
	if hf.ctx.TargetNo < hf.lastBlockInfo.No+hf.maxHashReq {
		count = hf.ctx.TargetNo - hf.lastBlockInfo.No
	}

	hf.reqCount = count
	hf.reqTime = time.Now()
	hf.isRequesting = true

	logger.Debug().Uint64("prev", hf.lastBlockInfo.No).Str("prevhash", enc.ToString(hf.lastBlockInfo.Hash)).Uint64("count", count).Msg("request hashset to peer")

	hf.compRequester.TellTo(message.P2PSvc, &message.GetHashes{Seq: hf.GetSeq(), ToWhom: hf.ctx.PeerID, PrevInfo: hf.lastBlockInfo, Count: count})
}

func (hf *HashFetcher) processHashSet(hashSet *HashSet) error {
	//get HashSet reply
	lastHash := hashSet.Hashes[len(hashSet.Hashes)-1]
	lastHashNo := hashSet.StartNo + uint64(hashSet.Count) - 1

	if lastHashNo > hf.ctx.TargetNo {
		logger.Error().Uint64("target", hf.ctx.TargetNo).Uint64("last", lastHashNo).Msg("invalid hashset reponse")
		return ErrInvalidHashSet
	}

	hf.lastBlockInfo = &types.BlockInfo{Hash: lastHash, No: lastHashNo}

	//total HashSet in memory can be 3 (network + resultCh + blockFetcher)
	select {
	case hf.resultCh <- hashSet:
	case <-hf.quitCh:
		logger.Info().Msg("hash fetcher quit while pushing result")
		return ErrQuitHashFetcher
	}

	hf.isRequesting = false

	logger.Debug().Uint64("target", hf.ctx.TargetNo).Uint64("start", hashSet.StartNo).Uint64("last", lastHashNo).Int("count", len(hashSet.Hashes)).Msg("push hashset to BlockFetcher")

	return nil
}

func (hf *HashFetcher) stop() {
	if hf == nil {
		return
	}

	if hf.isRunning {
		logger.Info().Msg("HashFetcher stop#1")

		close(hf.quitCh)
		logger.Info().Msg("HashFetcher close quitCh")

		close(hf.responseCh)

		hf.waitGroup.Wait()
		hf.isRunning = false
	}
	logger.Info().Msg("HashFetcher stopped")
}

func (hf *HashFetcher) isValidResponse(msg *message.GetHashesRsp) (bool, error) {
	isValid := true
	var err error

	if msg == nil {
		panic("nil message error")
	}

	if msg.Err != nil {
		logger.Error().Err(msg.Err).Msg("receive GetHashesRsp with error")
		err = msg.Err
		isValid = false
	}

	if !hf.lastBlockInfo.Equal(msg.PrevInfo) || hf.reqCount != msg.Count {
		isValid = false
	}

	if !isValid {
		logger.Error().Str("req prev", enc.ToString(hf.lastBlockInfo.Hash)).
			Str("msg prev", enc.ToString(msg.PrevInfo.Hash)).
			Uint64("req count", hf.reqCount).
			Uint64("msg count", msg.Count).
			Msg("invalid GetHashesRsp")
		return false, err
	}

	return true, nil
}

func (hf *HashFetcher) GetHahsesRsp(msg *message.GetHashesRsp) {
	if hf == nil {
		return
	}

	count := len(msg.Hashes)

	if count == 0 {
		logger.Error().Int("count", count).
			Uint64("prev", msg.PrevInfo.No).Msg("receive empty GetHashesRsp")
		return
	}

	logger.Debug().Int("count", count).
		Uint64("prev", msg.PrevInfo.No).
		Str("start", enc.ToString(msg.Hashes[0])).
		Str("end", enc.ToString(msg.Hashes[count-1])).Msg("receive GetHashesRsp")

	hf.responseCh <- msg
	return
}
