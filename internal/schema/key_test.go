package schema

import (
	"math"
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

// TODO
func TestKeyReceipts(t *testing.T) {
	for _, test := range []struct {
		blockHash []byte
		blockNo   types.BlockNo
		expectKey []byte
	}{
		{nil, 0, []byte{byte('r'), 0, 0, 0, 0, 0, 0, 0, 0}},
		{nil, 1, []byte{byte('r'), 1, 0, 0, 0, 0, 0, 0, 0}},
		{nil, 255, []byte{byte('r'), 255, 0, 0, 0, 0, 0, 0, 0}},
		{nil, math.MaxUint64, []byte{byte('r'), 255, 255, 255, 255, 255, 255, 255, 255}},
		{[]byte{1, 2, 3, 4}, 0, []byte{byte('r'), 1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0}},
		{[]byte{144, 75, 132, 157, 195, 199, 41, 233, 24, 89, 183, 252, 80, 151, 5, 244, 83, 64, 39, 204, 84, 37, 182, 61, 72, 248, 192, 223, 59, 216, 108, 240}, 0, []byte{byte('r'), 144, 75, 132, 157, 195, 199, 41, 233, 24, 89, 183, 252, 80, 151, 5, 244, 83, 64, 39, 204, 84, 37, 182, 61, 72, 248, 192, 223, 59, 216, 108, 240, 0, 0, 0, 0, 0, 0, 0, 0}},
	} {
		key := KeyReceipts(test.blockHash, test.blockNo)
		assert.Equal(t, test.expectKey, key, "TestKeyReceipts(%v, %v)", test.blockHash, test.blockNo)
	}
}

// raft
func TestKeyRaftEntry(t *testing.T) {

}

func TestKeyRaftEntryInvert(t *testing.T) {

}

func TestKeyRaftConfChangeProgress(t *testing.T) {

}

// governance
func TestKeyEnterpriseConf(t *testing.T) {

}

func TestKeyName(t *testing.T) {

}

func TestKeyParam(t *testing.T) {

}

func TestKeyStaking(t *testing.T) {

}

func TestKeyVote(t *testing.T) {

}

func TestKeyVoteSort(t *testing.T) {

}

func TestKeyVoteTotal(t *testing.T) {

}

func TestKeyVpr(t *testing.T) {

}
