package syncer

import (
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type Finder struct {
	hub      *component.ComponentHub //for communicate with other service
	anchorCh chan chain.ChainAnchor
	lScanCh  chan *types.BlockInfo
	fScanCh  chan []byte

	doneCh chan *FinderResult
	quitCh chan interface{}

	lastAnchor []byte //point last block during lightscan
	ctx        types.SyncContext

	dfltTimeout time.Duration

	waitGroup *sync.WaitGroup
}

type FinderResult struct {
	ancestor *types.BlockInfo
	err      error
}

var (
	ErrorFinderClosed           = errors.New("sync finder closed")
	ErrorGetSyncAncestorTimeout = errors.New("timeout for GetSyncAncestor")
	dfltTimeOut                 = time.Second * 180
)

func newFinder(ctx *types.SyncContext, hub *component.ComponentHub) *Finder {
	finder := &Finder{ctx: *ctx, hub: hub}

	finder.dfltTimeout = dfltTimeOut
	finder.quitCh = make(chan interface{})
	finder.doneCh = make(chan *FinderResult)
	finder.lScanCh = make(chan *types.BlockInfo)
	finder.fScanCh = make(chan []byte)

	finder.start()

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
				logger.Info().Msg("Finder finished")
				return
			case <-finder.quitCh:
				logger.Info().Msg("Finder exited")
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

func (finder *Finder) lightscan() (*types.BlockInfo, error) {
	var ancestor *types.BlockInfo

	anchors, err := finder.getAnchors()
	if err != nil {
		return nil, err
	}

	ancestor, err = finder.getAncestor(anchors)

	if ancestor == nil {
		logger.Debug().Msg("Syncer: not found ancestor in lightscan")
	}
	return ancestor, err
}

func (finder *Finder) getAnchors() ([][]byte, error) {
	result, err := finder.hub.RequestFuture(message.ChainSvc, &message.GetAnchors{}, finder.dfltTimeout, "Finder/getAnchors").Result()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get anchors")
		return nil, err
	}

	anchors := result.(message.GetAnchorsRsp).Hashes
	if len(anchors) > 0 {
		finder.ctx.LastAnchor = anchors[len(anchors)-1]
	}

	return anchors, nil
}

func (finder *Finder) getAncestor(anchors [][]byte) (*types.BlockInfo, error) {
	//	send remote Peer
	finder.hub.Tell(message.P2PSvc, &message.GetSyncAncestor{ToWhom: finder.ctx.PeerID, Hashes: anchors})

	timer := time.NewTimer(finder.dfltTimeout)

	// recieve Ancestor response
	for {
		select {
		case result := <-finder.lScanCh:
			return result, nil
		case <-timer.C:
			return nil, ErrorGetSyncAncestorTimeout
		case <-finder.quitCh:
			return nil, ErrorFinderClosed
		}
	}
}

//TODO binary search scan
func (finder *Finder) fullscan() (*types.BlockInfo, error) {
	logger.Debug().Msg("Finder fullscan")

	panic("not implemented")
	return nil, nil
}
