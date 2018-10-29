/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
)

const aergoP2PSub protocol.ID = "/aergop2p/0.2"

// PeerHandshaker works to handshake to just connected peer, it detect chain networks
// and protocol versions, and then select InnerHandshaker for that protocol version.
type PeerHandshaker struct {
	pm        PeerManager
	actorServ ActorService
	logger    *log.Logger
	peerID    peer.ID

	remoteStatus *types.Status
}

// InnerHandshaker do handshake work and msgreadwriter for a protocol version
type innerHandshaker interface {
	doForOutbound() (*types.Status, error)
	doForInbound() (*types.Status, error)
	GetMsgRW() MsgReadWriter
}

type hsResult struct {
	rw        MsgReadWriter
	statusMsg *types.Status
	err       error
}

func newHandshaker(pm PeerManager, actor ActorService, log *log.Logger, peerID peer.ID) *PeerHandshaker {
	return &PeerHandshaker{pm: pm, actorServ: actor, logger: log, peerID: peerID}
}

func (h *PeerHandshaker) handshakeOutboundPeerTimeout(r io.Reader, w io.Writer, ttl time.Duration) (MsgReadWriter, *types.Status, error) {
	ret, err := runFuncTimeout(func(doneChan chan<- interface{}) {
		rw, statusMsg, err := h.handshakeOutboundPeer(r, w)
		doneChan <- &hsResult{rw: rw, statusMsg: statusMsg, err: err}
	}, ttl)
	if err != nil {
		return nil, nil, err
	}
	return ret.(*hsResult).rw, ret.(*hsResult).statusMsg, ret.(*hsResult).err
}

func (h *PeerHandshaker) handshakeInboundPeerTimeout(r io.Reader, w io.Writer, ttl time.Duration) (MsgReadWriter, *types.Status, error) {
	ret, err := runFuncTimeout(func(doneChan chan<- interface{}) {
		rw, statusMsg, err := h.handshakeInboundPeer(r, w)
		doneChan <- &hsResult{rw: rw, statusMsg: statusMsg, err: err}
	}, ttl)
	if err != nil {
		return nil, nil, err
	}
	return ret.(*hsResult).rw, ret.(*hsResult).statusMsg, ret.(*hsResult).err
}

type targetFunc func(chan<- interface{})
type timeoutErr error

func runFuncTimeout(m targetFunc, ttl time.Duration) (interface{}, error) {
	done := make(chan interface{})
	go m(done)
	select {
	case hsResult := <-done:
		return hsResult, nil
	case <-time.NewTimer(ttl).C:
		return nil, fmt.Errorf("timeout").(timeoutErr)
	}
}

func (h *PeerHandshaker) handshakeOutboundPeer(r io.Reader, w io.Writer) (MsgReadWriter, *types.Status, error) {
	bufReader, bufWriter := bufio.NewReader(r), bufio.NewWriter(w)
	// send initial hsmessage
	hsHeader := HSHeader{Magic: MAGICTest, Version: P2PVersion030}
	sent, err := bufWriter.Write(hsHeader.Marshal())
	if err != nil {
		return nil, nil, err
	}
	if sent != len(hsHeader.Marshal()) {
		return nil, nil, fmt.Errorf("transport error")
	}
	// continue to handshake with innerHandshaker
	innerHS, err := h.selectProtocolVersion(hsHeader, bufReader, bufWriter)
	if err != nil {
		return nil, nil, err
	}
	status, err := innerHS.doForOutbound()
	h.remoteStatus = status
	return innerHS.GetMsgRW(), status, err
}

func (h *PeerHandshaker) handshakeInboundPeer(r io.Reader, w io.Writer) (MsgReadWriter, *types.Status, error) {
	var hsHeader HSHeader
	bufReader, bufWriter := bufio.NewReader(r), bufio.NewWriter(w)
	// wait initial hsmessage
	headBuf := make([]byte, 8)
	read, err := h.readToLen(bufReader, headBuf, 8)
	if err != nil {
		return nil, nil, err
	}
	if read != 8 {
		return nil, nil, fmt.Errorf("transport error")
	}
	hsHeader.Unmarshal(headBuf)

	// continue to handshake with innerHandshaker
	innerHS, err := h.selectProtocolVersion(hsHeader, bufReader, bufWriter)
	if err != nil {
		return nil, nil, err
	}
	status, err := innerHS.doForInbound()
	// send hsresponse
	h.remoteStatus = status
	return innerHS.GetMsgRW(), status, err
}

func (h *PeerHandshaker) readToLen(rd io.Reader, bf []byte, max int) (int, error) {
	remain := max
	offset := 0
	for remain > 0 {
		read, err := rd.Read(bf[offset:])
		if err != nil {
			return offset, err
		}
		remain -= read
		offset += read
	}
	return offset, nil
}

// doPostHandshake is additional work after peer is added.
func (h *PeerHandshaker) doInitialSync() {

	if chain.UseFastSyncer {
		h.actorServ.SendRequest(message.SyncerSvc, &message.SyncStart{PeerID: h.peerID, TargetNo: h.remoteStatus.BestHeight})
	} else {
		// sync block infos
		h.actorServ.SendRequest(message.ChainSvc, &message.SyncBlockState{PeerID: h.peerID, BlockNo: h.remoteStatus.BestHeight, BlockHash: h.remoteStatus.BestBlockHash})
	}

	// sync mempool tx infos
	// TODO add tx handling
}

func createStatusMsg(pm PeerManager, actorServ ActorService) (*types.Status, error) {
	// find my best block
	bestBlock, err := actorServ.GetChainAccessor().GetBestBlock()
	if err != nil {
		return nil, err
	}
	selfAddr := pm.SelfMeta().ToPeerAddress()
	// create message data
	statusMsg := &types.Status{
		Sender:        &selfAddr,
		BestBlockHash: bestBlock.BlockHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	return statusMsg, nil
}

func (h *PeerHandshaker) selectProtocolVersion(head HSHeader, r *bufio.Reader, w *bufio.Writer) (innerHandshaker, error) {
	switch head.Version {
	case P2PVersion030:
		v030 := newV030StateHS(h.pm, h.actorServ, h.logger, h.peerID, r, w)
		return v030, nil
	default:
		return nil, fmt.Errorf("not supported version")
	}
}

func (h *PeerHandshaker) checkProtocolVersion(versionStr string) error {
	// TODO modify interface and put check code here
	return nil
}

type HSHeader struct {
	Magic   uint32
	Version uint32
}

func (h HSHeader) Marshal() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b, h.Magic)
	binary.BigEndian.PutUint32(b[4:], h.Version)
	return b
}

func (h *HSHeader) Unmarshal(b []byte) {
	h.Magic = binary.BigEndian.Uint32(b)
	h.Version = binary.BigEndian.Uint32(b[4:])
}
