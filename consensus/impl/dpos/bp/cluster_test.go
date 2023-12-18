package bp

const (
	BlockProducers = 32
)

/* TODO: BP-related paramters eliminated. Rewrite test!

func TestNewClusterInvalid(t *testing.T) {
	randIds := func() []string {
		ids := make([]string, BlockProducers)
		for i := 0; i < BlockProducers; i++ {
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
		bpc, err := NewCluster(newConfig(tc.ids), nil)
		fmt.Println(tc.name, "--> ", err.Error())
		assert.NotNil(t, err)
		assert.Nil(t, bpc)
	}
}

func TestNewCluster(t *testing.T) {
	genID := func() string {
		_, pubKey, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		assert.Nil(t, err)
		b, err := types.IDFromPublicKey(pubKey)
		assert.Nil(t, err)
		return b.String()
	}

	genIds := func() []string {
		ids := make([]string, BlockProducers)
		for i := 0; i < BlockProducers; i++ {
			ids[i] = genID()
			fmt.Println(ids[i])
		}
		return ids
	}

	bpc, err := NewCluster(newConfig(genIds()), nil)
	assert.Nil(t, err)
	assert.NotNil(t, bpc, "Cluster alloc failed")
}

func newConfig(ids []string) *config.ConsensusConfig {
	return &config.ConsensusConfig{}
}
*/
