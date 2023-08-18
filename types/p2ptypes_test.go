/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"net"
	"reflect"
	"testing"

	"github.com/aergoio/aergo/v2/internal/network"
	"github.com/multiformats/go-multiaddr"
)

func TestParseMultiaddrWithResolve(t *testing.T) {
	tests := []struct {
		name      string
		str       string
		wantProto int
		wantIps   []net.IP // verify one of them
		wantPort  string
		wantErr   bool
	}{
		{"TIP4peerAddr", "/ip4/192.168.0.58/tcp/11002/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", multiaddr.P_IP4, []net.IP{net.ParseIP("192.168.0.58")}, "11002", false},
		{"TIP4AndPort", "/ip4/192.168.0.58/tcp/11002", multiaddr.P_IP4, []net.IP{net.ParseIP("192.168.0.58")}, "11002", false},
		{"TIP6peerAddr", "/ip6/FE80::0202:B3FF:FE1E:8329/tcp/11003/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", multiaddr.P_IP6, []net.IP{net.ParseIP("FE80::0202:B3FF:FE1E:8329")}, "11003", false},
		//FIXME
		{"TDomainName", "/dns/aergo.io/tcp/11004/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", multiaddr.P_DNS4, []net.IP{}, "11004", false},
		{"TInvalidDomain", "/dns4/!nowhere.a.aergo.io/tcp/11004/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", multiaddr.P_DNS4, []net.IP{}, "11004", false},
		{"TWrongProto", "/ipx/192.168.0.58/tcp/11004/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", multiaddr.P_DNS, []net.IP{}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMultiaddr(tt.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMultiaddrWithResolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Errorf("ParseMultiaddrWithResolve() is nil , want not nil")
				}

				ipStr, err := got.ValueForProtocol(tt.wantProto)
				if err != nil {
					t.Errorf("ParseMultiaddrWithResolve() no proto found , want %v", multiaddr.ProtocolWithCode(tt.wantProto))
				} else {

				}
				if len(tt.wantIps) > 0 {
					ip := net.ParseIP(ipStr)
					found := false
					for _, wantIp := range tt.wantIps {
						if reflect.DeepEqual(ip, wantIp) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("ParseMultiaddrWithResolve() = %v, want %v", ip, tt.wantIps)
					}
				}
				port, err := got.ValueForProtocol(multiaddr.P_TCP)
				if !reflect.DeepEqual(port, tt.wantPort) {
					t.Errorf("ParseMultiaddrWithResolve() = %v, want %v", port, tt.wantPort)
				}
			}
		})
	}
}

func TestPeerToMultiAddr(t *testing.T) {
	id1, id2 := RandomPeerID(), RandomPeerID()
	type args struct {
		address string
		port    uint32
		pid     PeerID
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		wantProto int
	}{
		{"t1", args{"192.168.1.2", 2, id1}, false, multiaddr.P_IP4},
		{"t2", args{"0.0.0.0", 2223, id2}, false, multiaddr.P_IP4},
		{"t3", args{"2001:0db8:85a3:08d3:1319:8a2e:0370:7334", 444, id1}, false, multiaddr.P_IP6},
		{"t4", args{"::ffff:192.0.1.2", 444, id1}, false, multiaddr.P_IP4},
		//{"t5", fields{"www.aergo.io", 444, "id0002"}, false, multiaddr.P_IP4},
		{"t6", args{"no1.blocko.com", 444, id1}, false, multiaddr.P_DNS4},
		{"tErrFormat", args{"dw::ffff:192.0.1.2", 444, id1}, true, multiaddr.P_IP4},
		{"tErrDomain", args{"dw!.com", 444, id1}, true, multiaddr.P_IP4},
		{"tErrWrongDomain", args{".google.com", 444, id1}, true, multiaddr.P_IP4},
		{"tErrWrongPID", args{"192.168.1.2", 2, "id0002"}, true, multiaddr.P_IP4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PeerToMultiAddr(tt.args.address, tt.args.port, tt.args.pid)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerToMultiAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				val, err := got.ValueForProtocol(tt.wantProto)
				if err != nil {
					t.Errorf("PeerToMultiAddr() missing proto = %v", multiaddr.ProtocolWithCode(tt.wantProto))
				}
				if !network.IsSameAddress(val, tt.args.address) {
					t.Errorf("PeerToMultiAddr() got = %v, want %v", got, tt.args.address)
				}
			}
		})
	}
}

func TestToMultiAddr(t *testing.T) {
	type args struct {
		address string
		port    uint32
	}
	tests := []struct {
		name string
		args args

		wantErr   bool
		wantProto int
	}{
		{"t1", args{"192.168.1.2", 2}, false, multiaddr.P_IP4},
		{"t2", args{"0.0.0.0", 2223}, false, multiaddr.P_IP4},
		{"t3", args{"2001:0db8:85a3:08d3:1319:8a2e:0370:7334", 444}, false, multiaddr.P_IP6},
		{"t4", args{"::ffff:192.0.1.2", 444}, false, multiaddr.P_IP4},
		//{"t5", fields{"www.aergo.io", 444, "id0002"}, false, multiaddr.P_IP4},
		{"t6", args{"no1.blocko.com", 444}, false, multiaddr.P_DNS4},
		{"tErrFormat", args{"dw::ffff:192.0.1.2", 444}, true, multiaddr.P_IP4},
		{"tErrDomain", args{"dw!.com", 444}, true, multiaddr.P_IP4},
		{"tErrWrongDomain", args{".google.com", 444}, true, multiaddr.P_IP4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToMultiAddr(tt.args.address, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToMultiAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				val, err := got.ValueForProtocol(tt.wantProto)
				if err != nil {
					t.Errorf("PeerToMultiAddr() missing proto = %v", multiaddr.ProtocolWithCode(tt.wantProto))
				}
				if !network.IsSameAddress(val, tt.args.address) {
					t.Errorf("PeerToMultiAddr() got = %v, want %v", got, tt.args.address)
				}
			}
		})
	}
}

func TestGetIPFromMultiaddr(t *testing.T) {
	tests := []struct {
		name      string
		arg       string
		wantProto int
		want      net.IP
	}{
		{"TIP4peerAddr", "/ip4/192.168.0.58/tcp/11002/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", multiaddr.P_IP4, net.ParseIP("192.168.0.58")},
		{"TIP4AndPort", "/ip4/192.168.0.58/tcp/11002", multiaddr.P_IP4, net.ParseIP("192.168.0.58")},
		{"TIP6peerAddr", "/ip6/FE80::0202:B3FF:FE1E:8329/tcp/11003/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", multiaddr.P_IP6, net.ParseIP("FE80::0202:B3FF:FE1E:8329")},
		{"TDomainName", "/dns/aergo.io/tcp/11004/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", multiaddr.P_DNS4, nil},
		{"TInvalidDomain", "/dns4/!nowhere.a.aergo.io/tcp/11004/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", multiaddr.P_DNS4, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma, _ := multiaddr.NewMultiaddr(tt.arg)
			got := GetIPFromMultiaddr(ma)

			if !got.Equal(tt.want) {
				t.Errorf("GetIPFromMultiaddr() = %v, want %v", got, tt.want)
			}
			if tt.want != nil {
				switch tt.wantProto {
				case multiaddr.P_IP4:
					if got.To4() == nil {
						t.Error("GetIPFromMultiaddr() return not ip4, but want")
					}
				case multiaddr.P_IP6:
					if got.To16() == nil {
						t.Error("GetIPFromMultiaddr() return not ip6, but want")
					}
				default:
					t.Error("Test assumption is wrong. input must be ip address")
				}
			}
		})
	}
}

func TestExtractIPAddress(t *testing.T) {
	tests := []struct {
		name  string
		mastr string

		want net.IP
	}{
		{"TIP4peerAddr", "/ip4/192.168.0.58/tcp/11002/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", net.ParseIP("192.168.0.58")},
		{"TIP4AndPort", "/ip4/192.168.0.58/tcp/11002", net.ParseIP("192.168.0.58")},
		{"TMissingAddr", "/tcp/11002", nil},
		{"TIP4Only", "/ip4/192.168.0.58", net.ParseIP("192.168.0.58")},
		{"TIP6peerAddr", "/ip6/FE80::0202:B3FF:FE1E:8329/tcp/11003/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", net.ParseIP("FE80::0202:B3FF:FE1E:8329")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma, _ := multiaddr.NewMultiaddr(tt.mastr)
			got := GetIPFromMultiaddr(ma)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractIPAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
