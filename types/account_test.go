package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnDecodeAddress(t *testing.T) {
	addr, err := DecodeAddress("aergo.system")
	assert.NoError(t, err, "decode aergo.system")
	assert.Equal(t, "aergo.system", string(addr), "decode aergo.system")

	addr, err = DecodeAddress("aergo.name")
	assert.NoError(t, err, "decode aergo.name")
	assert.Equal(t, "aergo.name", string(addr), "decode aergo.name")

	addr, err = DecodeAddress("aergo.enterprise")
	assert.NoError(t, err, "decode aergo.enterprise")
	assert.Equal(t, "aergo.enterprise", string(addr), "decode aergo.enterprise")

	addr, err = DecodeAddress("")
	assert.NoError(t, err, "decode aergo.name")
	assert.Equal(t, "", string(addr), "decode \"\"")

	encoded := EncodeAddress([]byte("aergo.system"))
	assert.Equal(t, "aergo.system", encoded, "encode aergo.system")

	encoded = EncodeAddress([]byte("aergo.name"))
	assert.Equal(t, "aergo.name", encoded, "encode aergo.name")

	encoded = EncodeAddress([]byte("aergo.enterprise"))
	assert.Equal(t, "aergo.enterprise", encoded, "encode aergo.enterprise")

	encoded = EncodeAddress(nil)
	assert.Equal(t, "", encoded, "encode nil")

	testAddress := []byte{3, 49, 133, 46, 196, 51, 37, 27, 66, 1, 191, 92, 224, 12, 248, 246, 21, 247, 79, 10, 164, 63, 34, 97, 238, 0, 145, 129, 206, 54, 218, 241, 84}
	encoded = EncodeAddress(testAddress)
	assert.Equal(t, "AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2", encoded, "encode test address")

	addr, err = DecodeAddress("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2")
	assert.NoError(t, err, "decode test address")
	assert.Equal(t, testAddress, addr, "decode test address")
}

func TestToAddress(t *testing.T) {
	addr := ToAddress("")
	assert.Equal(t, 0, len(addr), "nil")
}
