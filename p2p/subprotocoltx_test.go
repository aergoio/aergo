/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package p2p

import (
	"fmt"
	"testing"
	"time"

	"github.com/aergoio/aergo/internal/enc"

	"github.com/aergoio/aergo-actor/actor"
	uuid "github.com/satori/go.uuid"
)

var _ ActorService = (*mockIServ)(nil)

type mockIServ struct {
}

func (m mockIServ) SendRequest(actor string, msg interface{}) {

}

func (m mockIServ) CallRequest(actor string, msg interface{}) (interface{}, error) {
	return nil, nil
}
func (m mockIServ) FutureRequest(actor string, msg interface{}) *actor.Future {
	return nil
}

const hashSize = 32

func BenchmarkArrayKey(b *testing.B) {
	size := 100000
	var samples = make([]([hashSize]byte), size)
	for i := 0; i < size; i++ {
		copy(samples[i][:], uuid.Must(uuid.NewV4()).Bytes())
		copy(samples[i][16:], uuid.Must(uuid.NewV4()).Bytes())
	}

	b.Run("BArray", func(b *testing.B) {
		var keyArr [hashSize]byte
		startTime := time.Now()
		fmt.Printf("P1 in byte array\n")
		target := make(map[[hashSize]byte]int)
		for i := 0; i < size; i++ {
			copy(keyArr[:], samples[i][:])
			target[keyArr] = i
		}
		endTime := time.Now()
		fmt.Printf("Takes %f sec \n", endTime.Sub(startTime).Seconds())
	})

	b.Run("Bbase64", func(b *testing.B) {
		startTime := time.Now()
		fmt.Printf("P2 in base64\n")
		target2 := make(map[string]int)
		for i := 0; i < size; i++ {
			target2[enc.ToString(samples[i][:])] = i
		}
		endTime := time.Now()
		fmt.Printf("Takes %f sec\n", endTime.Sub(startTime).Seconds())

	})

}

func Test_bytesArrToString(t *testing.T) {
	t.SkipNow()
	type args struct {
		bbarray [][]byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "tsucc-01", args: args{[][]byte{[]byte("abcde"), []byte("12345")}}, want: "[\"YWJjZGU=\",\"MTIzNDU=\",]"},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bytesArrToString(tt.args.bbarray); got != tt.want {
				t.Errorf("bytesArrToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
