package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestDefaultGenesis(t *testing.T) {
	a := assert.New(t)
	g := GetDefaultGenesis()
	a.Equal(g.ID, defaultChainID)
}

func TestGenesisJSON(t *testing.T) {
	a := assert.New(t)
	g := GetDefaultGenesis()
	d := time.Unix(0, g.Timestamp)
	fmt.Println("timestamp", d)
	g.Balance = map[string]string{"abc": "1234"}
	b, err := json.Marshal(g)
	a.Nil(err)
	fmt.Println(string(b))
}

func TestGenesisChainID(t *testing.T) {
	a := assert.New(t)
	g := GetDefaultGenesis()
	chainID, err := g.ChainID()
	a.Nil(err)
	a.True(g.ID.Equals(&defaultChainID))
	fmt.Println("len:", len(chainID))
	fmt.Println("chain_id: ", enc.ToString(chainID))
}

func TestGenesisBytes(t *testing.T) {
	a := assert.New(t)
	g1 := GetDefaultGenesis()
	g1.Balance = map[string]string{"abc": "1234"}
	g1.BPs = []string{"xxx", "yyy", "zzz"}

	b := g1.Bytes()
	fmt.Println(spew.Sdump(g1))

	g2 := GetGenesisFromBytes(b)
	a.Nil(g2.Balance)
}

func TestCodecChainID(t *testing.T) {
	a := assert.New(t)
	id1 := NewChainID()

	id1.AsDefault()
	a.True(id1.Equals(&defaultChainID))

	b, err := id1.Bytes()
	a.Nil(err)

	id2 := NewChainID()
	err = id2.Read(b)
	a.Nil(err)
	a.True(id1.Equals(id2))
}

func TestUpdateChainIdVersion(t *testing.T) {
	g := GetMainNetGenesis()
	b := g.Block().GetHeader().GetChainID()
	cid0 := new(ChainID)
	cid0.Read(b)
	if cid0.Version != 0 {
		t.Errorf("version mismatch: 0 expected, but got %d", cid0.Version)
		t.Log(cid0.ToJSON())
	}
	updatedCID := MakeChainId(b, 0)
	if !bytes.Equal(b, updatedCID) {
		t.Error("chainid is not equal")
	}
	updatedCID = MakeChainId(b, 1)
	cid1 := new(ChainID)
	cid1.Read(updatedCID)
	if cid1.Version != 1 {
		t.Errorf("version mismatch: 1 expected, but got %d", cid1.Version)
		t.Log(cid1.ToJSON())
	}
}
