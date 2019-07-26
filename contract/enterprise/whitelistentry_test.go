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
		{"TAll", "16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt:172.21.3.35", false, true, true},
		{"TIDOnly", "16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt:", false, true, false},
		{"TIAddrOnly", ":172.21.3.35", false, false, true},
		{"TIAddrOnly2", ":::0123:4567:89ab:cdef:1234:5678", false, false, true},
		{"TIAddrRange", "16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt:172.21.3.35/24", false, true, true},
		{"TIAddrRange2", ":2001:0db8:0123:4567:89ab:cdef:1234:5678/96", false, false, true},
		{"TEmpty", ":", false, false, false},
		{"TWrongFormat", "16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", true, false, false},
		{"TWrongFormat2", "172.21.3.35/24", true, false, false},
		{"TWrongID", "16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBRf3mBFBJpz3te@GGt:", true, false, false},
		{"TWrongAddr", "16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt:172.21.3.355", true, false, false},
		{"TWrongAddr2", ":12001:0db8:0123:4567:89ab:cdef:1234:5678/96", true, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWhiteListEntry(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWhiteListEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if (got.IpNet != notSpecifiedAddr) != tt.wantAddr {
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
	IDOnly := "16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt:"
	AddrOnly := ":122.1.3.4"
	IDAddr := "16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt:122.1.3.4"
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
		{"TNotSpec1", ":", args{sampleID, sampleIP4}, true},
		{"TNotSpec2", ":", args{sampleID, sampleIP6}, true},
		{"TNotSpec3", ":", args{otherID, sampleIP4}, true},
		{"TNotSpec4", ":", args{otherID, sampleIP6}, true},

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
		{"T24T", AddrOnly + "/24", args{otherID, net.ParseIP("122.1.3.251")}, true},
		{"T24F", AddrOnly + "/24", args{otherID, net.ParseIP("122.1.4.251")}, false},

		// test everything. both id and addr should match
		{"TR24T", IDAddr + "/24", args{sampleID, net.ParseIP("122.1.3.251")}, true},
		{"TDiffID1", IDAddr + "/24", args{otherID, net.ParseIP("122.1.3.251")}, false},
		{"TR24F1", IDAddr + "/24", args{sampleID, net.ParseIP("122.1.2.4")}, false},
		{"TR24F2", IDAddr + "/24", args{sampleID, net.ParseIP("122.1.4.4")}, false},
		{"TR16T", IDAddr + "/16", args{sampleID, net.ParseIP("122.1.4.251")}, true},
		{"TR16F1", IDAddr + "/16", args{sampleID, net.ParseIP("122.2.3.251")}, false},
		{"TR16F2", IDAddr + "/16", args{sampleID, net.ParseIP("122.0.3.251")}, false},
		{"TR8T", IDAddr + "/8", args{sampleID, net.ParseIP("122.2.33.251")}, true},
		{"TR8F1", IDAddr + "/8", args{sampleID, net.ParseIP("121.1.3.251")}, false},
		{"TR8F2", IDAddr + "/8", args{sampleID, net.ParseIP("123.1.3.251")}, false},
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
