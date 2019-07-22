package system

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

const (
	vprMax = 50000
	// vprBucketsMax must be smaller than 256. A bigger number is regarded as
	// 256.
	vprBucketsMax = 71
)

var (
	vprKeyPrefix = []byte("VotingPowerBucket/")
	zeroValue    = &big.Int{}

	rank = newVpr()
)

type votingPower struct {
	addr  types.AccountID
	power *big.Int
}

func newVotingPower(addr types.AccountID, pow *big.Int) *votingPower {
	return &votingPower{addr: addr, power: pow}
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

func (vp *votingPower) marshal() []byte {
	var buf bytes.Buffer

	buf.Write(vp.addr[:])
	buf.Write(vp.getPower().Bytes())

	hdr := make([]byte, 4)
	binary.LittleEndian.PutUint32(hdr, uint32(buf.Len()))

	return append(hdr, buf.Bytes()...)
}

func (vp *votingPower) unmarshal(b []byte) uint32 {
	var n uint32

	r := bytes.NewReader(b)
	binary.Read(r, binary.LittleEndian, &n)

	vp.addr = types.AccountID(types.ToHashID(b[4:36]))
	if int(4+n) < len(b) {
		vp.setPower(new(big.Int).SetBytes(b[36 : 4+n]))
	} else {
		vp.setPower(new(big.Int).SetBytes(b[36:]))
	}

	return 4 + n
}

type vprBucket struct {
	buckets map[uint8]*list.List
	max     uint16
	cmp     func(pow *big.Int) func(v *votingPower) int
}

func newVprBucket(max uint16) *vprBucket {
	return &vprBucket{
		max:     max,
		buckets: make(map[uint8]*list.List, vprBucketsMax),
		cmp: func(pow *big.Int) func(v *votingPower) int {
			return func(v *votingPower) int {
				return v.cmp(pow)
			}
		},
	}
}

func (b *vprBucket) update(addr types.AccountID, pow *big.Int) (idx uint8) {
	idx = getBucketIdx(addr)

	var (
		bu    *list.List
		exist bool
	)

	if bu, exist = b.buckets[idx]; !exist {
		bu = list.New()
		b.buckets[idx] = bu
	}

	v := remove(bu, addr)
	if v != nil {
		v.setPower(pow)
	} else {
		v = newVotingPower(addr, pow)
	}

	if m := findInsPos(bu, b.cmp(pow)); m != nil {
		bu.InsertBefore(v, m)
	} else {
		bu.PushBack(v)
	}

	return
}

func (b *vprBucket) add(i uint8, vp *votingPower) {
	var l *list.List
	if l = b.buckets[i]; l == nil {
		l = list.New()
		b.buckets[i] = l
	}
	l.PushBack(vp)
}

type dataSetter interface {
	SetData(key, value []byte) error
}

type dataGetter interface {
	GetData(key []byte) ([]byte, error)
}

func (b *vprBucket) write(s dataSetter, i uint8) error {
	var buf bytes.Buffer

	l := b.buckets[i]
	for e := l.Front(); e != nil; e = e.Next() {
		buf.Write(e.Value.(*votingPower).marshal())
	}

	return s.SetData(vprKey(i), buf.Bytes())
}

func (b *vprBucket) read(s dataGetter, i uint8) ([]*votingPower, error) {
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

func remove(bu *list.List, addr types.AccountID) *votingPower {
	for e := bu.Front(); e != nil; e = e.Next() {
		if v := e.Value.(*votingPower); addr == v.addr {
			return bu.Remove(e).(*votingPower)
		}
	}
	return nil
}

func findInsPos(bu *list.List, fn func(v *votingPower) int) *list.Element {
	for e := bu.Front(); e != nil; e = e.Next() {
		v := e.Value.(*votingPower)
		ind := fn(v)
		if ind < 0 || ind == 0 {
			return e
		}
	}

	return nil
}

func getBucketIdx(addr types.AccountID) uint8 {
	return uint8(addr[0]) % vprBucketsMax
}

func (b *vprBucket) getBucket(addr types.AccountID) *list.List {
	if b, exist := b.buckets[getBucketIdx(addr)]; exist {
		return b
	}
	return nil
}

// Voters Power Ranking (VPR)
type vpr struct {
	votingPower map[types.AccountID]*big.Int
	changes     map[types.AccountID]*big.Int // temporary buffer for update
	table       *vprBucket
	totalPower  *big.Int
}

func (v *vpr) getTotalPower() *big.Int {
	return new(big.Int).Set(v.totalPower)
}

func newVpr() *vpr {
	return &vpr{
		votingPower: make(map[types.AccountID]*big.Int, vprMax),
		changes:     make(map[types.AccountID]*big.Int, vprMax),
		table:       newVprBucket(vprBucketsMax),
		totalPower:  new(big.Int),
	}
}

func loadVpr(s dataGetter) (*vpr, error) {
	v := newVpr()

	for i := uint8(0); i < vprBucketsMax; i++ {
		var (
			vps []*votingPower
			err error
		)
		if vps, err = v.table.read(s, i); err != nil {
			return nil, err
		}
		for _, vp := range vps {
			v.table.add(i, vp)
			v.setVotingPower(vp)
		}
	}

	return v, nil
}

func (v *vpr) setVotingPower(vp *votingPower) {
	var (
		pwr   = vp.getPower()
		total = v.totalPower
	)
	v.votingPower[vp.addr] = pwr
	total.Add(total, pwr)
}

func (v *vpr) votingPowerOf(address types.AccountID) *big.Int {
	return v.votingPower[address]
}

func (v *vpr) update(addr types.AccountID, fn func(lhs *big.Int)) {
	if _, exist := v.votingPower[addr]; !exist {
		v.votingPower[addr] = &big.Int{}
	}

	if _, exist := v.changes[addr]; !exist {
		v.changes[addr] = &big.Int{}
	}
	// Each entry of v.changes corresponds to the change (increment or
	// decrement) of voting power. It is added to later by calling the v.apply
	// method.
	ch := v.changes[addr]

	fn(ch)
}

func (v *vpr) add(addr types.AccountID, power *big.Int) {
	v.update(addr,
		func(lhs *big.Int) {
			lhs.Add(lhs, power)
		},
	)
}

func (v *vpr) sub(addr types.AccountID, power *big.Int) {
	v.update(addr,
		func(lhs *big.Int) {
			lhs.Sub(lhs, power)
		},
	)
}

func (v *vpr) apply(s *state.ContractState) (int, error) {
	var (
		nApplied = 0
		updRows  = make(map[uint8]interface{})
	)

	for key, pow := range v.changes {
		if pow.Cmp(zeroValue) != 0 {
			lhs := v.votingPower[key]
			lhs.Add(lhs, pow)
			v.totalPower.Add(v.totalPower, pow)

			if s != nil {
				i := v.table.update(key, lhs)
				if _, exist := updRows[i]; !exist {
					updRows[i] = struct{}{}
				}
			}

			delete(v.changes, key)
			nApplied++
		}
	}

	for i, _ := range updRows {
		if err := v.table.write(s, i); err != nil {
			return 0, err
		}
	}

	return nApplied, nil
}

// TESTING PURPOSE ONLY! Do not call this function during run time please.
func (v *vpr) checkValidity(t *testing.T) {
	sum1 := &big.Int{}
	sum2 := &big.Int{}

	for _, pow := range v.votingPower {
		sum1.Add(sum1, pow)
	}
	assert.True(t, sum1.Cmp(v.getTotalPower()) == 0, "voting power map inconsistent with total voting power")

	for _, l := range v.table.buckets {
		for e := l.Front(); e != nil; e = e.Next() {
			sum2.Add(sum2, toVotingPower(e).getPower())
		}
	}
	assert.True(t, sum2.Cmp(v.getTotalPower()) == 0, "voting power buckects inconsistent with total voting power")
}

func (v *vpr) Bingo(seed []byte) {
}

func vprKey(i uint8) []byte {
	var vk []byte = vprKeyPrefix
	return append(vk, []byte(fmt.Sprintf("%v", i))...)
}

func toVotingPower(e *list.Element) *votingPower {
	return e.Value.(*votingPower)
}
