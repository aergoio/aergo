package message

import (
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

const SyncerSvc = "SyncerSvc"

//Syncer
type SyncStart struct {
	PeerID   peer.ID
	TargetNo types.BlockNo
}

type FinderResult struct {
	Ancestor *types.BlockInfo
	Err      error
}

//HashDownloader
type StartFetch struct{}

type SyncStop struct {
	FromWho string
	Err     error
}

type CloseFetcher struct {
	FromWho string
}
