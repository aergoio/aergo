/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/funkygao/golib/rand"
	"github.com/libp2p/go-libp2p-core/crypto"
	"testing"

	"github.com/rs/zerolog"
)

func TestLogStringersMarshaler_MarshalZerologArray(t *testing.T) {

	sampleArr := make([]fmt.Stringer, 20)
	for i := 0; i < 20; i++ {
		sampleArr[i] = NumOrderer{i}
	}
	type fields struct {
		arr   []fmt.Stringer
		limit int
	}

	tests := []struct {
		name   string
		fields fields

		wantSize int
	}{
		{"TEmpty", fields{nil, 10}, 0},
		{"TOne", fields{sampleArr[:1], 10}, 1},
		{"TMid", fields{sampleArr[1:6], 10}, 5},
		{"TMax", fields{sampleArr[:10], 10}, 10},
		{"TOver", fields{sampleArr[0:11], 10}, 10},
		{"TOver2", fields{sampleArr, 10}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf1 := bytes.NewBuffer(nil)
			log1 := log.NewLogger("test.p2p").Output(buf1)
			buf2 := bytes.NewBuffer(nil)
			log2 := log.NewLogger("test.p2p").Output(buf2)

			m := NewLogStringersMarshaller(tt.fields.arr, tt.fields.limit)
			a := zerolog.Arr()
			m.MarshalZerologArray(a)
			log1.Info().Array("t", m).Msg("Print ")
			log2.Info().Array("t", a).Msg("Print ")

			if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
				t.Errorf("output is differ \n%v \n%v ", buf1.String(), buf2.String())
			} else {
				//fmt.Println(buf1.String())
			}
		})
	}
}

func TestLogPeerMetasMarshaler_MarshalZerologArray(t *testing.T) {
	sampleArr := make([]p2pcommon.PeerMeta, 20)
	for i := 0; i < 20; i++ {
		meta := p2pcommon.PeerMeta{}
		meta.ID = pseudoGenID()
		meta.IPAddress = fmt.Sprintf("192.168.0.%d", i)
		meta.Port = uint32(i * 1000)
		meta.Hidden = i%2 == 0
		sampleArr[i] = meta
	}
	type fields struct {
		arr   []p2pcommon.PeerMeta
		limit int
	}

	tests := []struct {
		name   string
		fields fields

		wantSize int
	}{
		{"TEmpty", fields{nil, 10}, 0},
		{"TOne", fields{sampleArr[:1], 10}, 1},
		{"TMid", fields{sampleArr[1:6], 10}, 5},
		{"TMax", fields{sampleArr[:10], 10}, 10},
		{"TOver", fields{sampleArr[0:11], 10}, 10},
		{"TOver2", fields{sampleArr, 10}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf1 := bytes.NewBuffer(nil)
			log1 := log.NewLogger("test.p2p").Output(buf1)
			buf2 := bytes.NewBuffer(nil)
			log2 := log.NewLogger("test.p2p").Output(buf2)

			m := NewLogPeerMetasMarshaller(tt.fields.arr, tt.fields.limit)
			a := zerolog.Arr()
			m.MarshalZerologArray(a)
			log1.Info().Array("t", m).Msg("Print ")
			log2.Info().Array("t", a).Msg("Print ")

			if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
				t.Errorf("output is differ \n%v \n%v ", buf1.String(), buf2.String())
			} else {
				//fmt.Println(buf1.String())
			}
		})
	}
}

func TestLogB58EncMarshaler_MarshalZerologArray(t *testing.T) {
	sampleArr := make([][]byte, 20)
	for i := 0; i < 20; i++ {
		sampleArr[i] = rand.RandomByteSlice(32)
	}
	type fields struct {
		arr   [][]byte
		limit int
	}

	tests := []struct {
		name   string
		fields fields

		wantSize int
	}{
		{"TEmpty", fields{nil, 10}, 0},
		{"TOne", fields{sampleArr[:1], 10}, 1},
		{"TMid", fields{sampleArr[1:6], 10}, 5},
		{"TMax", fields{sampleArr[:10], 10}, 10},
		{"TOver", fields{sampleArr[0:11], 10}, 10},
		{"TOver2", fields{sampleArr, 10}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf1 := bytes.NewBuffer(nil)
			log1 := log.NewLogger("test.p2p").Output(buf1)
			buf2 := bytes.NewBuffer(nil)
			log2 := log.NewLogger("test.p2p").Output(buf2)

			m := types.NewLogB58EncMarshaller(tt.fields.arr, tt.fields.limit)
			a := zerolog.Arr()
			m.MarshalZerologArray(a)
			log1.Info().Array("t", m).Msg("Print ")
			log2.Info().Array("t", a).Msg("Print ")

			if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
				t.Errorf("output is differ \n%v \n%v ", buf1.String(), buf2.String())
			} else {
				//fmt.Println(buf1.String())
			}
		})
	}
}

func pseudoGenID() types.PeerID {
	priv, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	id, _ := types.IDFromPrivateKey(priv)
	return id
}

type NumOrderer struct {
	num int
}

func (no NumOrderer) String() string {
	return fmt.Sprintf("I am no.%d", no.num)
}
