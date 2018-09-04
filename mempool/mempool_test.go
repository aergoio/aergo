/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package mempool

import (
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/types"
)

const (
	maxAccount   = 1000
	maxRecipient = 1000
)

var (
	pool      *MemPool
	account   [maxAccount][]byte
	recipient [maxRecipient][]byte
)

func _itobU32(argv uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, argv)
	return bs
}

func getAccount(tx *types.Tx) string {
	return hex.EncodeToString(tx.GetBody().GetAccount())
}

func simulateBlockGen(txs ...*types.Tx) error {
	lock.Lock()
	defer lock.Unlock()
	for _, tx := range txs {
		acc := getAccount(tx)
		n := tx.GetBody().GetNonce()
		nonce[acc] = n
		_, ok := balance[acc]
		if !ok {
			balance[acc] = defaultBalance
		}
		balance[acc] -= tx.GetBody().GetAmount()
	}
	bestBlockNo++
	return nil
}
func initTest() {
	serverCtx := config.NewServerContext("", "")
	cfg := serverCtx.GetDefaultConfig().(*config.Config)
	pool = NewMemPoolService(cfg)
	pool.testConfig = true
	pool.BeforeStart()

	for i := 0; i < maxAccount; i++ {
		account[i] = _itobU32(uint32(i))
	}
	for i := 0; i < maxRecipient; i++ {
		recipient[i] = _itobU32(uint32(i))
	}
}
func deinitTest() {

}

func sameTx(a *types.Tx, b *types.Tx) bool {
	return types.ToTxID(a.Hash) != types.ToTxID(b.Hash)
}
func sameTxs(a []*types.Tx, b []*types.Tx) bool {
	if len(a) != len(b) {
		return false
	}
	check := false
	for _, txa := range a {
		check = false
		for _, txb := range b {
			if sameTx(txa, txb) {
				check = true
				break
			}
		}
		if !check {
			break
		}
	}
	return check
}
func genTx(acc int, rec int, nonce uint64, amount uint64) *types.Tx {
	tx := types.Tx{
		Body: &types.TxBody{
			Nonce:     nonce,
			Account:   account[acc],
			Recipient: recipient[rec],
			Amount:    amount,
		},
	}
	tx.Hash = tx.CalculateTxHash()
	return &tx
}

// func TestInvalidTransaction(t *testing.T) {

// 	initTest()
// 	defer deinitTest()
// 	err := pool.put(genTx(0, 1, 1, defaultBalance*2))
// 	if err != ErrInsufficientBalance {
// 		t.Errorf("check valid failed, err != ErrInsufficientBalance, but %s", err)
// 	}

// 	err = pool.put(genTx(0, 1, 1, 1))
// 	if err != nil {
// 		t.Errorf("tx should be accepted, err:%s", err)
// 	}
// 	err = pool.put(genTx(0, 1, 1, 1))
// 	if err != ErrTxAlreadyInMempool {
// 		t.Errorf("tx should be denied /w ErrTxAlreadyInMempool, err:%s", err)
// 	}
// 	txs := []*types.Tx{genTx(0, 1, 1, 1)}
// 	simulateBlockGen(txs...)

// 	pool.removeOnBlockArrival(getCurrentBestBlockNoMock(), txs...)
// 	err = pool.put(genTx(0, 1, 1, 1))
// 	if err != ErrTxNonceTooLow {
// 		t.Errorf("tx should be denied /w ErrTxNonceTooLow, err:%s", err)
// 	}
// }

func TestOrphanTransaction(t *testing.T) {
	//	t.Errorf("Sum was incorrect, ")

	initTest()
	defer deinitTest()

	err := pool.put(genTx(0, 1, 1, 2))
	if err != nil {
		t.Error("put tx should be succeeded", err)
	}
	// tx inject order : 1 3 5 2 4 10 9 8 7 6
	// non-sequential nonce should be accepted (orphan) but not counted
	if err = pool.put(genTx(0, 1, 3, 2)); err != nil {
		t.Error("put tx should be succeeded", err)
	}
	if err = pool.put(genTx(0, 1, 5, 2)); err != nil {
		t.Error("put tx should be succeeded", err)
	}

	if p1, p2 := pool.Size(); !(p1 == 3 && p2 == 2) {
		t.Errorf("invalid count status pool:%d orphan:%d", p1, p2)
	}

	if err = pool.put(genTx(0, 1, 2, 2)); err != nil {
		t.Error("put tx should be succeeded", err)
	}
	if p1, p2 := pool.Size(); !(p1 == 4 && p2 == 1) {
		t.Errorf("invalid count status pool:%d orphan:%d", p1, p2)
	}
	if err = pool.put(genTx(0, 1, 4, 2)); err != nil {
		t.Error("put tx should be succeeded", err)
	}
	if p1, p2 := pool.Size(); !(p1 == 5 && p2 == 0) {
		t.Errorf("invalid count status pool:%d orphan:%d", p1, p2)
	}

	if err = pool.put(genTx(0, 1, 10, 2)); err != nil {
		t.Error("put tx should be succeeded", err)
	}
	if err = pool.put(genTx(0, 1, 9, 2)); err != nil {
		t.Error("put tx should be succeeded", err)
	}
	if err = pool.put(genTx(0, 1, 8, 2)); err != nil {
		t.Error("put tx should be succeeded", err)
	}
	if err = pool.put(genTx(0, 1, 7, 2)); err != nil {
		t.Error("put tx should be succeeded", err)
	}
	if p1, p2 := pool.Size(); !(p1 == 9 && p2 == 4) {
		t.Errorf("invalid count status pool:%d orphan:%d", p1, p2)
	}

	//pool.pool[getAccount(genTx(0, 1, 1, 1))].CheckSanity()
	//pool.pending[getAccount(genTx(0, 1, 1, 1))].CheckSanity()
	if err = pool.put(genTx(0, 1, 6, 2)); err != nil {
		t.Error("put tx should be succeeded", err)
	}
	if p1, p2 := pool.Size(); !(p1 == 10 && p2 == 0) {
		t.Errorf("invalid count status pool:%d orphan:%d", p1, p2)
	}

}
func TestBasics2(t *testing.T) {
	initTest()
	defer deinitTest()
	txs := make([]*types.Tx, 0)

	accCount := 1
	txCount := 100
	nonce := make([]uint64, txCount)
	for i := 0; i < txCount; i++ {
		nonce[i] = uint64(i + 1)
		//nonce[i] = uint64(txCount -i+1)
	}
	for i := 0; i < accCount; i++ {
		rand.Shuffle(txCount, func(i, j int) {
			nonce[i], nonce[j] = nonce[j], nonce[i]
		})
		for j := 0; j < txCount; j++ {
			tmp := genTx(i, 0, nonce[j], uint64(i+1))
			txs = append(txs, tmp)
		}
	}

	for _, tx := range txs {
		errs := pool.puts(tx)

		if errs[0] != nil {
			t.Errorf("th - tx should be added(%s),", errs)
		}
	}

	txsMempool, err := pool.get()
	if err != nil {
		t.Errorf("Getting tx should be succeeded, %s", err)
	}
	t.Log(len(txsMempool))
	if !sameTxs(txs, txsMempool) {
		t.Error("should be same")
	}
}

// gen sequential transactions
// check mempool internal states
func TestBasics(t *testing.T) {
	initTest()
	defer deinitTest()
	txs := make([]*types.Tx, 0)

	accCount := 500
	txCount := 1000
	nonce := make([]uint64, txCount)
	for i := 0; i < txCount; i++ {
		nonce[i] = uint64(i + 1)
		//nonce[i] = uint64(txCount -i+1)
	}
	for i := 0; i < accCount; i++ {
		rand.Shuffle(txCount, func(i, j int) {
			nonce[i], nonce[j] = nonce[j], nonce[i]
		})
		for j := 0; j < txCount; j++ {
			tmp := genTx(i, 0, nonce[j], uint64(i+1))
			txs = append(txs, tmp)
		}
	}
	errs := pool.puts(txs...)

	if len(errs) != accCount*txCount {
		t.Error("error count invalid", len(errs))
	}
	for i := 0; i < len(errs); i++ {
		if errs[i] != nil {
			t.Errorf("%dth - tx should be added(%s),", i, errs[i])
		}
	}
	txsMempool, err := pool.get()
	if err != nil {
		t.Errorf("Getting tx should be succeeded, %s", err)
	}
	t.Log(len(txsMempool))
	if !sameTxs(txs, txsMempool) {
		t.Error("should be same")
	}
}

func TestDeleteOTxs(t *testing.T) {
	initTest()
	defer deinitTest()
	txs := make([]*types.Tx, 0)
	for i := 0; i < 5; i++ {
		tmp := genTx(0, 0, uint64(i+1), uint64(i+1))
		txs = append(txs, tmp)
	}
	pool.puts(txs...)
	if ps, _ := pool.Size(); ps != 5 {
		t.Errorf("pool should contain 5 , %d", ps)
	}

	txs[4] = genTx(0, 1, 5, 150)
	simulateBlockGen(txs...)
	pool.removeOnBlockArrival(getCurrentBestBlockNoMock(), txs...)
	if r, o := pool.Size(); r != 0 || o != 0 {
		t.Error("pool should contain nothing", r, o)
	}
}

// add 100 sequential txs and simulate to generate block 10time.
// each block contains 10 txs
func TestBasicDeleteOnBlockConnect(t *testing.T) {
	initTest()
	defer deinitTest()
	txs := make([]*types.Tx, 0)

	for i := 0; i < 100; i++ {
		tmp := genTx(0, 0, uint64(i+1), uint64(i+1))
		txs = append(txs, tmp)
	}
	pool.puts(txs...)
	if ps, _ := pool.Size(); ps != 100 {
		t.Errorf("pool should contain 100 , %d", ps)
	}
	//suppose 10 txs are select into new block

	for j := 0; j < 10; j++ {
		simulateBlockGen(txs[:10]...)
		//	pool.removes(uint32(pool.curBestBlockNo+1), txs[:10]...)
		pool.removeOnBlockArrival(getCurrentBestBlockNoMock(), txs[:10]...)
		if ps, _ := pool.Size(); ps != 10*(9-j) {
			t.Errorf("pool should contain 90 , %d", ps)
		}

		removed := txs[:10]
		for _, tx := range removed {
			if pool.exists(tx.Hash) != nil {
				t.Errorf("wrong tx removed [%s]", tx.GetBody().String())
			}
		}

		leftover := txs[10:]
		for _, tx := range leftover {
			if pool.exists(tx.Hash) == nil {
				t.Errorf("wrong tx removed [%s]", tx.GetBody().String())
			}
		}
		txs = txs[10:]
	}

	if l, e := pool.get(); e != nil || len(l) != 0 {
		t.Fatalf("there's leftover")
	}
}

// suppose txs appended with orphan
//
func TestDeleteInvokeRearrange(t *testing.T) {

	initTest()
	defer deinitTest()
	txs := make([]*types.Tx, 0)

	missing := map[int]bool{
		7: true, 8: true, 9: true,
		17: true, 18: true, 19: true,
		27: true, 28: true, 29: true,
		33: true, 34: true, 35: true,
		50: true}

	for i := 1; i < 51; i++ {
		tmp := genTx(0, 0, uint64(i), uint64(i))
		txs = append(txs, tmp)
		if _, v := missing[i]; v {
			continue
		}
		if pool.put(tmp) != nil {
			t.Errorf("pool should accept tx")
			//			t.Errorf("???? %d %d", getNonce(tmp), tmp.GetBody().GetNonce())
		}
	}
	if ps, os := pool.Size(); ps != 37 || os != 31 {
		t.Errorf("pool should contain 100 , %d, %d", ps, os)
	}
	// txs currently
	// ready: 1~6 orphan: 10~16, 20~26, 30~32, 36~49
	// test senario : check boundary, middle, end of each tx chunk
	// 1. gen block including 1~4
	// 2. gen block including 5~8
	// 3. gen block including 9~13
	// 4. gen block including  14~28
	// 5. gen block including 29~30
	// 6. gen block including 31~32
	// 7. gen block including 33~35
	// 8. gen blocin including ~50
	start := []int{1, 5, 9, 14, 29, 31, 33, 36}
	end := []int{4, 8, 13, 28, 30, 32, 35, 50}
	for i := 0; i < len(start); i++ {
		s, e := start[i]-1, end[i]
		simulateBlockGen(txs[s:e]...)
		pool.removeOnBlockArrival(getCurrentBestBlockNoMock(), txs[s:e]...)

		//p1, p2 := pool.Size()
		//t.Errorf("%d, %d, %d", i, p1, p2)
		removed := txs[s:e]
		for _, tx := range removed {
			if pool.exists(tx.Hash) != nil {
				t.Errorf("wrong tx removed [%s]", tx.GetBody().String())
			}
		}

		leftover := txs[e:]
		for _, tx := range leftover {
			n := tx.GetBody().GetNonce()
			if _, v := missing[int(n)]; v {
				continue
			}
			if pool.exists(tx.Hash) == nil {
				t.Errorf("wrong tx removed [%s]", tx.GetBody().String())
			}
		}
	}
}

/*
func TestDeleteInvokePriceFilterOut(t *testing.T) {
	initTest()
	defer deinitTest()

	checkRemainder := func(total int, orphan int) {
		w, o := pool.Size()
		if w != total || o != orphan {
			t.Fatalf("pool should have %d tx(%d orphans)\n", total, orphan)
		}
	}
	pool.put(genTx(0, 0, 1, 3))
	pool.put(genTx(0, 0, 2, 10))
	pool.put(genTx(0, 0, 3, 3))
	checkRemainder(3, 0)
	pool.adjust(account[0], 0, 4)
	checkRemainder(2, 1)
	pool.adjust(account[0], 1, 2)
	checkRemainder(0, 0)
}
*/

func simplePut(n uint64, b *testing.B) {
	tx := genTx(0, 0, uint64(n), 10)
	err := pool.put(tx)
	if err != nil {
		b.Fatalf("tx should be inserted %s", err)
	}

}
func BenchmarkSequential(b *testing.B) {
	initTest()
	defer deinitTest()

	//txCount := 1000
	for t := 1; t <= b.N; t++ {
		simplePut(uint64(t+1), b)
	}

}
func BenchmarkReverse(b *testing.B) {
	initTest()
	defer deinitTest()

	//txCount := 1000
	for t := b.N + 1; t >= 1; t-- {
		simplePut(uint64(t), b)
	}

}
