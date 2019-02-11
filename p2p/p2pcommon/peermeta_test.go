/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
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
			ipAddr := tt.args.ip
			addr := &types.PeerAddress{Address: ipAddr, Port: tt.args.port, PeerID: []byte(tt.args.id)}
			actual := FromPeerAddress(addr)
			assert.Equal(t, ipAddr, actual.IPAddress)
			assert.Equal(t, tt.args.port, actual.Port)
			assert.Equal(t, tt.args.id, string(actual.ID))

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

