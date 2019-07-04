/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/gofrs/uuid"
	"strings"
	"testing"
)

func TestLogB58EncMarshaller_MarshalZerologArray(t *testing.T) {
	sampleID := make([][]byte,10)
	for i:=0; i<10; i++ {
		sampleID[i] = uuid.Must(uuid.NewV4()).Bytes()
	}
	type fields struct {
		arr   [][]byte
		limit int
	}
	tests := []struct {
		name   string
		fields fields

		wantNum int
	}{
		{"TSmall", fields{sampleID, 9}, 2},
		{"TSame", fields{sampleID, 10}, 0},
		{"TBig", fields{sampleID, 100}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			logger := log.NewLogger("test").Output(buf)
			m := NewLogB58EncMarshaller(tt.fields.arr, tt.fields.limit)

			logger.Warn().Array("target", m).Msg("do test")
			logStr := buf.String()

			if tt.wantNum>0 {
				moreMsg := fmt.Sprintf("%d more", tt.wantNum)
				if !strings.Contains(logStr, moreMsg) {
					t.Errorf("result %s, want exact %s items ",logStr, moreMsg)
				}
			} else {
				if strings.Contains(logStr,  "more") {
					t.Errorf("result %s, want no dropped item ",logStr)
				}
			}
		})
	}
}

func TestLogBlockHashMarshaller_MarshalZerologArray(t *testing.T) {
	sampleBlks := make([]*Block,10)
	for i:=0; i<10; i++ {
		sampleBlks[i] = &Block{Hash:uuid.Must(uuid.NewV4()).Bytes()}
	}
	type fields struct {
		arr   []*Block
		limit int
	}
	tests := []struct {
		name   string
		fields fields

		wantNum int
	}{
		{"TSmall", fields{sampleBlks, 9}, 2},
		{"TSame", fields{sampleBlks, 10}, 0},
		{"TBig", fields{sampleBlks, 100}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			logger := log.NewLogger("test").Output(buf)
			m := LogBlockHashMarshaller{
				arr:   tt.fields.arr,
				limit: tt.fields.limit,
			}
			logger.Warn().Array("target", m).Msg("do test")
			logStr := buf.String()
			if tt.wantNum>0 {
				moreMsg := fmt.Sprintf("%d more", tt.wantNum)
				if !strings.Contains(logStr, moreMsg) {
					t.Errorf("result %s , want exact %s items ",buf.String(), moreMsg)
				}
			} else {
				if strings.Contains(logStr,  "more") {
					t.Errorf("result %s , want no dropped item ", buf.String())
				}
			}
		})
	}
}
