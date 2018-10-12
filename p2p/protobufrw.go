/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"

	"github.com/aergoio/aergo/types"
	mc_pb "github.com/multiformats/go-multicodec/protobuf"
)

type bufMsgReadWriter struct {
	r *bufMsgReader
	w *bufMsgWriter
}

func newBufMsgReadWriter(r *bufio.Reader, w *bufio.Writer) *bufMsgReadWriter {
	return &bufMsgReadWriter{
		r: &bufMsgReader{r},
		w: &bufMsgWriter{w},
	}
}

func (rw *bufMsgReadWriter) ReadMsg() (Message, error) {
	return rw.r.ReadMsg()
}

func (rw *bufMsgReadWriter) WriteMsg(msg Message) error {
	return rw.w.WriteMsg(msg)
}

func newBufMsgReader(rd *bufio.Reader) *bufMsgReader {
	return &bufMsgReader{rd}
}

func newBufMsgWriter(wr *bufio.Writer) *bufMsgWriter {
	return &bufMsgWriter{wr}
}

type bufMsgReader struct {
	rd *bufio.Reader
}

func (r *bufMsgReader) ReadMsg() (Message, error) {
	msg := &types.P2PMessage{}
	decoder := mc_pb.Multicodec(nil).Decoder(r.rd)
	err := decoder.Decode(msg)
	if err != nil {
		return nil, err
	}
	return &V020Wrapper{P2PMessage:msg}, nil
}

type bufMsgWriter struct {
	wr *bufio.Writer
}

func (w *bufMsgWriter) WriteMsg(msg Message) error {
	err := SendProtoMessage(msg.(*V020Wrapper).P2PMessage, w.wr)
	return err
}
