/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"net"
	"reflect"
	"testing"
)

func TestFromMultiAddr(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		wantIp   net.IP // verify one of them
		wantPort int
		wantErr  bool
	}{
		{"TIP4peerAddr", "/ip4/192.168.0.58/tcp/11002/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", net.ParseIP("192.168.0.58"), 11002, false},
		{"TIP4MissingID", "/ip4/192.168.0.58/tcp/11002", net.ParseIP("192.168.0.58"), -1, true},
		{"TIP4MissingPort", "/ip4/192.168.0.58/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", net.ParseIP("192.168.0.58"), 11002, true},
		{"TIP6peerAddr", "/ip6/FE80::0202:B3FF:FE1E:8329/tcp/11003/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", net.ParseIP("FE80::0202:B3FF:FE1E:8329"), 11003, false},
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

			got2, err := ParseMultiAddrString(tt.str)
			if !reflect.DeepEqual(got2, got) {
				t.Errorf("result of FromMultiAddr and ParseMultiAddrString differ %v, want %v", got2, got)
			}
		})
	}
}

func TestPeerMeta_ToMultiAddr(t *testing.T) {
	type fields struct {
		IPAddress string
		Port      uint32
		ID        peer.ID
	}
	tests := []struct {
		name         string
		fields       fields
		wantErr      bool
		wantProtocol int
	}{
		{"t1", fields{"192.168.1.2", 2, "id0002"}, false, multiaddr.P_IP4},
		{"t2", fields{"0.0.0.0", 2223, "id2223"}, false, multiaddr.P_IP4},
		{"t3", fields{"2001:0db8:85a3:08d3:1319:8a2e:0370:7334", 444, "id0002"}, false, multiaddr.P_IP6},
		{"t4", fields{"::ffff:192.0.1.2", 444, "id0002"}, false, multiaddr.P_IP4},
		{"t5", fields{"iparkmac.blocko.io", 444, "id0002"}, false, multiaddr.P_IP4},
		{"t6", fields{"no1.blocko.com", 444, "id0002"}, true, multiaddr.P_IP4},
		{"tErr", fields{"dw::ffff:192.0.1.2", 444, "id0002"}, true, multiaddr.P_IP4},
		{"tErr2", fields{"dw!.com", 444, "id0002"}, true, multiaddr.P_IP4},
		{"tErr3", fields{".google.com", 444, "id0002"}, true, multiaddr.P_IP4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := PeerMeta{
				ID:        tt.fields.ID,
				IPAddress: tt.fields.IPAddress,
				Port:      tt.fields.Port,
			}
			got, err := PeerMetaToMultiAddr(m)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerMeta.PeerMetaToMultiAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				val, err := got.ValueForProtocol(tt.wantProtocol)
				assert.Nil(t, err)
				expIp := net.ParseIP(tt.fields.IPAddress)
				if expIp == nil {
					ips, _ := p2putil.ResolveHostDomain(tt.fields.IPAddress)
					expIp = ips[0]
				}
				assert.Equal(t, expIp, net.ParseIP(val))
			}
		})
	}
}

func TestParseMultiaddrWithResolve(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		wantIps  []net.IP // verify one of them
		wantPort string
		wantErr  bool
	}{
		{"TIP4peerAddr", "/ip4/192.168.0.58/tcp/11002/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", []net.IP{net.ParseIP("192.168.0.58")}, "11002", false},
		{"TIP4AndPort", "/ip4/192.168.0.58/tcp/11002", []net.IP{net.ParseIP("192.168.0.58")}, "11002", false},
		{"TIP6peerAddr", "/ip6/FE80::0202:B3FF:FE1E:8329/tcp/11003/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", []net.IP{net.ParseIP("FE80::0202:B3FF:FE1E:8329")}, "11003", false},
		{"TDomainName", "/dns/aergo.io/tcp/11004/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", []net.IP{net.ParseIP("104.20.161.59"), net.ParseIP("104.20.160.59")}, "11004", false},
		{"TInvalidDomain", "/dns/nowhere.a.aergo.io/tcp/11004/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", []net.IP{}, "", true},
		{"TWrongProto", "/ipx/192.168.0.58/tcp/11004/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", []net.IP{}, "", true},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMultiaddrWithResolve(tt.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMultiaddrWithResolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Errorf("ParseMultiaddrWithResolve() is nil , want not nil")
				}
				ipStr, err := got.ValueForProtocol(multiaddr.P_IP4)
				if err != nil {
					ipStr, _ = got.ValueForProtocol(multiaddr.P_IP6)
				}
				ip := net.ParseIP(ipStr)
				found := false
				for _, wantIp := range tt.wantIps {
					if reflect.DeepEqual(ip, wantIp) {
						found = true
						break;
					}
				}
				if !found {
					t.Errorf("ParseMultiaddrWithResolve() = %v, want %v", ip, tt.wantIps)
				}
				port, err := got.ValueForProtocol(multiaddr.P_TCP)
				if !reflect.DeepEqual(port, tt.wantPort) {
					t.Errorf("ParseMultiaddrWithResolve() = %v, want %v", port, tt.wantPort)
				}
			}
		})
	}
}
