// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raft

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"

	"github.com/aergoio/etcd/raft/raftpb"
)

type cluster struct {
	peers       []string
	commitC     []<-chan *string
	errorC      []<-chan error
	proposeC    []chan string
	confChangeC []chan raftpb.ConfChange

	rs []*raftServer
}

var dataDirBase = "./rafttest"

// newCluster creates a cluster of n nodes
func newCluster(n int, delayPromote bool) *cluster {
	peers := make([]string, n)
	for i := range peers {
		peers[i] = fmt.Sprintf("http://127.0.0.1:%d", 10000+i)
	}

	clus := &cluster{
		peers:       peers,
		commitC:     make([]<-chan *string, len(peers)),
		errorC:      make([]<-chan error, len(peers)),
		proposeC:    make([]chan string, len(peers)),
		confChangeC: make([]chan raftpb.ConfChange, len(peers)),
		rs:          make([]*raftServer, len(peers)),
	}

	os.RemoveAll(dataDirBase)

	for i := range clus.peers {
		waldir := fmt.Sprintf("%s/%d/wal", dataDirBase, i+1)
		snapdir := fmt.Sprintf("%s/%d/snap", dataDirBase, i+1)

		clus.proposeC[i] = make(chan string, 1)
		clus.confChangeC[i] = make(chan raftpb.ConfChange, 1)

		rs := newRaftServer(uint64(i+1), clus.peers, false, waldir, snapdir, "", "", nil, clus.proposeC[i], clus.confChangeC[i], delayPromote)
		clus.rs[i] = rs
		clus.commitC[i] = rs.commitC
		clus.errorC[i] = rs.errorC
	}

	return clus
}

// sinkReplay reads all commits in each node's local log.
func (clus *cluster) sinkReplay() {
	for i := range clus.peers {
		for s := range clus.commitC[i] {
			if s == nil {
				break
			}
		}
	}
}

// Close closes all cluster nodes and returns an error if any failed.
func (clus *cluster) Close() (err error) {
	for i := range clus.peers {
		close(clus.proposeC[i])
		for range clus.commitC[i] {
			// drain pending commits
		}
		// wait for channel to close
		if erri := <-clus.errorC[i]; erri != nil {
			err = erri
		}
	}

	os.RemoveAll(dataDirBase)

	return err
}

func (clus *cluster) closeNoErrors(t *testing.T) {
	if err := clus.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestProposeOnCommit starts three nodes and feeds commits back into the proposal
// channel. The intent is to ensure blocking on a proposal won't block raft progress.
func TestProposeOnCommit(t *testing.T) {
	clus := newCluster(3, false)
	defer clus.closeNoErrors(t)

	//wait creation of all Raft nodes
	clus.sinkReplay()

	donec := make(chan struct{})
	for i := range clus.peers {
		// feedback for "n" committed entries, then update donec
		go func(i int, pC chan<- string, cC <-chan *string, eC <-chan error) {
			for n := 0; n < 100; n++ {
				s, ok := <-cC
				if !ok {
					pC = nil
				}
				//t.Logf("raft node [%d][%d] commit", i, n)
				select {
				case pC <- *s:
					continue
				case err := <-eC:
					t.Fatalf("eC message (%v)", err)
				}
			}
			t.Logf("raft node [%d] done", i)
			donec <- struct{}{}
			for range cC {
				// acknowledge the commits from other nodes so
				// raft continues to make progress
			}
		}(i, clus.proposeC[i], clus.commitC[i], clus.errorC[i])

		// one message feedback per node
		go func(i int) { clus.proposeC[i] <- "foo" }(i)
	}

	for range clus.peers {
		<-donec
	}
}

// TestCloseProposerBeforeReplay tests closing the producer before raft starts.
func TestCloseProposerBeforeReplay(t *testing.T) {
	clus := newCluster(1, false)
	// close before replay so raft never starts
	defer clus.closeNoErrors(t)
}

// TestCloseProposerInflight tests closing the producer while
// committed messages are being published to the client.
func TestCloseProposerInflight(t *testing.T) {
	clus := newCluster(1, false)
	defer clus.closeNoErrors(t)

	clus.sinkReplay()

	// some inflight ops
	go func() {
		clus.proposeC[0] <- "foo"
		clus.proposeC[0] <- "bar"
	}()

	// wait for one message
	if c, ok := <-clus.commitC[0]; *c != "foo" || !ok {
		t.Fatalf("Commit failed")
	}
}

func TestRaftDelayPromotable(t *testing.T) {
	clus := newCluster(3, true)

	defer clus.closeNoErrors(t)

	//wait creation of all Raft nodes
	clus.sinkReplay()

	//2개 node가 promotable 되면 leader 결정
	t.Log("replay ready")

	checkHasLeader := func(has bool) {
		res := false
		for _, raftserver := range clus.rs {
			if raftserver.IsLeader() {
				res = true
			}
		}

		assert.Equal(t, has, res)
	}

	time.Sleep(time.Second * 1)
	checkHasLeader(false)

	clus.rs[0].SetPromotable(true)

	time.Sleep(time.Second * 5)
	checkHasLeader(true)
}

/*
func TestRaftIsLeader(t *testing.T) {
	clus := newCluster(1)
	defer clus.closeNoErrors(t)

	clus.sinkReplay()

	// check who is leader

}*/
