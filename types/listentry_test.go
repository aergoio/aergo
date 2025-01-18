/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"bytes"
	"fmt"
	"net"
	"testing"
)

type addrType int

const (
	none addrType = iota
	addr
	cidr
)

func TestNewListEntry(t *testing.T) {
	tests := []struct {
		name string
		arg  string

		wantErr  bool
		wantID   bool
		wantAddr addrType
	}{
		{"TAll", "{\"peerid\":\"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt\", \"address\":\"172.21.3.35\" }", false, true, addr},
		{"TAll2", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"172.21.3.35", "cidr":""}`, false, true, addr},
		//{"TAll2", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"", "cidr":"172.21.3.35/32"}`, false, true, true, false},

		{"TIDOnly", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt"}`, false, true, none},
		{"TIDOnly2", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":""}`, false, true, none},
		{"TIDOnly3", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"", "cidr":""}`, false, true, none},

		{"TIAddrOnly", `{"address":"172.21.3.35"}`, false, false, addr},
		{"TIAddrOnly2", `{"address":"::0123:4567:89ab:cdef:1234:5678"}`, false, false, addr},
		{"TIAddrRange", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "cidr":"172.21.3.35/24" }`, false, true, cidr},
		{"TIAddrRange2", `{"cidr":"2001:0db8:0123:4567:89ab:cdef:1234:5678/96"}`, false, false, cidr},
		{"TEmpty", ":", true, false, none},
		{"TWrongFormat", `"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt"`, true, false, none},
		{"TWrongFormat2", `"172.21.3.35/24"`, true, false, none},
		{"TWrongID", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBRf3mBFBJpz3te@GGt"}`, true, false, none},
		{"TWrongAddr", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"172.21.3.355"}`, true, false, none},
		{"TWrongAddr2", `{"cidr":":12001:0db8:0123:4567:89ab:cdef:1234:5678/96"}`, true, false, none},
		{"TWrongAddr3", `{"address":"2001:0db8:0123:4567:89ab:cdef:1234:5678/96"}`, true, false, none},
		{"TWrongAddr3", `{"address":"172.21.3.35/24"}`, true, false, none},
		{"TWrongEnt", `{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"172.21.3.35", "cidr":"172.21.3.35/24"}`, true, false, none},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseListEntry(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseListEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.wantAddr == none {
					if got.IpNet != notSpecifiedCIDR {
						t.Errorf("ParseListEntry().IpNet = %v, want not", got.IpNet)
					}
				} else {
					m, b := got.IpNet.Mask.Size()
					if (m == b) != (tt.wantAddr == addr) {
						t.Errorf("ParseListEntry().IpNet = %v, want type %v ", got.IpNet, tt.wantAddr)
					}
				}

				if (got.PeerID != NotSpecifiedID) != tt.wantID {
					t.Errorf("ParseListEntry().PeerID = %v, want not", got.PeerID.String())
				}
			}
		})
	}
}

func TestWhiteListEntry_Contains(t *testing.T) {
	sampleIDStr := "16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt"
	sampleID, _ := IDB58Decode(sampleIDStr)
	otherID, _ := IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
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
		pid  PeerID
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
			e, _ := ParseListEntry(tt.cParam)

			if got := e.Contains(tt.args.addr, tt.args.pid); got != tt.want {
				t.Errorf("WhiteListEntry.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIpNet(t *testing.T) {
	tests := []struct {
		name string

		in string

		wantErr bool
		wantIp  net.IP
	}{
		{"TIP4", "192.168.3.4/32", false, net.ParseIP("192.168.3.4")},
		{"TIP4R", "122.1.3.4/16", false, net.ParseIP("122.1.3.4")},
		{"TIP6R", "2001:0db8:0123:4567:89ab:cdef:1234:5678/96", false, net.ParseIP("2001:0db8:0123:4567:89ab:cdef:1234:5678")},
		{"TIP6R", "2001:0db8:0123:4567:89ab:cdef:1234:5678/128", false, net.ParseIP("2001:0db8:0123:4567:89ab:cdef:1234:5678")},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, n, err := net.ParseCIDR(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCIDR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !ip.Equal(tt.wantIp) {
				t.Errorf("IpNet IP = %v, want %v", ip.String(), tt.wantIp.String())
			}
			m, b := n.Mask.Size()
			t.Logf("mask size is mask %v and bits %v ", m, b)
		})
	}
}

func TestWriteEntries(t *testing.T) {
	eIDIP, _ := ParseListEntry(`{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"172.21.3.35" }`)
	eIDIR, _ := ParseListEntry(`{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "cidr":"172.21.3.35/16" }`)
	eID, _ := ParseListEntry(`{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt" }`)
	eIR, _ := ParseListEntry(`{"cidr":"172.21.3.35/16" }`)
	eIP6, _ := ParseListEntry(`{"address":"2001:0db8:0123:4567:89ab:cdef:1234:5678" }`)
	eIR6, _ := ParseListEntry(`{"cidr":"2001:0db8:0123:4567:89ab:cdef:1234:5678/96" }`)

	tests := []struct {
		name string
		args []WhiteListEntry

		wantErr bool
	}{
		{"TA", []WhiteListEntry{eIDIP, eIDIR, eID, eIR, eIP6, eIR6}, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wr := &bytes.Buffer{}
			err := WriteEntries(tt.args, wr)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("RESULT")
			fmt.Println(wr.String())
		})
	}
}
