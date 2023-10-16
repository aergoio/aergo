/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package rpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/message/messagemock"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/mr-tron/base58/base58"
)

func TestAergoRPCService_dummys(t *testing.T) {
	fmt.Println("dummyBlockHash")
	fmt.Printf("HEX : %s \n", hex.EncodeToString(dummyBlockHash))
	fmt.Printf("B64 : %s \n", enc.ToString(dummyBlockHash))
	fmt.Printf("B58 : %s \n", base58.Encode(dummyBlockHash))
	fmt.Println()
	fmt.Println("dummyTx")
	fmt.Printf("HEX : %s \n", hex.EncodeToString(dummyTxHash))
	fmt.Printf("B64 : %s \n", enc.ToString(dummyTxHash))
	fmt.Printf("B58 : %s \n", base58.Encode(dummyTxHash))
	fmt.Println()

	fmt.Println("Address1")
	fmt.Printf("HEX : %s \n", hex.EncodeToString(dummyWalletAddress))
	fmt.Printf("B64 : %s \n", enc.ToString(dummyWalletAddress))
	fmt.Printf("B58 : %s \n", base58.Encode(dummyWalletAddress))
	fmt.Println()

	fmt.Println("Address2")
	fmt.Printf("HEX : %s \n", hex.EncodeToString(dummyWalletAddress2))
	fmt.Printf("B64 : %s \n", enc.ToString(dummyWalletAddress2))
	fmt.Printf("B58 : %s \n", base58.Encode(dummyWalletAddress2))
	fmt.Println()

}

var dummyBlockHash, _ = hex.DecodeString("4f461d85e869ade8a0544f8313987c33a9c06534e50c4ad941498299579bd7ac")
var dummyBlockHeight uint32 = 100215
var dummyTxHash, _ = hex.DecodeString("218bdab4e87fb332b55eb89854ef553f9e3d440c81fff4161b672adede1261ee")

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgHelper := messagemock.NewHelper(ctrl)
	mockActorHelper := p2pmock.NewMockActorService(ctrl)

	dummyTxBody := types.TxBody{Account: dummyWalletAddress, Amount: new(big.Int).SetUint64(4332).Bytes(),
		Recipient: dummyWalletAddress2, Payload: dummyPayload}
	sampleTx := &types.Tx{Hash: dummyTxHash, Body: &dummyTxBody}
	mockActorHelper.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.Any()).Return(message.MemPoolGetRsp{}, nil)
	mockMsgHelper.EXPECT().ExtractTxFromResponse(gomock.AssignableToTypeOf(message.MemPoolGetRsp{})).Return(sampleTx, nil)
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
		{name: "T00", args: args{ctx: mockCtx, in: &types.SingleBytes{Value: append(dummyTxHash, 'b', 'd')}}, fields: fields{hubStub, mockActorHelper, mockMsgHelper},
			want: &types.Tx{Hash: dummyTxHash, Body: &dummyTxBody}, wantErr: false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpc := &AergoRPCService{
				hub: tt.fields.hub, actorHelper: mockActorHelper, msgHelper: mockMsgHelper,
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
