package bp

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/aergoio/aergo/consensus/impl/dpos/param"
	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
)

func TestNewClusterInvalid(t *testing.T) {
	randIds := func() []string {
		ids := make([]string, param.BlockProducers)
		for i := 0; i < param.BlockProducers; i++ {
			ids[i] = fmt.Sprintf("%v", rand.Int31())
		}
		return ids
	}

	tcs := []struct {
		name string
		ids  []string
	}{
		{"invalid size", []string{"x", "y"}},
		{"invalid IDs ", randIds()},
	}

	for _, tc := range tcs {
		bpc, err := NewCluster(tc.ids)
		fmt.Println(tc.name, "--> ", err.Error())
		assert.NotNil(t, err)
		assert.Nil(t, bpc)
	}
}

func TestNewCluster(t *testing.T) {
	genID := func() string {
		_, pubKey, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		assert.Nil(t, err)
		b, err := peer.IDFromPublicKey(pubKey)
		assert.Nil(t, err)
		return b.Pretty()
	}

	genIds := func() []string {
		ids := make([]string, param.BlockProducers)
		for i := 0; i < param.BlockProducers; i++ {
			ids[i] = genID()
			fmt.Println(ids[i])
		}
		return ids
	}

	bpc, err := NewCluster(genIds())
	assert.Nil(t, err)
	assert.NotNil(t, bpc, "Cluster alloc failed")
}
