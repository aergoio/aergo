/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import "testing"

func TestP2PVersion_String(t *testing.T) {
	tests := []struct {
		name string
		v    P2PVersion
		want string
	}{
		{"T030", P2PVersion030, "0.3.0"},
		{"T031", P2PVersion031, "0.3.1"},
		{"T032", P2PVersion032, "0.3.2"},
		{"T100", P2PVersion(0x010000), "1.0.0"},
		{"T101", P2PVersion(0x010001), "1.0.1"},
		{"T121", P2PVersion(0x010201), "1.2.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("P2PVersion.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
