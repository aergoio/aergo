/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package slot

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlotRotation(t *testing.T) {
	const (
		nSlots     = 5
		bpInterval = 1
	)

	Init(bpInterval, nSlots)

	ticker := time.NewTicker(time.Second)
	slots := make(map[int64]interface{}, nSlots)
	i := 0
	for now := range ticker.C {
		idx := Time(now).NextBpIndex()
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
	assert.Equal(t, nsToMs(slot.timeNs), slot.timeMs, "inconsistent slot members")
	fmt.Println(slot.timeNs, slot.timeMs)
}

func TestSlotValidNow(t *testing.T) {
	assert.True(t, Now().IsValidNow(), "invalid slot")
}
