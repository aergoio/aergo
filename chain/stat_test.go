package chain

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainStatBasic(t *testing.T) {
	var chk = assert.New(t)

	stats := newStats()
	chk.Equal(len(stats), statMax, "# of stat item is inconsistent.")
	for i, st := range stats {
		chk.Equal(int64(0), st.getCount(), "stat[%d] initial # of events must be 0.", i)
	}
}

func TestChainStatReorgBasic(t *testing.T) {
	var chk = assert.New(t)

	stats := newStats()
	i := statReorg
	chk.Equal(int64(0), stats.getCount(i), "reorg stat's initial # of events must be 0.")
	stats.updateEvent(i, "a", "b", "c")
	chk.Equal(int64(1), stats.getCount(i))
	stats.updateEvent(i, "a", "b", "c")
	stats.updateEvent(i, "a", "b", "c")
	stats.updateEvent(i, "a", "b", "c")
	chk.Equal(int64(4), stats.getCount(i))

	b, err := json.Marshal(stats.getLastestEvent(i))
	chk.Nil(err)
	fmt.Println(string(b))
}

func TestChainStatReorgClone(t *testing.T) {
	var chk = assert.New(t)

	stats := newStats()
	i := statReorg

	r := stats.clone(i)
	chk.NotNil(r)
	b, err := json.Marshal(r)
	chk.Nil(err)
	fmt.Println(string(b))

	stats.updateEvent(i, "a", "b", "c")
	stats.updateEvent(i, "a", "b", "c")
	stats.updateEvent(i, "a", "b", "c")
	r = stats.clone(i)
	chk.NotNil(r)
	b, err = json.Marshal(r)
	chk.Nil(err)
	fmt.Println(string(b))

}
