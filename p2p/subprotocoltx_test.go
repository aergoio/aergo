/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package p2p

import (
	"testing"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
	inet "github.com/libp2p/go-libp2p-net"
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

func Test_bytesArrToString(t *testing.T) {
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

func TestTxProtocol_onGetTXsRequest(t *testing.T) {
	type fields struct {
		iserv    ActorService
		ps       PeerManager
		requests map[string]*types.MessageData
		log      *log.Logger
	}
	type args struct {
		s inet.Stream
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &TxProtocol{
				iserv: tt.fields.iserv,
				ps:    tt.fields.ps,
				log:   tt.fields.log,
			}
			p.onGetTXsRequest(tt.args.s)
		})
	}
}
