package system

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

const (
	opAdd = iota
	opSub
)

var (
	valHundred = new(big.Int).SetUint64(100)
	valTen     = new(big.Int).SetUint64(10)

	vprOP = []func(types.AccountID, *big.Int){
		opAdd: func(addr types.AccountID, opr *big.Int) {
			rank.Add(addr, opr)
		},
		opSub: func(addr types.AccountID, opr *big.Int) {
			rank.Sub(addr, opr)
		},
	}

	vprChainStateDB     *state.ChainStateDB
	vprStateDB          *state.StateDB
	initializedVprtTest bool
)

type vprOpt struct {
	op  int
	arg *big.Int
}

type vprTC struct {
	addr types.AccountID
	ops  []vprOpt
	want *big.Int
}

func (tc *vprTC) run(t *testing.T, s *state.ContractState) {
	for _, o := range tc.ops {
		vprOP[o.op](tc.addr, o.arg)
	}
	rank.Apply(s)
	assert.True(t,
		rank.vp[tc.addr].Cmp(tc.want) == 0,
		"incorrect result: %s (must be %s)", rank.vp[tc.addr].String(), tc.want)

	if s != nil {
		b, err := s.GetRawKV(vprKey(tc.addr[:]))
		assert.NoError(t, err, "fail to get a voting power")
		v := new(big.Int).SetBytes(b)
		assert.True(t, v.Cmp(tc.want) == 0,
			"value mismatch: want: %s, actual: %s", tc.want, v)
	}
}

func initVprtTest(t *testing.T) {
	if isInitialized() {
		return
	}

	vprChainStateDB = state.NewChainStateDB()
	_ = vprChainStateDB.Init(string(db.BadgerImpl), "test", nil, false)
	vprStateDB = vprChainStateDB.GetStateDB()
	genesis := types.GetTestGenesis()

	err := vprChainStateDB.SetGenesis(genesis, nil)
	assert.NoError(t, err, "failed init")

	for i := int32(0); i < vprMax; i++ {
		rank.Set(genAddr(i), new(big.Int).SetUint64(10000))
		rank.Apply(nil)
	}

	initializedVprtTest = true
}

func finalizeTest() {
	_ = vprChainStateDB.Close()
	_ = os.RemoveAll("test")
}

func isInitialized() bool {
	return initializedVprtTest
}

func genAddr(i int32) types.AccountID {
	dig := sha256.New()
	binary.Write(dig, binary.LittleEndian, i)
	return types.ToAccountID(dig.Sum(nil))
}

func commit() error {
	return vprStateDB.Commit()
}

func TestVprtOp(t *testing.T) {
	initVprtTest(t)
	defer finalizeTest()

	testCases := []vprTC{
		{
			addr: genAddr(10),
			ops:  []vprOpt{{opAdd, valHundred}, {opSub, valTen}},
			want: new(big.Int).SetUint64(10090),
		},
		{
			addr: genAddr(11),
			ops:  []vprOpt{{opSub, valTen}, {opAdd, valHundred}},
			want: new(big.Int).SetUint64(10090),
		},
		{
			addr: genAddr(12),
			ops:  []vprOpt{{opAdd, valHundred}, {opAdd, valHundred}},
			want: new(big.Int).SetUint64(10200),
		},
		{
			addr: genAddr(13),
			ops:  []vprOpt{{opAdd, valTen}, {opAdd, valTen}},
			want: new(big.Int).SetUint64(10020),
		},
		{
			addr: genAddr(14),
			ops:  []vprOpt{{opSub, valTen}, {opSub, valTen}},
			want: new(big.Int).SetUint64(9980),
		},
	}

	s, err := vprStateDB.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	assert.NoError(t, err, "fail to open the system contract state")

	for _, tc := range testCases {
		tc.run(t, s)
	}
}
