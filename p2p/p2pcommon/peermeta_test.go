/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"bytes"
	"github.com/aergoio/aergo/internal/network"
	"github.com/multiformats/go-multiaddr"
	"testing"

	"github.com/aergoio/aergo/types"
)

func TestFromPeerAddressLegacy(t *testing.T) {
	type args struct {
		addr string
		id   string
	}
	tests := []struct {
		name string
		args args
	}{
		{"t1", args{"/ip4/192.168.1.2/tcp/2", "id0002"}},
		{"t2", args{"/ip4/0.0.0.0/tcp/2223", "id2223"}},
		{"t3", args{"/ip6/2001:0db8:85a3:08d3:1319:8a2e:0370:7334/tcp/444", "id0002"}},
		{"t4", args{"/ip6/::ffff:192.0.1.2/tcp/444", "id0002"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := &types.PeerAddress{Addresses: []string{tt.args.addr}, PeerID: []byte(tt.args.id)}
			actual := FromPeerAddress(addr)
			if len(actual.Addresses) != 1 {
				t.Fatalf("FromPeerAddress len(addrs) %v , want 1", len(actual.Addresses))
			}
			ma, err := types.ParseMultiaddr(tt.args.addr)
			if err != nil {
				t.Fatalf("Wrong test input, %v is not valid input for multiaddr: err %v", tt.args.addr, err.Error())
			}
			if !ma.Equal(actual.Addresses[0]) {
				t.Fatalf("FromPeerAddress Addresses %v , want %v", actual.Addresses[0].String(), ma.String())
			}

			if string(actual.ID) != tt.args.id {
				t.Fatalf("FromPeerAddress ID %v , want %v", string(actual.ID), tt.args.id)
			}

			actual2 := actual.ToPeerAddress()
			if !bytes.Equal(addr.PeerID, actual2.PeerID) {
				t.Fatalf("ToPeerAddress %v , want %v", actual2.String(), addr.String())
			}
		})
	}
}

func TestNewMetaFromStatus(t *testing.T) {
	type args struct {
		addr     string
		id       string
		noExpose bool
		outbound bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"TExpose", args{"/ip4/192.168.1.2/tcp/2", "id0002", false, false}},
		{"TNoExpose", args{"/ip4/0.0.0.0/tcp/2223", "id2223", true, false}},
		{"TOutbound", args{"/ip6/2001:0db8:85a3:08d3:1319:8a2e:0370:7334/tcp/444", "id0002", false, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &types.PeerAddress{Addresses: []string{tt.args.addr}, PeerID: []byte(tt.args.id)}
			status := &types.Status{Sender: sender, NoExpose: tt.args.noExpose}
			ma, err := types.ParseMultiaddr(tt.args.addr)
			if err != nil {
				t.Fatalf("Wrong test input, %v is not valid input for multiaddr: err %v", tt.args.addr, err.Error())
			}

			actual := NewMetaFromStatus(status, tt.args.outbound)
			if len(actual.Addresses) != 1 {
				t.Fatalf("FromPeerAddress len(addrs) %v , want 1", len(actual.Addresses))
			}

			if !ma.Equal(actual.Addresses[0]) {
				t.Fatalf("FromPeerAddress Addresses %v , want %v", actual.Addresses[0].String(), ma.String())
			}
			if string(actual.ID) != tt.args.id {
				t.Fatalf("FromPeerAddress ID %v , want %v", string(actual.ID), tt.args.id)
			}
		})
	}
}

func TestNewMetaWith1Addr(t *testing.T) {
	id1, id2 := types.RandomPeerID(), types.RandomPeerID()
	type fields struct {
		addr string
		port uint32
		id   types.PeerID
	}
	tests := []struct {
		name      string
		args      fields
		wantErr   bool
		wantProto int
	}{
		{"t1", fields{"192.168.1.2", 2, id1}, false, multiaddr.P_IP4},
		{"t2", fields{"0.0.0.0", 2223, id2}, false, multiaddr.P_IP4},
		{"t3", fields{"2001:0db8:85a3:08d3:1319:8a2e:0370:7334", 444, id1}, false, multiaddr.P_IP6},
		{"t4", fields{"::ffff:192.0.1.2", 444, id1}, false, multiaddr.P_IP4},
		//{"t5", fields{"www.aergo.io", 444, "id0002"}, false, multiaddr.P_IP4},
		{"t6", fields{"no1.blocko.com", 444, id1}, false, multiaddr.P_DNS4},
		{"tErrWrongPID", fields{"192.168.1.2", 2, "id0002"}, false, multiaddr.P_IP4},
		{"tErrFormat", fields{"dw::ffff:192.0.1.2", 444, id1}, true, multiaddr.P_IP4},
		{"tErrDomain", fields{"dw!.com", 444, id1}, true, multiaddr.P_IP4},
		{"tErrWrongDomain", fields{".google.com", 444, id1}, true, multiaddr.P_IP4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMetaWith1Addr(tt.args.id, tt.args.addr, tt.args.port)
			if !tt.wantErr {
				if !types.IsSamePeerID(got.ID,tt.args.id) {
					t.Errorf("NewMetaWith1Addr() ID = %v, want %v", got.ID, tt.args.id)
				}
				if !network.IsSameAddress(got.PrimaryAddress(),tt.args.addr) {
					t.Errorf("NewMetaWith1Addr() addr = %v, want %v", got.PrimaryAddress(), tt.args.addr)
				}
				if got.PrimaryPort() != tt.args.port {
					t.Errorf("NewMetaWith1Addr() port = %v, want %v", got.PrimaryPort(), tt.args.port)
				}
			} else {
				if len(got.Addresses) != 0 {
					t.Errorf("NewMetaWith1Addr() = %v, want error", got.Addresses[0].String())
				}
			}
		})
	}
}
