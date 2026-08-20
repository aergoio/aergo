package message

import (
	"errors"

	"github.com/aergoio/aergo/v2/types"
)

const SyncerSvc = "SyncerSvc"

// ErrSyncerBusy is returned (via NotifyC) when SyncStart is requested while a
// sync is already in progress. Callers should wait/retry rather than treating
// this as a fatal sync failure.
var ErrSyncerBusy = errors.New("syncer is already running")

// Syncer
type SyncStart struct {
	PeerID   types.PeerID
	TargetNo types.BlockNo
	NotifyC  chan error
}

type FinderResult struct {
	Seq      uint64
	Ancestor *types.BlockInfo
	Err      error
}

// HashDownloader
type SyncStop struct {
	Seq     uint64
	FromWho string
	Err     error
}

type CloseFetcher struct {
	Seq     uint64
	FromWho string
}
