package message

import (
	"github.com/aergoio/aergo/v2/types"
)

const SyncerSvc = "SyncerSvc"

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
