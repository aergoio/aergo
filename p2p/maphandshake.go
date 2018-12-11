/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

// InboundMapHSHandler works to handshake for aergomap
type InboundMapHSHandler struct {
	pm        PeerManager
	actorServ ActorService
	logger    *log.Logger
	peerID    peer.ID

	remoteStatus *types.Status
}

func (h *InboundMapHSHandler) Handle(r io.Reader, w io.Writer, ttl time.Duration) (MsgReadWriter, *types.Status, error) {
	ret, err := runFuncTimeout(func(doneChan chan<- interface{}) {
		rw, statusMsg, err := h.handshakeInboundPeer(r, w)
		doneChan <- &hsResult{rw: rw, statusMsg: statusMsg, err: err}
	}, ttl)
	if err != nil {
		return nil, nil, err
	}
	return ret.(*hsResult).rw, ret.(*hsResult).statusMsg, ret.(*hsResult).err
}

func (h *InboundMapHSHandler) handshakeInboundPeer(r io.Reader, w io.Writer) (MsgReadWriter, *types.Status, error) {
	var hsHeader HSHeader
	bufReader, bufWriter := bufio.NewReader(r), bufio.NewWriter(w)
	// wait initial hsmessage
	headBuf := make([]byte, 8)
	read, err := readToLen(bufReader, headBuf, 8)
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

func (h *InboundMapHSHandler) handshakeOutboundPeer(r io.Reader, w io.Writer) (MsgReadWriter, *types.Status, error) {
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


func readToLen(rd io.Reader, bf []byte, max int) (int, error) {
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

func (h *InboundMapHSHandler) selectProtocolVersion(head HSHeader, r *bufio.Reader, w *bufio.Writer) (innerHandshaker, error) {
	switch head.Version {
	case P2PVersion030:
		v030 := newV030StateHS(h.pm, h.actorServ, h.logger, h.peerID, r, w)
		return v030, nil
	default:
		return nil, fmt.Errorf("not supported version")
	}
}

func (h *InboundMapHSHandler) checkProtocolVersion(versionStr string) error {
	// TODO modify interface and put check code here
	return nil
}

