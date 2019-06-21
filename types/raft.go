package types

import (
	"fmt"
	"github.com/aergoio/etcd/raft/raftpb"
)

func (mc *MembershipChange) ToString() string {
	var buf string

	buf = fmt.Sprintf("type:%s,", MembershipChangeType_name[int32(mc.Type)])

	buf = buf + mc.Attr.ToString()
	return buf
}

func (mattr *MemberAttr) ToString() string {
	var buf string

	buf = fmt.Sprintf("{")

	if len(mattr.Name) > 0 {
		buf = buf + fmt.Sprintf("name=%s", mattr.Name)
	}
	if len(mattr.Url) > 0 {
		buf = buf + fmt.Sprintf("name=%s", mattr.Url)
	}
	if len(mattr.PeerID) > 0 {
		buf = buf + fmt.Sprintf("name=%s", PeerID(mattr.PeerID).Pretty())
	}
	if mattr.ID > 0 {
		buf = buf + fmt.Sprintf("name=%d", mattr.ID)
	}

	return buf
}

func (hs *HardStateInfo) ToString() string {
	return fmt.Sprintf("{ term=%d, commit=%d }", hs.Term, hs.Commit)
}

type JsonMemberAttr struct {
	ID     uint64 `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Name   string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Url    string `protobuf:"bytes,3,opt,name=url,proto3" json:"url,omitempty"`
	PeerID string `protobuf:"bytes,4,opt,name=peerID,proto3" json:"peerID,omitempty"`
}

func (mc *JsonMemberAttr) ToMemberAttr() *MemberAttr {
	decodedPeerID, err := IDB58Decode(mc.PeerID)
	if err != nil {
		return nil
	}
	return &MemberAttr{ID: mc.ID, Name: mc.Name, Url: mc.Url, PeerID: []byte(decodedPeerID)}
}

func RaftHardStateToString(hardstate raftpb.HardState) string {
	return fmt.Sprintf("term=%d, vote=%x, commit=%d", hardstate.Term, hardstate.Vote, hardstate.Commit)
}
