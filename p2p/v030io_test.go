/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
)

func init() {
	sampleTxs = make([][]byte, len(sampleTxsB58))
	sampleTxHashes = make([]types.TxID, len(sampleTxsB58))
	for i, hashb58 := range sampleTxsB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleTxs[i] = hash
		sampleTxHashes[i] = types.ToTxID(hash)
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
			sample := &V030Message{subProtocol: NewTxNotice, id: sampleID, timestamp: time.Now().UnixNano(), length: uint32(len(payload)), payload: payload}

			buf := bytes.NewBuffer(nil)
			target := NewV030Writer(bufio.NewWriter(buf))
			target.WriteMsg(sample)

			actual := buf.Bytes()
			assert.Equal(t, len(payload)+msgHeaderLength, len(actual))

			rd := NewV030Reader(bufio.NewReader(buf))
			readMsg, err := rd.ReadMsg()
			assert.Nil(t, err)
			assert.Equal(t, sample, readMsg)
			assert.True(t, bytes.Equal(sample.payload, readMsg.Payload()))

			// read error test
			buf2 := bytes.NewBuffer(actual)
			buf2.Truncate(buf2.Len() - 1 )
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
	//sample := &V030Message{subProtocol: NewTxNotice, id: sampleID, timestamp: time.Now().UnixNano(), length: uint32(len(payload)), payload: payload}
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
	smallMsg := &V030Message{id:sampleID, originalID:sampleID,timestamp:timestamp, subProtocol:NewTxNotice,payload:payload,length:uint32(len(payload))}

	bigHashes := make([][]byte, 0, len(sampleTxs)*10000)
	for i:=0; i<10000; i++ {
		bigHashes = append(bigHashes, sampleTxs...)
	}
	bigPData := &types.NewTransactionsNotice{TxHashes:bigHashes}
	payload, _ = proto.Marshal(bigPData)
	bigMsg := &V030Message{id:sampleID, originalID:sampleID,timestamp:timestamp, subProtocol:NewTxNotice,payload:payload,length:uint32(len(payload))}

	benchmarks := []struct {
		name string
		input *V030Message
		repeatCount int
	}{
		// write small
		{"BWSmall", smallMsg, 100, },
		// write big
		{"BWBig", bigMsg, 100, },
		////write small sign
		//{"BWSmallSign", smallMsg, 100, },
		////write big sign
		//{"BWBigSign", bigMsg, 100, },
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				w := ioutil.Discard
				target := NewV030Writer(	bufio.NewWriter(w)	)
				for j:=0; j<bm.repeatCount; j++ {
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
	smallMsg := &V030Message{id:sampleID, originalID:sampleID,timestamp:timestamp, subProtocol:NewTxNotice,payload:payload,length:uint32(len(payload))}
	smallBytes := getMashaledV030(smallMsg,100)

	bigHashes := make([][]byte, 0, len(sampleTxs)*10000)
	for i:=0; i<10000; i++ {
		bigHashes = append(bigHashes, sampleTxs...)
	}
	bigPData := &types.NewTransactionsNotice{TxHashes:bigHashes}
	payload, _ = proto.Marshal(bigPData)
	bigMsg := &V030Message{id:sampleID, originalID:sampleID,timestamp:timestamp, subProtocol:NewTxNotice,payload:payload,length:uint32(len(payload))}
	bigBytes := getMashaledV030(bigMsg,100)

	fmt.Printf("small : %d , big : %d \n", len(smallBytes), len(bigBytes) )

	benchmarks := []struct {
		name string
		input []byte
		repeatCount int
	}{
		// read small
		{"BRSmall", smallBytes, 100, },
		// read big
		{"BRBig", bigBytes, 100, },
		//// read small sign verify
		//{"BRSmallSign", smallBytes, 100, },
		//// read big sign verify
		//{"BRBigSign", bigBytes, 100, },


	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r := bytes.NewBuffer(bm.input)
				target := NewV030Reader(	bufio.NewReader(r)	)
				for j:=0; j<bm.repeatCount; j++ {
					actual, _ := target.ReadMsg()
					if actual.ID() != sampleID {
						b.Error()
					}
				}
			}
		})
	}
}

func getMashaledV030(m *V030Message,  repeat int) []byte {
	unitbuf := &bytes.Buffer{}
	writer := NewV030Writer(bufio.NewWriter(unitbuf) )
	writer.WriteMsg(m)
	unitbytes := unitbuf.Bytes()
	buf := make([]byte, 0, len(unitbytes) * repeat )
	for i:=0 ;i < repeat ; i++ {
		buf = append(buf, unitbytes...)
	}
	return buf
}
