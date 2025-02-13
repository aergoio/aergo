package msg

import (
	"bytes"
	"strings"
	"encoding/binary"
	"io"
	"net"
	"time"
	"errors"
	"fmt"
)

// SerializeMessage serializes a variable number of strings into a byte slice
func SerializeMessage(strings ...string) []byte {
	var buf bytes.Buffer

	// write number of strings
	binary.Write(&buf, binary.LittleEndian, uint32(len(strings)))

	// write each string's length and content
	for _, s := range strings {
		binary.Write(&buf, binary.LittleEndian, uint32(len(s)))
		buf.WriteString(s)
	}

	return buf.Bytes()
}

func SerializeMessageBytes(args ...[]byte) []byte {
	var buf bytes.Buffer

	// write number of strings
	binary.Write(&buf, binary.LittleEndian, uint32(len(args)))

	// write each string's length and content
	for _, b := range args {
		binary.Write(&buf, binary.LittleEndian, uint32(len(b)))
		buf.Write(b)
	}

	return buf.Bytes()
}

// DeserializeMessage deserializes a byte slice into an array of strings
func DeserializeMessage(data []byte) ([]string, error) {
	var strings []string
	buf := bytes.NewReader(data)

	// read number of strings
	var numStrings uint32
	if err := binary.Read(buf, binary.LittleEndian, &numStrings); err != nil {
		return nil, err
	}

	// read each string's length and content without making unnecessary copies,
	// by creating a slice that references the original buffer
	for i := uint32(0); i < numStrings; i++ {
		var length uint32
		if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
			return nil, err
		}

		// get the current position
		pos, err := buf.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, err
		}

		// create a slice that references the original buffer
		strBytes := data[pos : pos+int64(length)]
		buf.Seek(int64(length), io.SeekCurrent) // move the buffer position forward

		strings = append(strings, string(strBytes))
	}

	return strings, nil
}

func SendMessage(conn net.Conn, message []byte) (err error) {

	// send the length prefix
	lengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBytes, uint32(len(message)))
	_, err = conn.Write(lengthBytes)
	if err != nil {
		return err
	}

	// send the message
	_, err = conn.Write(message)
	if err != nil {
		return err
	}

	// attempt to flush the connection
	if flusher, ok := conn.(interface{ Flush() error }); ok {
		err = flusher.Flush()
		if err != nil {
			return fmt.Errorf("error flushing connection: %w", err)
		}
	}

	return nil
}

// waits for a full message (length prefix + data) from the abstract domain socket
func WaitForMessage(conn net.Conn, deadline time.Time) (msg []byte, err error) {

	if !deadline.IsZero() {
		// define the deadline for the connection
		conn.SetReadDeadline(deadline)
	}

	// read the length prefix
	length := make([]byte, 4)
	n, err := io.ReadFull(conn, length)
	if err != nil {
		if err == io.EOF && n == 0 {
			return nil, fmt.Errorf("connection closed before reading message length")
		}
		if strings.Contains(err.Error(), "i/o timeout") {
			return nil, errors.New("contract timeout during vm execution")
		}
		return nil, fmt.Errorf("error reading message length (read %d bytes): %w", n, err)
	}

	// use little endian to get the length
	msgLength := binary.LittleEndian.Uint32(length)

	// read the message
	msg = make([]byte, msgLength)
	n, err = io.ReadFull(conn, msg)
	if err != nil {
		if strings.Contains(err.Error(), "i/o timeout") {
			return nil, errors.New("contract timeout during vm execution")
		}
		return nil, fmt.Errorf("error reading message body (read %d/%d bytes): %w", n, msgLength, err)
	}

	// return the message
	return msg, nil
}
