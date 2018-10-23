/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"time"
)

// syncWorker do work
// current implementation only allow consuctive range of blocks
type syncWorker struct {
	ttl    time.Duration
	cancel chan interface{}
	retain chan interface{}
	finish chan interface{}

	sm         *syncManager
	peerID     peer.ID
	targetPeer RemotePeer

	requestMsgID  MsgID
	anchors       []message.BlockHash
	currentBest   BlkHash
	stopHash      BlkHash
	currentParent BlkHash
	expectedNo    types.BlockNo
}

var notDefinedYet BlkHash

func init() {
	for i := 0; i < len(notDefinedYet); i++ {
		notDefinedYet[i] = 0
	}
}

func newSyncWorker(sm *syncManager, peer RemotePeer, hashes []message.BlockHash, stopHash message.BlockHash) (*syncWorker) {
	var currentBest, stopH BlkHash
	copy(currentBest[:], hashes[0])
	copy(stopH[:], stopHash)

	sw := &syncWorker{ttl: SyncWorkTTL, cancel: make(chan interface{}), retain: make(chan interface{}), finish: make(chan interface{}),
		sm: sm, peerID: peer.ID(), targetPeer: peer, currentBest: currentBest, anchors: hashes, currentParent: notDefinedYet, stopHash: stopH}

	return sw
}

func (sw *syncWorker) Cancel() {
	sw.cancel <- struct{}{}
}

func (sw *syncWorker) runWorker() {
	defer sw.sm.removeWorker()

	hashes := make([][]byte, len(sw.anchors))
	for i, a := range sw.anchors {
		hashes[i] = message.BlockHash(a)
	}

	// create message data
	req := &types.GetMissingRequest{
		Hashes:   hashes,
		Stophash: sw.stopHash[:]}
	mo := sw.targetPeer.MF().newMsgRequestOrder(true, GetMissingRequest, req)
	sw.targetPeer.sendMessage(mo)
	sw.requestMsgID = mo.GetMsgID()

	timer := time.NewTimer(sw.ttl)
	RUNLOOP:
	for {
		select {
		case <-timer.C:
			sw.sm.logger.Debug().Str(LogPeerID,sw.peerID.Pretty()).Str("base_hash",sw.currentBest.String()).Msg("sync work timeout")
			break RUNLOOP
		case <-sw.cancel:
			sw.sm.logger.Debug().Str(LogPeerID,sw.peerID.Pretty()).Str("base_hash",sw.currentBest.String()).Msg("sync work cancelled")
			break RUNLOOP
		case <-sw.retain:
			if !timer.Stop() {
				// it's already timeout or finished.
				sw.sm.logger.Debug().Str(LogPeerID,sw.peerID.Pretty()).Str("base_hash",sw.currentBest.String()).Msg("failed to retain. already timeout or cancelled")
				break RUNLOOP
			}
			sw.sm.logger.Debug().Str(LogPeerID,sw.peerID.Pretty()).Str("base_hash",sw.currentBest.String()).Msg("retain sync work")
			timer.Reset(sw.ttl)
		case <-sw.finish:
			sw.sm.logger.Debug().Str(LogPeerID,sw.peerID.Pretty()).Str("base_hash",sw.currentBest.String()).Msg("sync work finished")
			// add code if finish is needed.
			break RUNLOOP
		}
	}
}

// putAddBlock check and send blocks to chainservice. Preventing empty blocks is the role of caller
func (sw *syncWorker) putAddBlock(msg Message, blocks []*types.Block, hasNext bool) {
	if msg.OriginalID() != sw.requestMsgID {
		// ignore from other peers
		return
	}
	// TODO fine tune later
	// get first response. it can be previous block if node is forked.
	if sw.currentParent != notDefinedYet {
		// if response is not expected blocks, cancel sync.
		parentHash := blocks[0].Header.PrevBlockHash
		if !bytes.Equal(parentHash, sw.currentBest[:]) {
			// TODO cancel sync
			sw.cancel <- struct{}{}
			return
		}
	}
	// send to chainservice if no actor is found.
	for _, block := range blocks {
		sw.sm.actor.TellRequest(message.ChainSvc, &message.AddBlock{PeerID: sw.peerID, Block: block, Bstate: nil})
	}

	if hasNext {
		sw.retain <- struct{}{}
		lastBlock := blocks[len(blocks)-1]
		copy(sw.currentParent[:], lastBlock.GetHash())
	} else {
		sw.finish <- struct{}{}
	}

}
