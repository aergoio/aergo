package raft

import (
	"fmt"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/p2p"
	"github.com/libp2p/go-libp2p-peer"
	"sync"
)

// raft cluster membership
// copy from dpos/bp
// TODO refactoring
// Cluster represents a cluster of block producers.
type Index uint16

type Cluster struct {
	sync.Mutex

	ID     uint64
	Size   uint16
	Member map[uint64]*blockProducer
	BPUrls []string //for raft server

	cdb consensus.ChainDB
}

type blockProducer struct {
	raftID uint64
	url    string
	peerID peer.ID
}

func (bp *blockProducer) isDifferent(x *blockProducer) bool {
	if bp.raftID == x.raftID || bp.url == x.url || bp.peerID == x.peerID {
		return false
	}

	return true
}

func (cc *Cluster) addMember(id uint64, url string, p2pID peer.ID) error {
	//check unique
	bp := &blockProducer{raftID: id, url: url, peerID: p2pID}

	if cc.ID == id {
		bp.peerID = p2p.NodeID()
	}

	for prevID, prevBP := range cc.Member {
		if prevID == id {
			return ErrDupBP
		}

		if !prevBP.isDifferent(bp) {
			return ErrDupBP
		}
	}

	cc.Member[id] = bp

	cc.BPUrls[id-1] = bp.url

	return nil
}

func (cc *Cluster) toString() string {
	var buf string

	buf = fmt.Sprintf("raft cluster configure: total=%d, curid=%d, bps=[", cc.ID, cc.Size)
	for _, bp := range cc.Member {
		bpbuf := fmt.Sprintf("{ id:%d, url:%s, peerID:%s }", bp.raftID, bp.url, bp.peerID)
		buf += bpbuf
	}
	fmt.Sprintf("]")

	return buf
}

// current config
