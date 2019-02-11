/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParseSubProtocol(t *testing.T) {
	lastVal := 0x50 // TODO should change value if protocol is changed.
	expectedProtocolCount := 23
	actualCount := 0

	for i:=0 ; i < lastVal ; i++ {
		input := p2pcommon.SubProtocol(i)
		str := input.String()

		actual, found := ParseSubProtocol(str)
		actual2, found2 := FindSubProtocol(input.Uint32())
		if strings.HasPrefix(str, "SubProtocol(") {
			assert.False(t, found)
			assert.False(t, found2)
		} else {
			assert.Truef(t, found, "should be found subprotocol %s ",str)
			assert.True(t, found2,"should be found subprotocol %s by code %d ",str,input.Uint32())
			actualCount++
			assert.Equal(t, input, actual)
			assert.Equal(t, input, actual2)
		}
	}
	assert.Equal(t, expectedProtocolCount, actualCount)
}

