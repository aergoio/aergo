/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAddress(t *testing.T) {
	for i := 0; i < 100; i++ {
		key, err := btcec.NewPrivateKey()
		assert.NoError(t, err, "could not create private key")

		address := GenerateAddress(key.PubKey().ToECDSA())
		assert.Equalf(t, types.AddressLength, len(address), "wrong address length : %s", address)
		assert.Equal(t, key.PubKey().SerializeCompressed(), address, "wrong address contents")
	}
}
