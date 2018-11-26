package types

import (
	"encoding/json"
	fmt "fmt"
	"testing"

	"github.com/aergoio/aergo/internal/enc"
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
	g1.Balance = map[string]*State{"abc": &State{Balance: 1234}}
	b := g1.Bytes()
	g2 := GetGenesisFromBytes(b)
	a.Nil(g2.Balance)
}
