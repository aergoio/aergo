package syncer

import (
	"bytes"
	"sync"
	"time"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/pkg/errors"
)

type Finder struct {
	compRequester component.IComponentRequester //for communicate with other service
	chain         types.ChainAccessor

	anchorCh chan chain.ChainAnchor
	lScanCh  chan *types.BlockInfo
	fScanCh  chan *message.GetHashByNoRsp

	quitCh chan interface{}

	lastAnchor []byte //point last block during lightscan
	ctx        types.SyncContext

	dfltTimeout time.Duration

	cfg *SyncerConfig

	isRunning bool
	waitGroup *sync.WaitGroup
}

type FinderResult struct {
	ancestor *types.BlockInfo
	err      error
}

var (
	ErrFinderQuit               = errors.New("sync finder quit")
	ErrorGetSyncAncestorTimeout = errors.New("timeout for GetSyncAncestor")
	ErrFinderTimeout            = errors.New("Finder timeout")
	ErrAlreadySyncDone          = errors.New("Already sync done")
)

func newFinder(ctx *types.SyncContext, compRequester component.IComponentRequester, chain types.ChainAccessor, cfg *SyncerConfig) *Finder {
	finder := &Finder{ctx: *ctx, compRequester: compRequester, chain: chain, cfg: cfg}

	finder.dfltTimeout = cfg.fetchTimeOut
	finder.quitCh = make(chan interface{})
	finder.lScanCh = make(chan *types.BlockInfo)
	finder.lScanCh = make(chan *types.BlockInfo)
	finder.fScanCh = make(chan *message.GetHashByNoRsp)

	return finder
}

// TODO refactoring: move logic to SyncContext (sync Object)
func (finder *Finder) start() {
	finder.waitGroup = &sync.WaitGroup{}
	finder.waitGroup.Add(1)
	finder.isRunning = true

	run := func() {
		var ancestor *types.BlockInfo
		var err error

		defer RecoverSyncer(NameFinder, finder.GetSeq(), finder.compRequester, func() { finder.waitGroup.Done() })

		logger.Debug().Msg("start to find common ancestor")

		//1. light sync
		//   gather summary of my chain nodes, runTask searching ancestor to remote node
		ancestor, err = finder.lightscan()

		//2. heavy sync
		//	 full binary search in my chain
		if ancestor == nil && err == nil {
			ancestor, err = finder.fullscan()
		}

		if err != nil {
			logger.Debug().Msg("quit finder")
			stopSyncer(finder.compRequester, finder.GetSeq(), NameFinder, err)
			return
		}

		finder.compRequester.TellTo(message.SyncerSvc, &message.FinderResult{Seq: finder.GetSeq(), Ancestor: ancestor, Err: nil})
		logger.Info().Msg("stopped finder successfully")
	}

	go run()
}

func (finder *Finder) stop() {
	if finder == nil {
		return
	}

	logger.Info().Msg("finder stop#1")

	if finder.isRunning {
		logger.Debug().Msg("finder closed quitChannel")

		close(finder.quitCh)
		finder.isRunning = false
	}

	finder.waitGroup.Wait()

	logger.Info().Msg("finder stop#2")
}

func (finder *Finder) GetSeq() uint64 {
	return finder.ctx.Seq
}

func (finder *Finder) GetHashByNoRsp(rsp *message.GetHashByNoRsp) {
	finder.fScanCh <- rsp
}

func (finder *Finder) lightscan() (*types.BlockInfo, error) {
	if finder.cfg.useFullScanOnly {
		finder.ctx.LastAnchor = finder.ctx.BestNo + 1
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
		logger.Info().Str("hash", enc.ToString(ancestor.Hash)).Uint64("no", ancestor.No).Msg("find ancestor in lightscan")

		if ancestor.No >= finder.ctx.TargetNo {
			logger.Info().Msg("already synchronized")
			return nil, ErrAlreadySyncDone
		}
	}

	return ancestor, err
}

func (finder *Finder) getAnchors() ([][]byte, error) {
	result, err := finder.compRequester.RequestToFutureResult(message.ChainSvc, &message.GetAnchors{Seq: finder.GetSeq()}, finder.dfltTimeout, "Finder/getAnchors")
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
	logger.Debug().Stringer("peer", types.LogPeerShort(finder.ctx.PeerID)).Msg("send GetAncestor message to peer")
	finder.compRequester.TellTo(message.P2PSvc, &message.GetSyncAncestor{Seq: finder.GetSeq(), ToWhom: finder.ctx.PeerID, Hashes: anchors})

	timer := time.NewTimer(finder.dfltTimeout)

	for {
		select {
		case result := <-finder.lScanCh:
			//valid response
			if result == nil || result.No >= finder.ctx.LastAnchor {
				return result, nil
			}
		case <-timer.C:
			logger.Error().Float64("sec", finder.dfltTimeout.Seconds()).Msg("get ancestor response timeout")
			return nil, ErrorGetSyncAncestorTimeout
		case <-finder.quitCh:
			return nil, ErrFinderQuit
		}
	}
}

// TODO binary search scan
func (finder *Finder) fullscan() (*types.BlockInfo, error) {
	logger.Debug().Msg("finder fullscan")

	ancestor, err := finder.binarySearch(0, finder.ctx.LastAnchor-1)
	if err != nil {
		logger.Error().Err(err).Msg("finder fullscan failed")
		return nil, err
	}

	if ancestor == nil {
		logger.Info().Msg("failed to search ancestor in fullscan")
	} else {
		logger.Info().Uint64("no", ancestor.No).Str("hash", enc.ToString(ancestor.Hash)).Msg("find ancestor in fullscan")
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
	finder.compRequester.TellTo(message.P2PSvc, &message.GetHashByNo{Seq: finder.GetSeq(), ToWhom: finder.ctx.PeerID, BlockNo: no})

	recvHashRsp := func() (*message.GetHashByNoRsp, error) {
		timer := time.NewTimer(finder.dfltTimeout)

		for {
			select {
			case result := <-finder.fScanCh:
				return result, result.Err
			case <-timer.C:
				logger.Error().Float64("sec", finder.dfltTimeout.Seconds()).Msg("finder get response timeout")
				return nil, ErrFinderTimeout
			case <-finder.quitCh:
				return nil, ErrFinderQuit
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
