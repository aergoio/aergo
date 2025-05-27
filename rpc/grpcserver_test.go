/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package rpc

import (
	"context"
	"fmt"
	"github.com/aergoio/aergo/v2/types/utils"
	"math/big"
	"reflect"
	"testing"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/internal/enc/base64"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/message"
	"github.com/aergoio/aergo/v2/types/message/messagemock"
	"github.com/golang/mock/gomock"
)

func TestAergoRPCService_dummys(t *testing.T) {
	fmt.Println("dummyBlockHash")
	fmt.Printf("HEX : %s \n", hex.Encode(dummyBlockHash))
	fmt.Printf("B64 : %s \n", base64.Encode(dummyBlockHash))
	fmt.Printf("B58 : %s \n", base58.Encode(dummyBlockHash))
	fmt.Println()
	fmt.Println("dummyTx")
	fmt.Printf("HEX : %s \n", hex.Encode(dummyTxHash))
	fmt.Printf("B64 : %s \n", base64.Encode(dummyTxHash))
	fmt.Printf("B58 : %s \n", base58.Encode(dummyTxHash))
	fmt.Println()

	fmt.Println("Address1")
	fmt.Printf("HEX : %s \n", hex.Encode(dummyWalletAddress))
	fmt.Printf("B64 : %s \n", base64.Encode(dummyWalletAddress))
	fmt.Printf("B58 : %s \n", base58.Encode(dummyWalletAddress))
	fmt.Println()

	fmt.Println("Address2")
	fmt.Printf("HEX : %s \n", hex.Encode(dummyWalletAddress2))
	fmt.Printf("B64 : %s \n", base64.Encode(dummyWalletAddress2))
	fmt.Printf("B58 : %s \n", base58.Encode(dummyWalletAddress2))
	fmt.Println()

}

var dummyBlockHash, _ = hex.Decode("4f461d85e869ade8a0544f8313987c33a9c06534e50c4ad941498299579bd7ac")
var dummyBlockHeight uint32 = 100215
var dummyTxHash, _ = hex.Decode("218bdab4e87fb332b55eb89854ef553f9e3d440c81fff4161b672adede1261ee")

// base64 encoding of dummyTxHash is ""
var dummyWalletAddress, _ = base58.Decode("1Ee8uhLFXzkSRRU1orBpgXFAPpVi64aSYo")
var dummyWalletAddress2, _ = base58.Decode("16Uiu2HAkwgfFvViH6j2QpQYKtGKKdveEKZvU2T5mRkqFLTZKU4Vp")
var dummyPayload = []byte("OPreturn I am groooot")

var hubStub *component.ComponentHub
var mockCtx context.Context
var mockMsgHelper *messagemock.Helper
var mockActorHelper *MockActorService

func init() {
	hubStub = &component.ComponentHub{}

	mockCtx = &Context{}
}

func TestAergoRPCService_GetTX(t *testing.T) {
	dummyTxBody := types.TxBody{Account: dummyWalletAddress, Amount: new(big.Int).SetUint64(4332).Bytes(),
		Recipient: dummyWalletAddress2, Payload: dummyPayload}
	sampleTx := &types.Tx{Hash: dummyTxHash, Body: &dummyTxBody}

	type fields struct {
		hub         *component.ComponentHub
		actorHelper p2pcommon.ActorService
		msgHelper   message.Helper
	}
	type args struct {
		ctx context.Context
		in  *types.SingleBytes
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.Tx
		wantErr bool
	}{
		{name: "TNormal", args: args{ctx: mockCtx, in: &types.SingleBytes{Value: dummyTxHash}},
			want: &types.Tx{Hash: dummyTxHash, Body: &dummyTxBody}, wantErr: false},
		// TODO the malformed hash is allowed until v2.8.x , but should return error at v2.9.0
		{name: "TMalformedHash", args: args{ctx: mockCtx, in: &types.SingleBytes{Value: append(dummyTxHash, 'b', 'd')}},
			want: &types.Tx{Hash: dummyTxHash, Body: &dummyTxBody}, wantErr: false},
		{name: "TEmptyHash", args: args{ctx: mockCtx, in: &types.SingleBytes{Value: []byte{}}},
			want: &types.Tx{Hash: dummyTxHash, Body: &dummyTxBody}, wantErr: false},
		{name: "TNilHash", args: args{ctx: mockCtx, in: &types.SingleBytes{Value: nil}},
			want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMsgHelper := messagemock.NewHelper(ctrl)
			mockActorHelper := p2pmock.NewMockActorService(ctrl)
			mockActorHelper.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.Any()).Return(message.MemPoolGetRsp{}, nil).AnyTimes()
			mockMsgHelper.EXPECT().ExtractTxFromResponse(gomock.AssignableToTypeOf(message.MemPoolGetRsp{})).Return(sampleTx, nil).AnyTimes()

			rpc := &AergoRPCService{
				hub: hubStub, actorHelper: mockActorHelper, msgHelper: mockMsgHelper,
			}
			got, err := rpc.GetTX(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("AergoRPCService.GetTX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AergoRPCService.GetTX() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAergoRPCService_NodeState(t *testing.T) {
	var emptyMap = []byte("{}")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgHelper := messagemock.NewHelper(ctrl)
	mockActorHelper := p2pmock.NewMockActorService(ctrl)

	dummyResult := make(map[string]*component.CompStatRsp)
	mockActorHelper.EXPECT().CallRequestDefaultTimeout(gomock.Any(), gomock.Any()).Return(dummyResult, nil).AnyTimes()
	type fields struct {
		hub         *component.ComponentHub
		actorHelper p2pcommon.ActorService
		msgHelper   message.Helper
	}
	type args struct {
		ctx context.Context
		in  *types.NodeReq
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.SingleBytes
		wantErr bool
	}{
		{name: "normal", args: args{ctx: mockCtx,
			in: &types.NodeReq{Timeout: utils.ToByteArrayOrEmpty(int64(5)), Component: nil}}, fields: fields{hubStub, mockActorHelper, mockMsgHelper},
			want: &types.SingleBytes{Value: emptyMap}, wantErr: false},
		{name: "nilTime", args: args{ctx: mockCtx,
			in: &types.NodeReq{Timeout: nil, Component: nil}}, fields: fields{hubStub, mockActorHelper, mockMsgHelper},
			want: &types.SingleBytes{Value: emptyBytes}, wantErr: true},
		{name: "shortData", args: args{ctx: mockCtx,
			in: &types.NodeReq{Timeout: []byte("short"), Component: nil}}, fields: fields{hubStub, mockActorHelper, mockMsgHelper},
			want: &types.SingleBytes{Value: emptyMap}, wantErr: true},
		{name: "wrongComponent", args: args{ctx: mockCtx,
			in: &types.NodeReq{Timeout: utils.ToByteArrayOrEmpty(int64(5)), Component: []byte("nope")}}, fields: fields{hubStub, mockActorHelper, mockMsgHelper},
			want: &types.SingleBytes{Value: emptyBytes}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpc := &AergoRPCService{
				hub: tt.fields.hub, actorHelper: mockActorHelper, msgHelper: mockMsgHelper,
			}
			got, err := rpc.NodeState(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("AergoRPCService.GetTX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// do not check return value when error is expected
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AergoRPCService.GetTX() = %v, want %v", got, tt.want)
			}
		})
	}
}

type FutureStub struct {
	actor.Future
	dumbResult interface{}
}

func (fs *FutureStub) Result() interface{} {
	return fs.dumbResult
}

func NewFutureStub(result interface{}) FutureStub {
	return FutureStub{dumbResult: result}
}

func TestAergoRPCService_GetBlockIncompleteArg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgHelper := messagemock.NewHelper(ctrl)
	mockActorHelper := p2pmock.NewMockActorService(ctrl)

	dummyResult := make(map[string]*component.CompStatRsp)
	mockActorHelper.EXPECT().CallRequestDefaultTimeout(gomock.Any(), gomock.Any()).Return(dummyResult, nil).AnyTimes()
	type fields struct {
		hub         *component.ComponentHub
		actorHelper p2pcommon.ActorService
		msgHelper   message.Helper
	}
	type args struct {
		ctx context.Context
		in  *types.SingleBytes
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "nilValue", fields: fields{hubStub, mockActorHelper, mockMsgHelper},
			args: args{mockCtx, &types.SingleBytes{}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpc := &AergoRPCService{
				hub: tt.fields.hub, actorHelper: mockActorHelper, msgHelper: mockMsgHelper,
			}
			_, err := rpc.GetBlock(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
