package p2putil

import (
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
)

func TestPeerMeta_String(t *testing.T) {
	sampleID1 := types.RandomPeerID()
	sampleID2 := types.RandomPeerID()

	type fields struct {
		ip   string
		port uint32
		id   types.PeerID
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"t1", fields{"192.168.1.2", 2, sampleID1}},
		{"t2", fields{"0.0.0.0", 2223, sampleID2}},
		{"t3", fields{"2001:0db8:85a3:08d3:1319:8a2e:370:7334", 444, sampleID1}},
		{"t4", fields{"::ffff:192.0.1.2", 444, sampleID1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := p2pcommon.NewMetaWith1Addr(tt.fields.id, tt.fields.ip, tt.fields.port, "v2.0.0")
			actual := ShortMetaForm(m)
			wantIP := net.ParseIP(tt.fields.ip)
			if !strings.Contains(actual, wantIP.String()) {
				t.Errorf("ShortForm() %v,  want contains %v ", actual, wantIP)
			}
			if !strings.Contains(actual, strconv.Itoa(int(tt.fields.port))) {
				t.Errorf("ShortForm() %v,  want contains %v ", actual, tt.fields.port)
			}
			if !strings.Contains(actual, ShortForm(m.ID)) {
				t.Errorf("ShortForm() %v,  want contains %v ", actual, ShortForm(m.ID))
			}
		})
	}
}
