package system

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"reflect"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	rb "github.com/emirpasic/gods/trees/redblacktree"
	jsoniter "github.com/json-iterator/go"
	"github.com/sanity-io/litter"
)

// ********************************** WARNING **********************************
// All the functions and the methods below must be properly synchronized by
// acquiring the chain lock (chain.Lock / chain.Unlock) when they are called.
// *****************************************************************************

const (
	vprMax = uint32(50000)
	// vprBucketsMax must be smaller than 256. A bigger number is regarded as
	// 256.
	vprBucketsMax = 71
)

var (
	ErrNoVotingRewardWinner = errors.New("voting reward: no winner")
	ErrNoVotingRewardRank   = errors.New("voting reward rank: not initialized")

	zeroValue     = types.NewZeroAmount()
	vprKeyPrefix  = []byte("VotingPowerBucket/")
	million       = types.NewAmount(1e6, types.Aer)          // 1,000,000 Aer
	annualRewardM = types.NewAmount(5045760000, types.Gaer)  // 5,045,760,000 Gaer
	annualReward  = new(big.Int).Mul(annualRewardM, million) // 5,045,760 AERGO
	oneYIS        = new(big.Int).SetUint64(365 * 24 * 60 * 60)
	defaultReward = new(big.Int).Div(annualReward, oneYIS) // 0.16 AERGO per sec
	binSize       = types.NewAmount(1e4, types.Aergo)      // 10,000 AERGO

	votingPowerRank *vpr

	vprLogger = log.NewLogger("vpr")
	jsonIter  = jsoniter.ConfigCompatibleWithStandardLibrary
)

type dataSetter interface {
	SetData(key, value []byte) error
}

type dataGetter interface {
	GetData(key []byte) ([]byte, error)
}

// InitVotingPowerRank reads the stored data from s and initializes the Voting
// Power Rank, which contains each voters's voting power.
func InitVotingPowerRank(s dataGetter) (err error) {
	votingPowerRank, err = loadVpr(s)

	return
}

type votingPower struct {
	id    types.AccountID
	addr  types.Address
	power *big.Int
}

func newVotingPower(addr []byte, id types.AccountID, pow *big.Int) *votingPower {
	return &votingPower{addr: addr, id: id, power: pow}
}

func (vp *votingPower) getAddr() []byte {
	return vp.addr
}

func (vp *votingPower) setAddr(addr []byte) {
	vp.addr = addr
}

func (vp *votingPower) getID() types.AccountID {
	return vp.id
}

func (vp *votingPower) idBytes() []byte {
	return vp.id[:]
}

func (vp *votingPower) setID(id types.AccountID) {
	vp.id = id
}

func (vp *votingPower) getPower() *big.Int {
	return vp.power
}

func (vp *votingPower) setPower(pow *big.Int) {
	vp.power = pow
}

func (vp *votingPower) cmp(pow *big.Int) int {
	return vp.getPower().Cmp(pow)
}

func (vp *votingPower) lt(rhs *votingPower) bool {
	return vp.cmp(rhs.getPower()) < 0
}

func (vp *votingPower) le(rhs *votingPower) bool {
	return vp.lt(rhs) || vp.cmp(rhs.getPower()) == 0
}

func (vp *votingPower) isZero() bool {
	return vp.getPower().Cmp(zeroValue) == 0
}

func (vp *votingPower) marshal() []byte {
	var buf bytes.Buffer

	// Account ID
	buf.Write(vp.idBytes()) // 32 bytes

	// Address
	binary.Write(&buf, binary.LittleEndian, uint16(len(vp.addr)))
	buf.Write(vp.addr)

	// Voting Power
	pwr := vp.getPower().Bytes()
	binary.Write(&buf, binary.LittleEndian, uint16(len(pwr)))
	buf.Write(pwr)

	return buf.Bytes()
}

func (vp *votingPower) unmarshal(b []byte) uint32 {
	var sz1, sz2 uint16

	r := bytes.NewReader(b)
	r.Seek(int64(32), 0)
	binary.Read(r, binary.LittleEndian, &sz1)
	// <Accound ID (32)> + <address length (2)> + <address>
	r.Seek(int64(32+2+sz1), 0)
	binary.Read(r, binary.LittleEndian, &sz2)

	vp.setID(types.AccountID(types.ToHashID(b[:32])))
	vp.setAddr(b[34 : 34+sz1])

	if int(36+sz1+sz2) < len(b) {
		vp.setPower(new(big.Int).SetBytes(b[36+sz1 : 36+sz1+sz2]))
	} else {
		vp.setPower(new(big.Int).SetBytes(b[36+sz1:]))
	}

	return 36 + uint32(sz1) + uint32(sz2)
}

func (vp *votingPower) toJSON() ([]byte, error) {
	b, err := jsonIter.Marshal(&struct {
		Address string
		Power   string
	}{types.EncodeAddress(vp.getAddr()), vp.getPower().String()})
	if err != nil {
		return nil, err
	}
	return b, nil
}

type vprStore struct {
	buckets map[uint8]*list.List
	cmp     func(lhs *votingPower, rhs *votingPower) int
}

func newVprStore(bucketsMax uint32) *vprStore {
	return &vprStore{
		buckets: make(map[uint8]*list.List, bucketsMax),
		cmp: func(lhs *votingPower, rhs *votingPower) int {
			return bytes.Compare(lhs.idBytes(), rhs.idBytes())
		},
	}
}

func (b *vprStore) equals(rhs *vprStore) bool {
	if !reflect.DeepEqual(b.buckets, rhs.buckets) {
		return false
	}
	return true
}

func (b *vprStore) update(vp *votingPower) (idx uint8) {
	var (
		id  = vp.getID()
		pow = vp.getPower()
	)

	idx = getBucketIdx(id)

	var (
		bu    *list.List
		exist bool
	)

	if bu, exist = b.buckets[idx]; !exist {
		bu = list.New()
		b.buckets[idx] = bu
	}

	v := remove(bu, id)
	if vp.isZero() {
		// A zero power voter must be removed.
		return
	}

	if v != nil {
		v.setPower(pow)
	} else {
		v = vp
	}

	orderedListAdd(bu, v,
		func(e *list.Element) bool {
			ind := b.cmp(toVotingPower(e), v)
			return ind <= 0
		},
	)

	return
}

type predicate func(e *list.Element) bool

func search(l *list.List, match func(e *list.Element) bool) *list.Element {
	for e := l.Front(); e != nil; e = e.Next() {
		if match(e) {
			return e
		}
	}
	return nil
}

func orderedListMove(l *list.List, e *list.Element, match predicate) {
	if m := search(l, match); m != nil {
		l.MoveBefore(e, m)
	} else {
		l.MoveToBack(e)
	}
}

func orderedListAdd(l *list.List, e *votingPower, match predicate) *list.Element {
	var voter *list.Element

	if m := search(l, match); m != nil {
		voter = l.InsertBefore(e, m)
	} else {
		voter = l.PushBack(e)
	}

	return voter
}

func (b *vprStore) addTail(i uint8, vp *votingPower) {
	var l *list.List
	if l = b.buckets[i]; l == nil {
		l = list.New()
		b.buckets[i] = l
	}
	l.PushBack(vp)
}

func (b *vprStore) write(s dataSetter, i uint8) error {
	var buf bytes.Buffer

	l := b.buckets[i]
	for e := l.Front(); e != nil; e = e.Next() {
		buf.Write(toVotingPower(e).marshal())
	}

	return s.SetData(vprKey(i), buf.Bytes())
}

func (b *vprStore) read(s dataGetter, i uint8) ([]*votingPower, error) {
	buf, err := s.GetData(vprKey(i))
	if err != nil {
		return nil, err
	}
	vps := make([]*votingPower, 0, 10)
	for off := 0; off < len(buf); {
		vp := &votingPower{}
		off += int(vp.unmarshal(buf[off:]))
		vps = append(vps, vp)
	}
	return vps, nil
}

func remove(bu *list.List, id types.AccountID) *votingPower {
	for e := bu.Front(); e != nil; e = e.Next() {
		if v := toVotingPower(e); id == v.getID() {
			return bu.Remove(e).(*votingPower)
		}
	}
	return nil
}

func getBucketIdx(addr types.AccountID) uint8 {
	return uint8(addr[0]) % vprBucketsMax
}

func (b *vprStore) getBucket(addr types.AccountID) *list.List {
	if b, exist := b.buckets[getBucketIdx(addr)]; exist {
		return b
	}
	return nil
}

type topVoters struct {
	members *rb.Tree
	max     uint32
	powers  map[types.AccountID]*votingPower
}

func newTopVoters(max uint32) *topVoters {
	cmp := func(lhs, rhs interface{}) int {
		var (
			l = lhs.(*votingPower)
			r = rhs.(*votingPower)
		)

		pwrInd := l.getPower().Cmp(r.getPower())
		if pwrInd == 0 {
			return -bytes.Compare(l.idBytes(), r.idBytes())
		}
		return -pwrInd
	}

	return &topVoters{
		members: rb.NewWith(cmp),
		max:     max,
		powers:  make(map[types.AccountID]*votingPower),
	}
}

func (tv *topVoters) equals(rhs *topVoters) bool {
	if !reflect.DeepEqual(tv.powers, rhs.powers) {
		return false
	}

	if !reflect.DeepEqual(tv.members.Keys(), rhs.members.Keys()) {
		return false
	}

	if tv.max != rhs.max {
		return false
	}

	return true
}

func (tv *topVoters) Count() int {
	return len(tv.powers)
}

func (tv *topVoters) get(id types.AccountID) *votingPower {
	return tv.powers[id]
}

func (tv *topVoters) set(id types.AccountID, v *votingPower) {
	tv.powers[id] = v
}

func (tv *topVoters) del(id types.AccountID) {
	delete(tv.powers, id)
}

func (tv *topVoters) getVotingPower(addr types.AccountID) *votingPower {
	return tv.powers[addr]
}

func (tv *topVoters) powerOf(addr types.AccountID) *big.Int {
	if vp := tv.getVotingPower(addr); vp != nil {
		return vp.getPower()
	}
	return nil
}

func (tv *topVoters) lowest() *votingPower {
	lowest := tv.members.Right()
	if lowest == nil {
		return nil
	}
	return lowest.Value.(*votingPower)
}

func (tv *topVoters) update(v *votingPower) (vp *votingPower) {
	// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
	// TODO: Maintain len(tv.powers) <= tv.max
	//
	// * Rejection & Eviction Rule:
	//
	// 1. Reject if the new voter has a voting power lesser than the lowest.
	//
	// 2. Evict the lowest if the new voter has a voting power larger than
	//    anyone among the current voting power rankers.
	//
	// 3. Randomly select & evict one among the lowest voters if the new voter
	//    has the same voting power as the lowest voter.
	//
	// -------------------------------------------------------------------------
	//
	// ISSUE - There exists some unfair case as follows: The VPR slots are
	// fully occupied (len(tv.powers) == tv.max) so that a voter A is rejected
	// by the rule above. Afterwards, one voter cancels his staking and is
	// removed from the VPR. In such a situtation, any voter cating a vote will
	// be unconditionally included into the VPR since one slot is available for
	// him even if his voting power is less than the aforementioned voter A.
	// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
	if vp = tv.get(v.getID()); vp != nil {
		tv.members.Remove(vp)
		if vp.isZero() {
			tv.del(vp.getID())
		} else {
			vp.setPower(v.getPower())
			tv.members.Put(vp, vp)
		}
	} else {
		vp = v

		tv.members.Put(v, v)
		tv.set(v.getID(), v)
	}

	return
}

func (tv *topVoters) dump(w io.Writer, topN int) error {
	if tv == nil {
		fmt.Fprintf(w, "nothing to dump!")
		return nil
	}

	endingIndex := func() int {
		size := tv.members.Size()

		if topN < size && topN > 0 {
			return topN - 1
		}
		return size - 1
	}

	end := endingIndex()

	fmt.Fprint(w, "[\n")
	for i, m := range tv.members.Values() {
		if topN > 0 && i >= topN {
			break
		}
		if vp, ok := m.(*votingPower); ok {
			if b, err := vp.toJSON(); err == nil {
				fmt.Fprintf(w, "%s", string(b))
				if i != end {
					fmt.Fprint(w, ",\n")
				}
			} else {
				return err
			}
		} else {
			vprLogger.Error().Str("content", litter.Sdump(m)).Msg("invalid type of member")
		}
	}
	fmt.Fprint(w, "\n]\n")

	return nil
}

func (tv *topVoters) addVotingPower(id types.AccountID, delta *deltaVP) *votingPower {

	vp := tv.getVotingPower(id)

	if vp != nil {
		pwr := vp.getPower()
		pwr.Add(pwr, delta.getAmount())
	} else {
		vp = newVotingPower(delta.getAddr(), id, delta.getAmount())
	}
	return tv.update(vp)
}

// Voters Power Ranking (VPR)
type vpr struct {
	voters     *topVoters
	store      *vprStore
	totalPower *big.Int
	lowest     *votingPower

	changes map[types.AccountID]*deltaVP // temporary buffer for update
}

func newVpr() *vpr {
	return &vpr{
		voters:     newTopVoters(vprMax),
		store:      newVprStore(vprBucketsMax),
		totalPower: new(big.Int),
		changes:    make(map[types.AccountID]*deltaVP, vprMax),
	}
}

func loadVpr(s dataGetter) (*vpr, error) {
	v := newVpr()

	for i := uint8(0); i < vprBucketsMax; i++ {
		var (
			vps []*votingPower
			err error
		)
		if vps, err = v.store.read(s, i); err != nil {
			return nil, err
		}
		for _, vp := range vps {
			rv := v.voters.update(vp)
			v.store.addTail(i, rv)

			v.updateLowest(vp)
			v.addTotal(rv.getPower())
		}
	}

	return v, nil
}

func (v *vpr) equals(rhs *vpr) bool {
	if !reflect.DeepEqual(v.getTotalPower(), rhs.getTotalPower()) {
		return false
	}

	if !reflect.DeepEqual(v.getLowest(), rhs.getLowest()) {
		return false
	}

	if !v.voters.equals(rhs.voters) {
		return false
	}

	if len(v.changes) != len(v.changes) {
		return false
	}

	return true
}

func (v *vpr) getTotalPower() *big.Int {
	if v == nil {
		return nil
	}
	return new(big.Int).Set(v.totalPower)
}

func (v *vpr) getLowest() *votingPower {
	return v.lowest
}

func (v *vpr) updateLowest(vp *votingPower) {
	if vp.isZero() {
		v.resetLowest()
		return
	}

	if v.lowest == nil {
		v.lowest = vp
	} else if vp.lt(v.lowest) {
		v.lowest = vp
	}
}

func (v *vpr) setLowest(vp *votingPower) {
	v.lowest = vp
}

func (v *vpr) addTotal(delta *big.Int) {
	total := v.totalPower
	total.Add(total, delta)
}

func (v *vpr) votingPowerOf(addr types.AccountID) *big.Int {
	return v.voters.powerOf(addr)
}

func (v *vpr) prepare(id types.AccountID, addr types.Address, fn func(lhs *big.Int)) {
	if _, exist := v.changes[id]; !exist {
		v.changes[id] = newDeltaVP(addr, new(big.Int))
	}
	// Each entry of v.changes corresponds to the change (increment or
	// decrement) of voting power. It is added to later by calling the v.apply
	// method.
	ch := v.changes[id]

	fn(ch.getAmount())
}

func (v *vpr) add(id types.AccountID, addr []byte, power *big.Int) {
	if v == nil || power == nil || power.Cmp(zeroValue) == 0 {
		return
	}

	v.prepare(id, addr,
		func(lhs *big.Int) {
			if vprLogger.IsDebugEnabled() {
				vprLogger.Debug().
					Str("op", "add").
					Str("addr", enc.ToString(addr)).
					Str("orig", lhs.String()).
					Str("delta", power.String()).
					Msg("prepare voting power change")
			}
			lhs.Add(lhs, power)
		},
	)
}

func (v *vpr) sub(id types.AccountID, addr []byte, power *big.Int) {
	if v == nil || v.voters.powers[id] == nil {
		return
	}

	v.prepare(id, addr,
		func(lhs *big.Int) {
			if vprLogger.IsDebugEnabled() {
				vprLogger.Debug().
					Str("op", "sub").
					Str("addr", enc.ToString(addr)).
					Str("orig", lhs.String()).
					Str("delta", power.String()).
					Msg("prepare voting power change")
			}
			lhs.Sub(lhs, power)
		},
	)
}

func (v *vpr) apply(s *state.ContractState) (int, error) {
	if v == nil || len(v.changes) == 0 {
		return 0, nil
	}

	var (
		nApplied = 0
		updRows  = make(map[uint8]interface{})
	)

	for id, delta := range v.changes {
		if delta.cmp(zeroValue) != 0 {
			vp := v.voters.addVotingPower(id, delta)
			if s != nil {
				i := v.store.update(vp)
				if _, exist := updRows[i]; !exist {
					updRows[i] = struct{}{}
				}
			}

			v.updateLowest(vp)
			v.addTotal(delta.getAmount())

			delete(v.changes, id)
			// TODO: Remove a victim chosen above from the VPR bucket.
			nApplied++
		}
	}

	for i := range updRows {
		if err := v.store.write(s, i); err != nil {
			return 0, err
		}
	}

	return nApplied, nil
}

func (v *vpr) resetLowest() {
	v.setLowest(v.voters.lowest())
}

func PickVotingRewardWinner(seed int64) (types.Address, error) {
	return votingPowerRank.pickVotingRewardWinner(seed)
}

func (v *vpr) pickVotingRewardWinner(seed int64) (types.Address, error) {
	if v == nil {
		return nil, ErrNoVotingRewardRank
	}

	if v.getTotalPower().Cmp(zeroValue) == 0 {
		return nil, ErrNoVotingRewardWinner
	}

	totalVp := v.getTotalPower()
	r := new(big.Int).Rand(
		rand.New(rand.NewSource(seed)),
		totalVp)
	for i := uint8(0); i < vprBucketsMax; i++ {
		if l := v.store.buckets[i]; l != nil && l.Len() > 0 {
			for e := l.Front(); e != nil; e = e.Next() {
				vp := toVotingPower(e)
				if r.Sub(r, vp.getPower()).Cmp(zeroValue) < 0 {
					winner := vp.getAddr()

					if vprLogger.IsDebugEnabled() {
						vprLogger.Debug().
							Str("total voting power", totalVp.String()).
							Str("addr", enc.ToString(winner)).
							Msg("pick voting reward winner")
					}

					return winner, nil
				}
			}
		}
	}

	return nil, ErrNoVotingRewardWinner
}

func vprKey(i uint8) []byte {
	var vk []byte = vprKeyPrefix
	return append(vk, []byte(fmt.Sprintf("%v", i))...)
}

func toVotingPower(e *list.Element) *votingPower {
	return e.Value.(*votingPower)
}

type deltaVP struct {
	addr_  types.Address
	amount *big.Int
}

func newDeltaVP(addr []byte, amount *big.Int) *deltaVP {
	return &deltaVP{addr_: addr, amount: amount}
}

func (dv *deltaVP) getAddr() types.Address {
	return dv.addr_
}

func (dv *deltaVP) getAmount() *big.Int {
	return dv.amount
}

func (dv *deltaVP) cmp(rhs *big.Int) int {
	return dv.getAmount().Cmp(rhs)
}

func GetVotingRewardAmount() *big.Int {
	return defaultReward
}

func GetTotalVotingPower() *big.Int {
	return votingPowerRank.getTotalPower()
}

func DumpVotingPowerRankers(w io.Writer, topN int) error {
	if votingPowerRank == nil {
		return errors.New("not supported")
	}

	return votingPowerRank.voters.dump(w, topN)
}
