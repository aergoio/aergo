/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package v030

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

var sampleTxsB58 = []string{
	"4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45",
	"6xfk39kuyDST7NwCu8tx3wqwFZ5dwKPDjxUS14tU7NZb8",
	"E8dbBGe9Hnuhk35cJoekPjL3VoL4xAxtnRuP47UoxzHd",
	"HB7Hg5GUbHuxwe8Lp5PcYUoAaQ7EZjRNG6RuvS6DnDRf",
	"BxKmDg9VbWHxrWnStEeTzJ2Ze7RF7YK4rpyjcsWSsnxs",
	"DwmGqFU4WgADpYN36FXKsYxMjeppvh9Najg4KxJ8gtX3",
}

var sampleTxs [][]byte
var sampleTxHashes []types.TxID

var sampleBlksB58 = []string{
	"v6zbuQ4aVSdbTwQhaiZGp5pcL5uL55X3kt2wfxor5W6",
	"2VEPg4MqJUoaS3EhZ6WWSAUuFSuD4oSJ645kSQsGV7H9",
	"AtzTZ2CZS45F1276RpTdLfYu2DLgRcd9HL3aLqDT1qte",
	"2n9QWNDoUvML756X7xdHWCFLZrM4CQEtnVH2RzG5FYAw",
	"6cy7U7XKYtDTMnF3jNkcJvJN5Rn85771NSKjc5Tfo2DM",
	"3bmB8D37XZr4DNPs64NiGRa2Vw3i8VEgEy6Xc2XBmRXC",
}
var sampleBlks [][]byte
var sampleBlksHashes []types.BlockID

func init() {
	sampleTxs = make([][]byte, len(sampleTxsB58))
	sampleTxHashes = make([]types.TxID, len(sampleTxsB58))
	for i, hashb58 := range sampleTxsB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleTxs[i] = hash
		copy(sampleTxHashes[i][:], hash)
	}

	sampleBlks = make([][]byte, len(sampleBlksB58))
	sampleBlksHashes = make([]types.BlockID, len(sampleBlksB58))
	for i, hashb58 := range sampleTxsB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleBlks[i] = hash
		copy(sampleBlksHashes[i][:], hash)
	}
}

func Test_ReadWrite(t *testing.T) {
	var sampleID p2pcommon.MsgID
	sampleUUID, _ := uuid.NewV4()
	copy(sampleID[:], sampleUUID[:])

	tests := []struct {
		name string
		ids  [][]byte
	}{
		{"TEmpty", nil},
		{"TSingle", sampleTxs[:1]},
		{"TSmall", sampleTxs},
		{"TBig", func() [][]byte {
			toreturn := make([][]byte, 0, len(sampleTxs)*1000)
			for i := 0; i < 1000; i++ {
				toreturn = append(toreturn, sampleTxs...)
			}
			return toreturn
		}()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sizeChecker, sizeChecker2 := ioSum{}, ioSum{}
			samplePData := &types.NewTransactionsNotice{TxHashes: test.ids}
			payload, _ := proto.Marshal(samplePData)
			sample := p2pcommon.NewMessageValue(p2pcommon.NewTxNotice, sampleID, p2pcommon.EmptyID, time.Now().UnixNano(), payload)

			buf := bytes.NewBuffer(nil)
			target := NewV030ReadWriter(nil, buf, nil)
			target.AddIOListener(&sizeChecker)

			target.WriteMsg(sample)

			actual := buf.Bytes()
			assert.Equal(t, len(payload)+msgHeaderLength, len(actual))
			assert.Equal(t, len(actual), sizeChecker.writeN)

			rd := NewV030ReadWriter(bufio.NewReader(buf), ioutil.Discard, nil)
			rd.AddIOListener(&sizeChecker2)

			readMsg, err := rd.ReadMsg()

			assert.Nil(t, err)
			assert.NotNil(t, readMsg)
			assert.Equal(t, sample, readMsg)
			assert.True(t, bytes.Equal(sample.Payload(), readMsg.Payload()))
			assert.Equal(t, sizeChecker.writeN, sizeChecker2.readN)

			// read error test
			buf2 := bytes.NewBuffer(actual)
			buf2.Truncate(buf2.Len() - 1)
			rd2 := NewV030ReadWriter(bufio.NewReader(buf), ioutil.Discard, nil)
			readMsg, err = rd2.ReadMsg()
			assert.NotNil(t, err)
		})
	}
}

type ioSum struct {
	readN  int
	writeN int
}

func (s *ioSum) OnRead(protocol p2pcommon.SubProtocol, read int) {
	s.readN += read
}

func (s *ioSum) OnWrite(protocol p2pcommon.SubProtocol, write int) {
	s.writeN += write
}

func TestV030Writer_WriteError(t *testing.T) {
	//var sampleID MsgID
	//sampleUUID, _ := uuid.NewRandom()
	//copy(sampleID[:], sampleUUID[:])
	//samplePData := &types.NewTransactionsNotice{TxHashes:sampleTxs}
	//payload, _ := proto.Marshal(samplePData)
	//sample := &MessageValue{subProtocol: subproto.NewTxNotice, id: sampleID, timestamp: time.Now().UnixNano(), length: uint32(len(payload)), payload: payload}
	//mockWriter := make(MockWriter)
	//mockWriter.On("Write", mock.Anything).Return(fmt.Errorf("writer error"))
	//target := NewV030Writer(bufio.NewWriter(mockWriter))
	//err := target.WriteMsg(sample)
	//assert.NotNil(t, err)
}

func BenchmarkV030Writer_WriteMsg(b *testing.B) {
	var sampleID p2pcommon.MsgID
	sampleUUID, _ := uuid.NewV4()
	copy(sampleID[:], sampleUUID[:])
	timestamp := time.Now().UnixNano()

	smallPData := &types.NewTransactionsNotice{}
	payload, _ := proto.Marshal(smallPData)
	smallMsg := p2pcommon.NewMessageValue(p2pcommon.NewTxNotice, sampleID, p2pcommon.EmptyID, timestamp, payload)
	bigHashes := make([][]byte, 0, len(sampleTxs)*10000)
	for i := 0; i < 10000; i++ {
		bigHashes = append(bigHashes, sampleTxs...)
	}
	bigPData := &types.NewTransactionsNotice{TxHashes: bigHashes}
	payload, _ = proto.Marshal(bigPData)
	bigMsg := p2pcommon.NewMessageValue(p2pcommon.NewTxNotice, sampleID, p2pcommon.EmptyID, timestamp, payload)

	benchmarks := []struct {
		name        string
		rw          p2pcommon.MsgReadWriter
		input       *p2pcommon.MessageValue
		repeatCount int
	}{
		// write small
		{"BWSmall", NewV030ReadWriter(nil, ioutil.Discard, nil), smallMsg, 100},
		// write big
		{"BWBig", NewV030ReadWriter(nil, ioutil.Discard, nil), bigMsg, 100},
		////write small with rw
		//{"BWSmallRW", NewV030ReadWriter(nil, ioutil.Discard, nil), smallMsg, 100},
		//// write big with rw
		//{"BWBigRW", NewV030ReadWriter(nil, ioutil.Discard, nil), bigMsg, 100},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				target := bm.rw
				for j := 0; j < bm.repeatCount; j++ {
					target.WriteMsg(bm.input)
				}
			}
		})
	}
}

func BenchmarkV030Reader_ReadMsg(b *testing.B) {
	var sampleID p2pcommon.MsgID
	sampleUUID, _ := uuid.NewV4()
	copy(sampleID[:], sampleUUID[:])
	timestamp := time.Now().UnixNano()

	smallPData := &types.NewTransactionsNotice{}
	payload, _ := proto.Marshal(smallPData)
	smallMsg := p2pcommon.NewMessageValue(p2pcommon.NewTxNotice, sampleID, p2pcommon.EmptyID, timestamp, payload)
	smallBytes := getMarshaledV030(smallMsg, 1)
	bigHashes := make([][]byte, 0, len(sampleTxs)*10000)
	for i := 0; i < 10000; i++ {
		bigHashes = append(bigHashes, sampleTxs...)
	}
	bigPData := &types.NewTransactionsNotice{TxHashes: bigHashes}
	payload, _ = proto.Marshal(bigPData)
	bigMsg := p2pcommon.NewMessageValue(p2pcommon.NewTxNotice, sampleID, p2pcommon.EmptyID, timestamp, payload)
	bigBytes := getMarshaledV030(bigMsg, 1)

	fmt.Printf("small : %d , big : %d \n", len(smallBytes), len(bigBytes))

	benchmarks := []struct {
		name        string
		rd          p2pcommon.MsgReadWriter
		repeatCount int
	}{
		// read big with rw
		{"BRBigRW", NewV030ReadWriter(NewRepeatedBuffer(bigBytes), nil, nil), 100},
		// read small with rw
		{"BRSmallRW", NewV030ReadWriter(NewRepeatedBuffer(smallBytes), nil, nil), 100},
		//// read small
		//{"BRSmall", NewV030Reader(NewRepeatedBuffer(smallBytes)), 100},
		//// read big
		//{"BRBig", NewV030Reader(NewRepeatedBuffer(bigBytes)), 100},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				target := bm.rd
				for j := 0; j < bm.repeatCount; j++ {
					actual, err := target.ReadMsg()
					if err != nil {
						b.Fatal("err while reading on lap", j, err.Error())
					} else if actual.ID() != sampleID {
						b.Fatal()
					}
				}
			}
		})
	}
}

func getMarshaledV030(m *p2pcommon.MessageValue, repeat int) []byte {
	unitbuf := &bytes.Buffer{}
	writer := NewV030ReadWriter(nil, bufio.NewWriter(unitbuf), nil)
	writer.WriteMsg(m)
	unitbytes := unitbuf.Bytes()
	buf := make([]byte, 0, len(unitbytes)*repeat)
	for i := 0; i < repeat; i++ {
		buf = append(buf, unitbytes...)
	}
	return buf
}

type RepeatedBuffer struct {
	offset, size int
	buf          []byte
}

func (r *RepeatedBuffer) Read(p []byte) (n int, err error) {
	copied := copy(p, r.buf[r.offset:])
	r.offset += copied
	if r.offset >= r.size {
		r.offset = 0
	}
	return copied, nil
}

func NewRepeatedBuffer(buf []byte) *RepeatedBuffer {
	return &RepeatedBuffer{buf: buf, size: len(buf)}
}
