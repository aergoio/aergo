package types

import (
	"encoding/json"
	fmt "fmt"
	"math/big"
	"testing"

	"github.com/aergoio/aergo/internal/enc"
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
	g.Balance = map[string]*State{"abc": &State{Balance: new(big.Int).SetUint64(1234).Bytes()}}
	b, err := json.Marshal(g)
	a.Nil(err)
	fmt.Println(string(b))
}

func TestGenesisChainID(t *testing.T) {
	a := assert.New(t)
	g := GetDefaultGenesis()
	chainID := g.ChainID()
	a.True(g.ID.Equals(&defaultChainID))
	fmt.Println("len:", len(chainID))
	fmt.Println("chain_id: ", enc.ToString(chainID))
}

func TestGenesisBytes(t *testing.T) {
	a := assert.New(t)
	g1 := GetDefaultGenesis()
	g1.Balance = map[string]*State{"abc": &State{Balance: new(big.Int).SetUint64(1234).Bytes()}}
	g1.BPs = []string{"xxx", "yyy", "zzz"}
	b := g1.Bytes()
	fmt.Println(spew.Sdump(g1))

	g2 := GetGenesisFromBytes(b)
	fmt.Println(spew.Sdump(g2))
	a.Nil(g2.Balance)
}
