package p2p

import (
	"bufio"

	"github.com/aergoio/aergo/types"
	"github.com/multiformats/go-multicodec/protobuf"
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

func newBufMsgRW(rw *bufio.ReadWriter) *bufMsgReadWriter {
	return newBufMsgReadWriter(rw.Reader, rw.Writer)
}

func (rw *bufMsgReadWriter) ReadMsg() (*types.P2PMessage, error) {
	return rw.r.ReadMsg()
}

func (rw *bufMsgReadWriter) WriteMsg(msg *types.P2PMessage) error {
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

func (r *bufMsgReader) ReadMsg() (*types.P2PMessage, error) {
	msg := &types.P2PMessage{}
	decoder := mc_pb.Multicodec(nil).Decoder(r.rd)
	err := decoder.Decode(msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

type bufMsgWriter struct {
	wr *bufio.Writer
}

func (w *bufMsgWriter) WriteMsg(msg *types.P2PMessage) error {
	err := SendProtoMessage(msg, w.wr)
	return err
}
