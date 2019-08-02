/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package enterprise

import (
	"net"
	"testing"

	"github.com/aergoio/aergo/types"
)

func TestNewWhiteListEntry(t *testing.T) {
	tests := []struct {
		name string
		arg  string

		wantErr  bool
		wantID   bool
		wantAddr bool
	}{
		{"TAll", "{\"peerid\":\"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt\", \"address\":\"172.21.3.35\" }", false, true, true},
		{"TAll2", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"172.21.3.35", "cidr":""}`, false, true, true},
		//{"TAll2", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"", "cidr":"172.21.3.35/32"}`, false, true, true, false},

		{"TIDOnly", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt"}`, false, true, false},
		{"TIDOnly2", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":""}`, false, true,  false},
		{"TIDOnly3", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"", "cidr":""}`, false, true,  false},

		{"TIAddrOnly", `{"address":"172.21.3.35"}`, false, false, true},
		{"TIAddrOnly2", `{"address":"::0123:4567:89ab:cdef:1234:5678"}`, false, false, true},
		{"TIAddrRange", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "cidr":"172.21.3.35/24" }`, false, true, true },
		{"TIAddrRange2", `{"cidr":"2001:0db8:0123:4567:89ab:cdef:1234:5678/96"}`, false, false, true},
		{"TEmpty", ":", true, false, false},
		{"TWrongFormat", `"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt"`, true, false, false},
		{"TWrongFormat2", `"172.21.3.35/24"`, true, false, false},
		{"TWrongID", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBRf3mBFBJpz3te@GGt"}`, true, false, false},
		{"TWrongAddr", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"172.21.3.355"}`, true, false, false},
		{"TWrongAddr2", `{"cidr":":12001:0db8:0123:4567:89ab:cdef:1234:5678/96"}`, true, false, false},
		{"TWrongEnt", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"172.21.3.35", "cidr":"172.21.3.35/24"}`, true, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWhiteListEntry(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWhiteListEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if (got.IpNet != notSpecifiedCIDR) != tt.wantAddr {
					t.Errorf("NewWhiteListEntry().IpNet = %v, want not", got.IpNet)
				}
				if (got.PeerID != NotSpecifiedID) != tt.wantID {
					t.Errorf("NewWhiteListEntry().PeerID = %v, want not", got.PeerID.Pretty())
				}
			}
		})
	}
}

func TestWhiteListEntry_Contains(t *testing.T) {
	sampleIDStr := "16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt"
	sampleID, _ := types.IDB58Decode(sampleIDStr)
	otherID, _ := types.IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
	sampleIP4Str := "122.1.3.4"
	sampleIP4 := net.ParseIP(sampleIP4Str)
	sampleIP6 := net.ParseIP("2001:0db8:0123:4567:89ab:cdef:1234:5678")
	IDOnly := `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt"}`
	AddrOnly := `{"address":"122.1.3.4"}`
	AddrRange := `{"peerid":"", "cidr":"122.1.3.4/24"}`
	IDAddr := `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"122.1.3.4"}`
	IDAdRange24 := `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "cidr":"122.1.3.4/24"}`
	IDAdRange16 := `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "cidr":"122.1.3.4/16"}`
	IDAdRange8 := `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "cidr":"122.1.3.4/8"}`
	type args struct {
		pid  types.PeerID
		addr net.IP
	}
	tests := []struct {
		name   string
		cParam string
		args   args
		want   bool
	}{
		// match everything
		{"TSIDAndAddr", IDAddr, args{sampleID, sampleIP4}, true},
		{"TSIDAndAddr", IDAddr, args{sampleID, sampleIP6}, false},
		{"TSIDAndAddr", IDAddr, args{otherID, sampleIP4}, false},

		// test only id,  not care Address
		{"TSIDMatch", IDOnly, args{sampleID, sampleIP4}, true},
		{"TSIDMatch2", IDOnly, args{sampleID, sampleIP6}, true},
		{"TSIDNoMatch", IDOnly, args{otherID, sampleIP4}, false},

		// test only addr. not care PeerID
		{"TSAddrMatch", AddrOnly, args{sampleID, sampleIP4}, true},
		{"TSAddrMatch2", AddrOnly, args{otherID, sampleIP4}, true},
		{"TSAddrMatch3", AddrOnly + "/32", args{otherID, sampleIP4}, true},
		{"TSAddrNoMatch", AddrOnly, args{otherID, sampleIP6}, false},
		// test addr range.  not care PeerID
		{"T24T", AddrRange, args{otherID, net.ParseIP("122.1.3.251")}, true},
		{"T24F", AddrRange, args{otherID, net.ParseIP("122.1.4.251")}, false},

		// test everything. both id and addr should match
		{"TR24T", IDAdRange24, args{sampleID, net.ParseIP("122.1.3.251")}, true},
		{"TDiffID1", IDAdRange24, args{otherID, net.ParseIP("122.1.3.251")}, false},
		{"TR24F1", IDAdRange24, args{sampleID, net.ParseIP("122.1.2.4")}, false},
		{"TR24F2", IDAdRange24, args{sampleID, net.ParseIP("122.1.4.4")}, false},
		{"TR16T", IDAdRange16, args{sampleID, net.ParseIP("122.1.4.251")}, true},
		{"TR16F1", IDAdRange16, args{sampleID, net.ParseIP("122.2.3.251")}, false},
		{"TR16F2", IDAdRange16, args{sampleID, net.ParseIP("122.0.3.251")}, false},
		{"TR8T", IDAdRange8, args{sampleID, net.ParseIP("122.2.33.251")}, true},
		{"TR8F1", IDAdRange8, args{sampleID, net.ParseIP("121.1.3.251")}, false},
		{"TR8F2", IDAdRange8, args{sampleID, net.ParseIP("123.1.3.251")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, _ := NewWhiteListEntry(tt.cParam)

			if got := e.Contains(tt.args.addr, tt.args.pid); got != tt.want {
				t.Errorf("WhiteListEntry.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
