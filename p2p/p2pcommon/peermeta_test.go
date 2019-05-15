/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"testing"

	"github.com/aergoio/aergo/types"
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

func TestNewMetaFromStatus(t *testing.T) {
	type args struct {
		ip       string
		port     uint32
		id       string
		noExpose bool
		outbound bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"TExpose", args{"192.168.1.2", 2, "id0002", false, false}},
		{"TNoExpose", args{"0.0.0.0", 2223, "id2223", true, false}},
		{"TOutbound", args{"2001:0db8:85a3:08d3:1319:8a2e:0370:7334", 444, "id0002", false, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &types.PeerAddress{Address: tt.args.ip, Port: tt.args.port, PeerID: []byte(tt.args.id)}
			status := &types.Status{Sender: sender, NoExpose: tt.args.noExpose}
			actual := NewMetaFromStatus(status, tt.args.outbound)
			assert.Equal(t, tt.args.ip, actual.IPAddress)
			assert.Equal(t, tt.args.port, actual.Port)
			assert.Equal(t, tt.args.id, string(actual.ID))
			assert.Equal(t, tt.args.noExpose, actual.Hidden)
			assert.Equal(t, tt.args.outbound, actual.Outbound)

			actual2 := actual.ToPeerAddress()
			assert.Equal(t, *sender, actual2)
		})
	}
}
