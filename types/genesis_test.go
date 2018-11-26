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
