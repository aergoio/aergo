package dpos

import (
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/stretchr/testify/assert"
)

type testChain struct {
	chain  []*types.Block
	status *Status
	bestNo types.BlockNo

	bpid          string
	lpb           map[string]types.BlockNo
	bpKey         []crypto.PrivKey
	bpClusterSize uint16
}

type testCluster struct {
	size uint16
}

func (c *testCluster) Size() uint16 {
	return c.size
}

func (c *testCluster) Update(ids []string) error {
	return nil
}

func newTestChain(clusterSize uint16) (*testChain, error) {
	bpKey := make([]crypto.PrivKey, int(clusterSize))
	for i := 0; i < int(clusterSize); i++ {
		var err error
		bpKey[i], _, err = crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		if err != nil {
			return nil, err
		}
	}

	b, err := crypto.MarshalPublicKey(bpKey[0].GetPublic())
	if err != nil {
		return nil, err
	}

	tc := &testChain{
		chain:         make([]*types.Block, 0),
		status:        NewStatus(&testCluster{size: clusterSize}, nil, nil, 0),
		bpid:          base58.Encode(b),
		lpb:           make(map[string]types.BlockNo),
		bpKey:         bpKey,
		bpClusterSize: clusterSize,
	}
	tc.setGenesis(newBlock(0))

	// Prevent DB access
	tc.status.done = true

	return tc, nil
}

func (tc *testChain) setGenesis(block *types.Block) {
	if block.BlockNo() != 0 {
		logger.Panic().Msg("invalid genesis block: non-zero block no")
	}
	tc.status.libState.genesisInfo = &blockInfo{BlockHash: block.ID(), BlockNo: 0}
	tc.status.bestBlock = block
	tc.chain = append(tc.chain, block)
}

func (tc *testChain) addBlock(i types.BlockNo) error {
	pk := tc.getBpKey(i % types.BlockNo(tc.bpClusterSize))
	b, err := crypto.MarshalPrivateKey(pk)
	if err != nil {
		return err
	}
	spk := base58.Encode(b)

	prevBlock := tc.chain[len(tc.chain)-1]
	block := newBlockFromPrev(prevBlock, 0, types.DummyBlockVersionner(0))

	confirmNo := func(no types.BlockNo) (confirms types.BlockNo) {
		lpb := types.BlockNo(0)
		if v, exist := tc.lpb[spk]; exist {
			lpb = v
		}
		confirms = no - lpb

		return
	}
	block.SetConfirms(confirmNo(block.BlockNo()))

	if err = block.Sign(pk); err != nil {
		return err
	}

	tc.lpb[spk] = block.BlockNo()

	tc.chain = append(tc.chain, block)
	tc.bestNo = types.BlockNo(len(tc.chain) - 1)
	tc.status.Update(block)

	return nil
}

func (tc *testChain) getBpKey(i types.BlockNo) crypto.PrivKey {
	return tc.bpKey[i%types.BlockNo(tc.bpClusterSize)]
}

func TestTestChain(t *testing.T) {
	const (
		clusterSize = 3
		maxBlockNo  = types.BlockNo(clusterSize) * 20
	)

	a := assert.New(t)
	tc, err := newTestChain(clusterSize)
	a.Nil(err)

	for i := types.BlockNo(1); i <= maxBlockNo; i++ {
		a.Nil(tc.addBlock(i))
		logger.Info().Uint64("LIB:", tc.status.libState.Lib.BlockNo).Msg("lib")
	}

	a.Equal(tc.bestNo, maxBlockNo)
	a.Equal(tc.status.libState.Lib.BlockNo, maxBlockNo-clusterSize-1)
}

func TestNumLimitGC(t *testing.T) {
	const clusterSize = 23

	a := assert.New(t)

	ls := newLibStatus(clusterSize)

	for i := 1; i <= clusterSize*3; i++ {
		ls.confirms.PushBack(
			&confirmInfo{
				blockInfo: &blockInfo{BlockNo: types.BlockNo(i)},
			})
	}

	ls.gc(nil)
	a.True(ls.confirms.Len() <= ls.gcNumLimit())
}

func TestLibGC(t *testing.T) {
	const (
		clusterSize = 23
		libNo       = 3
	)

	a := assert.New(t)

	ls := newLibStatus(clusterSize)
	ls.Lib = &blockInfo{BlockNo: libNo}

	for i := 1; i <= clusterSize*3; i++ {
		ls.confirms.PushBack(
			&confirmInfo{
				blockInfo: &blockInfo{BlockNo: types.BlockNo(i)},
			})
	}

	ls.gc(nil)
	a.True(cInfo(ls.confirms.Front()).blockInfo.BlockNo > libNo)
}
