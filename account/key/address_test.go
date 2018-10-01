package key

import (
	"testing"

	"github.com/btcsuite/btcd/btcec"
)

func TestGenerateAddress(t *testing.T) {
	for i := 0; i < 1000; i++ {
		key, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			t.Fatal("could not create private key")
		}
		address := GenerateAddress(&key.PublicKey)
		t.Logf("trace: %d, len(%d)", key.PublicKey.X, len(key.PublicKey.X.Bytes()))
		if len(address) != addressLength {
			t.Errorf("wrong address length : %d", len(address))
		}
	}
}
