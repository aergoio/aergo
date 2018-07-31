/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package mempool

import (
	"container/heap"
	"sort"
	"sync"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
)

type nonceHeap []uint64

func (h nonceHeap) Len() int           { return len(h) }
func (h nonceHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h nonceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *nonceHeap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}

func (h *nonceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// MemPoolList is main structure for managing transaction list
type MemPoolList struct {
	sync.RWMutex
	items map[uint64]*types.Tx
	heap  *nonceHeap
	//maxNonce   uint64
	minNonce   uint64
	checkNonce bool
}

// NewMemPoolList creates new object for MemPoolList
func NewMemPoolList(nonce uint64, check bool) *MemPoolList {
	return &MemPoolList{
		items: map[uint64]*types.Tx{},
		heap:  new(nonceHeap),
		//maxNonce:   nonce,
		minNonce:   nonce,
		checkNonce: check,
	}
}

// Len returns number of maintaining transactions
func (mpl *MemPoolList) Len() int {
	return mpl.heap.Len()
}

func (mpl *MemPoolList) updateMinNonce() {
	if mpl.Len() > 0 {
		mpl.minNonce = (*mpl.heap)[0]
	}
}
func (mpl *MemPoolList) putLocked(tx *types.Tx) {
	nonce := tx.GetBody().GetNonce()
	heap.Push(mpl.heap, nonce)
	mpl.items[nonce] = tx
	mpl.updateMinNonce()
}

// Put inserts given transaction if not already exists
func (mpl *MemPoolList) Put(tx *types.Tx) error {
	defer mpl.checkSanity()
	mpl.Lock()
	defer mpl.Unlock()
	nonce := tx.GetBody().GetNonce()
	if _, ok := mpl.items[nonce]; ok {
		return message.ErrTxAlreadyInMempool
	}

	if mpl.checkNonce && mpl.minNonce+uint64(mpl.Len()) != nonce {
		return message.ErrTxNonceToohigh
	}
	mpl.putLocked(tx)
	return nil
}

// Merge moves transactions of given MemPoolList to itself
func (mpl *MemPoolList) Merge(l *MemPoolList) (int, error) {
	defer mpl.checkSanity()
	defer l.checkSanity()
	mpl.Lock()
	defer mpl.Unlock()
	added := 0
	for l.Len() > 0 {
		min := l.GetMinNonce()
		if mpl.checkNonce && mpl.minNonce+uint64(mpl.Len()) != min {
			break
		}
		// TODO lock should be held for entire loop? could be delt with min nonce value
		tx := l.Del(min)
		mpl.putLocked(tx)
		added++
	}
	return added, nil
}

// FilterByPrice removes transactions which has amount larger than given balance
func (mpl *MemPoolList) FilterByPrice(bal uint64) int {
	defer mpl.checkSanity()
	mpl.Lock()
	defer mpl.Unlock()
	// TODO need refactoring
	delList := make([]uint64, 0)
	for i := 0; i < mpl.Len(); i++ {
		n := (*mpl.heap)[i]
		tx := mpl.items[n]
		if tx.GetBody().GetAmount() <= bal {
			continue
		}
		(*mpl.heap)[i] = 0
		delList = append(delList, n)
	}
	if len(delList) == 0 {
		return 0
	}
	heap.Init(mpl.heap)
	for _, v := range delList {
		delete(mpl.items, v)
		heap.Pop(mpl.heap)
	}
	mpl.updateMinNonce()
	return len(delList)
}

// SetMinNonce set minimum nonce of all maintaining transactions
func (mpl *MemPoolList) SetMinNonce(nonce uint64) int {
	defer mpl.checkSanity()
	mpl.Lock()
	defer mpl.Unlock()
	evict := 0
	for mpl.Len() > 0 {
		min := mpl.minNonce
		if min > nonce {
			break
		}
		mpl.delLocked(min)
		evict++
	}
	return evict
}

// Get retrieves single transaction which has same nonce with given parameter
func (mpl *MemPoolList) Get(nonce uint64) *types.Tx {
	mpl.RLock()
	defer mpl.RUnlock()
	return mpl.items[nonce]
}
func (mpl *MemPoolList) delLocked(nonce uint64) *types.Tx {
	rv, ok := mpl.items[nonce]
	if !ok {
		return nil
	}
	tmp := heap.Pop(mpl.heap)
	//TODO delete below code, it is for test
	{
		if tmp != nonce {
			/*
				for k, v := range mpl.items {
					logger.Info("#############", k, "::::", v.String())

				}
				logger.Info("#############you are not deleting min nonce in the list ", tmp, "::", nonce)
			*/
			panic("")
		}
	}
	delete(mpl.items, nonce)
	mpl.updateMinNonce()
	return rv
}

// Del removes single transaction which has exactly same nonce with given parameter
func (mpl *MemPoolList) Del(nonce uint64) *types.Tx {
	defer mpl.checkSanity()
	mpl.Lock()
	defer mpl.Unlock()
	return mpl.delLocked(nonce)
}

// GetAll retreives all transactions
func (mpl *MemPoolList) GetAll() []*types.Tx {
	mpl.RLock()
	defer mpl.RUnlock()
	if mpl.Len() == 0 {
		return nil
	}
	//TODO remove heap use list
	sort.Sort(mpl.heap)
	var val []*types.Tx
	//for _, v := range mpl.items {
	for i := 0; i < mpl.Len(); i++ {
		n := (*mpl.heap)[i]
		val = append(val, mpl.items[n])
	}
	heap.Init(mpl.heap)
	return val
}

// GetMinNonce returns minimum nonce of maintaining transactions
func (mpl *MemPoolList) GetMinNonce() uint64 {
	mpl.RLock()
	defer mpl.RUnlock()
	return mpl.minNonce
}

func (mpl *MemPoolList) checkSanity() {
	/*
		if len(mpl.items) != mpl.Len() {
			panic("mmmpooll panic : len(heap) != len(items)")
		}
		if mpl.Len() == 0 {
			return
		}
		var acc string

		for i := 0; mpl.Len() > i; i++ {
			v := (*mpl.heap)[i]
			tx, ok := mpl.items[v]
			if !ok {
				panic("mempooll panic : heap, items mismatch")
			}
			if tx.GetBody().GetNonce() != v {
				panic("mempooll panic : heap, items mismatch")
			}
			if v < mpl.minNonce {
				panic("mempooll panic : min wrong")
			}
			if mpl.checkNonce && v > mpl.minNonce+uint64(mpl.Len()) {
				panic("mempooll panic : max wrong")
			}
			if acc == "" {
				acc = getAccount(tx)
			} else if acc != getAccount(tx) {
				panic("mempooll panic : 2 different account in list")
			}
		}*/
}
