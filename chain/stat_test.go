package chain

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

var block = types.NewBlock(types.EmptyBlockHeaderInfo, nil, nil, nil, nil, nil)

func TestChainStatReorgClone(t *testing.T) {
	var chk = assert.New(t)

	stats := newStats()
	i := ReorgStat

	r := stats.clone(i)
	chk.NotNil(r)
	b, err := json.Marshal(r)
	chk.Nil(err)
	fmt.Println(string(b))

	stats.updateEvent(i, time.Second*10, block, block, block)
	stats.updateEvent(i, time.Second*10, block, block, block)
	stats.updateEvent(i, time.Second*10, block, block, block)
	r = stats.clone(i)
	chk.NotNil(r)
	b, err = json.Marshal(r)
	chk.Nil(err)
	fmt.Println(string(b))
}

func TestChainStatJSON(t *testing.T) {
	var chk = assert.New(t)

	stats := newStats()
	i := ReorgStat
	stats.updateEvent(i, time.Second*10, block, block, block)
	stats.updateEvent(i, time.Second*10, block, block, block)
	stats.updateEvent(i, time.Second*10, block, block, block)

	s := stats.JSON()
	chk.NotZero(len(s))
	fmt.Println(s)
}
