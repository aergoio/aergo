/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"testing"
)


func Test_MarshalTxResp(t *testing.T) {
	dummyTx := &types.Tx{Hash:dummyTxHash, Body:&types.TxBody{Payload:[]byte("It's a good day to die.")}}
	txMarshaled,_ := proto.Marshal(dummyTx)
	txSize := len(dummyTxHash)+2 + len(txMarshaled)+2 // hash+ field desc of hash + tx+field desc of tx
	//fmt.Println("TX   : ",hex.EncodeToString(txMarshaled))
	emptyMarshaled,_ := proto.Marshal(&types.GetTransactionsResponse{})
	emptySize := len(emptyMarshaled)
	//fmt.Println("EMPTY: ",hex.EncodeToString(emptyMarshaled))
	//fmt.Printf("Size of All nil: %d , txsize: %d ",emptySize, txSize)
	tests := []struct {
		name string
		itemSize int
		expectedSize int
	}{
		// empty
		{"TEmpty", 0, emptySize},
		// single
		{"TSingle", 1, emptySize+txSize},
		// small multi
		{"T10", 10, emptySize+txSize*10},
		// big
		// boundary
		{"T50000", 50000, emptySize+txSize*50000},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hashSlice := make([][]byte,0,10)
			txSlice := make([]*types.Tx,0,10)
			for i := 0 ; i < test.itemSize; i++ {
				hashSlice = append(hashSlice, dummyTxHash)
				txSlice = append(txSlice, dummyTx)
			}
			sampleRsp := &types.GetTransactionsResponse{Hashes:hashSlice, Txs:txSlice}
			actual, err := proto.Marshal(sampleRsp)
			if err != nil {
				t.Errorf("Invalid proto error %s",err.Error())
			}
			actualSize := len(actual)
			cut := 80
			if actualSize < cut {
				cut = actualSize
			}
			//fmt.Println("ACTUA: ",hex.EncodeToString(actual[:cut]))

			assert.Equal(t, test.expectedSize, actualSize)

		})
	}
}

func Test_calculateFieldDesc(t *testing.T) {
	sampleSlize := make([]byte, 2<<25 )
	tests := []struct {
		name string
		valueSize int
		expected int
	}{
		{"TZero",0, 0},
		{"TSmall",127, 2},
		{"TMedium",128, 3},
		{"TLarge",16384, 4},
		{"TVeryL",10000000, 5},
		{"TOverflow",2000000000, 6},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, calculateFieldDescSize(test.valueSize) )
			if test.valueSize <= len(sampleSlize) {
				inputBytes := sampleSlize[:test.valueSize]
				dummy := &types.GetBlockHeadersRequest{Hash:inputBytes}
				realSize := proto.Size(dummy)
				assert.Equal(t, realSize, calculateFieldDescSize(test.valueSize)+len(inputBytes))
			} else {
				fmt.Println(test.name , " is too big to make real ")
			}
		})
	}
}
