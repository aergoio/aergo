package p2putil

import (
	"github.com/aergoio/aergo/types"
	"strconv"
	"strings"
	"testing"

	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/stretchr/testify/assert"
)

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
			m := p2pcommon.PeerMeta{
				IPAddress:  tt.fields.ip,
				Port:       tt.fields.port,
				ID:         types.PeerID(tt.fields.id),
				Designated: false,
				Outbound:   false,
			}
			actual := ShortMetaForm(m)
			assert.True(t, strings.Contains(actual, tt.fields.ip))
			assert.True(t, strings.Contains(actual, strconv.Itoa(int(tt.fields.port))))
			assert.True(t, strings.Contains(actual, m.ID.Pretty()))
			m2 := m
			m2.Designated = true
			assert.Equal(t, actual, ShortMetaForm(m2))
			m3 := m
			m3.Outbound = true
			assert.Equal(t, actual, ShortMetaForm(m3))
		})
	}
}
