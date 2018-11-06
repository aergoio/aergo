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
	quitCh     chan interface{}
	resultCh   chan *HashSet //BlockFetcher input channel (-> BlockFetcher)

	lastBlockInfo *types.BlockInfo
	reqCount      uint64

	name string
}

type HashSet struct {
	Count  int
	Hashes []message.BlockHash
}

var (
	dfltTimeout = time.Second * 180
	HashSetSize = uint64(128)
)

var (
	ErrInvalidHashSet = errors.New("Invalid Hash set reply")
)

func newHashFetcher(ctx *types.SyncContext, hub *component.ComponentHub, bfCh chan *HashSet) *HashFetcher {
	hf := &HashFetcher{ctx: ctx, hub: hub, name: "HashFetcher"}

	hf.quitCh = make(chan interface{})
	hf.responseCh = make(chan *HashSet)

	hf.resultCh = bfCh

	hf.lastBlockInfo = ctx.CommonAncestor
	return hf
}

func (hf *HashFetcher) Start() {
	run := func() {
		hf.requestHashSet()

		for {
			select {
			case HashSet := <-hf.responseCh:
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

func (hf *HashFetcher) requestHashSet() {
	count := HashSetSize
	if hf.ctx.TargetNo < hf.lastBlockInfo.No+HashSetSize {
		count = hf.ctx.TargetNo - hf.lastBlockInfo.No
	}

	hf.reqCount = count

	hf.hub.Tell(message.P2PSvc, &message.GetHashes{ToWhom: hf.ctx.PeerID, PrevInfo: hf.lastBlockInfo, Count: count})
}

func (hf *HashFetcher) processHashSet(hashSet *HashSet) error {
	//get HashSet reply
	lastHash := hashSet.Hashes[len(hashSet.Hashes)-1]
	lastHashNo := hf.lastBlockInfo.No + uint64(hashSet.Count)

	if lastHashNo > hf.ctx.TargetNo {
		logger.Error().Uint64("target", hf.ctx.TargetNo).Uint64("last", lastHashNo).Msg("invalid HashSet reponse")
		return ErrInvalidHashSet
	}

	hf.lastBlockInfo = &types.BlockInfo{Hash: lastHash, No: lastHashNo}
	hf.resultCh <- hashSet

	return nil
}

func (hf *HashFetcher) stop() {
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
		logger.Error().Str("req prev", enc.ToString(hf.lastBlockInfo.Hash)).
			Str("msg prev", enc.ToString(msg.PrevInfo.Hash)).
			Uint64("req count", hf.reqCount).
			Uint64("msg count", msg.Count).
			Msg("invalid GetHashesRsp in HashFetcher")
		return false
	}

	return true
}

func (hf *HashFetcher) handleGetHahsesRsp(msg *message.GetHashesRsp) {
	if !hf.isValidResponse(msg) {
		return
	}

	count := len(msg.Hashes)
	logger.Debug().Int("count", count).
		Str("start", enc.ToString(msg.Hashes[0])).
		Str("end", enc.ToString(msg.Hashes[count-1])).Msg("receive GetHashesRsp")

	hf.responseCh <- &HashSet{Count: len(msg.Hashes), Hashes: msg.Hashes}
	return
}
