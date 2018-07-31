/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package mempool

import (
	"math/rand"
	"testing"
)

func TestListBasic(t *testing.T) {
	initTest()
	defer deinitTest()
	mpl := NewMemPoolList(uint64(0), false)

	count := 1000
	nonce := make([]int, count)
	for i := 0; i < count; i++ {
		nonce[i] = i + 1
	}
	rand.Shuffle(count, func(i, j int) {
		nonce[i], nonce[j] = nonce[j], nonce[i]
	})
	for i := 0; i < count; i++ {
		mpl.Put(genTx(0, 0, uint64(nonce[i]), 0))
	}

	for i := 1; i < count+1; i++ {
		if mpl.Del(uint64(i)) == nil {
			t.Error("delete failed")
		}
	}
}
func TestListDel(t *testing.T) {
	initTest()
	defer deinitTest()
	mpl := NewMemPoolList(uint64(0), false)
	count := 100

	//put 1~10 excpet 4 6 8
	for i := 0; i < count; i++ {
		if i == 3 || i == 5 || i == 7 {
			continue
		}
		mpl.Put(genTx(0, 0, uint64(i+1), 0))
	}
	mpl.SetMinNonce(uint64(2))
	if mpl.Len() != count-5 {
		t.Error(mpl.Len())
	}
	mpl.SetMinNonce(uint64(4))
	if mpl.Len() != count-6 {
		t.Error(mpl.Len())
	}

	mpl.SetMinNonce(uint64(8))
	if mpl.Len() != count-8 {
		t.Error(mpl.Len())
	}

	mpl.SetMinNonce(uint64(15))
	if mpl.Len() != count-15 {
		t.Error(mpl.Len())
	}

}

func TestListPriceFilter(t *testing.T) {
	initTest()
	defer deinitTest()
	mpl := NewMemPoolList(uint64(0), false)

	mpl.Put(genTx(0, 0, uint64(1), 2))
	mpl.Put(genTx(0, 0, uint64(3), 8))
	mpl.Put(genTx(0, 0, uint64(5), 4))
	mpl.Put(genTx(0, 0, uint64(6), 5))

	if mpl.Len() != 4 {
		t.Error(mpl.Len())
	}

	mpl.FilterByPrice(4) // -> 2nd tx should be filtered out
	if mpl.Len() != 2 {
		t.Error(mpl.Len())
	}
	if mpl.Get(1) == nil || mpl.Get(5) == nil || mpl.Get(3) != nil {
		t.Error("")
	}
	mpl.FilterByPrice(4)
	if mpl.Len() != 2 {
		t.Error(mpl.Len())
	}
	mpl.FilterByPrice(8)
	if mpl.Len() != 2 {
		t.Error(mpl.Len())
	}
	mpl.FilterByPrice(3)
	if mpl.Len() != 1 {
		t.Error(mpl.Len())
	}

}
