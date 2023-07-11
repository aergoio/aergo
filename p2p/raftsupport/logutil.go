/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"github.com/aergoio/aergo/v2/consensus/impl/raftv2"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/rs/zerolog"
)

type RaftMsgMarshaller struct {
	*raftpb.Message
}

func (m RaftMsgMarshaller) MarshalZerologObject(e *zerolog.Event) {
	e.Str("from", raftv2.EtcdIDToString(m.From)).Str("to", raftv2.EtcdIDToString(m.To)).Str("type", m.Type.String()).Uint64("term", m.Term).Uint64("index", m.Index)
}
