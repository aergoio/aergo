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
	hub *component.ComponentHub //for communicate with other service

	ctx *types.SyncContext

	responseCh chan *HashSet //HashSet response channel (<- Syncer)
	resultCh   chan *HashSet //BlockFetcher input channel (-> BlockFetcher)
	//HashFetcher can wait in resultCh
	quitCh chan interface{}

	lastBlockInfo *types.BlockInfo
	reqCount      uint64
	reqTime       time.Time

	maxReqSize uint64
	name       string
}

type HashSet struct {
	Count  int
	Hashes []message.BlockHash
}

type HashRequest struct {
	prevInfo *types.BlockInfo
	count    uint64
}

var (
	dfltTimeout    = time.Second * 180
	MaxHashSetSize = uint64(128)
)

var (
	ErrInvalidHashSet     = errors.New("Invalid hash set reply")
	ErrGetHashesRspError  = errors.New("GetHashesRsp error received")
	ErrHashFetcherTimeout = errors.New("HashFetcher response timeout")
)

func newHashFetcher(ctx *types.SyncContext, hub *component.ComponentHub, bfCh chan *HashSet, maxReqSize uint64) *HashFetcher {
	hf := &HashFetcher{ctx: ctx, hub: hub, name: "HashFetcher"}

	hf.quitCh = make(chan interface{})
	hf.responseCh = make(chan *HashSet)

	hf.resultCh = bfCh

	hf.lastBlockInfo = ctx.CommonAncestor

	hf.maxReqSize = maxReqSize

	return hf
}

func (hf *HashFetcher) Start() {
	run := func() {
		timer := time.NewTimer(dfltTimeout)

		hf.requestHashSet()

		for {
			select {
			case HashSet := <-hf.responseCh:
				timer.Stop()

				if err := hf.processHashSet(HashSet); err != nil {
					//TODO send errmsg to syncer & stop sync
					logger.Error().Err(err).Msg("error! process hash chunk, HashFetcher exited")
					stopSyncer(hf.hub, hf.name, err)
					return
				}

				if hf.isFinished(HashSet) {
					logger.Info().Msg("HashFetcher finished")
					return
				}
				hf.requestHashSet()

				//timer restart
				timer.Reset(dfltTimeout)
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
	return time.Now().Sub(hf.reqTime) > dfltTimeout
}

func (hf *HashFetcher) requestHashSet() {
	count := MaxHashSetSize
	if hf.ctx.TargetNo < hf.lastBlockInfo.No+MaxHashSetSize {
		count = hf.ctx.TargetNo - hf.lastBlockInfo.No
	}

	hf.reqCount = count
	hf.reqTime = time.Now()

	logger.Debug().Uint64("prev", hf.lastBlockInfo.No).Str("prevhash", enc.ToString(hf.lastBlockInfo.Hash)).Uint64("count", count).Msg("request hashset to peer")

	hf.hub.Tell(message.P2PSvc, &message.GetHashes{ToWhom: hf.ctx.PeerID, PrevInfo: hf.lastBlockInfo, Count: count})
}

func (hf *HashFetcher) processHashSet(hashSet *HashSet) error {
	//get HashSet reply
	lastHash := hashSet.Hashes[len(hashSet.Hashes)-1]
	lastHashNo := hf.lastBlockInfo.No + uint64(hashSet.Count)

	if lastHashNo > hf.ctx.TargetNo {
		logger.Error().Uint64("target", hf.ctx.TargetNo).Uint64("last", lastHashNo).Msg("invalid hashset reponse")
		return ErrInvalidHashSet
	}

	hf.lastBlockInfo = &types.BlockInfo{Hash: lastHash, No: lastHashNo}
	hf.resultCh <- hashSet

	logger.Debug().Uint64("target", hf.ctx.TargetNo).Uint64("last", lastHashNo).Int("count", len(hashSet.Hashes)).Msg("push hashset to BlockFetcher")

	return nil
}

func (hf *HashFetcher) stop() {
	logger.Info().Msg("HashFetcher stopped")

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
	if !hf.lastBlockInfo.Equal(msg.PrevInfo) || hf.reqCount != msg.Count {
		return false
	}

	return true
}

func (hf *HashFetcher) GetHahsesRsp(msg *message.GetHashesRsp) {
	if !hf.isValidResponse(msg) {
		logger.Error().Str("req prev", enc.ToString(hf.lastBlockInfo.Hash)).
			Str("msg prev", enc.ToString(msg.PrevInfo.Hash)).
			Uint64("req count", hf.reqCount).
			Uint64("msg count", msg.Count).
			Msg("invalid GetHashesRsp")
		return
	}

	if msg.Err != nil {
		logger.Error().Err(msg.Err).Msg("receive GetHashesRsp with error")
		stopSyncer(hf.hub, hf.name, ErrGetHashesRspError)
		return
	}

	count := len(msg.Hashes)
	logger.Debug().Int("count", count).
		Str("start", enc.ToString(msg.Hashes[0])).
		Str("end", enc.ToString(msg.Hashes[count-1])).Msg("receive GetHashesRsp")

	hf.responseCh <- &HashSet{Count: len(msg.Hashes), Hashes: msg.Hashes}
	return
}
