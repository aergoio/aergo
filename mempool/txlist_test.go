/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package mempool

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/fee"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

var dummyMempool = &MemPool{
	cfg: &config.Config{
		Hardfork: &config.HardforkConfig{
			V2: 0,
		},
	},
	bestBlockInfo: getCurrentBestBlockInfoMock(),
}

func NewState(nonce uint64, bal uint64) *types.State {
	return &types.State{Nonce: nonce, Balance: new(big.Int).SetUint64(bal).Bytes()}
}

func TestListPutBasic(t *testing.T) {
	initTest(t)
	defer deinitTest()
	mpl := newTxList(nil, NewState(0, 0), dummyMempool)

	count := 100
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

	ret := mpl.Get()
	if len(ret) != count {
		t.Error("put failed", len(ret), count)
	}
}

func TestListPutBasicOrphan(t *testing.T) {
	initTest(t)
	defer deinitTest()
	mpl := newTxList(nil, NewState(0, 0), dummyMempool)

	count := 20
	nonce := make([]int, count)
	for i := 0; i < count; i++ {
		nonce[i] = i + 1
	}
	rand.Shuffle(count, func(i, j int) {
		nonce[i], nonce[j] = nonce[j], nonce[i]
	})
	nonce = nonce[count/2:]

	for i := 0; i < len(nonce); i++ {
		mpl.Put(genTx(0, 0, uint64(nonce[i]), 0))
	}

	ret := mpl.GetAll()
	if len(ret) != len(nonce) {
		t.Error("put failed", len(ret), len(nonce))
	}
}

func TestListPutErrors(t *testing.T) {
	initTest(t)
	defer deinitTest()
	mpl := newTxList(nil, NewState(9, 0), dummyMempool)
	added, err := mpl.Put(genTx(0, 0, uint64(1), 0))
	if added != 0 || err != types.ErrTxNonceTooLow {
		t.Errorf("put should be failed with ErrTxNonceTooLow, but %s", err)
	}

	added, err = mpl.Put(genTx(0, 0, uint64(10), 0))
	if added != 0 || err != nil || len(mpl.list) != 1 {
		t.Errorf("put should be not failed, but (%d)%s", added, err)
	}

	added, err = mpl.Put(genTx(0, 0, uint64(10), 0))
	if added != 0 || err != types.ErrSameNonceAlreadyInMempool {
		t.Errorf("put should be failed with ErrSameNonceAlreadyInMempool, but %s", err)
	}

}
func TestListDel(t *testing.T) {
	initTest(t)
	defer deinitTest()
	mpl := newTxList(nil, NewState(0, 0), dummyMempool)

	fee.EnableZeroFee()
	ret, txs := mpl.FilterByState(NewState(2, 100))
	if ret != 0 || mpl.Len() != 0 || len(txs) != 0 {
		t.Error(ret, mpl.Len(), len(txs))
	}

	ret, txs = mpl.FilterByState(NewState(0, 100))
	if ret != 0 || mpl.Len() != 0 || len(txs) != 0 {
		t.Error(ret, mpl.Len(), len(txs))
	}

	count := 100

	//put 1~10 excpet 4 6 8
	for i := 0; i < count; i++ {
		if i == 3 || i == 5 || i == 7 {
			continue
		}
		mpl.Put(genTx(0, 0, uint64(i+1), 0))
	}
	// 1, |2, 3, | x, 5, x, 7, | x, 9... 14, |15... 100
	ret, txs = mpl.FilterByState(NewState(0, 100))
	if ret != 0 || mpl.Len() != 3 || len(txs) != 0 {
		t.Error(ret, mpl.Len(), len(txs))
	}

	ret, txs = mpl.FilterByState(NewState(1, 100))
	if ret != 0 || mpl.Len() != 2 || len(txs) != 1 {
		t.Error(ret, mpl.Len(), len(txs))
	}

	ret, txs = mpl.FilterByState(NewState(3, 100))
	if ret != 0 || mpl.Len() != 0 || len(txs) != 2 {
		t.Error(ret, mpl.Len(), len(txs))
	}

	ret, txs = mpl.FilterByState(NewState(7, 100))
	if ret != 2 || mpl.Len() != 0 || len(txs) != 2 {
		t.Error(ret, mpl.Len(), len(txs))
	}

	ret, txs = mpl.FilterByState(NewState(14, 100))
	if ret != 92 || mpl.Len() != count-14 || len(txs) != 6 {
		t.Error(ret, mpl.Len(), len(txs))
	}

	txs = mpl.Get()
	if txs[0].GetBody().GetNonce() != 15 {
		t.Error(txs[0].GetBody().GetNonce())
	}

}

func TestListDelMiddle(t *testing.T) {
	initTest(t)
	defer deinitTest()
	mpl := newTxList(nil, NewState(3, 0), dummyMempool)

	mpl.Put(genTx(0, 0, uint64(4), 0))
	mpl.Put(genTx(0, 0, uint64(5), 0))
	mpl.Put(genTx(0, 0, uint64(6), 0))

	if mpl.Len() != 3 {
		t.Error("should be 3 not ", len(mpl.list))
	}
	fee.EnableZeroFee()
	ret, txs := mpl.FilterByState(NewState(1, 100))
	if ret != -3 || mpl.Len() != 0 || len(txs) != 0 {
		t.Error(ret, mpl.Len(), len(txs))
	}
	ret, txs = mpl.FilterByState(NewState(4, 100))
	if ret != 3 || mpl.Len() != 2 || len(txs) != 1 {
		t.Error(ret, mpl.Len(), len(txs))
	}

}

func TestListPutRandom(t *testing.T) {
	initTest(t)
	defer deinitTest()
	mpl := newTxList(nil, NewState(0, 0), dummyMempool)

	count := 100
	txs := make([]types.Transaction, count)
	for i := 0; i < count; i++ {
		txs[i] = genTx(0, 0, uint64(i+1), 0)
	}
	rand.Shuffle(count, func(i, j int) {
		txs[i], txs[j] = txs[j], txs[i]
	})
	for i := 0; i < count; i++ {
		mpl.Put(txs[i])
		if !sameTxs(mpl.GetAll(), txs[:i+1]) {
			t.Error("GetAll returns unproperly")
		}
	}
	ret := mpl.Get()
	if len(ret) != count {
		t.Error("put failed", len(ret), count)
	}
}

func TestListRemoveTx(t *testing.T) {
	initTest(t)
	defer deinitTest()
	mpl := newTxList(nil, NewState(3, 0), dummyMempool)

	four := genTx(0, 0, uint64(4), 0)
	five := genTx(0, 0, uint64(5), 0)
	six := genTx(0, 0, uint64(6), 0)
	eight := genTx(0, 0, uint64(8), 0)
	nine := genTx(0, 0, uint64(9), 0)
	ten := genTx(0, 0, uint64(10), 0)
	mpl.Put(four)
	mpl.Put(five)
	mpl.Put(six)
	mpl.Put(eight)
	mpl.Put(nine)
	mpl.Put(ten)

	if mpl.Len() != 3 {
		t.Error("should be 3 not ", len(mpl.list))
	}

	changedOrphan, tx := mpl.RemoveTx(four.GetTx())
	assert.Equal(t, 5, mpl.allLen(), "list length")
	assert.Equal(t, 2, changedOrphan, "new orphan count")
	assert.Equal(t, four.GetTx().GetHash(), tx.GetHash(), "removed tx")
	mpl.Put(four)

	changedOrphan, tx = mpl.RemoveTx(five.GetTx())
	assert.Equal(t, 1, changedOrphan, "new orphan count")
	assert.Equal(t, five.GetTx().GetHash(), tx.GetHash(), "removed tx")
	mpl.Put(five)

	changedOrphan, tx = mpl.RemoveTx(six.GetTx())
	assert.Equal(t, 0, changedOrphan, "new orphan count")
	assert.Equal(t, six.GetTx().GetHash(), tx.GetHash(), "removed tx")
	mpl.Put(six)

	changedOrphan, tx = mpl.RemoveTx(nine.GetTx())
	assert.Equal(t, -1, changedOrphan, "new orphan count")
	assert.Equal(t, nine.GetTx().GetHash(), tx.GetHash(), "removed tx")
	mpl.Put(six)
}
