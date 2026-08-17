package raftv2

import (
	"errors"
	"testing"

	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/types"
	raftlib "github.com/aergoio/etcd/raft"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

func TestRaftConfigEnablesPreVote(t *testing.T) {
	if config := makeConfig(1, raftlib.NewMemoryStorage()); !config.PreVote {
		t.Fatal("raft pre-vote must be enabled to prevent disruptive term inflation after partitions")
	}
}

func TestRaftMessagePeerBinding(t *testing.T) {
	senderPeerID := types.RandomPeerID()
	localPeerID := types.RandomPeerID()

	cluster := NewCluster([]byte("chain"), nil, "local", localPeerID, 0, nil)
	sender := &consensus.Member{MemberAttr: types.MemberAttr{
		ID:      11,
		Name:    "sender",
		Address: "/ip4/127.0.0.1/tcp/11001",
		PeerID:  []byte(senderPeerID),
	}}
	local := &consensus.Member{MemberAttr: types.MemberAttr{
		ID:      22,
		Name:    "local",
		Address: "/ip4/127.0.0.1/tcp/11002",
		PeerID:  []byte(localPeerID),
	}}
	cluster.Members().add(sender)
	cluster.Members().add(local)
	cluster.SetNodeID(local.ID)

	accessor := &raftHttpWrapper{raftServer: &raftServer{cluster: cluster}}
	valid := raftpb.Message{From: sender.ID, To: local.ID, Type: raftpb.MsgHeartbeat}

	tests := []struct {
		name   string
		peerID types.PeerID
		msg    raftpb.Message
		want   error
	}{
		{name: "valid member", peerID: senderPeerID, msg: valid},
		{name: "unknown peer", peerID: types.RandomPeerID(), msg: valid, want: ErrUnknownRaftPeer},
		{name: "spoofed sender", peerID: senderPeerID, msg: raftpb.Message{From: local.ID, To: local.ID, Type: raftpb.MsgHeartbeat}, want: ErrRaftSenderMismatch},
		{name: "wrong target", peerID: senderPeerID, msg: raftpb.Message{From: sender.ID, To: sender.ID, Type: raftpb.MsgHeartbeat}, want: ErrRaftTargetMismatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := accessor.ValidateMessage(tt.peerID, tt.msg)
			if !errors.Is(err, tt.want) {
				t.Fatalf("ValidateMessage() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestPublishFirstBlockDoesNotDereferenceEmptyProgress(t *testing.T) {
	block := types.NewBlock(types.EmptyBlockHeaderInfo, nil, nil, nil, nil, nil)
	data, err := marshalEntryData(block)
	if err != nil {
		t.Fatal(err)
	}

	server := &raftServer{
		commitC: make(chan *commitEntry, 1),
		stopc:   make(chan struct{}),
	}
	if ok := server.publishEntries([]raftpb.Entry{{
		Index: 1,
		Term:  1,
		Type:  raftpb.EntryNormal,
		Data:  data,
	}}); !ok {
		t.Fatal("publishEntries() unexpectedly stopped")
	}
}

func TestMalformedConfChangeReturnsErrorWithoutPanic(t *testing.T) {
	server := &raftServer{cluster: NewCluster([]byte("chain"), nil, "local", types.RandomPeerID(), 0, nil)}

	if _, _, err := server.ValidateConfChangeEntry(&raftpb.Entry{
		Index: 1,
		Term:  1,
		Type:  raftpb.EntryConfChange,
		Data:  []byte{0xff},
	}); err == nil {
		t.Fatal("ValidateConfChangeEntry() accepted malformed protobuf")
	}

	data, err := (&raftpb.ConfChange{NodeID: 1, Type: raftpb.ConfChangeAddNode}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := server.ValidateConfChangeEntry(&raftpb.Entry{
		Index: 1,
		Term:  1,
		Type:  raftpb.EntryConfChange,
		Data:  data,
	}); !errors.Is(err, ErrCCMemberIsNil) {
		t.Fatalf("ValidateConfChangeEntry() error = %v, want %v", err, ErrCCMemberIsNil)
	}
}

func TestRaftBlockSignerMustBeCurrentMember(t *testing.T) {
	memberKey, memberPub, err := crypto.GenerateKeyPair(crypto.Secp256k1, -1)
	if err != nil {
		t.Fatal(err)
	}
	memberPeerID, err := peer.IDFromPublicKey(memberPub)
	if err != nil {
		t.Fatal(err)
	}
	cluster := NewCluster([]byte("chain"), nil, "local", memberPeerID, 0, nil)
	cluster.Members().add(&consensus.Member{MemberAttr: types.MemberAttr{
		ID:      1,
		Name:    "member",
		Address: "/ip4/127.0.0.1/tcp/11001",
		PeerID:  []byte(memberPeerID),
	}})
	factory := &BlockFactory{bpc: cluster}

	authorized := types.NewBlock(types.EmptyBlockHeaderInfo, nil, nil, nil, nil, nil)
	if err := authorized.Sign(memberKey); err != nil {
		t.Fatal(err)
	}
	if err := factory.IsBlockValid(authorized, nil); err != nil {
		t.Fatalf("IsBlockValid() rejected current member: %v", err)
	}

	otherKey, _, err := crypto.GenerateKeyPair(crypto.Secp256k1, -1)
	if err != nil {
		t.Fatal(err)
	}
	unauthorized := types.NewBlock(types.EmptyBlockHeaderInfo, nil, nil, nil, nil, nil)
	if err := unauthorized.Sign(otherKey); err != nil {
		t.Fatal(err)
	}
	if err := factory.IsBlockValid(unauthorized, nil); err == nil {
		t.Fatal("IsBlockValid() accepted signer outside current membership")
	}
}

func TestSnapshotMembershipMustMatchConfState(t *testing.T) {
	member := &consensus.Member{MemberAttr: types.MemberAttr{
		ID:      1,
		Name:    "member",
		Address: "/ip4/127.0.0.1/tcp/11001",
		PeerID:  []byte(types.RandomPeerID()),
	}}
	block := types.NewBlock(types.EmptyBlockHeaderInfo, nil, nil, nil, nil, nil)
	data, err := consensus.NewSnapshotData([]*consensus.Member{member}, nil, block).Encode()
	if err != nil {
		t.Fatal(err)
	}
	snapshot := &raftpb.Snapshot{
		Data: data,
		Metadata: raftpb.SnapshotMetadata{
			Index:     1,
			Term:      1,
			ConfState: raftpb.ConfState{Nodes: []uint64{member.ID}},
		},
	}
	if err := validateSnapshotConsistency(snapshot); err != nil {
		t.Fatalf("validateSnapshotConsistency() rejected valid snapshot: %v", err)
	}

	snapshot.Metadata.ConfState.Nodes[0] = 2
	if err := validateSnapshotConsistency(snapshot); !errors.Is(err, ErrSnapshotMembershipMismatch) {
		t.Fatalf("validateSnapshotConsistency() error = %v, want %v", err, ErrSnapshotMembershipMismatch)
	}
}
