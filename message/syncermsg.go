package message

import (
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

const SyncerSvc = "SyncerSvc"

//Syncer
type SyncRequest struct {
	TargetNo types.BlockNo
	PeerID   peer.ID
}

//Finder
type FindAncestor struct {
	Ctx *types.SyncContext
}

type FindAncestorRsp struct {
	Ancestor *types.BlockInfo
	Err      error
}

//HashDownloader
type StartFetch struct{}
