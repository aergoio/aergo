package system

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

var initializedVprtTest bool

func TestVprtOp(t *testing.T) {
	initVprtTest()

	var (
		hundred = new(big.Int).SetUint64(100)
		ten     = new(big.Int).SetUint64(10)
	)

	const (
		opAdd = iota
		opSub
	)

	op := []func(types.AccountID, *big.Int){
		opAdd: func(addr types.AccountID, opr *big.Int) {
			rank.Add(addr, opr)
		},
		opSub: func(addr types.AccountID, opr *big.Int) {
			rank.Sub(addr, opr)
		},
	}

	type opt struct {
		op  int
		arg *big.Int
	}

	testCases := []struct {
		addr types.AccountID
		ops  []opt
		want *big.Int
	}{
		{
			addr: genAddr(10),
			ops:  []opt{{opAdd, hundred}, {opSub, ten}},
			want: new(big.Int).SetUint64(10090),
		},
		{
			addr: genAddr(11),
			ops:  []opt{{opSub, ten}, {opAdd, hundred}},
			want: new(big.Int).SetUint64(10090),
		},
		{
			addr: genAddr(12),
			ops:  []opt{{opAdd, hundred}, {opAdd, hundred}},
			want: new(big.Int).SetUint64(10200),
		},
		{
			addr: genAddr(13),
			ops:  []opt{{opAdd, ten}, {opAdd, ten}},
			want: new(big.Int).SetUint64(10020),
		},
		{
			addr: genAddr(14),
			ops:  []opt{{opSub, ten}, {opSub, ten}},
			want: new(big.Int).SetUint64(9980),
		},
	}

	for _, tc := range testCases {
		for _, o := range tc.ops {
			op[o.op](tc.addr, o.arg)
		}
		rank.Apply(nil)
		assert.True(t,
			rank.vp[tc.addr].Cmp(tc.want) == 0,
			"incorrect result: %s (must be %s)", rank.vp[tc.addr].String(), tc.want)
	}
}

func initVprtTest() {
	if isInitialized() {
		return
	}

	for i := int32(0); i < vprMax; i++ {
		rank.Set(genAddr(i), new(big.Int).SetUint64(10000))
		rank.Apply(nil)
	}

	initializedVprtTest = true
}

func isInitialized() bool {
	return initializedVprtTest
}

func genAddr(i int32) types.AccountID {
	dig := sha256.New()
	binary.Write(dig, binary.LittleEndian, i)
	return types.ToAccountID(dig.Sum(nil))
}
