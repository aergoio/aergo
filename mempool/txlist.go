/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package mempool

import (
	"sort"
	"sync"

	"github.com/aergoio/aergo/types"
)

// TxList is internal struct for transactions per account
type TxList struct {
	sync.RWMutex
	min     uint64
	account []byte
	ready   int
	list    []*types.Tx // nonce-ordered tx list
}

// NewTxList creates new TxList with given nonce as min
func NewTxList(acc []byte, nonce uint64) *TxList {
	return &TxList{
		min:     nonce,
		account: acc,
	}
}

func (tl *TxList) GetAccount() []byte {
	return tl.account
}

// Len returns number of transactios which are ready to be processed
func (tl *TxList) Len() int {
	tl.RLock()
	defer tl.RUnlock()
	return tl.ready
}

// Empty check TxList is empty including orphan
func (tl *TxList) Empty() bool {
	tl.RLock()
	defer tl.RUnlock()
	return len(tl.list) == 0
}

func (tl *TxList) search(tx *types.Tx) (int, bool) {
	key := tx.GetBody().GetNonce()
	ind := sort.Search(len(tl.list), func(i int) bool {
		comp := tl.list[i].GetBody().GetNonce()
		return comp >= key
	})
	if ind < len(tl.list) && tl.compare(tx, ind) {
		return ind, true
	}
	return ind, false
}
func (tl *TxList) compare(tx *types.Tx, index int) bool {
	if tx.GetBody().GetNonce() == tl.list[index].GetBody().GetNonce() {
		return true
	}
	return false
}

func (tl *TxList) continuous(index int) bool {
	l := tl.min
	r := tl.list[index].GetBody().GetNonce()
	if tl.ready > 0 {
		l = tl.list[tl.ready-1].GetBody().GetNonce()
	}

	if l+1 == r {
		return true
	}
	return false
}

// Put inserts transaction into TxList
// if transaction is processible, it is appended to list
// if not, transaction is managed as orphan
func (tl *TxList) Put(tx *types.Tx) (int, error) {
	tl.Lock()
	defer tl.Unlock()

	nonce := tx.GetBody().GetNonce()
	if nonce <= tl.min {
		return 0, types.ErrTxNonceTooLow
	}

	index, found := tl.search(tx)
	if found == true { // exact match
		return 0, types.ErrSameNonceAlreadyInMempool
	}

	oldCnt := len(tl.list) - tl.ready

	if index < len(tl.list) {
		tl.list = append(tl.list[:index], append([]*types.Tx{tx},
			tl.list[index:]...)...)
	} else {
		tl.list = append(tl.list, tx)
	}

	for ; index < len(tl.list); index++ {
		if !tl.continuous(index) {
			break
		}
		tl.ready++
	}
	newCnt := len(tl.list) - tl.ready

	return oldCnt - newCnt, nil
}

// SetMinNonce sets new minimum nonce for TxList
// evict on some transactions is possible due to minimum nonce
func (tl *TxList) FilterByState(st *types.State) (int, []*types.Tx) {
	tl.Lock()
	defer tl.Unlock()

	oldCnt := len(tl.list) - tl.ready

	tl.min = st.Nonce
	var left []*types.Tx
	removed := tl.list[:0]
	for _, x := range tl.list {
		if err := x.ValidateWithSenderState(st); err == nil || err == types.ErrTxNonceToohigh {
			left = append(left, x)
		} else {
			removed = append(removed, x)
		}
	}
	tl.list = left
	tl.ready = 0
	for i := 0; i < len(tl.list); i++ {
		if !tl.continuous(i) {
			break
		}
		tl.ready++
	}
	newCnt := len(tl.list) - tl.ready

	return oldCnt - newCnt, removed
}

// FilterByPrice will evict transactions that needs more amount than balance
/*
func (tl *TxList) FilterByPrice(balance uint64) error {
	tl.Lock()
	defer tl.Unlock()
	return nil
}
*/

// Get returns processible transactions
func (tl *TxList) Get() []*types.Tx {
	tl.RLock()
	defer tl.RUnlock()
	return tl.list[:tl.ready]
}

// GetAll returns all transactions including orphans
func (tl *TxList) GetAll() []*types.Tx {
	tl.Lock()
	defer tl.Unlock()
	return tl.list

}

func (tl *TxList) len() int {
	return len(tl.list)
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
func (tl *TxList) printList() {
	fmt.Printf("\t\t")
	for i := 0; i < len(tl.list); i++ {
		cur := tl.list[i].GetBody().GetNonce()
		fmt.Printf("%d, ", cur)
	}
	fmt.Printf("done ready:%d min:%d\n", tl.ready, tl.min)

}

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
