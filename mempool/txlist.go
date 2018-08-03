/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package mempool

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"sort"
	"sync"
)

func (tl *TxList) search(key uint64) (int, *types.Tx) {
	found := sort.Search(tl.len(), func(i int) bool {
		return tl.list[i].GetBody().GetNonce() >= key
	})
	if found < tl.len() && tl.list[found].GetBody().GetNonce() == key {
		return found, tl.list[found]
	}
	return found, nil

}
func (tl *TxList) len() int {
	return len(tl.list)
}

func (tl *TxList) putOrphan(tx *types.Tx) {
	n := tx.GetBody().GetNonce()
	var tmp []*types.Tx
	tmp = append(tmp, tx)
	v, ok := tl.deps[n]
	if ok {
		delete(tl.deps, n)
		tmp = append(tmp, v...)
	}
	lastN := tmp[len(tmp)-1].GetBody().GetNonce()

	parent, dok := tl.parent[n-1]
	if !dok {
		tl.deps[n-1] = tmp
		tl.parent[lastN] = n - 1
		return
	}

	pList := tl.deps[parent]
	oldlast := pList[len(pList)-1].GetBody().GetNonce()

	tl.deps[parent] = append(pList, tmp...)
	delete(tl.parent, oldlast)
	tl.parent[lastN] = parent
}
func (tl *TxList) getOrphans(nonce uint64) []*types.Tx {

	v, ok := tl.deps[nonce]
	if !ok {
		return nil
	}
	lastN := v[len(v)-1].GetBody().GetNonce()

	delete(tl.deps, nonce)
	delete(tl.parent, lastN)
	return v
}

type TxList struct {
	sync.RWMutex
	min    uint64
	list   []*types.Tx
	deps   map[uint64][]*types.Tx // <empty key, aggregated slice>
	parent map[uint64]uint64      //<lastkey, empty key>
}

func NewTxList(nonce uint64) *TxList {
	return &TxList{
		min:    nonce,
		deps:   map[uint64][]*types.Tx{},
		parent: map[uint64]uint64{},
	}
}
func (tl *TxList) Len() int {
	tl.RLock()
	defer tl.RUnlock()
	return tl.len()
}

func (tl *TxList) Put(tx *types.Tx) (error, int) {
	tl.Lock()
	defer tl.Unlock()
	nonce := tx.GetBody().GetNonce()
	if nonce < tl.min {
		return message.ErrTxNonceTooLow, 0
	}
	if nonce < uint64(tl.len())+tl.min {
		return message.ErrTxAlreadyInMempool, 0
	}
	if uint64(tl.len())+tl.min != nonce {
		tl.putOrphan(tx)
		return nil, -1
	}

	tl.list = append(tl.list, tx)
	tmp := tl.getOrphans(nonce)
	tl.list = append(tl.list, tmp...)

	return nil, len(tmp)
}

func (tl *TxList) SetMinNonce(n uint64) int {
	tl.Lock()
	defer tl.Unlock()
	defer func() { tl.min = n }()

	delcnt := 0
	processed := n - tl.min
	if processed < uint64(tl.len()) {
		tl.list = tl.list[processed:]
		return delcnt
	}

	tl.list = nil
	for k, v := range tl.deps {
		l := v[len(v)-1].GetBody().GetNonce()
		if l < n {
			delete(tl.deps, k)
			delete(tl.parent, l)
			delcnt += len(v)
		}
		if k < n && n <= l {
			delete(tl.deps, k)
			delete(tl.parent, l)
			tl.list = v[n-k:]
			delcnt += len(v)
			break
		}
	}
	return delcnt
}

func (tl *TxList) FilterByPrice(balance uint64) error {
	tl.Lock()
	defer tl.Unlock()
	return nil
}
func (tl *TxList) Get() []*types.Tx {
	tl.RLock()
	defer tl.RUnlock()
	return tl.list
}

func (tl *TxList) checkSanity() bool {
	prev := uint64(0)
	for _, v := range tl.list {
		x := v.GetBody().GetNonce()
		if prev >= x {
			return false
		}
		prev = x
	}
	return true
}

/*
func (tl *TxList) printList() {

	var l uint64
	if tl.list != nil {
		l = tl.list[len(tl.list)-1].GetBody().GetNonce()
	}
	fmt.Printf("ready(%d)(min:%d)(last:%d)", len(tl.list), tl.min, l)

	for i := 0; i < len(tl.list); i++ {
		fmt.Printf("%d,", tl.list[i].GetBody().GetNonce())
	}

	fmt.Println()
	fmt.Printf("deps 1st(%d):", len(tl.deps))
	for k, v := range tl.deps {
		l := v[len(v)-1].GetBody().GetNonce()
		fmt.Printf("(%d, %d)", k, l)
	}
	fmt.Println()

	fmt.Printf("dep parent(%d):", len(tl.parent))
	for k, v := range tl.parent {
		fmt.Printf("(%d, %d)", k, v)
	}
	fmt.Println()

}
*/
