/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"bytes"
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
	factory := &pbMOFactory{signer: &dummySigner{}}
	sampleObj := factory.newMsgRequestOrder(true, PingRequest, &types.Ping{BestBlockHash: dummyBlockHash, BestHeight:123}).(*pbRequestOrder).message
	buf := bytes.NewBuffer(nil)
	encoder := mc_pb.Multicodec(nil).Encoder(buf)
	encoder.Encode(sampleObj)
	sampleBytes := buf.Bytes()
	wrongBytes := []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0xaa, 0x1, 0x2, 0x3, 0x4, 0x5, 0xaa, 0x1, 0x2, 0x3, 0x4, 0x5, 0xaa}

	tests := []struct {
		name    string
		input   []byte
		want    *types.P2PMessage
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
			if tt.want != nil {
				assert.Equal(t, tt.want.Header.ClientVersion, got.Header.ClientVersion)
				assert.Equal(t, tt.want.Header.Timestamp, got.Header.Timestamp)
				assert.Equal(t, tt.want.Header.Id, got.Header.Id)
				assert.Equal(t, tt.want.Header.Subprotocol, got.Header.Subprotocol)
				assert.Equal(t, tt.want.Header.Length, got.Header.Length)
				assert.Equal(t, tt.want.Data, got.Data)
				
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

		assert.Equal(t, sampleObj.Header.ClientVersion, readBackObj.Header.ClientVersion)
		assert.Equal(t, sampleObj.Header.Timestamp, readBackObj.Header.Timestamp)
		assert.Equal(t, sampleObj.Header.Id, readBackObj.Header.Id)
		assert.Equal(t, sampleObj.Header.Subprotocol, readBackObj.Header.Subprotocol)
		assert.Equal(t, sampleObj.Header.Length, readBackObj.Header.Length)
		assert.Equal(t, sampleObj.Data, readBackObj.Data)

}
