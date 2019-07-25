/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"github.com/multiformats/go-multiaddr"
	"net"
	"reflect"
	"testing"
)

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
		//FIXME
		// skip case, since it depend on external environment. uncomment it if really need, and comment back after finishing test
		//{"TDomainName", "/dns/aergo.io/tcp/11004/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", []net.IP{net.ParseIP("104.20.161.59"), net.ParseIP("104.20.160.59")}, "11004", false},
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
						break
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
