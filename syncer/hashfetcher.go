package syncer

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/pkg/errors"
	"time"
)

type HashFetcher struct {
	hub *component.ComponentHub //for communicate with other service

	ctx *types.SyncContext

	curRequest int

	responseCh chan *HashSet //HashSet response channel (<- Syncer)
	quitCh     chan interface{}
	resultCh   chan *HashSet //BlockFetcher input channel (-> BlockFetcher)

	lastBlockInfo *types.BlockInfo
}

type HashSet struct {
	Count  int
	Hashes []message.BlockHash
}

var (
	dfltTimeout = time.Second * 100
	HashSetSize = uint64(128)
)

var (
	ErrInvalidHashSet = errors.New("Invalid Hash set reply")
)

func newHashFetcher(ctx *types.SyncContext, hub *component.ComponentHub, bfCh chan *HashSet) *HashFetcher {
	hf := &HashFetcher{ctx: ctx, hub: hub}

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
					logger.Panic().Err(err).Msg("error! process hash chunk")
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
	close(hf.quitCh)
}

func (hf *HashFetcher) isValidResponse(msg *message.GetHashesRsp) bool {
	if !hf.lastBlockInfo.Equal(msg.PrevInfo) {
		return false
	}

	return true
}

func (hf *HashFetcher) setResult(msg *message.GetHashesRsp) error {
	if !hf.isValidResponse(msg) {
		return ErrInvalidHashSet
	}

	hf.responseCh <- &HashSet{Count: len(msg.Hashes), Hashes: msg.Hashes}
	return nil
}
