/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package mempool

import (
	"sync"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
)

// TxList is internal struct for transactions per account
type TxList struct {
	sync.RWMutex
	min     uint64
	account []byte
	list    []*types.Tx
	deps    map[uint64][]*types.Tx // <empty key, aggregated slice>
	parent  map[uint64]uint64      //<lastkey, empty key>
}

// NewTxList creates new TxList with given nonce as min
func NewTxList(acc []byte, nonce uint64) *TxList {
	return &TxList{
		min:     nonce,
		account: acc,
		deps:    map[uint64][]*types.Tx{},
		parent:  map[uint64]uint64{},
	}
}

func (tl *TxList) GetAccount() []byte {
	return tl.account
}

// Len returns number of transactios which are ready to be processed
func (tl *TxList) Len() int {
	tl.RLock()
	defer tl.RUnlock()
	return tl.len()
}

// Empty check TxList is empty including orphan
func (tl *TxList) Empty() bool {
	tl.RLock()
	defer tl.RUnlock()
	return tl.len() == 0 && len(tl.deps) == 0
}

// Put inserts transaction into TxList
// if transaction is processible, it is appended to list
// if not, transaction is managed as orphan
func (tl *TxList) Put(tx *types.Tx) (int, error) {
	tl.Lock()
	defer tl.Unlock()
	nonce := tx.GetBody().GetNonce()
	if nonce < tl.min {
		return 0, message.ErrTxNonceTooLow
	}
	if nonce < uint64(tl.len())+tl.min {
		return 0, message.ErrTxAlreadyInMempool
	}
	if uint64(tl.len())+tl.min != nonce {
		tl.putOrphan(tx)
		return -1, nil
	}

	tl.list = append(tl.list, tx)
	tmp := tl.getOrphans(nonce)
	tl.list = append(tl.list, tmp...)

	return len(tmp), nil
}

// SetMinNonce sets new minimum nonce for TxList
// evict on some transactions is possible due to minimum nonce
func (tl *TxList) SetMinNonce(n uint64) (int, []*types.Tx) {
	tl.Lock()
	defer tl.Unlock()
	defer func() { tl.min = n }()
	if tl.min == n {
		return 0, nil
	}
	if tl.min > n {
		neworphan := len(tl.list)
		l := tl.list[neworphan-1].GetBody().GetNonce()
		tl.deps[n] = tl.list
		tl.parent[l] = n
		tl.list = nil
		return -neworphan, nil
	}
	delOrphan := 0
	var delTxs []*types.Tx
	processed := n - tl.min
	if processed < uint64(tl.len()) {
		delTxs = tl.list[0:processed]
		tl.list = tl.list[processed:]
		return delOrphan, delTxs
	}

	delTxs = append(delTxs, tl.list...)
	tl.list = nil

	for k, v := range tl.deps {
		l := v[len(v)-1].GetBody().GetNonce()
		if l < n {
			delete(tl.deps, k)
			delete(tl.parent, l)
			delOrphan += len(v)
			delTxs = append(delTxs, v...)
		}
		if k < n && n <= l {
			delete(tl.deps, k)
			delete(tl.parent, l)
			tl.list = v[n-k-1:]
			delTxs = append(delTxs, v[0:n-k-1]...)
			delOrphan += len(v)
		}
	}
	return delOrphan, delTxs
}

// FilterByPrice will evict transactions that needs more amount than balance
func (tl *TxList) FilterByPrice(balance uint64) error {
	tl.Lock()
	defer tl.Unlock()
	return nil
}

// Get returns processible transactions
func (tl *TxList) Get() []*types.Tx {
	tl.RLock()
	defer tl.RUnlock()
	return tl.list
}

// GetAll returns all transactions including orphans
func (tl *TxList) GetAll() []*types.Tx {
	tl.Lock()
	defer tl.Unlock()
	var all []*types.Tx
	all = append(all, tl.list...)
	for _, v := range tl.deps {
		all = append(all, v...)
	}
	return all

}

func (tl *TxList) len() int {
	return len(tl.list)
}

func (tl *TxList) putOrphan(tx *types.Tx) {
	n := tx.GetBody().GetNonce()
	lastN := n
	var tmp []*types.Tx
	tmp = append(tmp, tx)
	v, ok := tl.deps[n]
	if ok {
		delete(tl.deps, n)
		tmp = append(tmp, v...)
		lastN = tmp[len(tmp)-1].GetBody().GetNonce()
	}

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

/*
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
*/
/*
func (tl *TxList) printList() {

	var f, l, before uint64
	if tl.list != nil {
		f = tl.list[0].GetBody().GetNonce()
		l = tl.list[len(tl.list)-1].GetBody().GetNonce()
	}
	fmt.Printf("min: %d ready(nr:%d)[%d~%d]", tl.min, len(tl.list), f, l)

	for i := 0; i < len(tl.list); i++ {
		cur := tl.list[i].GetBody().GetNonce()
		if i != 0 && before+1 != cur {
			fmt.Printf("WARN: List is not sequential")
		}
		before = cur
	}

	fmt.Println()
	fmt.Printf("deps 1st(%d):", len(tl.deps))
	for k, v := range tl.deps {
		f := v[0].GetBody().GetNonce()
		l := v[len(v)-1].GetBody().GetNonce()
		fmt.Printf("%d=>[%d~%d],", k, f, l)
	}
	fmt.Println()

	fmt.Printf("dep parent(%d):", len(tl.parent))
	for k, v := range tl.parent {
		fmt.Printf("(%d, %d)", k, v)
	}
	fmt.Println()
}
*/
