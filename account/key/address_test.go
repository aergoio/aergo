package key

import (
	"bytes"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
)

func TestGenerateAddress(t *testing.T) {
	for i := 0; i < 1000; i++ {
		key, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			t.Fatal("could not create private key")
		}
		address := GenerateAddress(&key.PublicKey)
		if len(address) != types.AddressLength {
			t.Errorf("wrong address length : %d", len(address))
		}

		if !bytes.Equal(key.PubKey().SerializeCompressed(), address) {
			t.Errorf("wrong address contents : %v, %v",
				key.PubKey().SerializeCompressed(), address)
		}
	}
}
