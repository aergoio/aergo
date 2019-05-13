package types

import (
	"fmt"
	"github.com/libp2p/go-libp2p-peer"
)

func (mc *MembershipChange) ToString() string {
	var buf string

	buf = fmt.Sprintf("type:%s,", MembershipChangeType_name[int32(mc.Type)])

	buf = buf + mc.Attr.ToString()
	return buf
}

func (mattr *MemberAttr) ToString() string {
	var buf string

	buf = fmt.Sprintf("{ name=%s, url=%s, peerid=%s, id=%x }", mattr.Name, mattr.Url, peer.ID(mattr.PeerID).Pretty(), mattr.ID)
	return buf
}
