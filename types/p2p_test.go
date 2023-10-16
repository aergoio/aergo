/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalSize(t *testing.T) {
	var dummyTxHash, _ = enc.ToBytes("4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45")
	fmt.Println("Hash: ", hex.EncodeToString(dummyTxHash))

	sample := &NewTransactionsNotice{}

	expectedLen := proto.Size(sample)
	actual, err := proto.Marshal(sample)
	assert.Nil(t, err)
	fmt.Println("Empty notice size ", len(actual))
	assert.Equal(t, expectedLen, len(actual))

	// single member
	hashes := make([][]byte, 0, 10)
	hashes = append(hashes, dummyTxHash)
	sample.TxHashes = hashes
	expectedLen = proto.Size(sample)
	actual, err = proto.Marshal(sample)
	assert.Nil(t, err)
	fmt.Println("Single hash notice size ", len(actual))
	fmt.Println("Hex: ", hex.EncodeToString(actual))
	assert.Equal(t, expectedLen, len(actual))

	// 100 hashes
	hashes = make([][]byte, 100)
	for i := 0; i < 100; i++ {
		hashes[i] = dummyTxHash
	}
	sample.TxHashes = hashes
	expectedLen = proto.Size(sample)
	actual, err = proto.Marshal(sample)
	assert.Nil(t, err)
	fmt.Println("Hundred hashes notice size ", len(actual))
	fmt.Println("Hex: ", hex.EncodeToString(actual[0:40]))
	assert.Equal(t, expectedLen, len(actual))

	// 1000 hashes
	hashes = make([][]byte, 1000)
	for i := 0; i < 1000; i++ {
		hashes[i] = dummyTxHash
	}
	sample.TxHashes = hashes
	expectedLen = proto.Size(sample)
	actual, err = proto.Marshal(sample)
	assert.Nil(t, err)
	fmt.Println("Thousand hashes notice size ", len(actual))
	fmt.Println("Hex: ", hex.EncodeToString(actual[0:40]))
	assert.Equal(t, expectedLen, len(actual))

}
