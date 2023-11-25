/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"bytes"
	"testing"

	"github.com/aergoio/aergo/v2/internal/network"
	"github.com/aergoio/aergo/v2/types"
	"github.com/multiformats/go-multiaddr"
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
		version  string
		noExpose bool
		outbound bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"TExpose", args{"/ip4/192.168.1.2/tcp/2", "id0002", "v1.3.0", false, false}},
		{"TNoExpose", args{"/ip4/0.0.0.0/tcp/2223", "id2223", "v1.3.0", true, false}},
		{"TOutbound", args{"/ip6/2001:0db8:85a3:08d3:1319:8a2e:0370:7334/tcp/444", "id0002", "v1.3.0", false, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &types.PeerAddress{Addresses: []string{tt.args.addr}, PeerID: []byte(tt.args.id), Version: tt.args.version}
			status := &types.Status{Sender: sender, NoExpose: tt.args.noExpose}
			ma, err := types.ParseMultiaddr(tt.args.addr)
			if err != nil {
				t.Fatalf("Wrong test input, %v is not valid input for multiaddr: err %v", tt.args.addr, err.Error())
			}

			actual := NewMetaFromStatus(status)
			if len(actual.Addresses) != 1 {
				t.Fatalf("NewMetaFromStatus len(addrs) %v , want 1", len(actual.Addresses))
			}

			if !ma.Equal(actual.Addresses[0]) {
				t.Fatalf("NewMetaFromStatus Addresses %v , want %v", actual.Addresses[0].String(), ma.String())
			}
			if string(actual.ID) != tt.args.id {
				t.Fatalf("NewMetaFromStatus ID %v , want %v", string(actual.ID), tt.args.id)
			}
			// Hidden property is not in peer address
			conv2 := FromPeerAddress(sender)
			conv2.Hidden = status.NoExpose
			if !actual.Equals(conv2) {
				t.Fatalf("FromPeerAddress ID %v , want %v", actual, FromPeerAddress(sender))
			}
		})
	}
}

func TestNewMetaWith1Addr(t *testing.T) {
	id1, id2 := types.RandomPeerID(), types.RandomPeerID()
	v1 := "v2.0.0"
	type fields struct {
		addr    string
		port    uint32
		id      types.PeerID
		version string
	}
	tests := []struct {
		name      string
		args      fields
		wantErr   bool
		wantProto int
	}{
		{"t1", fields{"192.168.1.2", 2, id1, v1}, false, multiaddr.P_IP4},
		{"t2", fields{"0.0.0.0", 2223, id2, v1}, false, multiaddr.P_IP4},
		{"t3", fields{"2001:0db8:85a3:08d3:1319:8a2e:0370:7334", 444, id1, v1}, false, multiaddr.P_IP6},
		{"t4", fields{"::ffff:192.0.1.2", 444, id1, v1}, false, multiaddr.P_IP4},
		//{"t5", fields{"www.aergo.io", 444, "id0002", v1}, false, multiaddr.P_IP4},
		{"t6", fields{"no1.blocko.com", 444, id1, v1}, false, multiaddr.P_DNS4},
		{"tErrWrongPID", fields{"192.168.1.2", 2, "id0002", v1}, false, multiaddr.P_IP4},
		{"tErrFormat", fields{"dw::ffff:192.0.1.2", 444, id1, v1}, true, multiaddr.P_IP4},
		{"tErrDomain", fields{"dw!.com", 444, id1, v1}, true, multiaddr.P_IP4},
		{"tErrWrongDomain", fields{".google.com", 444, id1, v1}, true, multiaddr.P_IP4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMetaWith1Addr(tt.args.id, tt.args.addr, tt.args.port, "v2.0.0")
			if !tt.wantErr {
				if !types.IsSamePeerID(got.ID, tt.args.id) {
					t.Errorf("NewMetaWith1Addr() ID = %v, want %v", got.ID, tt.args.id)
				}
				if !network.IsSameAddress(got.PrimaryAddress(), tt.args.addr) {
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

func TestPeerMeta_ToPeerAddress(t *testing.T) {
	id1, id2, id3 := types.RandomPeerID(), types.RandomPeerID(), types.RandomPeerID()
	v1 := "v1.3.0"
	v2 := "v2.0.0"
	ma1, _ := types.ParseMultiaddr("/ip4/192.168.1.2/tcp/7846")
	ma2, _ := types.ParseMultiaddr("/ip6/2001:0db8:85a3:08d3:1319:8a2e:0370:7334/tcp/7846")
	ma3, _ := types.ParseMultiaddr("/ip4/192.168.1.2/tcp/7846")
	type fields struct {
		ID          types.PeerID
		Role        types.PeerRole
		ProducerIDs []types.PeerID
		Addresses   []types.Multiaddr
		Version     string
		Hidden      bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"TBP", fields{id1, types.PeerRole_Producer, nil, []multiaddr.Multiaddr{ma1}, v1, false}},
		{"TAent", fields{id1, types.PeerRole_Agent, []types.PeerID{id2, id3}, []multiaddr.Multiaddr{ma1, ma2}, v1, false}},
		{"TBP", fields{id1, types.PeerRole_Watcher, nil, []multiaddr.Multiaddr{ma1}, v1, false}},
		{"TDiffVer", fields{id1, types.PeerRole_Producer, nil, []multiaddr.Multiaddr{ma3}, v2, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := PeerMeta{
				ID:          tt.fields.ID,
				Role:        tt.fields.Role,
				ProducerIDs: tt.fields.ProducerIDs,
				Addresses:   tt.fields.Addresses,
				Version:     tt.fields.Version,
				Hidden:      tt.fields.Hidden,
			}
			got := m.ToPeerAddress()
			gotID := types.PeerID(got.PeerID)
			if !types.IsSamePeerID(gotID, m.ID) {
				t.Errorf("ToPeerAddress() version = %v, want %v", got.Version, m.Version)
			}
			if got.Role != m.Role {
				t.Errorf("ToPeerAddress() version = %v, want %v", got.Version, m.Version)
			}
			if got.Version != m.Version {
				t.Errorf("ToPeerAddress() version = %v, want %v", got.Version, m.Version)
			}
			got2 := FromPeerAddress(&got)
			if !got2.Equals(m) {
				t.Errorf("FromPeerAddress() = %v, want %v", got2, m)

			}
		})
	}
}
