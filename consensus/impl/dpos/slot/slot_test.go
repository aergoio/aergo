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

const (
	nSlots     = 5
	bpInterval = 1
)

func TestSlotRotation(t *testing.T) {
	Init(bpInterval)

	ticker := time.NewTicker(time.Second)
	slots := make(map[int64]interface{}, nSlots)
	i := 0
	for now := range ticker.C {
		idx := Time(now).NextBpIndex(nSlots)
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
	Init(bpInterval)

	slot := Now()
	assert.Equal(t, nsToMs(slot.timeNs), slot.timeMs, "inconsistent slot members")
	fmt.Println(slot.timeNs, slot.timeMs)
}

func TestSlotValidNow(t *testing.T) {
	Init(bpInterval)

	assert.True(t, Now().IsValidNow(), "invalid slot")
}

func TestSlotFuture(t *testing.T) {
	Init(bpInterval)

	assert.True(t, !Time(time.Now().Add(-time.Second)).IsFuture(), "must not be a future slot")
	assert.True(t, !Now().IsFuture(), "must not be a future slot")
	assert.True(t, !Time(time.Now().Add(time.Second)).IsFuture(), "must not be a future slot")
	assert.True(t, Time(time.Now().Add(2*time.Second)).IsFuture(), "must be a future slot")
	assert.True(t, Time(time.Now().Add(3*time.Second)).IsFuture(), "must be a future slot")
}
