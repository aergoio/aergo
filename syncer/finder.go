package syncer

import (
	"bytes"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type Finder struct {
	hub   component.ICompRequester //for communicate with other service
	chain types.ChainAccessor

	anchorCh chan chain.ChainAnchor
	lScanCh  chan *types.BlockInfo
	fScanCh  chan *message.GetHashByNoRsp

	doneCh chan *FinderResult
	quitCh chan interface{}

	lastAnchor []byte //point last block during lightscan
	ctx        types.SyncContext

	dfltTimeout time.Duration

	debugFullScanOnly bool

	waitGroup *sync.WaitGroup
}

type FinderResult struct {
	ancestor *types.BlockInfo
	err      error
}

var (
	ErrorFinderClosed           = errors.New("sync finder closed")
	ErrorGetSyncAncestorTimeout = errors.New("timeout for GetSyncAncestor")
	ErrorFinderTimeout          = errors.New("Finder timeout")
	dfltTimeOut                 = time.Second * 60
)

func newFinder(ctx *types.SyncContext, hub component.ICompRequester, chain types.ChainAccessor) *Finder {
	finder := &Finder{ctx: *ctx, hub: hub, chain: chain}

	finder.dfltTimeout = dfltTimeOut
	finder.quitCh = make(chan interface{})
	finder.doneCh = make(chan *FinderResult)
	finder.lScanCh = make(chan *types.BlockInfo)
	finder.fScanCh = make(chan *message.GetHashByNoRsp)

	return finder
}

//TODO refactoring: move logic to SyncContext (sync Object)
func (finder *Finder) start() {
	finder.waitGroup = &sync.WaitGroup{}
	finder.waitGroup.Add(2)

	scanFn := func() {
		var ancestor *types.BlockInfo
		var err error

		defer finder.waitGroup.Done()

		logger.Debug().Msg("start to find common ancestor")

		//1. light sync
		//   gather summary of my chain nodes, runTask searching ancestor to remote node
		ancestor, err = finder.lightscan()

		//2. heavy sync
		//	 full binary search in my chain
		if ancestor == nil && err == nil {
			ancestor, err = finder.fullscan()
		}

		finder.doneCh <- &FinderResult{ancestor, err}
		logger.Debug().Msg("stop to find common ancestor")
	}

	go scanFn()

	go func() {
		defer finder.waitGroup.Done()
		for {
			select {
			case result := <-finder.doneCh:
				finder.hub.Tell(message.SyncerSvc, &message.FinderResult{result.ancestor, result.err})
				logger.Info().Msg("finder finished")
				return
			case <-finder.quitCh:
				logger.Info().Msg("finder exited")
				return
			}
		}
	}()
}

func (finder *Finder) stop() {
	if finder == nil {
		return
	}

	logger.Info().Msg("finder stop#1")

	if finder.quitCh != nil {
		logger.Debug().Msg("finder closed quitChannel")

		close(finder.quitCh)
		finder.quitCh = nil
	}

	finder.waitGroup.Wait()

	logger.Info().Msg("finder stop#2")
}

//for debugging
func (finder *Finder) setFullScanOnly(lastAnchor types.BlockNo) {
	finder.debugFullScanOnly = true
	finder.ctx.LastAnchor = lastAnchor
	if finder.ctx.BestNo+1 < lastAnchor {
		panic("set invalid last anchor")
	}
}

func (finder *Finder) GetHashByNoRsp(rsp *message.GetHashByNoRsp) {
	finder.fScanCh <- rsp
}

func (finder *Finder) lightscan() (*types.BlockInfo, error) {
	if finder.debugFullScanOnly {
		return nil, nil
	}

	var ancestor *types.BlockInfo

	anchors, err := finder.getAnchors()
	if err != nil {
		return nil, err
	}

	ancestor, err = finder.getAncestor(anchors)

	if ancestor == nil {
		logger.Debug().Msg("not found ancestor in lightscan")
	} else {
		logger.Debug().Str("hash", enc.ToString(ancestor.Hash)).Uint64("no", ancestor.No).Msg("receive ancestor in lightscan")
	}

	return ancestor, err
}

func (finder *Finder) getAnchors() ([][]byte, error) {
	result, err := finder.hub.RequestFutureResult(message.ChainSvc, &message.GetAnchors{}, finder.dfltTimeout, "Finder/getAnchors")
	if err != nil {
		logger.Error().Err(err).Msg("failed to get anchors")
		return nil, err
	}

	anchors := result.(message.GetAnchorsRsp).Hashes
	if len(anchors) > 0 {
		finder.ctx.LastAnchor = result.(message.GetAnchorsRsp).LastNo
	}

	logger.Info().Str("start", enc.ToString(anchors[0])).Int("count", len(anchors)).Uint64("last", finder.ctx.LastAnchor).Msg("get anchors from chain")

	return anchors, nil
}

func (finder *Finder) getAncestor(anchors [][]byte) (*types.BlockInfo, error) {
	//	send remote Peer
	logger.Debug().Msg("send GetAncestor message to peer")
	finder.hub.Tell(message.P2PSvc, &message.GetSyncAncestor{ToWhom: finder.ctx.PeerID, Hashes: anchors})

	timer := time.NewTimer(finder.dfltTimeout)

	for {
		select {
		case result := <-finder.lScanCh:
			return result, nil
		case <-timer.C:
			logger.Error().Float64("sec", finder.dfltTimeout.Seconds()).Msg("get ancestor response timeout")
			return nil, ErrorGetSyncAncestorTimeout
		case <-finder.quitCh:
			return nil, ErrorFinderClosed
		}
	}
}

//TODO binary search scan
func (finder *Finder) fullscan() (*types.BlockInfo, error) {
	logger.Debug().Msg("finder fullscan")

	ancestor, err := finder.binarySearch(0, finder.ctx.LastAnchor-1)
	if err != nil {
		logger.Error().Err(err).Msg("finder fullscan failed")
		return nil, err
	}

	if ancestor == nil {
		logger.Info().Msg("finder failed to search ancestor in fullscan")
	} else {
		logger.Info().Uint64("no", ancestor.No).Str("hash", enc.ToString(ancestor.Hash)).Msg("finder found ancestor")
	}

	return ancestor, err
}

func (finder *Finder) binarySearch(left uint64, right uint64) (*types.BlockInfo, error) {
	var mid uint64
	var lastMatch *types.BlockInfo
	for left <= right {
		// get median
		mid = (left + right) / 2
		// request hash of median from remote
		logger.Debug().Uint64("left", left).Uint64("right", right).Uint64("mid", mid).Msg("finder scan")

		midHash, err := finder.chain.GetHashByNo(mid)
		if err != nil {
			logger.Error().Uint64("no", mid).Err(err).Msg("finder failed to get local hash")
			return nil, err
		}

		exist, err := finder.hasSameHash(mid, midHash)
		if err != nil {
			logger.Error().Err(err).Msg("finder failed to check remote hash")
			return nil, err
		}

		if exist {
			left = mid + 1

			lastMatch = &types.BlockInfo{Hash: midHash, No: mid}
			logger.Debug().Uint64("mid", mid).Msg("matched")
		} else {
			if mid == 0 {
				break
			} else {
				right = mid - 1
			}
		}
	}

	return lastMatch, nil
}

func (finder *Finder) hasSameHash(no types.BlockNo, localHash []byte) (bool, error) {
	finder.hub.Tell(message.P2PSvc, &message.GetHashByNo{ToWhom: finder.ctx.PeerID, BlockNo: no})

	recvHashRsp := func() (*message.GetHashByNoRsp, error) {
		timer := time.NewTimer(finder.dfltTimeout)

		for {
			select {
			case result := <-finder.fScanCh:
				return result, result.Err
			case <-timer.C:
				logger.Error().Float64("sec", finder.dfltTimeout.Seconds()).Msg("finder get response timeout")
				return nil, ErrorFinderTimeout
			case <-finder.quitCh:
				return nil, ErrorFinderClosed
			}
		}
	}

	rspMsg, err := recvHashRsp()
	if err != nil || rspMsg.BlockHash == nil {
		logger.Error().Err(err).Msg("finder failed to get remote hash")
		return false, err
	}

	if bytes.Equal(localHash, rspMsg.BlockHash) {
		logger.Debug().Uint64("no", no).Msg("exist hash")
		return true, nil
	} else {
		logger.Debug().Uint64("no", no).Msg("not exist hash")
		return false, nil
	}
}
