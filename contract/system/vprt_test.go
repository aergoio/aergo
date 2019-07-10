package system

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func TestVprtAddSub(t *testing.T) {
	for i := int32(0); i < vprMax; i++ {
		rank.Set(genAddr(i), new(big.Int).SetUint64(10000))
		rank.Apply(nil)
	}

	var (
		hundred = new(big.Int).SetUint64(100)
		ten     = new(big.Int).SetUint64(10)
	)

	addr1 := genAddr(10)
	rank.Add(addr1, hundred)
	rank.Sub(addr1, ten)
	rank.Apply(nil)

	assert.True(t,
		rank.vp[addr1].Cmp(new(big.Int).SetUint64(10090)) == 0,
		"incorrect result: %s", rank.vp[addr1].String())
}

func genAddr(i int32) types.AccountID {
	dig := sha256.New()
	binary.Write(dig, binary.LittleEndian, i)
	return types.ToAccountID(dig.Sum(nil))
}
