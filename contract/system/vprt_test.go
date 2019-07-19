package system

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/internal/enc"
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
			rank.add(addr, opr)
		},
		opSub: func(addr types.AccountID, opr *big.Int) {
			rank.sub(addr, opr)
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
	rank.apply(s)
	assert.True(t,
		rank.votingPowerOf(tc.addr).Cmp(tc.want) == 0,
		"incorrect result: %s (must be %s)", rank.votingPowerOf(tc.addr).String(), tc.want)
}

func initVprtTest(t *testing.T, initTable func(rankMax int32)) {
	vprChainStateDB = state.NewChainStateDB()
	_ = vprChainStateDB.Init(string(db.BadgerImpl), "test", nil, false)
	vprStateDB = vprChainStateDB.GetStateDB()
	genesis := types.GetTestGenesis()

	err := vprChainStateDB.SetGenesis(genesis, nil)
	assert.NoError(t, err, "failed init")

	initTable(vprMax)
}

func finalizeVprtTest() {
	_ = vprChainStateDB.Close()
	_ = os.RemoveAll("test")
}

func initRankTable(rankMax int32) {
	for i := int32(0); i < rankMax; i++ {
		rank.add(genAddr(i), new(big.Int).SetUint64(10000))
		rank.apply(nil)
	}
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

func TestVprOp(t *testing.T) {
	initVprtTest(t, initRankTable)
	defer finalizeVprtTest()

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
	fmt.Printf(
		"(before) state, contract: %s, %s\n",
		enc.ToString(vprStateDB.GetRoot()),
		enc.ToString(s.GetStorageRoot()))

	for _, tc := range testCases {
		tc.run(t, s)
	}

	err = vprStateDB.StageContractState(s)
	assert.NoError(t, err, "fail to stage")
	err = vprStateDB.Update()
	assert.NoError(t, err, "fail to update")
	err = vprStateDB.Commit()
	assert.NoError(t, err, "fail to commit")

	s, err = vprStateDB.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	assert.NoError(t, err, "fail to open the system contract state")
	fmt.Printf(
		"(after) state, contract: %s, %s\n",
		enc.ToString(vprStateDB.GetRoot()),
		enc.ToString(s.GetStorageRoot()))
}

func TestVprTable(t *testing.T) {
	initVprtTest(t, initRankTableRand)
	defer finalizeVprtTest()

	for i, l := range rank.table.buckets {
		for e := l.Front(); e.Next() != nil; e = e.Next() {
			curr := e.Value.(*votingPower)
			next := e.Next().Value.(*votingPower)
			assert.True(t, curr.addr != next.addr, "duplicate elems")
			cmp := curr.power.Cmp(next.power)
			assert.True(t, cmp == 0 || cmp == 1, "unordered bucket found: idx = %v", i)
		}
	}
}

func initRankTableRand(rankMax int32) {
	rank = newVpr()
	max := new(big.Int).SetUint64(20000)
	src := rand.New(rand.NewSource(0))
	for i := int32(0); i < rankMax; i++ {
		rank.add(genAddr(i), new(big.Int).Rand(src, max))
		rank.apply(nil)
	}
}

func TestVotingPowerCodec(t *testing.T) {

	conv := func(s string) *big.Int {
		p, ok := new(big.Int).SetString(s, 10)
		assert.True(t, ok, "number conversion failed")
		return p
	}

	tcs := []struct {
		pow    *big.Int
		expect int
	}{
		{
			pow:    conv("500000000000000000000000000"),
			expect: 48,
		},
		{
			pow:    conv("5000000000000"),
			expect: 42,
		},
	}

	for _, tc := range tcs {
		orig := newVotingPower(genAddr(0), tc.pow)
		b := orig.marshal()
		assert.Equal(t, tc.expect, len(b))

		dec := &votingPower{}
		n := dec.unmarshal(b)
		assert.Equal(t, len(b)-4, int(n))

		assert.Equal(t, orig.addr, dec.addr)
		assert.True(t, orig.power.Cmp(dec.power) == 0, "fail to decode")
	}
}
