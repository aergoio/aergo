package syncer

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/pkg/errors"
		"time"
	"sync"
)

var (
	ErrorFinderClosed = errors.New("sync finder closed")
)

func newFinder(ctx *types.SyncContext, hub *component.ComponentHub) *Finder {
	finder := &Finder{ctx: *ctx, hub: hub}

	finder.dfltTimeout = time.Second * 100
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
		//   gather summary of my chain nodes, request searching ancestor to remote node
		ancestor, err = finder.lightscan()

		//2. heavy sync
		//	 full binary search in my chain
		if ancestor == nil && err == nil {
			ancestor, err = finder.fullscan()
		}

		finder.doneCh <- &FinderResult{ancestor, err}
	}

	go scanFn()

	go func() {
		defer finder.waitGroup.Done()

		timer := time.NewTimer(finder.dfltTimeout)

		for {
			select {
			case result := <-finder.doneCh:
				finder.hub.Tell(message.SyncerSvc, &message.FinderResult{result.ancestor, result.err})
			case <-timer.C:
				close(finder.quitCh)
			case <-finder.quitCh:
				return
			}
		}
	}()
}

func (finder *Finder) stop() {
	logger.Info().Msg("finder stopped")
	if finder == nil {
		return
	}

	if finder.quitCh != nil {
		logger.Debug().Msg("finder closed quitChannel")

		close(finder.quitCh)
		finder.quitCh = nil
	}

	finder.waitGroup.Wait()
}

func (finder *Finder) lightscan() (*types.BlockInfo, error) {
	var ancestor *types.BlockInfo

	anchors, err := finder.getAnchors()
	if err != nil {
		return nil, err
	}

	ancestor, err = finder.getAncestor(anchors)

	return ancestor, err
}

func (finder *Finder) getAnchors() ([][]byte, error) {
	result, err := finder.hub.RequestFuture(message.ChainSvc, message.GetAnchors{}, finder.dfltTimeout,
		"syncer/finder/lightscan").Result()
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
	finder.hub.Tell(message.P2PSvc, message.GetSyncAncestor{ToWhom: finder.ctx.PeerID, Hashes: anchors})

	// recieve Ancestor response
	for {
		select {
		case result := <-finder.lScanCh:
			return result, nil
		case <-finder.quitCh:
			return nil, ErrorFinderClosed
		}
	}
}

//TODO binary search scan
func (finder *Finder) fullscan() (*types.BlockInfo, error) {
	panic("not implemented")
	return nil, nil
}
