package system

import (
	"container/list"
	"math/big"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const vprMax = 50000

var rank = newVpr()

// Voters Power Ranking (VPRT)
type vpr struct {
	vp      map[types.AccountID]*big.Int
	ch      map[types.AccountID]*big.Int
	ranking *list.List
}

func newVpr() *vpr {
	return &vpr{
		vp:      make(map[types.AccountID]*big.Int, vprMax),
		ch:      make(map[types.AccountID]*big.Int, vprMax),
		ranking: list.New(),
	}
}

func (v *vpr) update(addr types.AccountID, fn func(lhs *big.Int)) {
	if _, exist := v.vp[addr]; !exist {
		v.vp[addr] = new(big.Int)
	}

	if _, exist := v.ch[addr]; !exist {
		v.ch[addr] = new(big.Int).Set(v.vp[addr])
	}
	ch := v.ch[addr]

	fn(ch)
}

func (v *vpr) Set(addr types.AccountID, power *big.Int) {
	v.update(addr,
		func(lhs *big.Int) {
			lhs.Set(power)
		},
	)
}

func (v *vpr) Add(addr types.AccountID, power *big.Int) {
	v.update(addr,
		func(lhs *big.Int) {
			lhs.Add(lhs, power)
		},
	)
}

func (v *vpr) Sub(addr types.AccountID, power *big.Int) {
	v.update(addr,
		func(lhs *big.Int) {
			lhs.Sub(lhs, power)
		},
	)
}

func (v *vpr) Apply(s *state.ContractState) {
	for key, pow := range v.ch {
		if curPow := v.vp[key]; curPow.Cmp(pow) != 0 {
			v.vp[key] = pow
			delete(v.ch, key)
		}
	}
}
