package types

import (
	"encoding/json"
	"fmt"
	"github.com/aergoio/etcd/raft/raftpb"
	"strconv"
)

type ChangeClusterStatus struct {
	State   string        `json:"status"`
	Error   string        `json:"error"`
	Members []*MemberAttr `json:"members"`
}

type EnterpriseTxStatus struct {
	Status   string               `json:"status"`
	Ret      string               `json:"ret"`
	CCStatus *ChangeClusterStatus `json:"change_cluster,omitempty"`
}

func (ccProgress *ConfChangeProgress) ToString() string {
	mbrsJson, err := json.Marshal(ccProgress.GetMembers())
	if err != nil {
		mbrsJson = []byte("")
	}

	return fmt.Sprintf("State=%s, Error=%s, cluster=%s", ConfChangeState_name[int32(ccProgress.State)], ccProgress.Err, string(mbrsJson))
}

func (ccProgress *ConfChangeProgress) ToPrintable() *ChangeClusterStatus {
	return &ChangeClusterStatus{State: ConfChangeState_name[int32(ccProgress.State)], Error: ccProgress.Err, Members: ccProgress.Members}
}

func RaftConfChangeToString(cc *raftpb.ConfChange) string {
	return fmt.Sprintf("requestID=%d, type=%s, nodeid=%d", cc.ID, raftpb.ConfChangeType_name[int32(cc.Type)], cc.NodeID)
}

func (mc *MembershipChange) ToString() string {
	var buf string

	buf = fmt.Sprintf("requestID=%d, type=%s,", mc.GetRequestID(), MembershipChangeType_name[int32(mc.Type)])

	buf = buf + mc.Attr.ToString()
	return buf
}

func (mattr *MemberAttr) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID      string `json:"id,omitempty"`
		Name    string `json:"name,omitempty"`
		Address string `json:"address,omitempty"`
		PeerID  string `json:"peerid,omitempty"`
	}{
		ID:      Uint64ToHexaString(mattr.ID),
		Name:    mattr.Name,
		Address: mattr.Address,
		PeerID:  IDB58Encode(PeerID(mattr.PeerID)),
	})
}

func (mattr *MemberAttr) UnmarshalJSON(data []byte) error {
	var (
		err    error
		peerID PeerID
	)

	aux := &struct {
		ID      string `json:"id,omitempty"`
		Name    string `json:"name,omitempty"`
		Address string `json:"address,omitempty"`
		PeerID  string `json:"peerid,omitempty"`
	}{}

	if err = json.Unmarshal(data, aux); err != nil {
		return err
	}

	if len(aux.ID) > 0 {
		mattr.ID, err = strconv.ParseUint(aux.ID, 16, 64)
		if err != nil {
			return err
		}
	}

	if len(aux.PeerID) > 0 {
		peerID, err = IDB58Decode(aux.PeerID)
		if err != nil {
			return err
		}

		mattr.PeerID = []byte(peerID)
	}
	mattr.Name = aux.Name
	mattr.Address = aux.Address

	return nil
}

func (mattr *MemberAttr) ToString() string {
	data, err := json.Marshal(mattr)
	if err != nil {
		return "{ \"name\": \"error to json\" }"
	}

	return string(data)
}

func (hs *HardStateInfo) ToString() string {
	if hs == nil {
		return fmt.Sprintf("hardstateinfo is nil")
	}
	return fmt.Sprintf("{ term=%d, commit=%d }", hs.Term, hs.Commit)
}

func RaftHardStateToString(hardstate raftpb.HardState) string {
	return fmt.Sprintf("term=%d, vote=%x, commit=%d", hardstate.Term, hardstate.Vote, hardstate.Commit)
}

func RaftEntryToString(entry *raftpb.Entry) string {
	return fmt.Sprintf("term=%d, index=%d, type=%s", entry.Term, entry.Index, raftpb.EntryType_name[int32(entry.Type)])
}

func Uint64ToHexaString(id uint64) string {
	return fmt.Sprintf("%x", id)
}
