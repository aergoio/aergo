package rpc

import (
	"reflect"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
)

func TestListBlockStream_Send(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type fields struct {
		id uint32
	}
	type args struct {
		block  *types.Block
		repeat int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"1", fields{1}, args{&types.Block{}, 1}, false},
		{"5", fields{1}, args{&types.Block{}, 5}, false},
		{"6", fields{1}, args{&types.Block{}, 6}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRPCServer := p2pmock.NewMockAergoRPCService_ListBlockStreamServer(ctrl)
			s := NewListBlockStream(tt.fields.id, mockRPCServer)
			var err error
			for i := 0; i < tt.args.repeat; i++ {
				err = s.Send(tt.args.block)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListBlockStream_StartSend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type fields struct {
		id uint32
	}
	type args struct {
		goawayCnt int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFinish bool
	}{
		{"no", fields{1}, args{0}, false},
		{"once", fields{1}, args{1}, true},
		{"multiple", fields{1}, args{2}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRPCServer := p2pmock.NewMockAergoRPCService_ListBlockStreamServer(ctrl)
			s := NewListBlockStream(tt.fields.id, mockRPCServer)
			indicatorChan := make(chan bool)
			finished := false
			go func() {
				s.StartSend()
				indicatorChan <- true
			}()
			time.Sleep(100 * time.Millisecond)
			for i := 0; i < tt.args.goawayCnt; i++ {
				s.finishSend <- true
			}
			select {
			case <-indicatorChan:
				finished = true
			case <-time.NewTimer(200 * time.Millisecond).C:
				finished = false
			}
			if finished != tt.wantFinish {
				t.Errorf("StartSend() trifer finish = %v, want %v", finished, tt.wantFinish)
			}
		})
	}
}

func TestNewListBlockStream(t *testing.T) {
	type args struct {
		id     uint32
		stream types.AergoRPCService_ListBlockStreamServer
	}
	tests := []struct {
		name string
		args args
		want *ListBlockStream
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewListBlockStream(tt.args.id, tt.args.stream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewListBlockStream() = %v, want %v", got, tt.want)
			}
		})
	}
}
