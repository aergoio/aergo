package types

import (
	"fmt"
	"github.com/aergoio/etcd/raft/raftpb"
)

type ConfChangeProgressPrintable struct {
	State string
	Error string
}

func (ccProgress *ConfChangeProgress) ToString() string {
	return fmt.Sprintf("State=%s, Error=%s", ConfChangeState_name[int32(ccProgress.State)], ccProgress.Err)
}

func (ccProgress *ConfChangeProgress) ToPrintable() *ConfChangeProgressPrintable {
	return &ConfChangeProgressPrintable{State: ConfChangeState_name[int32(ccProgress.State)], Error: ccProgress.Err}
}

func RaftConfChangeToString(cc *raftpb.ConfChange) string {
	return fmt.Sprintf("request id=%d, type=%s, nodeid=%d", cc.ID, raftpb.ConfChangeType_name[int32(cc.Type)], cc.NodeID)
}

func (mc *MembershipChange) ToString() string {
	var buf string

	buf = fmt.Sprintf("requestID:%d, type:%s,", mc.GetRequestID(), MembershipChangeType_name[int32(mc.Type)])

	buf = buf + mc.Attr.ToString()
	return buf
}

func (mattr *MemberAttr) ToString() string {
	var buf string

	buf = fmt.Sprintf("{")

	if len(mattr.Name) > 0 {
		buf = buf + fmt.Sprintf("name=%s, ", mattr.Name)
	}
	if len(mattr.Url) > 0 {
		buf = buf + fmt.Sprintf("url=%s, ", mattr.Url)
	}
	if len(mattr.PeerID) > 0 {
		buf = buf + fmt.Sprintf("peerID=%s, ", PeerID(mattr.PeerID).Pretty())
	}
	if mattr.ID > 0 {
		buf = buf + fmt.Sprintf("memberID=%d", mattr.ID)
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

func RaftEntryToString(entry *raftpb.Entry) string {
	return fmt.Sprintf("term=%d, index=%d, type=%s", entry.Term, entry.Index, raftpb.EntryType_name[int32(entry.Type)])
}
