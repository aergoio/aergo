/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package v030

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"io/ioutil"
	"testing"
	"time"

	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/subproto"
	"github.com/aergoio/aergo/types"
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
			samplePData := &types.NewTransactionsNotice{TxHashes: test.ids}
			payload, _ := proto.Marshal(samplePData)
			sample := p2pcommon.NewMessageValue(subproto.NewTxNotice, sampleID,  p2pcommon.EmptyID, time.Now().UnixNano(),payload)

			buf := bytes.NewBuffer(nil)
			target := NewV030Writer(bufio.NewWriter(buf))
			target.WriteMsg(sample)

			actual := buf.Bytes()
			assert.Equal(t, len(payload)+msgHeaderLength, len(actual))

			rd := NewV030Reader(bufio.NewReader(buf))
			readMsg, err := rd.ReadMsg()
			assert.Nil(t, err)
			assert.Equal(t, sample, readMsg)
			assert.True(t, bytes.Equal(sample.Payload(), readMsg.Payload()))

			// read error test
			buf2 := bytes.NewBuffer(actual)
			buf2.Truncate(buf2.Len() - 1)
			rd2 := NewV030Reader(bufio.NewReader(buf))
			readMsg, err = rd2.ReadMsg()
			assert.NotNil(t, err)
		})
	}
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
	smallMsg := p2pcommon.NewMessageValue(subproto.NewTxNotice, sampleID,  p2pcommon.EmptyID, timestamp,payload)
	bigHashes := make([][]byte, 0, len(sampleTxs)*10000)
	for i := 0; i < 10000; i++ {
		bigHashes = append(bigHashes, sampleTxs...)
	}
	bigPData := &types.NewTransactionsNotice{TxHashes: bigHashes}
	payload, _ = proto.Marshal(bigPData)
	bigMsg := p2pcommon.NewMessageValue(subproto.NewTxNotice, sampleID,  p2pcommon.EmptyID, timestamp,payload)

	benchmarks := []struct {
		name        string
		input       *p2pcommon.MessageValue
		repeatCount int
	}{
		// write small
		{"BWSmall", smallMsg, 100},
		// write big
		{"BWBig", bigMsg, 100},
		////write small sign
		//{"BWSmallSign", smallMsg, 100, },
		////write big sign
		//{"BWBigSign", bigMsg, 100, },
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				w := ioutil.Discard
				target := NewV030Writer(bufio.NewWriter(w))
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
	smallMsg := p2pcommon.NewMessageValue(subproto.NewTxNotice, sampleID,  p2pcommon.EmptyID, timestamp,payload)
	smallBytes := getMashaledV030(smallMsg, 100)

	bigHashes := make([][]byte, 0, len(sampleTxs)*10000)
	for i := 0; i < 10000; i++ {
		bigHashes = append(bigHashes, sampleTxs...)
	}
	bigPData := &types.NewTransactionsNotice{TxHashes: bigHashes}
	payload, _ = proto.Marshal(bigPData)
	bigMsg := p2pcommon.NewMessageValue(subproto.NewTxNotice, sampleID,  p2pcommon.EmptyID, timestamp,payload)
	bigBytes := getMashaledV030(bigMsg, 100)

	fmt.Printf("small : %d , big : %d \n", len(smallBytes), len(bigBytes))

	benchmarks := []struct {
		name        string
		input       []byte
		repeatCount int
	}{
		// read small
		{"BRSmall", smallBytes, 100},
		// read big
		{"BRBig", bigBytes, 100},
		//// read small sign verify
		//{"BRSmallSign", smallBytes, 100, },
		//// read big sign verify
		//{"BRBigSign", bigBytes, 100, },

	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r := bytes.NewBuffer(bm.input)
				target := NewV030Reader(bufio.NewReader(r))
				for j := 0; j < bm.repeatCount; j++ {
					actual, _ := target.ReadMsg()
					if actual.ID() != sampleID {
						b.Error()
					}
				}
			}
		})
	}
}

func getMashaledV030(m *p2pcommon.MessageValue, repeat int) []byte {
	unitbuf := &bytes.Buffer{}
	writer := NewV030Writer(bufio.NewWriter(unitbuf))
	writer.WriteMsg(m)
	unitbytes := unitbuf.Bytes()
	buf := make([]byte, 0, len(unitbytes)*repeat)
	for i := 0; i < repeat; i++ {
		buf = append(buf, unitbytes...)
	}
	return buf
}
