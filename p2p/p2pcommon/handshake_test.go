/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestHSHeader_Marshal(t *testing.T) {
	type fields struct {
		Magic   uint32
		Version P2PVersion
	}
	tests := []struct {
		name   string
		fields fields
		wantLen int
	}{
		{"T1", fields{MAGICTest, P2PVersion031}, 8},
		{"T3", fields{MAGICMain, P2PVersion033}, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := HSHeader{
				Magic:   tt.fields.Magic,
				Version: tt.fields.Version,
			}
			got := h.Marshal()
			if !reflect.DeepEqual(len(got), tt.wantLen) {
				t.Errorf("HSHeader.Marshal() = %v, want %v", len(got), tt.wantLen)
			}
			got2 := HSHeader{}
			got2.Unmarshal(got)

			if !reflect.DeepEqual(got2, h) {
				t.Errorf("HSHeader.Unmarshal() = %v, want %v", got2, h)
			}

		})
	}
}

func TestHSHeader_Marshal2(t *testing.T) {
	tests := []struct {
		name            string
		input           []byte
		expectedNetwork uint32
		expectedVersion P2PVersion
	}{
		{"TMain033", []byte{0x047, 0x041, 0x68, 0x41, 0, 0, 3, 3}, MAGICMain, P2PVersion033},
		{"TMain020", []byte{0x02e, 0x041, 0x54, 0x29, 0, 1, 3, 5}, MAGICTest, 0x010305},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hs := HSHeader{}
			hs.Unmarshal(test.input)
			assert.Equal(t, test.expectedNetwork, hs.Magic)
			assert.Equal(t, test.expectedVersion, hs.Version)

			actualBytes := hs.Marshal()
			assert.True(t, bytes.Equal(test.input, actualBytes))
		})
	}
}

func TestOutHSHeader_Marshal(t *testing.T) {
	type fields struct {
		Magic    uint32
		Versions []P2PVersion
	}
	tests := []struct {
		name   string
		fields fields
		wantLen int
	}{
		{"TEmpty", fields{MAGICMain, nil}, 8},
		{"TSingle", fields{MAGICMain, []P2PVersion{P2PVersion030}}, 12},
		{"TSingle", fields{MAGICMain, []P2PVersion{0x033333, 0x092fa10, P2PVersion031,P2PVersion030}}, 24},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := HSHeadReq{
				Magic:    tt.fields.Magic,
				Versions: tt.fields.Versions,
			}

			got := h.Marshal()
			if !reflect.DeepEqual(len(got), tt.wantLen) {
				t.Errorf("HSHeader.Marshal() = %v, want %v", len(got), tt.wantLen)
			}
			//got2 := HSHeadReq{}
			//got2.Unmarshal(got)
			//if !reflect.DeepEqual(got2, got) {
			//	t.Errorf("HSHeader.Unmarshal() = %v, want %v", got2, got)
			//}
		})
	}
}
