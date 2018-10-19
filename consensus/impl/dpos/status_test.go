package dpos

import (
	"fmt"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/stretchr/testify/assert"
)

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
	block := types.NewBlock(prev, nil, make([]*types.Tx, 0), nil, 10)
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
