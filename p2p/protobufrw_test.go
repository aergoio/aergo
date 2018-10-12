/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/multiformats/go-multicodec/protobuf"
	"github.com/stretchr/testify/assert"
)

func Test_newBufMsgReadWriter(t *testing.T) {
	mockReader := new(MockReader)
	mockWriter := new(MockWriter)
	r := bufio.NewReader(mockReader)
	w := bufio.NewWriter(mockWriter)
	got := newBufMsgReadWriter(r, w)

	assert.NotNil(t, got)
	assert.NotNil(t, got.r)
	assert.NotNil(t, got.w)
	assert.Equal(t, got.r.rd, r)
	assert.Equal(t, got.w.wr, w)
}

func Test_bufMsgReadWriter_ReadMsg(t *testing.T) {
}

func Test_bufMsgReader_ReadMsg(t *testing.T) {
	factory := &pbMOFactory{signer:&dummySigner{}}
	sampleObj := factory.newMsgRequestOrder(true, PingRequest, &types.Ping{BestBlockHash: dummyBlockHash, BestHeight:123}).(*pbRequestOrder).message
	buf := bytes.NewBuffer(nil)
	encoder := mc_pb.Multicodec(nil).Encoder(buf)
	encoder.Encode(sampleObj)
	sampleBytes := buf.Bytes()
	wrongBytes := []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0xaa, 0x1, 0x2, 0x3, 0x4, 0x5, 0xaa, 0x1, 0x2, 0x3, 0x4, 0x5, 0xaa}

	tests := []struct {
		name    string
		input   []byte
		want    Message
		wantErr bool
	}{
		{"TSucc", sampleBytes, sampleObj, false},
		{"TFail1", wrongBytes, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readerStub := bytes.NewReader(tt.input)
			br := bufio.NewReader(readerStub)
			r := newBufMsgReader(br)

			got, err := r.ReadMsg()
			if (err != nil) != tt.wantErr {
				t.Errorf("bufMsgReader.ReadMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			actual := got
			if tt.want != nil {
				assert.Equal(t, tt.want.Timestamp(), actual.Timestamp())
				assert.Equal(t, tt.want.ID(), actual.ID())
				assert.Equal(t, tt.want.Subprotocol(), actual.Subprotocol())
				assert.Equal(t, tt.want.Length(), actual.Length())
				assert.Equal(t, tt.want.Payload(), actual.Payload())
				
			}
		})
	}
}


func Test_bufMsgWriter_WriteMsg(t *testing.T) {
	factory := &pbMOFactory{signer: &dummySigner{}}
	sampleObj := factory.newMsgRequestOrder(true, PingRequest, &types.Ping{BestBlockHash: dummyBlockHash, BestHeight:123}).(*pbRequestOrder).message
	buf := bytes.NewBuffer(nil)
	encoder := mc_pb.Multicodec(nil).Encoder(buf)
	encoder.Encode(sampleObj)
	sampleBytes := buf.Bytes()

	wbuf := bytes.NewBuffer(nil)
	wr := bufio.NewWriter(wbuf)
	w := newBufMsgWriter(wr)

	w.WriteMsg(sampleObj)
	actualBytes := wbuf.Bytes()

	assert.Equal(t, sampleBytes, actualBytes)

	r := newBufMsgReader(bufio.NewReader(bytes.NewReader(actualBytes)))
	readBackObj, err := r.ReadMsg()
	assert.Nil(t, err)

		//assert.Equal(t, sampleObj.Header.ClientVersion, readBackObj.Header.ClientVersion)
		assert.Equal(t, sampleObj.Timestamp(), readBackObj.Timestamp())
		assert.Equal(t, sampleObj.ID(), readBackObj.ID())
		assert.Equal(t, sampleObj.Subprotocol(), readBackObj.Subprotocol())
		assert.Equal(t, sampleObj.Length(), readBackObj.Length())
		assert.Equal(t, sampleObj.Payload(), readBackObj.Payload())

}

func BenchmarkBufMsgWriter_WriteMsg(b *testing.B) {
	dummySign := &dummySigner{}
	//realSign := newDefaultMsgSigner(sampleKey1Priv, sampleKey1Pub, sampleKey1ID )
	mofactory := pbMOFactory{signer: dummySign}
	dummyMo := mofactory.newMsgTxBroadcastOrder(&types.NewTransactionsNotice{}).(*pbTxNoticeOrder)
	smallMsg := dummyMo.message

	bigHashes := make([][]byte, 0, len(sampleTxs)*10000)
	for i:=0; i<10000; i++ {
		bigHashes = append(bigHashes, sampleTxs...)
	}
	dummyMo2 := mofactory.newMsgTxBroadcastOrder(&types.NewTransactionsNotice{TxHashes:bigHashes}).(*pbTxNoticeOrder)
	bigMsg := dummyMo2.message

	benchmarks := []struct {
		name string
		input Message
		repeatCount int
		signer msgSigner
	}{
		// write small
		{"BWSmall", smallMsg, 100, dummySign},
		// write big
		{"BWBig", bigMsg, 100, dummySign},
		////write small sign
		//{"BWSmallSign", smallMsg, 100, realSign},
		////write big sign
		//{"BWBigSign", bigMsg, 100, realSign},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				w := ioutil.Discard
				target := newBufMsgWriter(	bufio.NewWriter(w)	)
				for j:=0; j<bm.repeatCount; j++ {
					target.WriteMsg(bm.input)
				}
			}
		})
	}
}

func BenchmarkBufMsgReader_ReadMsg(b *testing.B) {
	dummySign := &dummySigner{}
	//realSign := newDefaultMsgSigner(sampleKey1Priv, sampleKey1Pub, sampleKey1ID )
	mofactory := pbMOFactory{signer: dummySign}
	dummyMo := mofactory.newMsgTxBroadcastOrder(&types.NewTransactionsNotice{}).(*pbTxNoticeOrder)
	//realSign.signMsg(dummyMo.message)
	smallBytes := getV020Mashaled(dummyMo.message,100)

	bigHashes := make([][]byte, 0, len(sampleTxs)*10000)
	for i:=0; i<10000; i++ {
		bigHashes = append(bigHashes, sampleTxs...)
	}
	dummyMo2 := mofactory.newMsgTxBroadcastOrder(&types.NewTransactionsNotice{TxHashes:bigHashes}).(*pbTxNoticeOrder)
	//realSign.signMsg(dummyMo2.message)
	bigBytes := getV020Mashaled(dummyMo2.message,100)

	fmt.Printf("small : %d , big : %d \n", len(smallBytes), len(bigBytes) )


	benchmarks := []struct {
		name string
		input []byte
		repeatCount int
		signer msgSigner
	}{
		// read small
		{"BRSmall", smallBytes, 100, dummySign},
		// read big
		{"BRBig", bigBytes, 100, dummySign},
		//// read small sign verify
		//{"BRSmallSign", smallBytes, 100, realSign},
		//// read big sign verify
		//{"BRBigSign", bigBytes, 100, realSign},


	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r := bytes.NewBuffer(bm.input)
				target := newBufMsgReader(	bufio.NewReader(r)	)
				for j:=0; j<bm.repeatCount; j++ {
					actual, _ := target.ReadMsg()
					bm.signer.verifyMsg(actual.(*V020Wrapper).P2PMessage, sampleKey1ID)
				}
			}
		})
	}
}

func getV020Mashaled(m Message,  repeat int) []byte {
	unitbuf := &bytes.Buffer{}
	writer := newBufMsgWriter(	bufio.NewWriter(unitbuf) )
	writer.WriteMsg(m)
	unitbytes := unitbuf.Bytes()
	buf := make([]byte, 0, len(unitbytes) * repeat )
	for i:=0 ;i < repeat ; i++ {
		buf = append(buf, unitbytes...)
	}
	return buf
}