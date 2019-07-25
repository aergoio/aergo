package system

import (
	"container/list"
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
			votingPowerRank.add(addr, opr)
		},
		opSub: func(addr types.AccountID, opr *big.Int) {
			votingPowerRank.sub(addr, opr)
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

func (tc *vprTC) run(t *testing.T) {
	for _, o := range tc.ops {
		vprOP[o.op](tc.addr, o.arg)
	}
}

func (tc *vprTC) check(t *testing.T) {
	assert.True(t,
		votingPowerRank.votingPowerOf(tc.addr).Cmp(tc.want) == 0,
		"incorrect result: %s (must be %s)", votingPowerRank.votingPowerOf(tc.addr).String(), tc.want)
}

func initVprtTest(t *testing.T, initTable func()) {
	initDB(t)
	initTable()
}

func initVprtTestWithSc(t *testing.T, initTable func(*state.ContractState)) {
	initDB(t)

	s, err := vprStateDB.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	assert.NoError(t, err, "fail to open the system contract state")

	initTable(s)

	err = vprStateDB.StageContractState(s)
	assert.NoError(t, err, "fail to stage")
	err = vprStateDB.Update()
	assert.NoError(t, err, "fail to update")
	err = vprStateDB.Commit()
	assert.NoError(t, err, "fail to commit")
}

func initDB(t *testing.T) {
	vprChainStateDB = state.NewChainStateDB()
	_ = vprChainStateDB.Init(string(db.BadgerImpl), "test", nil, false)
	vprStateDB = vprChainStateDB.GetStateDB()
	genesis := types.GetTestGenesis()

	err := vprChainStateDB.SetGenesis(genesis, nil)
	assert.NoError(t, err, "failed init")
}

func initRankTableRandSc(rankMax uint32, s *state.ContractState) {
	votingPowerRank = newVpr()
	max := new(big.Int).SetUint64(20000)
	src := rand.New(rand.NewSource(0))
	for i := uint32(0); i < rankMax; i++ {
		votingPowerRank.add(genAddr(i), new(big.Int).Rand(src, max))
	}
	votingPowerRank.apply(s)
}

func initRankTableRand(rankMax uint32) {
	initRankTableRandSc(rankMax, nil)
}

func openSystemAccount(t *testing.T) *state.ContractState {
	s, err := vprStateDB.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	assert.NoError(t, err, "fail to open the system contract state")
	fmt.Printf(
		"(after) state, contract: %s, %s\n",
		enc.ToString(vprStateDB.GetRoot()),
		enc.ToString(s.GetStorageRoot()))

	return s
}

func finalizeVprtTest() {
	_ = vprChainStateDB.Close()
	_ = os.RemoveAll("test")
}

func initRankTable(rankMax uint32) {
	votingPowerRank = newVpr()
	for i := uint32(0); i < rankMax; i++ {
		votingPowerRank.add(genAddr(i), new(big.Int).SetUint64(10000))
		votingPowerRank.apply(nil)
	}
}

func isInitialized() bool {
	return initializedVprtTest
}

func genAddr(i uint32) types.AccountID {
	dig := sha256.New()
	binary.Write(dig, binary.LittleEndian, i)
	return types.ToAccountID(dig.Sum(nil))
}

func commit() error {
	return vprStateDB.Commit()
}

func TestVprOp(t *testing.T) {
	initVprtTest(t, func() { initRankTable(vprMax) })
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
		{
			addr: genAddr(15),
			ops:  []vprOpt{{opSub, valTen}, {opSub, valTen}, {opSub, valTen}},
			want: new(big.Int).SetUint64(9970),
		},
	}

	s := openSystemAccount(t)

	for _, tc := range testCases {
		tc.run(t)
	}
	n, err := votingPowerRank.apply(s)
	assert.NoError(t, err, "fail to update the voting power ranking")
	for _, tc := range testCases {
		tc.check(t)
	}

	err = vprStateDB.StageContractState(s)
	assert.NoError(t, err, "fail to stage")
	err = vprStateDB.Update()
	assert.NoError(t, err, "fail to update")
	err = vprStateDB.Commit()
	assert.NoError(t, err, "fail to commit")

	s = openSystemAccount(t)

	lRank, err := loadVpr(s)
	assert.NoError(t, err, "fail to load")
	assert.Equal(t, n, lRank.voters.Count(), "size mismatch: voting power")
}

func TestVprTable(t *testing.T) {
	initVprtTest(t, func() { initRankTableRand(vprMax) })
	defer finalizeVprtTest()

	for i, l := range votingPowerRank.store.buckets {
		for e := l.Front(); e.Next() != nil; e = e.Next() {
			curr := e.Value.(*votingPower)
			next := e.Next().Value.(*votingPower)
			assert.True(t, curr.getAddr() != next.getAddr(), "duplicate elems")
			cmp := curr.getPower().Cmp(next.getPower())
			assert.True(t, cmp == 0 || cmp == 1, "unordered bucket found: idx = %v", i)
		}
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
		assert.Equal(t, len(b), int(n))

		assert.Equal(t, orig.getAddr(), dec.getAddr())
		assert.True(t, orig.getPower().Cmp(dec.getPower()) == 0, "fail to decode")
	}
}

func TestVprLoader(t *testing.T) {
	const nVoters = 100

	initVprtTestWithSc(t, func(s *state.ContractState) { initRankTableRandSc(nVoters, s) })
	defer finalizeVprtTest()
	assert.Equal(t, nVoters, votingPowerRank.voters.Count(), "size mismatch: voting powers")
	assert.Equal(t, nVoters,
		func() int {
			sum := 0
			for i := uint8(0); i < vprBucketsMax; i++ {
				if l := votingPowerRank.store.buckets[i]; l != nil {
					sum += l.Len()
				}
			}
			return sum
		}(),
		"size mismatch: voting powers")

	s := openSystemAccount(t)
	r, err := loadVpr(s)
	assert.NoError(t, err, "fail to load")
	assert.Equal(t, nVoters, r.voters.Count(), "size mismatch: voting powers")

	r.checkValidity(t)
}

func TestVprTotalPower(t *testing.T) {
	const nVoters = 1000

	initVprtTestWithSc(t, func(s *state.ContractState) { initRankTableRandSc(nVoters, s) })
	defer finalizeVprtTest()

	testCases := []vprTC{
		{
			addr: genAddr(10),
			ops:  []vprOpt{{opAdd, valHundred}, {opSub, valTen}},
		},
		{
			addr: genAddr(11),
			ops:  []vprOpt{{opSub, valTen}, {opAdd, valHundred}},
		},
		{
			addr: genAddr(12),
			ops:  []vprOpt{{opAdd, valHundred}, {opAdd, valHundred}},
		},
		{
			addr: genAddr(13),
			ops:  []vprOpt{{opAdd, valTen}, {opAdd, valTen}},
		},
		{
			addr: genAddr(14),
			ops:  []vprOpt{{opSub, valTen}, {opSub, valTen}},
		},
		{
			addr: genAddr(15),
			ops:  []vprOpt{{opSub, valTen}, {opSub, valTen}, {opSub, valTen}},
		},
	}

	s := openSystemAccount(t)

	for _, tc := range testCases {
		tc.run(t)
	}
	_, err := votingPowerRank.apply(s)
	assert.NoError(t, err, "fail to update the voting power ranking")

	votingPowerRank.checkValidity(t)
}

func (v *vpr) checkValidity(t *testing.T) {
	sum1 := &big.Int{}
	sum2 := &big.Int{}

	low := v.lowest().getPower()
	for _, e := range v.voters.powers {
		pow := toVotingPower(e).getPower()
		sum1.Add(sum1, pow)
		assert.True(t, low.Cmp(pow) <= 0, "invalid lowest power voter")
	}
	assert.True(t, sum1.Cmp(v.getTotalPower()) == 0, "voting power map inconsistent with total voting power")

	for i, l := range v.store.buckets {
		var next *list.Element
		for e := l.Front(); e != nil; e = next {
			if next = e.Next(); next != nil {
				ind := v.store.cmp(toVotingPower(e), toVotingPower(next))
				assert.True(t, ind > 0, "bucket[%v] not ordered", i)
			}
			sum2.Add(sum2, toVotingPower(e).getPower())
		}
	}
	assert.True(t, sum2.Cmp(v.getTotalPower()) == 0, "voting power buckects inconsistent with total voting power")
}
