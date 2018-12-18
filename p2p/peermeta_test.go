package p2p

import (
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
)

func TestFromPeerAddress(t *testing.T) {
	type args struct {
		ip   string
		port uint32
		id   string
	}
	tests := []struct {
		name string
		args args
	}{
		{"t1", args{"192.168.1.2", 2, "id0002"}},
		{"t2", args{"0.0.0.0", 2223, "id2223"}},
		{"t3", args{"2001:0db8:85a3:08d3:1319:8a2e:0370:7334", 444, "id0002"}},
		{"t4", args{"::ffff:192.0.1.2", 444, "id0002"}},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipAddr := net.ParseIP(tt.args.ip)
			addrStr := ipAddr.String()
			addrBytes, _ := ipAddr.MarshalText()
			assert.Equal(t, addrStr, string(addrBytes))
			addr := &types.PeerAddress{Address: ipAddr, Port: tt.args.port, PeerID: []byte(tt.args.id)}
			actual := FromPeerAddress(addr)
			actualAddr := net.ParseIP(actual.IPAddress)
			assert.Equal(t, addrStr, actual.IPAddress)
			assert.Equal(t, tt.args.port, actual.Port)
			assert.Equal(t, tt.args.id, string(actual.ID))
			assert.True(t, ipAddr.Equal(actualAddr))

			actual2 := actual.ToPeerAddress()
			assert.Equal(t, *addr, actual2)
		})
	}
}

func TestPeerMeta_String(t *testing.T) {
	type fields struct {
		ip   string
		port uint32
		id   string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"t1", fields{"192.168.1.2", 2, "id0002"}},
		{"t2", fields{"0.0.0.0", 2223, "id2223"}},
		{"t3", fields{"2001:0db8:85a3:08d3:1319:8a2e:0370:7334", 444, "id0002"}},
		{"t4", fields{"::ffff:192.0.1.2", 444, "id0002"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := PeerMeta{
				IPAddress:  tt.fields.ip,
				Port:       tt.fields.port,
				ID:         peer.ID(tt.fields.id),
				Designated: false,
				Outbound:   false,
			}
			actual := m.String()
			assert.True(t, strings.Contains(actual, tt.fields.ip))
			assert.True(t, strings.Contains(actual, strconv.Itoa(int(tt.fields.port))))
			assert.True(t, strings.Contains(actual, m.ID.Pretty()))
			m2 := m
			m2.Designated = true
			assert.Equal(t, actual, m2.String())
			m3 := m
			m3.Outbound = true
			assert.Equal(t, actual, m3.String())
		})
	}
}

func TestFromMultiAddr(t *testing.T) {
	tests := []struct {
		name    string
		str    string
		wantIp  net.IP // verify one of them
		wantPort int
		wantErr bool
	}{
		{"TIP4peerAddr","/ip4/192.168.0.58/tcp/11002/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh",net.ParseIP("192.168.0.58"), 11002, false},
		{"TIP4MissingID","/ip4/192.168.0.58/tcp/11002",net.ParseIP("192.168.0.58"), -1, true},
		{"TIP4MissingPort","/ip4/192.168.0.58/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh",net.ParseIP("192.168.0.58"), 11002, true},
		{"TIP6peerAddr","/ip6/FE80::0202:B3FF:FE1E:8329/tcp/11003/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh",net.ParseIP("FE80::0202:B3FF:FE1E:8329"), 11003, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma, _ := ParseMultiaddrWithResolve(tt.str)
			got, err := FromMultiAddr(ma)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromMultiAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				ip := net.ParseIP(got.IPAddress)
				if !reflect.DeepEqual(ip, tt.wantIp) {
					t.Errorf("FromMultiAddr() = %v, want %v", ip.String(), tt.wantIp)
				}
				if !reflect.DeepEqual(got.Port, uint32(tt.wantPort)) {
					t.Errorf("FromMultiAddr() = %v, want %v", got.Port, tt.wantPort)
				}
			}

			got2, err := FromMultiAddrString(tt.str)
			if !reflect.DeepEqual(got2, got) {
				t.Errorf("result of FromMultiAddr and FromMultiAddrString differ %v, want %v", got2, got)
			}
		})
	}
}
