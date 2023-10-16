/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"context"
	"encoding/binary"
	"io"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	rtypes "github.com/aergoio/etcd/pkg/types"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/golang/protobuf/proto"
)

const (
	SnapRespHeaderLength = 4
)

// TODO consider the scope of type
type snapshotReceiver struct {
	logger *log.Logger
	pm     p2pcommon.PeerManager
	rAcc   consensus.AergoRaftAccessor
	peer   p2pcommon.RemotePeer
	rwc    io.ReadWriteCloser
}

func newSnapshotReceiver(logger *log.Logger, pm p2pcommon.PeerManager, rAcc consensus.AergoRaftAccessor, peer p2pcommon.RemotePeer, sender io.ReadWriteCloser) *snapshotReceiver {
	return &snapshotReceiver{logger: logger, pm: pm, rAcc: rAcc, peer: peer, rwc: sender}
}

func (s *snapshotReceiver) Receive() {
	resp := &types.SnapshotResponse{Status: types.ResultStatus_OK}
	defer s.sendResp(s.rwc, resp)

	dec := &RaftMsgDecoder{r: s.rwc}
	// let snapshots be very large since they can exceed 512MB for large installations
	m, err := dec.DecodeLimit(uint64(1 << 63))
	from := rtypes.ID(m.From).String()
	if err != nil {
		s.logger.Error().Str(p2putil.LogPeerName, s.peer.Name()).Err(err).Msg("failed to decode raft message")
		resp.Status = types.ResultStatus_INVALID_ARGUMENT
		resp.Message = "malformed message"
		// TODO return error
		//recvFailures.WithLabelValues(rwc.RemoteAddr).Inc()
		//snapshotReceiveFailures.WithLabelValues(from).Inc()
		return
	}

	//receivedBytes.WithLabelValues(from).Add(float64(m.Size()))

	if m.Type != raftpb.MsgSnap {
		s.logger.Error().Str("type", m.Type.String()).Msg("unexpected raft message type on snapshot path")
		resp.Status = types.ResultStatus_INVALID_ARGUMENT
		resp.Message = "invalid message type"

		//http.Error(w, "wrong raft message type", http.StatusBadRequest)
		//snapshotReceiveFailures.WithLabelValues(from).Inc()
		return
	}

	s.logger.Info().Uint64("index", m.Snapshot.Metadata.Index).Str("from", from).Msg("receiving database snapshot")
	// save incoming database snapshot.
	_, err = s.rAcc.SaveFromRemote(s.rwc, m.Snapshot.Metadata.Index, m)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to save KV snapshot")
		resp.Status = types.ResultStatus_INTERNAL

		//http.Error(w, msg, http.StatusInternalServerError)
		//snapshotReceiveFailures.WithLabelValues(from).Inc()
		return
	}
	//receivedBytes.WithLabelValues(from).Add(float64(n))
	s.logger.Info().Str(p2putil.LogPeerName, s.peer.Name()).Uint64("index", m.Snapshot.Metadata.Index).Str("from", from).Msg("received and saved database snapshot successfully")

	if err := s.rAcc.Process(context.TODO(), s.peer.ID(), m); err != nil {
		switch v := err.(type) {
		// Process may return codeError error when doing some
		// additional checks before calling raft.Node.Step.
		case codeError:
			// TODO get resp
			resp.Status = v.Status()
			resp.Message = v.Message()
		default:
			s.logger.Warn().Err(err).Msg("failed to process raft message")
			resp.Status = types.ResultStatus_UNKNOWN
			//http.Error(w, msg, http.StatusInternalServerError)
			//snapshotReceiveFailures.WithLabelValues(from).Inc()
		}
		return
	}
	// Write StatusNoContent header after the message has been processed by
	// raft, which facilitates the client to report MsgSnap status.
	//w.WriteHeader(http.StatusNoContent)

	//snapshotReceive.WithLabelValues(from).Inc()
	//snapshotReceiveSeconds.WithLabelValues(from).Observe(time.Since(start).Seconds())
}

func (s *snapshotReceiver) sendResp(w io.Writer, resp *types.SnapshotResponse) {
	b, err := proto.Marshal(resp)
	if err == nil {
		bytebuf := make([]byte, SnapRespHeaderLength)
		binary.BigEndian.PutUint32(bytebuf, uint32(len(b)))
		w.Write(bytebuf)
		w.Write(b)
	} else {
		s.logger.Info().Err(err).Msg("Failed to write snapshot response")
	}
}

type codeError interface {
	Status() types.ResultStatus
	Message() string
}
