/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package slot

import (
	"fmt"
	"testing"
	"time"

	"github.com/aergoio/aergo/consensus/impl/dpos/param"
	"github.com/stretchr/testify/assert"
)

func TestSlotRotation(t *testing.T) {
	const (
		nSlots = param.BlockProducers / 4
	)

	ticker := time.NewTicker(param.BlockInterval)
	slots := make(map[int64]interface{})
	i := 0
	for now := range ticker.C {
		idx := Time(now).nextBpIndex()
		slots[idx] = struct{}{}
		fmt.Printf("[%v]", idx)
		if i > nSlots {
			break
		}
		i++
	}
	fmt.Println()
	assert.True(t, nSlots <= len(slots), "invalid slot index")
}

func TestSlotConversion(t *testing.T) {
	slot := Now()
	assert.Equal(t, slot.timeSec, slot.timeMs/1000, "inconsistent slot members")
}

func TestSlotValidNow(t *testing.T) {
	assert.True(t, Now().IsValidNow(), "invalid slot")
}
