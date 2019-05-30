package types

import (
	"fmt"
)

func (mc *MembershipChange) ToString() string {
	var buf string

	buf = fmt.Sprintf("type:%s,", MembershipChangeType_name[int32(mc.Type)])

	buf = buf + mc.Attr.ToString()
	return buf
}

func (mattr *MemberAttr) ToString() string {
	var buf string

	buf = fmt.Sprintf("{ name=%s, url=%s, peerid=%s, id=%x }", mattr.Name, mattr.Url, PeerID(mattr.PeerID).Pretty(), mattr.ID)
	return buf
}

func (hs *HardStateInfo) ToString() string {
	return fmt.Sprintf("{ term=%d, commit=%d }", hs.Term, hs.Commit)
}
