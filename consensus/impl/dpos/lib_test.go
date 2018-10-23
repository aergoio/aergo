package dpos

import (
	"fmt"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/stretchr/testify/assert"
)

func TestBlockInfo(t *testing.T) {
	var bi *blockInfo
	a := assert.New(t)
	a.Equal("(nil)", bi.Hash())
}

func TestPlibCodec(t *testing.T) {
	a := assert.New(t)

	pl1, dpl1 := newSimplePlib(a)
	pl2, dpl2 := newPlibUndo(a)

	tests := []struct {
		name string
		src  interface{}
		dst  interface{}
	}{
		{name: "BP to *blockInfo", src: pl1, dst: dpl1},
		{name: "BP to []*blockInfo", src: pl2, dst: dpl2},
	}

	for i, test := range tests {
		fmt.Printf("*** test[%v]: %v ***\n", i, test.name)

		// Encode
		buf, err := encode(test.src)
		a.Nil(err)
		fmt.Println("gob size =", len(buf.Bytes()))

		// Decode
		err = decode(&buf, test.dst)
		a.Nil(err)
		fmt.Println(spew.Sdump(test.dst))
	}
}

func newSimplePlib(a *assert.Assertions) (interface{}, interface{}) {
	block1 := newSignedBlock(a, nil)
	block2 := newSignedBlock(a, block1)

	pl := newPlibMap()
	addBlockInfo(pl, block1)
	addBlockInfo(pl, block2)

	dpl := newPlibMap()

	return &pl, &dpl
}

type pLibUndo map[string][]*blockInfo

func newPlibUndo(a *assert.Assertions) (interface{}, interface{}) {
	block := newSignedBlock(a, nil)
	pu := make(pLibUndo)
	addBlockInfoAsUndo(pu, block)

	dpu := make(pLibUndo)

	return &pu, &dpu
}

func addBlockInfoAsUndo(p pLibUndo, block *types.Block) {
	bpID := block.BPID2Str()
	if _, exist := p[bpID]; !exist {
		p[bpID] = make([]*blockInfo, 0)
	}
	p[bpID] = append(p[bpID], newBlockInfo(block))
	p[bpID] = append(p[bpID], newBlockInfo(block))
	p[bpID] = append(p[bpID], newBlockInfo(block))
}

func genKeyPair(assert *assert.Assertions) (crypto.PrivKey, crypto.PubKey) {
	privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	assert.Nil(err)

	return privKey, pubKey
}

func newSignedBlock(a *assert.Assertions, prev *types.Block) *types.Block {
	block := types.NewBlock(prev, nil, make([]*types.Receipt, 0), make([]*types.Tx, 0), nil, 10)
	priv, _ := genKeyPair(a)
	err := block.Sign(priv)
	a.Nil(err)

	return block
}

func addBlockInfo(pLib map[string]*blockInfo, block *types.Block) {
	pLib[block.BPID2Str()] = newBlockInfo(block)
}

func newPlibMap() map[string]*blockInfo {
	return make(map[string]*blockInfo)
}

func TestEmbededMap(t *testing.T) {
	type X struct {
		M map[string]int
		V int
	}

	a := assert.New(t)
	x := &X{M: make(map[string]int), V: 10}
	x.M["a"] = 1
	x.M["b"] = 2
	x.M["c"] = 3

	orgLen := len(x.M)
	buf, err := encode(x)
	a.Nil(err)

	y := &X{M: make(map[string]int), V: 10}
	err = decode(&buf, y)
	a.Nil(err)
	a.Equal(orgLen, len(y.M))
	fmt.Println(len(y.M))

	l := newLibStatus(defaultConsensusCount)
	l.Prpsd["a"] = &plInfo{&blockInfo{}, &blockInfo{}}
	l.Prpsd["b"] = &plInfo{&blockInfo{}, &blockInfo{}}
	l.Prpsd["c"] = &plInfo{&blockInfo{}, &blockInfo{}}
	buf, err = encode(l)
	a.Nil(err)
	orgLen = len(l.Prpsd)

	m := newLibStatus(defaultConsensusCount)
	err = decode(&buf, m)
	a.Nil(err)
	a.Equal(len(m.Prpsd), orgLen)
}
