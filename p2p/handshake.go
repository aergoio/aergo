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
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

// HSHandlerFactory is creator of HSHandler
type HSHandlerFactory interface {
	CreateHSHandler(outbound bool, pm PeerManager, actor ActorService, log *log.Logger, pid peer.ID) HSHandler
}

// HSHandler will do handshake with remote peer
type HSHandler interface {
	// Handle peer handshake till ttl, and return msgrw for this connection, and status of remote peer.
	Handle(r io.Reader, w io.Writer, ttl time.Duration) (MsgReadWriter, *types.Status, error)
}

type InboundHSHandler struct {
	*PeerHandshaker
}

func (ih *InboundHSHandler) Handle(r io.Reader, w io.Writer, ttl time.Duration) (MsgReadWriter, *types.Status, error) {
	return ih.handshakeInboundPeerTimeout(r, w, ttl)
}

type OutboundHSHandler struct {
	*PeerHandshaker
}

func (oh *OutboundHSHandler) Handle(r io.Reader, w io.Writer, ttl time.Duration) (MsgReadWriter, *types.Status, error) {
	return oh.handshakeOutboundPeerTimeout(r, w, ttl)
}

// PeerHandshaker works to handshake to just connected peer, it detect chain networks
// and protocol versions, and then select InnerHandshaker for that protocol version.
type PeerHandshaker struct {
	pm        PeerManager
	actorServ ActorService
	logger    *log.Logger
	peerID    peer.ID
	// check if is it adhoc
	localChainID *types.ChainID

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

func newHandshaker(pm PeerManager, actor ActorService, log *log.Logger, chainID *types.ChainID, peerID peer.ID) *PeerHandshaker {
	return &PeerHandshaker{pm: pm, actorServ: actor, logger: log, localChainID: chainID, peerID: peerID}
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

func runFuncTimeout(m targetFunc, ttl time.Duration) (interface{}, error) {
	done := make(chan interface{})
	go m(done)
	select {
	case hsResult := <-done:
		return hsResult, nil
	case <-time.NewTimer(ttl).C:
		return nil, TimeoutError
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

func createStatusMsg(pm PeerManager, actorServ ActorService, chainID *types.ChainID) (*types.Status, error) {
	// find my best block
	bestBlock, err := actorServ.GetChainAccessor().GetBestBlock()
	if err != nil {
		return nil, err
	}
	selfAddr := pm.SelfMeta().ToPeerAddress()
	chainIDbytes, err := chainID.Bytes()
	if err != nil {
		return nil, err
	}
	// create message data
	statusMsg := &types.Status{
		Sender:        &selfAddr,
		ChainID:       chainIDbytes,
		BestBlockHash: bestBlock.BlockHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	return statusMsg, nil
}

func (h *PeerHandshaker) selectProtocolVersion(head HSHeader, r *bufio.Reader, w *bufio.Writer) (innerHandshaker, error) {
	switch head.Version {
	case P2PVersion030:
		v030 := newV030StateHS(h.pm, h.actorServ, h.logger, h.localChainID, h.peerID, r, w)
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
