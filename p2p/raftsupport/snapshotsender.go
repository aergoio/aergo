/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/internal/enc/proto"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	pioutil "github.com/aergoio/etcd/pkg/ioutil"
	"github.com/aergoio/etcd/raft"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/aergoio/etcd/snap"
	core "github.com/libp2p/go-libp2p/core"
)

type snapshotSender struct {
	logger   *log.Logger
	nt       p2pcommon.NetworkTransport
	rAcc     consensus.AergoRaftAccessor
	stopChan chan interface{}

	peer p2pcommon.RemotePeer
}

func newSnapshotSender(logger *log.Logger, nt p2pcommon.NetworkTransport, rAcc consensus.AergoRaftAccessor, peer p2pcommon.RemotePeer) *snapshotSender {
	return &snapshotSender{logger: logger, nt: nt, rAcc: rAcc, stopChan: make(chan interface{}), peer: peer}
}

func (s *snapshotSender) Send(snapMsg *snap.Message) {
	peer := s.peer

	// 1. connect to target peer with snap protocol
	stream, err := s.getSnapshotStream(peer.Meta())
	if err != nil {
		s.logger.Warn().Str(p2putil.LogPeerName, peer.Name()).Err(err).Msg("failed to send snapshot")
		s.rAcc.ReportUnreachable(peer.ID())
		s.rAcc.ReportSnapshot(peer.ID(), raft.SnapshotFailure)
		snapMsg.CloseWithError(errUnreachableMember)
		return
	}

	//
	m := snapMsg.Message
	body := s.createSnapBody(*snapMsg)
	defer body.Close()

	s.logger.Info().Uint64("index", m.Snapshot.Metadata.Index).Str(p2putil.LogPeerName, peer.Name()).Msg("start to send database snapshot")

	// send bytes to target peer
	err = s.pushBMsg(body, stream)
	defer snapMsg.CloseWithError(err)
	if err != nil {
		s.logger.Warn().Uint64("index", m.Snapshot.Metadata.Index).Str(p2putil.LogPeerName, peer.Name()).Err(err).Msg("database snapshot failed to be sent out")

		// errMemberRemoved is a critical error since a removed member should
		// always be stopped. So we use reportCriticalError to report it to errorc.
		//if err == errMemberRemoved {
		//	reportCriticalError(err, s.errorc)
		//}

		// TODO set peer status not healthy
		s.rAcc.ReportUnreachable(peer.ID())
		// report SnapshotFailure to raft state machine. After raft state
		// machine knows about it, it would pause a while and retry sending
		// new snapshot message.
		s.rAcc.ReportSnapshot(peer.ID(), raft.SnapshotFailure)
		//sentFailures.WithLabelValues(to).Inc()
		//snapshotSendFailures.WithLabelValues(to).Inc()
		return
	}

	s.rAcc.ReportSnapshot(peer.ID(), raft.SnapshotFinish)
	s.logger.Info().Uint64("index", m.Snapshot.Metadata.Index).Str(p2putil.LogPeerName, peer.Name()).Msg("database snapshot [index: %d, to: %s] sent out successfully")

	//sentBytes.WithLabelValues(to).Add(float64(merged.TotalSize))
	//snapshotSend.WithLabelValues(to).Inc()
	//snapshotSendSeconds.WithLabelValues(to).Observe(time.Since(start).Seconds())

}
func (s *snapshotSender) pushBMsg(body io.Reader, to io.ReadWriteCloser) error {
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	const (
		WholeTimeLimit  = time.Hour * 24 * 30 // just make indefinitely long term.
		ProcessingLimit = time.Minute * 20    // receiving peer should complete and response within after receiving whole snapshot data.
	)

	wErr := make(chan error, 1)
	rResult := make(chan error, 1)
	t := time.NewTimer(WholeTimeLimit)
	// write snapshot bytes
	go func() {
		_, err := io.Copy(to, body)
		if err != nil {
			wErr <- err
		}
		// renew timer if timer is not expired yet.
		if !t.Stop() {
			<-t.C
		}
		t.Reset(ProcessingLimit)
	}()

	// read response of receiver
	go func() {
		resp, err := readWireHSResp(to)
		if err == nil {
			if resp.Status == types.ResultStatus_OK {
				err = nil
			} else {
				err = fmt.Errorf("error code: %v, msg: %s", resp.Status.String(), resp.Message)
			}
		}
		rResult <- err
	}()

	select {
	case <-s.stopChan:
		return errors.New("stopped")
	case <-t.C:
		return errors.New("timeout")
	case r := <-wErr:
		return r
	case r := <-rResult:
		return r
	}
}

func (s *snapshotSender) getSnapshotStream(meta p2pcommon.PeerMeta) (core.Stream, error) {
	// try connect peer with possible versions
	stream, err := s.nt.GetOrCreateStream(meta, p2pcommon.RaftSnapSubAddr)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func readWireHSResp(rd io.Reader) (resp types.SnapshotResponse, err error) {
	bytebuf := make([]byte, SnapRespHeaderLength)
	readn, err := p2putil.ReadToLen(rd, bytebuf)
	if err != nil {
		return
	}
	if readn != SnapRespHeaderLength {
		err = fmt.Errorf("wrong header length")
		return
	}

	respLen := binary.BigEndian.Uint32(bytebuf)
	bodyBuf := make([]byte, respLen)
	readn, err = p2putil.ReadToLen(rd, bodyBuf)
	if err != nil {
		return
	}
	if readn != int(respLen) {
		err = fmt.Errorf("wrong body length")
		return
	}

	err = proto.Decode(bodyBuf, &resp)
	return
}
func (s *snapshotSender) createSnapBody(merged snap.Message) io.ReadCloser {
	buf := new(bytes.Buffer)
	enc := &RaftMsgEncoder{w: buf}
	// encode raft message
	if err := enc.Encode(&merged.Message); err != nil {
		s.logger.Panic().Err(err).Msg("encode raft message error")
	}

	return &pioutil.ReaderAndCloser{
		Reader: io.MultiReader(buf, merged.ReadCloser),
		Closer: merged.ReadCloser,
	}
}

// RaftMsgEncoder is encode raftpb.Message itt result will be same as rafthttp.messageEncoder
type RaftMsgEncoder struct {
	w io.Writer
}

func (enc *RaftMsgEncoder) Encode(m *raftpb.Message) error {
	if err := binary.Write(enc.w, binary.BigEndian, uint64(m.Size())); err != nil {
		return err
	}
	bytes, err := p2putil.MarshalMessageBody(m)
	if err != nil {
		return err
	}
	_, err = enc.w.Write(bytes)
	return err
}

type RaftMsgDecoder struct {
	r io.Reader
}

var (
	readBytesLimit     uint64 = 512 * 1024 * 1024 // 512 MB
	ErrExceedSizeLimit        = errors.New("raftsupport: error limit exceeded")
)

func (dec *RaftMsgDecoder) Decode() (raftpb.Message, error) {
	return dec.DecodeLimit(readBytesLimit)
}

func (dec *RaftMsgDecoder) DecodeLimit(numBytes uint64) (raftpb.Message, error) {
	var m raftpb.Message
	var l uint64
	if err := binary.Read(dec.r, binary.BigEndian, &l); err != nil {
		return m, err
	}
	if l > numBytes {
		return m, ErrExceedSizeLimit
	}
	buf := make([]byte, int(l))
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return m, err
	}
	return m, m.Unmarshal(buf)
}
