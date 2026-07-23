package raftv2

import (
	"errors"
	"testing"

	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/types"
	"github.com/aergoio/etcd/raft/raftpb"
)

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
