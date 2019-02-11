/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
)

const msgHeaderLength int = 48

type V030ReadWriter struct {
	r *V030Reader
	w *V030Writer
}

func NewV030ReadWriter(r *bufio.Reader, w *bufio.Writer) *V030ReadWriter {
	return &V030ReadWriter{
		r: &V030Reader{rd:r},
		w: &V030Writer{wr:w},
	}
}

func (rw *V030ReadWriter) ReadMsg() (p2pcommon.Message, error) {
	return rw.r.ReadMsg()
}

func (rw *V030ReadWriter) WriteMsg(msg p2pcommon.Message) error {
	return rw.w.WriteMsg(msg)
}

func NewV030Reader(rd *bufio.Reader) *V030Reader {
	return &V030Reader{rd:rd}
}

func NewV030Writer(wr *bufio.Writer) *V030Writer {
	return &V030Writer{wr:wr}
}

type V030Reader struct {
	rd *bufio.Reader
	headBuf [msgHeaderLength]byte
}

// ReadMsg() must be used in single thread
func (r *V030Reader) ReadMsg() (p2pcommon.Message, error) {
	// fill data
	read, err := r.readToLen(r.headBuf[:], msgHeaderLength)
	if err != nil {
		return nil, err
	}
	if read != msgHeaderLength {
		return nil, fmt.Errorf("invalid msgHeader")
	}

	msg := parseHeader(r.headBuf)
	if msg.length > MaxPayloadLength {
		return nil, fmt.Errorf("too big payload")
	}
	payload := make([]byte, msg.length)
	read, err = r.readToLen(payload, int(msg.length))
	if err != nil {
		return nil, fmt.Errorf("failed to read paylod of msg %s %s : %s", msg.subProtocol.String(), msg.id, err.Error())
	}
	if read != int(msg.length) {
		return nil, fmt.Errorf("failed to read paylod of msg %s %s : payload length mismatch", msg.subProtocol.String(), msg.id)
	}

	msg.payload = payload
	return msg, nil
}

func (r *V030Reader)readToLen(bf []byte, max int ) (int, error) {
	remain := max
	offset := 0
	for remain>0 {
		read, err := r.rd.Read(bf[offset:])
		if err != nil {
			return offset, err
		}
		remain -= read
		offset += read
	}
	return offset, nil
}

type V030Writer struct {
	wr *bufio.Writer
	headBuf [msgHeaderLength]byte
}

// WriteMsg() must be used in single thread
func (w *V030Writer) WriteMsg(msg p2pcommon.Message) error {
	if msg.Length() != uint32(len(msg.Payload())) {
		return fmt.Errorf("Invalid payload size")
	}
	if msg.Length() > MaxPayloadLength {
		return fmt.Errorf("too big payload")
	}

	w.marshalHeader(msg)
	written, err := w.wr.Write(w.headBuf[:])
	if err != nil {
		return err
	}
	if written != msgHeaderLength {
		return fmt.Errorf("header is not written")
	}
	written, err = w.wr.Write(msg.Payload())
	if err != nil {
		return err
	}
	if written !=  int(msg.Length())  {
		return fmt.Errorf("wrong write")
	}
	w.wr.Flush()
	return nil
}

func parseHeader(buf [msgHeaderLength]byte) *V030Message {
	m := &V030Message{}
	m.subProtocol = p2pcommon.SubProtocol(binary.BigEndian.Uint32(buf[0:4]))
	m.length =  binary.BigEndian.Uint32(buf[4:8])
	m.timestamp =  int64(binary.BigEndian.Uint64(buf[8:16]))
	copy(m.id[:], buf[16:32])
	copy(m.originalID[:], buf[32:48])
	return m
}

func (w *V030Writer)marshalHeader(m p2pcommon.Message) {
	binary.BigEndian.PutUint32(w.headBuf[0:4], m.Subprotocol().Uint32())
	binary.BigEndian.PutUint32(w.headBuf[4:8], m.Length())
	binary.BigEndian.PutUint64(w.headBuf[8:16], uint64(m.Timestamp()))

	msgID := m.ID()
	copy(w.headBuf[16:32], msgID[:])
	msgID = m.OriginalID()
	copy(w.headBuf[32:48], msgID[:])
}