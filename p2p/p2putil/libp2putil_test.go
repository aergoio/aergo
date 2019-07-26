/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/aergoio/aergo/internal/network"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
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
			ma, _ := types.ParseMultiaddrWithResolve(tt.str)
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

func TestFromMultiAddrStringWithPID(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		idStr    string
		wantIp   net.IP // verify one of them
		wantPort int
		wantErr  bool
	}{
		{"TIP4peerAddr", "/ip4/192.168.0.58/tcp/11002", "16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", net.ParseIP("192.168.0.58"), 11002, false},
		{"TMissingAddr", "/tcp/11002", "16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", net.ParseIP("192.168.0.58"), 11002, true},
		{"TIP4MissingPort", "/ip4/192.168.0.58", "16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", net.ParseIP("192.168.0.58"), 11002, true},
		{"TIP6peerAddr", "/ip6/FE80::0202:B3FF:FE1E:8329/tcp/11003", "16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", net.ParseIP("FE80::0202:B3FF:FE1E:8329"), 11003, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := types.IDB58Decode(tt.idStr)
			if err != nil {
				t.Fatalf("parse id error %v ", err)
			}
			got, err := FromMultiAddrStringWithPID(tt.str, id)
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

			ma, _ := types.ParseMultiaddrWithResolve(tt.str + "/p2p/" + tt.idStr)
			got2, err := FromMultiAddr(ma)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromMultiAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got2, got) {
				t.Errorf("result of FromMultiAddrStringWithPID and FromMultiAddr differ %v, want %v", got2, got)
			}

		})
	}
}

func TestPeerMeta_ToMultiAddr(t *testing.T) {
	type fields struct {
		IPAddress string
		Port      uint32
		ID        types.PeerID
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
		//{"t5", fields{"www.aergo.io", 444, "id0002"}, false, multiaddr.P_IP4},
		{"t6", fields{"no1.blocko.com", 444, "id0002"}, true, multiaddr.P_IP4},
		{"tErr", fields{"dw::ffff:192.0.1.2", 444, "id0002"}, true, multiaddr.P_IP4},
		{"tErr2", fields{"dw!.com", 444, "id0002"}, true, multiaddr.P_IP4},
		{"tErr3", fields{".google.com", 444, "id0002"}, true, multiaddr.P_IP4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := p2pcommon.PeerMeta{
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
					ips, _ := network.ResolveHostDomain(tt.fields.IPAddress)
					expIp = ips[0]
				}
				assert.Equal(t, expIp, net.ParseIP(val))
			}
		})
	}
}

func TestGenerateKeyFile(t *testing.T) {
	// this test should not be run by root
	testDir := "_tmp"
	existDir := "_tmp/holder.key"
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Skip("can't test since test dir already exists")
	} else {
		err := os.MkdirAll(testDir, os.ModePerm)
		if err != nil {
			t.Skip("can't test. permission error. " + err.Error())
		}
		defer func() {
			os.RemoveAll(testDir)
		}()
		err = os.MkdirAll(existDir, os.ModePerm)
		if err != nil {
			t.Skip("can't test. permission error. " + err.Error())
		}
	}
	type args struct {
		dir    string
		prefix string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"TSucc", args{filepath.Join(testDir, "abc"), "testkey"}, false},
		{"TPermission", args{"/sbin/abc", "testkey"}, true},
		{"TDir", args{testDir, "holder"}, true},
		//{"TNotExist", args{}, true},
		//{"TInvalidKey", args{}, false},
		//{"TInvalidKey", args{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPriv, gotPub, err := GenerateKeyFile(tt.args.dir, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateKeyFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			file := filepath.Join(tt.args.dir, tt.args.prefix+p2pcommon.DefaultPkKeyExt)
			if !tt.wantErr {
				if !gotPriv.GetPublic().Equals(gotPub) {
					t.Errorf("priv key and pub key check failed")
					return
				}

				ldPriv, ldPub, actErr := LoadKeyFile(file)
				if actErr != nil {
					t.Errorf("LoadKeyFile() should not return error, but get %v, ", actErr)
				}
				if !ldPriv.Equals(gotPriv) {
					t.Errorf("GenerateKeyFile() and LoadKeyFile() private key is differ.")
					return
				}
				if !ldPub.Equals(gotPub) {
					t.Errorf("GenerateKeyFile() and LoadKeyFile() public key is differ.")
					return
				}
			}
		})
	}
}

func TestExtractIPAddress(t *testing.T) {
	tests := []struct {
		name string
		mastr string

		want net.IP
	}{
		{"TIP4peerAddr", "/ip4/192.168.0.58/tcp/11002/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", net.ParseIP("192.168.0.58") },
		{"TIP4AndPort", "/ip4/192.168.0.58/tcp/11002", net.ParseIP("192.168.0.58") },
		{"TMissingAddr", "/tcp/11002", nil},
		{"TIP4Only", "/ip4/192.168.0.58", net.ParseIP("192.168.0.58"),},
		{"TIP6peerAddr", "/ip6/FE80::0202:B3FF:FE1E:8329/tcp/11003/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh", net.ParseIP("FE80::0202:B3FF:FE1E:8329")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma, _ := types.ParseMultiaddrWithResolve(tt.mastr)
			if got := ExtractIPAddress(ma); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractIPAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
