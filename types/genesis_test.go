package types

import (
	"encoding/json"
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultGenesis(t *testing.T) {
	a := assert.New(t)
	g := GetDefaultGenesis()
	a.Equal(g.ID, defaultChainID)
	b, err := json.Marshal(g)
	a.Nil(err)
	fmt.Println(string(b))
}
