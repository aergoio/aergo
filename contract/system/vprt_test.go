package system

import (
	"container/list"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

const (
	testVprMax = vprMax

	opAdd = iota
	opSub
)

var (
	valHundred = new(big.Int).SetUint64(100)
	valTen     = new(big.Int).SetUint64(10)

	vprOP = []func(types.AccountID, types.Address, *big.Int){
		opAdd: func(id types.AccountID, addr types.Address, opr *big.Int) {
			votingPowerRank.add(id, addr, opr)
		},
		opSub: func(id types.AccountID, addr types.Address, opr *big.Int) {
			votingPowerRank.sub(id, addr, opr)
		},
	}

	vprChainStateDB     *state.ChainStateDB
	vprStateDB          *state.StateDB
	initializedVprtTest bool
)

func (v *vpr) checkValidity(t *testing.T) {
	sum1 := &big.Int{}
	sum2 := &big.Int{}
	sum3 := &big.Int{}

	low := v.getLowest().getPower()
	for _, e := range v.voters.powers {
		pow := e.getPower()
		sum1.Add(sum1, pow)
		assert.True(t, low.Cmp(pow) <= 0, "invalid lowest power voter")
	}
	assert.True(t, sum1.Cmp(v.getTotalPower()) == 0, "voting power map inconsistent with total voting power")

	for _, i := range v.voters.members.Values() {
		sum2.Add(sum2, i.(*votingPower).getPower())
	}

	for i, l := range v.store.buckets {
		var next *list.Element
		for e := l.Front(); e != nil; e = next {
			if next = e.Next(); next != nil {
				ind := v.store.cmp(toVotingPower(e), toVotingPower(next))
				assert.True(t, ind > 0, "bucket[%v] not ordered", i)
			}
			sum3.Add(sum3, toVotingPower(e).getPower())
		}
	}
	assert.True(t, sum3.Cmp(v.getTotalPower()) == 0, "voting power buckects inconsistent with total voting power")
}

func initVpr() {
	votingPowerRank = newVpr()
}

func defaultVpr() *vpr {
	return votingPowerRank
}

func store(t *testing.T, s *state.ContractState) {
	err := vprStateDB.StageContractState(s)
	assert.NoError(t, err, "fail to stage")
	err = vprStateDB.Update()
	assert.NoError(t, err, "fail to update")
	err = vprStateDB.Commit()
	assert.NoError(t, err, "fail to commit")
}

type vprOpt struct {
	op  int
	arg *big.Int
}

type vprTC struct {
	seed uint32
	ops  []vprOpt
	want *big.Int
}

func (tc *vprTC) run(t *testing.T) {
	for _, o := range tc.ops {
		id, addr := genAddr(tc.seed)
		vprOP[o.op](id, addr, o.arg)
	}
}

func (tc *vprTC) check(t *testing.T) {
	id, _ := genAddr(tc.seed)

	assert.True(t,
		votingPowerRank.votingPowerOf(id).Cmp(tc.want) == 0,
		"incorrect result: %s (must be %s)", votingPowerRank.votingPowerOf(id).String(), tc.want)
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

	store(t, s)
}

func initDB(t *testing.T) {
	vprChainStateDB = state.NewChainStateDB()
	_ = vprChainStateDB.Init(string(db.BadgerImpl), "test", nil, false)
	vprStateDB = vprChainStateDB.GetStateDB()
	genesis := types.GetTestGenesis()

	err := vprChainStateDB.SetGenesis(genesis, nil)
	assert.NoError(t, err, "failed init")
}

func getStateRoot() []byte {
	return vprStateDB.GetRoot()
}

func openSystemAccountWith(root []byte) *state.ContractState {
	s, err := vprChainStateDB.OpenNewStateDB(root).OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	if err != nil {
		return nil
	}

	return s
}

func initRankTableRandSc(rankMax uint32, s *state.ContractState) {
	votingPowerRank = newVpr()
	max := new(big.Int).Mul(binSize, new(big.Int).SetUint64(5))
	src := rand.New(rand.NewSource(0))
	for i := uint32(0); i < rankMax; i++ {
		id, addr := genAddr(i)
		votingPowerRank.add(id, addr, new(big.Int).Rand(src, max))
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
		id, addr := genAddr(i)
		votingPowerRank.add(id, addr, binSize)
		votingPowerRank.apply(nil)
	}
}

func isInitialized() bool {
	return initializedVprtTest
}

func genAddr(i uint32) (types.AccountID, types.Address) {
	s := fmt.Sprintf("aergo.%v", i)
	addr, _ := types.DecodeAddress(s)
	return types.ToAccountID(addr), addr
}

func commit() error {
	return vprStateDB.Commit()
}

func TestValidateInitVprt(t *testing.T) {
	assert.Equal(t, "1000000", million.String(), "million is not valid. check contract/system/vprt.go")
	assert.Equal(t, "5045760000000000000", annualRewardM.String(), "annualRewardM is not valid. check contract/system/vprt.go")
	assert.Equal(t, "5045760000000000000000000", annualReward.String(), "annualReward is not valid. check contract/system/vprt.go")
	assert.Equal(t, "160000000000000000", defaultReward.String(), "defaultReward is not valid. check contract/system/vprt.go")
	assert.Equal(t, "10000000000000000000000", binSize.String(), "binSize is not valid. check contract/system/vprt.go")
}

func TestVprOp(t *testing.T) {
	initVprtTest(t, func() { initRankTable(testVprMax) })
	defer finalizeVprtTest()

	rValue := func(rhs int64) *big.Int {
		defVal := new(big.Int).Set(binSize)
		return new(big.Int).Set(defVal).Add(defVal, new(big.Int).SetInt64(rhs))
	}

	testCases := []vprTC{
		{
			seed: 10,
			ops:  []vprOpt{{opAdd, valHundred}, {opSub, valTen}},
			want: rValue(90),
		},
		{
			seed: 11,
			ops:  []vprOpt{{opSub, valTen}, {opAdd, valHundred}},
			want: rValue(90),
		},
		{
			seed: 12,
			ops:  []vprOpt{{opAdd, valHundred}, {opAdd, valHundred}},
			want: rValue(200),
		},
		{
			seed: 13,
			ops:  []vprOpt{{opAdd, valTen}, {opAdd, valTen}},
			want: rValue(20),
		},
		{
			seed: 14,
			ops:  []vprOpt{{opSub, valTen}, {opSub, valTen}},
			want: rValue(-20),
		},
		{
			seed: 15,
			ops:  []vprOpt{{opSub, valTen}, {opSub, valTen}, {opSub, valTen}},
			want: rValue(-30),
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

	store(t, s)

	s = openSystemAccount(t)

	lRank, err := loadVpr(s)
	assert.NoError(t, err, "fail to load")
	assert.Equal(t, n, lRank.voters.Count(), "size mismatch: voting power")
}

func TestVprTable(t *testing.T) {
	initVprtTest(t, func() { initRankTableRand(testVprMax) })
	defer finalizeVprtTest()

	for i, l := range votingPowerRank.store.buckets {
		for e := l.Front(); e.Next() != nil; e = e.Next() {
			curr := e.Value.(*votingPower)
			next := e.Next().Value.(*votingPower)
			assert.True(t, curr.getID() != next.getID(), "duplicate elems")
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
			expect: 55,
		},
		{
			pow:    conv("5000000000000"),
			expect: 49,
		},
	}

	for _, tc := range tcs {
		id, addr := genAddr(0)
		orig := newVotingPower(addr, id, tc.pow)
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
			seed: 10,
			ops:  []vprOpt{{opAdd, valHundred}, {opSub, valTen}},
		},
		{
			seed: 11,
			ops:  []vprOpt{{opSub, valTen}, {opAdd, valHundred}},
		},
		{
			seed: 12,
			ops:  []vprOpt{{opAdd, valHundred}, {opAdd, valHundred}},
		},
		{
			seed: 13,
			ops:  []vprOpt{{opAdd, valTen}, {opAdd, valTen}},
		},
		{
			seed: 14,
			ops:  []vprOpt{{opSub, valTen}, {opSub, valTen}},
		},
		{
			seed: 15,
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
	votingPowerRank.voters.dump(os.Stdout, 10)
	votingPowerRank.voters.dump(os.Stdout, 0)
}

func TestVprSingleWinner(t *testing.T) {
	const nVoters = 1

	initVprtTestWithSc(t, func(s *state.ContractState) { initRankTableRandSc(nVoters, s) })
	defer finalizeVprtTest()

	stat := make(map[types.AccountID]uint16)

	for i := int64(0); i < 1000; i++ {
		addr, err := votingPowerRank.pickVotingRewardWinner(i)
		assert.NoError(t, err)
		id := types.ToAccountID(addr)
		count := stat[id]
		stat[id] = count + 1
	}

	for addr, count := range stat {
		fmt.Printf("%v: pwr = %v, wins # = %v\n",
			addr, votingPowerRank.votingPowerOf(addr), count)
	}
}

func TestVprPickWinner(t *testing.T) {
	const nVoters = 1000

	initVprtTestWithSc(t, func(s *state.ContractState) { initRankTableRandSc(nVoters, s) })
	defer finalizeVprtTest()

	stat := make(map[types.AccountID]uint16)

	for i := int64(0); i < nVoters; i++ {
		addr, err := votingPowerRank.pickVotingRewardWinner(i)
		assert.NoError(t, err)
		id := types.ToAccountID(addr)
		count := stat[id]
		stat[id] = count + 1
	}

	for addr, count := range stat {
		fmt.Printf("%v: pwr = %v, wins # = %v\n",
			addr, votingPowerRank.votingPowerOf(addr), count)
	}
}

func TestVprZeroPowerVoter(t *testing.T) {
	testCases := []struct {
		pwr *big.Int
		chk func(*testing.T)
	}{
		{
			pwr: new(big.Int).SetUint64(0),
			chk: func(t *testing.T) {
				assert.Nil(t, votingPowerRank.getLowest(), "zero power votier must not be added.")
			},
		},
		{
			pwr: new(big.Int).SetUint64(1),
			chk: func(t *testing.T) {
				assert.True(t, votingPowerRank.getLowest().cmp(new(big.Int).SetUint64(1)) == 0, "invalid lowest power voter.")
			},
		},
		{
			pwr: new(big.Int).SetUint64(10),
			chk: func(t *testing.T) {
				assert.True(t, votingPowerRank.getLowest().cmp(new(big.Int).SetUint64(1)) == 0, "invalid lowest power voter.")
			},
		},
		{
			pwr: new(big.Int).SetUint64(100),
			chk: func(t *testing.T) {
				assert.True(t, votingPowerRank.getLowest().cmp(new(big.Int).SetUint64(1)) == 0, "invalid lowest power voter.")
			},
		},
	}

	initVprtTestWithSc(t, func(s *state.ContractState) {
		votingPowerRank = newVpr()
		for i, tc := range testCases {
			fmt.Printf("idx: %v, pwd: %v\n", i, tc.pwr)
			id, addr := genAddr(uint32(i))
			votingPowerRank.add(id, addr, tc.pwr)
			votingPowerRank.apply(s)
			tc.chk(t)
		}
	})

	defer finalizeVprtTest()

	s := openSystemAccount(t)

	idx := 1
	id, addr := genAddr(uint32(idx))
	votingPowerRank.sub(id, addr, new(big.Int).SetUint64(1))
	votingPowerRank.apply(s)
	assert.True(t, votingPowerRank.getLowest().cmp(new(big.Int).SetUint64(10)) == 0,
		"invalid lowest power(%v) voter.", votingPowerRank.getLowest().getPower())

	store(t, s)

	s = openSystemAccount(t)
	lRank, err := loadVpr(s)
	assert.NoError(t, err, "fail to load")
	assert.Equal(t, votingPowerRank.voters.Count(), lRank.voters.Count(), "size mismatch: voting power")

	assert.True(t, votingPowerRank.equals(lRank), "VPR mismatch")
}

func TestVprReorg(t *testing.T) {
	type testCase struct {
		pwr *big.Int
		chk func(*testing.T)
	}

	doTest := func(i int, tc testCase, s *state.ContractState) {
		fmt.Printf("idx: %v, pwd: %v\n", i, tc.pwr)
		id, addr := genAddr(uint32(i))
		votingPowerRank.add(id, addr, tc.pwr)
		votingPowerRank.apply(s)
		store(t, s)
		tc.chk(t)
	}

	initVprtTestWithSc(t, func(s *state.ContractState) {
		initVpr()
	})
	defer finalizeVprtTest()

	testCases := []testCase{

		{
			pwr: new(big.Int).SetUint64(1),
			chk: func(t *testing.T) {
				assert.True(t, votingPowerRank.getLowest().cmp(new(big.Int).SetUint64(1)) == 0, "invalid lowest power voter.")
			},
		},
		{
			pwr: new(big.Int).SetUint64(10),
			chk: func(t *testing.T) {
				assert.True(t, votingPowerRank.getLowest().cmp(new(big.Int).SetUint64(1)) == 0, "invalid lowest power voter.")
			},
		},
	}

	sRoots := make([][]byte, len(testCases))
	totalPowers := make([]*big.Int, len(testCases))

	for i, tc := range testCases {
		s := openSystemAccount(t)
		doTest(i, tc, s)
		sRoots[i] = getStateRoot()
		totalPowers[i] = defaultVpr().getTotalPower()
	}

	for i, root := range sRoots {
		s := openSystemAccountWith(root)
		assert.NotNil(t, s, "failed to open the system account")
		InitVotingPowerRank(s)
		assert.Equal(t, defaultVpr().getTotalPower(), totalPowers[i], "invalid total voting power")
	}
}
