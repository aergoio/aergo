/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package v030

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/aergoio/aergo/p2p/p2pcommon"
)

const msgHeaderLength int = 48

type V030ReadWriter struct {
	r        *bufio.Reader
	readBuf  [msgHeaderLength]byte
	w        *bufio.Writer
	writeBuf [msgHeaderLength]byte

	c io.Closer
}

func NewV030MsgPipe(s io.ReadWriteCloser) *V030ReadWriter {
	return NewV030ReadWriter(s, s, s)
}
func NewV030ReadWriter(r io.Reader, w io.Writer, c io.Closer) *V030ReadWriter {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}
	return &V030ReadWriter{
		r: br,
		w: bw,
		c: c,
	}
}

func (rw *V030ReadWriter) Close() error {
	return rw.c.Close()
}

// ReadMsg() must be used in single thread
func (r *V030ReadWriter) ReadMsg() (p2pcommon.Message, error) {
	// fill data
	read, err := r.readToLen(r.readBuf[:], msgHeaderLength)
	if err != nil {
		return nil, err
	}
	if read != msgHeaderLength {
		return nil, fmt.Errorf("invalid msgHeader")
	}

	msg, bodyLen := parseHeader(r.readBuf)
	if bodyLen > p2pcommon.MaxPayloadLength {
		return nil, fmt.Errorf("too big payload")
	}
	payload := make([]byte, bodyLen)
	read, err = r.readToLen(payload, int(bodyLen))
	if err != nil {
		return nil, fmt.Errorf("failed to read paylod of msg %s %s : %s", msg.Subprotocol().String(), msg.ID(), err.Error())
	}
	if read != int(bodyLen) {
		return nil, fmt.Errorf("failed to read paylod of msg %s %s : payload length mismatch", msg.Subprotocol().String(), msg.ID())
	}

	msg.SetPayload(payload)
	return msg, nil
}

func (r *V030ReadWriter) readToLen(bf []byte, max int) (int, error) {
	remain := max
	offset := 0
	for remain > 0 {
		read, err := r.r.Read(bf[offset:])
		if err != nil {
			return offset, err
		}
		remain -= read
		offset += read
	}
	return offset, nil
}

// WriteMsg() must be used in single thread
func (w *V030ReadWriter) WriteMsg(msg p2pcommon.Message) error {
	if msg.Length() != uint32(len(msg.Payload())) {
		return fmt.Errorf("Invalid payload size")
	}
	if msg.Length() > p2pcommon.MaxPayloadLength {
		return fmt.Errorf("too big payload")
	}

	w.marshalHeader(msg)
	written, err := w.w.Write(w.writeBuf[:])
	if err != nil {
		return err
	}
	if written != msgHeaderLength {
		return fmt.Errorf("header is not written")
	}
	written, err = w.w.Write(msg.Payload())
	if err != nil {
		return err
	}
	if written != int(msg.Length()) {
		return fmt.Errorf("wrong write")
	}
	return w.w.Flush()
}

func parseHeader(buf [msgHeaderLength]byte) (*p2pcommon.MessageValue, uint32) {
	subProtocol := p2pcommon.SubProtocol(binary.BigEndian.Uint32(buf[0:4]))
	length := binary.BigEndian.Uint32(buf[4:8])
	timestamp := int64(binary.BigEndian.Uint64(buf[8:16]))
	msgID := p2pcommon.MustParseBytes(buf[16:32])
	orgID := p2pcommon.MustParseBytes(buf[32:48])
	return p2pcommon.NewLiteMessageValue(subProtocol, msgID, orgID, timestamp), length
}

func (w *V030ReadWriter) marshalHeader(m p2pcommon.Message) {
	binary.BigEndian.PutUint32(w.writeBuf[0:4], m.Subprotocol().Uint32())
	binary.BigEndian.PutUint32(w.writeBuf[4:8], m.Length())
	binary.BigEndian.PutUint64(w.writeBuf[8:16], uint64(m.Timestamp()))

	msgID := m.ID()
	copy(w.writeBuf[16:32], msgID[:])
	msgID = m.OriginalID()
	copy(w.writeBuf[32:48], msgID[:])
}
