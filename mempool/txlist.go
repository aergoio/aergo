/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package mempool

import (
	"bytes"
	"sort"
	"sync"
	"time"

	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/types"
)

// txList is internal struct for transactions per account
type txList struct {
	sync.RWMutex
	base     *types.State
	lastTime time.Time
	account  []byte
	ready    int
	list     []types.Transaction // nonce-ordered tx list
	mp       *MemPool
}

// newTxList creates new txList with given State
func newTxList(acc []byte, st *types.State, mp *MemPool) *txList {
	return &txList{
		base:    st,
		account: acc,
		mp:      mp,
	}
}

func (tl *txList) GetLastModifiedTime() time.Time {
	return tl.lastTime
}

func (tl *txList) GetAccount() []byte {
	return tl.account
}

// Len returns number of transactions which are ready to be processed
func (tl *txList) Len() int {
	tl.RLock()
	defer tl.RUnlock()
	return tl.ready
}

// Empty check txList is empty including orphan
func (tl *txList) Empty() bool {
	tl.RLock()
	defer tl.RUnlock()
	return len(tl.list) == 0
}

func (tl *txList) search(tx types.Transaction) (int, bool) {
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
func (tl *txList) compare(tx types.Transaction, index int) bool {
	if tx.GetBody().GetNonce() == tl.list[index].GetBody().GetNonce() {
		return true
	}
	return false
}

func (tl *txList) continuous(index int) bool {
	l := tl.base.Nonce
	r := tl.list[index].GetBody().GetNonce()
	if tl.ready > 0 {
		l = tl.list[tl.ready-1].GetBody().GetNonce()
	}

	if l+1 == r {
		return true
	}
	return false
}

// Put inserts transaction into txList
// if transaction is processable, it is appended to list
// if not, transaction is managed as orphan
func (tl *txList) Put(tx types.Transaction) (int, error) {
	tl.Lock()
	defer tl.Unlock()

	nonce := tx.GetBody().GetNonce()
	if nonce <= tl.base.Nonce {
		return 0, types.ErrTxNonceTooLow
	}

	index, found := tl.search(tx)
	if found == true { // exact match
		return 0, types.ErrSameNonceAlreadyInMempool
	}

	oldCnt := len(tl.list) - tl.ready

	if index < len(tl.list) {
		tl.list = append(tl.list[:index], append([]types.Transaction{tx},
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

	tl.lastTime = time.Now()
	return oldCnt - newCnt, nil
}

func (tl *txList) FilterByState(st *types.State) (int, []types.Transaction) {
	tl.Lock()
	defer tl.Unlock()

	var balCheck bool

	if tl.base.Nonce == st.Nonce {
		tl.base = st
		return 0, nil
	}

	if tl.base.GetBalanceBigInt().Cmp(st.GetBalanceBigInt()) > 0 {
		balCheck = true
	}
	tl.base = st

	oldCnt := len(tl.list) - tl.ready
	var left []types.Transaction
	removed := tl.list[:0]
	for i, x := range tl.list {
		err := x.ValidateWithSenderState(st, system.GetGasPrice(), tl.mp.nextBlockVersion())
		if err == nil || err == types.ErrTxNonceToohigh {
			if err != nil && !balCheck {
				left = append(left, tl.list[i:]...)
				break
			}
			left = append(left, x)
		} else {
			removed = append(removed, x)
		}
	}

	tl.list = left
	tl.updateReady()
	newCnt := len(tl.list) - tl.ready

	tl.lastTime = time.Now()
	return oldCnt - newCnt, removed
}

func (tl *txList) updateReady() {
	tl.ready = 0
	for i := 0; i < len(tl.list); i++ {
		if !tl.continuous(i) {
			break
		}
		tl.ready++
	}
}

// RemoveTx will remove a transaction in the list and return the number of changed orphan, removed transaction
func (tl *txList) RemoveTx(tx *types.Tx) (int, types.Transaction) {
	oldLen := tl.Len()
	for i, x := range tl.list {
		if bytes.Equal(tx.GetHash(), x.GetTx().GetHash()) {
			tl.list = append(tl.list[:i], tl.list[i+1:]...)
			tl.updateReady()
			tl.lastTime = time.Now()
			return oldLen - tl.Len() - 1, x
		}
	}
	return 0, nil
}

func (tl *txList) pooled() []types.Transaction {
	return tl.list[:tl.ready]
}

func (tl *txList) orphaned() []types.Transaction {
	return tl.list[tl.ready:]
}

// FilterByPrice will evict transactions that needs more amount than balance
/*
func (tl *txList) FilterByPrice(balance uint64) error {
	tl.Lock()
	defer tl.Unlock()
	return nil
}
*/

// Get returns processible transactions
func (tl *txList) Get() []types.Transaction {
	tl.RLock()
	defer tl.RUnlock()
	return tl.list[:tl.ready]
}

// GetAll returns all transactions including orphans
func (tl *txList) GetAll() []types.Transaction {
	tl.Lock()
	defer tl.Unlock()
	return tl.list

}

func (tl *txList) allLen() int {
	return len(tl.list)
}

/*

func (tl *txList) printList() {
	fmt.Printf("\t\t")
	for i := 0; i < len(tl.list); i++ {
		cur := tl.list[i].GetBody().GetNonce()
		fmt.Printf("%d, ", cur)
	}
	fmt.Printf("done ready:%d n:%d, b:%d\n", tl.ready, tl.base.Nonce, tl.base.Balance)

}

func (tl *txList) checkSanity() bool {
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
func (tl *txList) printList() {

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
