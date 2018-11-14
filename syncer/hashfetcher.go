package syncer

import (
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/pkg/errors"
	"time"
)

type HashFetcher struct {
	hub component.ICompRequester //for communicate with other service

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

	finished bool
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
	dfltTimeout = time.Second * 180
	MaxHashReq  = uint64(128)
)

var (
	ErrInvalidHashSet     = errors.New("Invalid hash set reply")
	ErrGetHashesRspError  = errors.New("GetHashesRsp error received")
	ErrHashFetcherTimeout = errors.New("HashFetcher response timeout")
)

func newHashFetcher(ctx *types.SyncContext, hub component.ICompRequester, bfCh chan *HashSet, maxHashReq uint64) *HashFetcher {
	hf := &HashFetcher{ctx: ctx, hub: hub, name: NameHashFetcher}

	hf.quitCh = make(chan interface{})
	hf.responseCh = make(chan *message.GetHashesRsp)

	hf.resultCh = bfCh

	hf.lastBlockInfo = ctx.CommonAncestor

	hf.maxHashReq = maxHashReq

	hf.timeout = dfltTimeout

	return hf
}

func (hf *HashFetcher) setTimeout(timeout time.Duration) {
	hf.timeout = timeout
}

func (hf *HashFetcher) Start() {
	run := func() {
		logger.Debug().Msg("start hash fetcher")

		timer := time.NewTimer(hf.timeout)

		hf.requestHashSet()

		for {
		NEXT:
			select {
			case msg := <-hf.responseCh:
				logger.Debug().Msg("process GetHashesRsp")

				timer.Stop()
				if !hf.isValidResponse(msg) {
					goto NEXT
				}

				HashSet := &HashSet{Count: len(msg.Hashes), Hashes: msg.Hashes, StartNo: msg.PrevInfo.No + 1}

				if err := hf.processHashSet(HashSet); err != nil {
					//TODO send errmsg to syncer & stop sync
					logger.Error().Err(err).Msg("error! process hash chunk, HashFetcher exited")
					stopSyncer(hf.hub, hf.name, err)
					return
				}

				if hf.isFinished(HashSet) {
					hf.finished = true
					closeFetcher(hf.hub, hf.name)
					logger.Info().Msg("HashFetcher finished")
					return
				}
				hf.requestHashSet()

				//timer restart
				timer.Reset(hf.timeout)
			case <-timer.C:
				if hf.requestTimeout() {
					logger.Error().Msg("HashFetcher response timeout.")
					stopSyncer(hf.hub, hf.name, ErrHashFetcherTimeout)
				}

			case <-hf.quitCh:
				logger.Info().Msg("HashFetcher exited")
				return
			}
		}
	}

	go run()
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

	hf.hub.Tell(message.P2PSvc, &message.GetHashes{ToWhom: hf.ctx.PeerID, PrevInfo: hf.lastBlockInfo, Count: count})
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
	hf.resultCh <- hashSet
	hf.isRequesting = false

	logger.Debug().Uint64("target", hf.ctx.TargetNo).Uint64("start", hashSet.StartNo).Uint64("last", lastHashNo).Int("count", len(hashSet.Hashes)).Msg("push hashset to BlockFetcher")

	return nil
}

func (hf *HashFetcher) stop() {
	logger.Info().Msg("HashFetcher finished")

	if hf == nil {
		return
	}

	if hf.quitCh != nil {
		close(hf.quitCh)
		hf.quitCh = nil

		close(hf.responseCh)
		hf.responseCh = nil
	}
}

func (hf *HashFetcher) isValidResponse(msg *message.GetHashesRsp) bool {
	isValid := true

	if msg.Err != nil {
		logger.Error().Err(msg.Err).Msg("receive GetHashesRsp with error")
		stopSyncer(hf.hub, hf.name, ErrGetHashesRspError)
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
	}

	return true
}

func (hf *HashFetcher) GetHahsesRsp(msg *message.GetHashesRsp) {
	count := len(msg.Hashes)
	logger.Debug().Int("count", count).
		Uint64("prev", msg.PrevInfo.No).
		Str("start", enc.ToString(msg.Hashes[0])).
		Str("end", enc.ToString(msg.Hashes[count-1])).Msg("receive GetHashesRsp")

	hf.responseCh <- msg
	return
}
