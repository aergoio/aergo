package dpos

import (
	"fmt"
	"testing"

	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-core/crypto"
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

	b, err := bpKey[0].GetPublic().Bytes()
	if err != nil {
		return nil, err
	}

	tc := &testChain{
		chain:         make([]*types.Block, 0),
		status:        NewStatus(&testCluster{size: clusterSize}, nil, nil, 0),
		bpid:          enc.ToString(b),
		lpb:           make(map[string]types.BlockNo),
		bpKey:         bpKey,
		bpClusterSize: clusterSize,
	}
	tc.setGenesis(types.NewBlock(nil, nil, nil, nil, nil, 0))

	// Prevent DB access
	tc.status.done = true

	return tc, nil
}

func (tc *testChain) setGenesis(block *types.Block) {
	if block.BlockNo() != 0 {
		panic("invalid genesis block: non-zero block no")
	}
	tc.status.libState.genesisInfo = &blockInfo{BlockHash: block.ID(), BlockNo: 0}
	tc.status.bestBlock = block
	tc.chain = append(tc.chain, block)
}

func (tc *testChain) addBlock(i types.BlockNo) error {
	pk := tc.getBpKey(i % types.BlockNo(tc.bpClusterSize))
	b, err := pk.Bytes()
	if err != nil {
		return err
	}
	spk := enc.ToString(b)

	prevBlock := tc.chain[len(tc.chain)-1]
	block := types.NewBlock(prevBlock, nil, nil, nil, nil, 0)

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
		fmt.Println("LIB:", tc.status.libState.Lib.BlockNo)
	}

	a.Equal(tc.bestNo, maxBlockNo)
	a.Equal(tc.status.libState.Lib.BlockNo, maxBlockNo-clusterSize-1)
}

func TestNumLimitGC(t *testing.T) {
	const (
		clusterSize    = 23
		consensusCount = clusterSize*2/3 + 1
	)

	a := assert.New(t)

	ls := newLibStatus(consensusCount)

	for i := 1; i <= clusterSize*3; i++ {
		ls.confirms.PushBack(
			&confirmInfo{
				blockInfo: &blockInfo{BlockNo: types.BlockNo(i)},
			})
	}

	ls.gc()
	a.True(ls.confirms.Len() <= ls.gcNumLimit())
}

func TestLibGC(t *testing.T) {
	const (
		clusterSize    = 23
		consensusCount = clusterSize*2/3 + 1
		libNo          = 3
	)

	a := assert.New(t)

	ls := newLibStatus(consensusCount)
	ls.Lib = &blockInfo{BlockNo: libNo}

	for i := 1; i <= clusterSize*3; i++ {
		ls.confirms.PushBack(
			&confirmInfo{
				blockInfo: &blockInfo{BlockNo: types.BlockNo(i)},
			})
	}

	ls.gc()
	a.True(cInfo(ls.confirms.Front()).blockInfo.BlockNo > libNo)
}
