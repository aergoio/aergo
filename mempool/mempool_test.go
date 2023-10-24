/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package mempool

import (
	"encoding/binary"
	"math/big"
	"math/rand"
	"os"
	"sync/atomic"
	"testing"
	"time"

	crypto "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/aergoio/aergo/v2/cmd/aergocli/util/encoding/json"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/assert"
)

const (
	maxAccount       = 1000
	maxRecipient     = 1000
	maxBlockBodySize = 10485760
)

var (
	pool      *MemPool
	accs      [maxAccount][]byte
	sign      [maxAccount]*btcec.PrivateKey
	recipient [maxRecipient][]byte
)

func _itobU32(argv uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, argv)
	return bs
}

func getAccount(tx *types.Tx) string {
	ab := tx.GetBody().GetAccount()
	aid := types.ToAccountID(ab)
	as := aid.String()
	return as
}

func simulateBlockGen(txs ...types.Transaction) error {
	lock.Lock()
	inblock := make([]*types.Tx, 0)
	for _, tx := range txs {
		inblock = append(inblock, tx.GetTx())
		acc := getAccount(tx.GetTx())
		n := tx.GetBody().GetNonce()
		nonce[acc] = n
		_, ok := balance[acc]
		if !ok {
			balance[acc] = defaultBalance
		}
		balance[acc] -= tx.GetBody().GetAmountBigInt().Uint64()
	}
	lock.Unlock()
	pool.removeOnBlockArrival(
		&types.Block{
			Body: &types.BlockBody{
				Txs: inblock,
			}})

	//bestBlockNo++
	return nil
}
func initTest(t *testing.T) {
	pool = newTestPool()

	for i := 0; i < maxAccount; i++ {
		privkey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			t.Fatalf("failed to init test (%s)", err)
		}
		//gen new address
		accs[i] = crypto.GenerateAddress(&privkey.PublicKey)
		sign[i] = privkey
		recipient[i] = _itobU32(uint32(i))
	}
}

func newTestPool() *MemPool {
	serverCtx := config.NewServerContext("", "")
	cfg := serverCtx.GetDefaultConfig().(*config.Config)
	mp := NewMemPoolService(cfg, nil)
	mp.testConfig = true
	mp.BeforeStart()
	return mp
}

func deinitTest() {

}

func sameTx(a *types.Tx, b *types.Tx) bool {
	return types.ToTxID(a.Hash) == types.ToTxID(b.Hash)
}
func sameTxs(a []types.Transaction, b []types.Transaction) bool {
	if len(a) != len(b) {
		return false
	}
	check := false
	for _, txa := range a {
		check = false
		for _, txb := range b {
			if sameTx(txa.GetTx(), txb.GetTx()) {
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
func genTx(acc int, rec int, nonce uint64, amount uint64) types.Transaction {
	tx := types.Tx{
		Body: &types.TxBody{
			Nonce:     nonce,
			Account:   accs[acc],
			Recipient: recipient[rec],
			Amount:    new(big.Int).SetUint64(amount).Bytes(),
		},
	}
	tx.Hash = tx.CalculateTxHash()
	//key.SignTx(&tx, sign[acc])
	return types.NewTransaction(&tx)
}

/*
func TestTxSize(t *testing.T) {
	initTest(t)
	defer deinitTest()

	var b []byte
	b = make([]byte, txMaxSize)
	tx := &types.Tx{
		Body: &types.TxBody{
			Nonce:     1,
			Account:   accs[0],
			Recipient: recipient[0],
			Amount:    new(big.Int).SetUint64(1).Bytes(),
			Payload:   b,
		},
	}
	tx.Hash = tx.CalculateTxHash()
	err := pool.put(tx)
	assert.EqualError(t, err, types.ErrTxSizeExceedLimit.Error(), "wrong err")
}
*/

func TestInvalidTransaction(t *testing.T) {

	initTest(t)
	defer deinitTest()
	err := pool.put(genTx(0, 1, 1, defaultBalance*2))
	assert.EqualError(t, err, types.ErrInsufficientBalance.Error(), "wrong err")

	err = pool.put(genTx(0, 1, 1, 1))
	assert.NoError(t, err, "tx should be accepted")

	err = pool.put(genTx(0, 1, 1, 1))
	assert.EqualError(t, err, types.ErrTxAlreadyInMempool.Error(), "tx should be denied")

	txs := []types.Transaction{genTx(0, 1, 1, 1)}
	simulateBlockGen(txs...)

	err = pool.put(genTx(0, 1, 1, 1))
	assert.EqualError(t, err, types.ErrTxNonceTooLow.Error(), "tx should be denied")
}

/*
func TestInvalidTransactions(t *testing.T) {
	initTest(t)
	defer deinitTest()
	tx := genTx(0, 1, 1, 1)

	key.SignTx(tx, sign[1])
	err := pool.put(tx)
	if err == nil {
		t.Errorf("put invalid tx should be failed")
	}

	tx.Body.Sign = nil
	tx.Hash = tx.CalculateTxHash()

	err = pool.put(tx)
	if err == nil {
		t.Errorf("put invalid tx should be failed")
	}
}
*/

func TestOrphanTransaction(t *testing.T) {

	initTest(t)
	defer deinitTest()

	err := pool.put(genTx(0, 1, 1, 2))
	assert.NoError(t, err, "tx should be accepted")

	// tx inject order : 1 3 5 2 4 10 9 8 7 6
	// non-sequential nonce should be accepted (orphan) but not counted
	err = pool.put(genTx(0, 1, 3, 2))
	assert.NoError(t, err, "tx should be accepted")

	err = pool.put(genTx(0, 1, 5, 2))
	assert.NoError(t, err, "tx should be accepted")

	total, orphan := pool.Size()
	assert.EqualValuesf(t, []int{total, orphan}, []int{3, 2}, "wrong mempool stat")

	err = pool.put(genTx(0, 1, 2, 2))
	assert.NoError(t, err, "tx should be accepted")

	total, orphan = pool.Size()
	assert.EqualValuesf(t, []int{total, orphan}, []int{4, 1}, "wrong mempool stat")

	err = pool.put(genTx(0, 1, 4, 2))
	assert.NoError(t, err, "tx should be accepted")

	total, orphan = pool.Size()
	assert.EqualValuesf(t, []int{total, orphan}, []int{5, 0}, "wrong mempool stat")

	err = pool.put(genTx(0, 1, 10, 2))
	assert.NoError(t, err, "tx should be accepted")

	err = pool.put(genTx(0, 1, 9, 2))
	assert.NoError(t, err, "tx should be accepted")

	err = pool.put(genTx(0, 1, 8, 2))
	assert.NoError(t, err, "tx should be accepted")

	err = pool.put(genTx(0, 1, 7, 2))
	assert.NoError(t, err, "tx should be accepted")

	total, orphan = pool.Size()
	assert.EqualValuesf(t, []int{total, orphan}, []int{9, 4}, "wrong mempool stat")

	err = pool.put(genTx(0, 1, 6, 2))
	assert.NoError(t, err, "tx should be accepted")

	total, orphan = pool.Size()
	assert.EqualValuesf(t, []int{total, orphan}, []int{10, 0}, "wrong mempool stat")
}

func TestBasics2(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	initTest(t)
	defer deinitTest()
	txs := make([]*types.Tx, 0)

	accCount := 1000
	txCount := 1000
	nonce := make([]uint64, txCount)
	for i := 0; i < txCount; i++ {
		nonce[i] = uint64(i + 1)
	}

	for i := 0; i < accCount; i++ {
		rand.Shuffle(txCount, func(i, j int) {
			nonce[i], nonce[j] = nonce[j], nonce[i]
		})
		for j := 0; j < txCount; j++ {
			tmp := genTx(i, 0, nonce[j], uint64(i+1))
			txs = append(txs, tmp.GetTx())
		}
	}

	for _, tx := range txs {
		err := pool.put(types.NewTransaction(tx))
		assert.NoError(t, err, "tx should be accepted")
	}

	txsMempool, err := pool.get(maxBlockBodySize * 10)
	assert.NoError(t, err, "get failed")
	assert.Equal(t, len(txsMempool), len(txs))
}

// gen sequential transactions
// check mempool internal states
func TestBasics(t *testing.T) {
	initTest(t)
	defer deinitTest()
	txs := make([]types.Transaction, 0)

	accCount := 10
	txCount := 10
	nonce := make([]uint64, txCount)
	for i := 0; i < txCount; i++ {
		nonce[i] = uint64(i + 1)
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
	assert.Equal(t, len(errs), accCount*txCount, "error length is different")

	for i := 0; i < len(errs); i++ {
		assert.NoError(t, errs[i], "%dth tx failed", i)
	}

	txsMempool, err := pool.get(maxBlockBodySize)
	assert.NoError(t, err, "get failed")
	assert.Equal(t, len(txsMempool), len(txs))
}

func TestDeleteOTxs(t *testing.T) {
	initTest(t)
	defer deinitTest()
	txs := make([]types.Transaction, 0)
	for i := 0; i < 5; i++ {
		tmp := genTx(0, 0, uint64(i+1), uint64(i+1))
		txs = append(txs, tmp)
	}
	pool.puts(txs...)

	total, orphan := pool.Size()
	assert.EqualValuesf(t, []int{total, orphan}, []int{5, 0}, "wrong mempool stat")

	txs[4] = genTx(0, 1, 5, 150)
	simulateBlockGen(txs...)

	total, orphan = pool.Size()
	assert.EqualValuesf(t, []int{total, orphan}, []int{0, 0}, "wrong mempool stat")
}

// add 100 sequential txs and simulate to generate block 10time.
// each block contains 10 txs
func TestBasicDeleteOnBlockConnect(t *testing.T) {
	initTest(t)
	defer deinitTest()
	txs := make([]types.Transaction, 0)

	for i := 0; i < 100; i++ {
		tmp := genTx(0, 0, uint64(i+1), uint64(i+1))
		txs = append(txs, tmp)
	}
	pool.puts(txs...)

	total, orphan := pool.Size()
	assert.EqualValuesf(t, []int{total, orphan}, []int{100, 0}, "wrong mempool stat")

	//suppose 10 txs are select into new block
	for j := 0; j < 10; j++ {
		simulateBlockGen(txs[:10]...)

		total, orphan := pool.Size()
		assert.EqualValuesf(t, []int{total, orphan}, []int{10 * (9 - j), 0}, "wrong mempool stat")

		removed := txs[:10]

		for _, tx := range removed {
			found := pool.exist(tx.GetHash())
			assert.Nil(t, found, "wrong transaction removed")
		}

		leftover := txs[10:]
		for _, tx := range leftover {
			found := pool.exist(tx.GetHash())
			assert.NotNil(t, found, "wrong transaction removed")
		}
		txs = txs[10:]
	}

	l, e := pool.get(maxBlockBodySize)
	assert.NoError(t, e, "get should succeed")
	assert.Equalf(t, len(l), 0, "leftover found")
}

func TestDeleteInvokeRearrange(t *testing.T) {

	initTest(t)
	defer deinitTest()
	txs := make([]types.Transaction, 0)

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
		assert.NoError(t, pool.put(tmp), "tx should be accepted")
	}

	total, orphan := pool.Size()
	assert.EqualValuesf(t, []int{total, orphan}, []int{37, 31}, "wrong mempool stat")

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

		//p1, p2 := pool.Size()
		//t.Errorf("%d, %d, %d", i, p1, p2)
		removed := txs[s:e]
		for _, tx := range removed {
			found := pool.exist(tx.GetHash())
			assert.Nil(t, found, "wrong transaction removed")
		}

		leftover := txs[e:]
		for _, tx := range leftover {
			n := tx.GetBody().GetNonce()
			if _, v := missing[int(n)]; v {
				continue
			}
			if pool.exist(tx.GetHash()) == nil {
				t.Errorf("wrong tx removed [%s]", tx.GetBody().String())
			}
		}
	}
}

func TestSwitchingBestBlock(t *testing.T) {
	initTest(t)
	defer deinitTest()

	txs := make([]types.Transaction, 0)
	tx0 := genTx(0, 1, 1, 1)
	tx1 := genTx(0, 1, 2, 1)
	txs = append(txs, tx0, tx1)

	err := pool.puts(txs...)
	if len(err) != 2 || err[0] != nil || err[1] != nil {
		t.Errorf("put should succeed, %s", err)
	}
	simulateBlockGen(txs...)

	tx2 := genTx(0, 1, 3, 1)
	if err := pool.put(tx2); err != nil {
		t.Errorf("put should succeed, %s", err)
	}
	ready, orphan := pool.Size()
	if ready != 1 || orphan != 0 {
		t.Errorf("size wrong:%d, %d", ready, orphan)
	}

	simulateBlockGen(txs[:1]...)

	ready, orphan = pool.Size()
	if ready != 1 || orphan != 1 {
		t.Errorf("size wrong:%d, %d", ready, orphan)
	}

	tx4 := genTx(0, 1, 5, 1)
	if err := pool.put(tx4); err != nil {
		t.Errorf("put should succeed, %s", err.Error())
	}

	ready, orphan = pool.Size()
	if ready != 2 || orphan != 2 {
		t.Errorf("size wrong:%d, %d", ready, orphan)
	}

	if err := pool.put(tx1); err != nil {
		t.Errorf("put should succeed, %s", err.Error())
	}
	ready, orphan = pool.Size()
	if ready != 3 || orphan != 1 {
		t.Errorf("size wrong:%d, %d", ready, orphan)
	}
}

func TestDumpAndLoad(t *testing.T) {
	initTest(t)
	//set temporary path for test
	pool.dumpPath = "./mempool_dump_test"
	txs := make([]*types.Tx, 0)

	if _, err := os.Stat(pool.dumpPath); os.IsExist(err) {
		if os.Remove(pool.dumpPath) != nil {
			t.Errorf("init test failed (rm %s failed)", pool.dumpPath)
		}
	}

	pool.dumpTxsToFile()
	if _, err := os.Stat(pool.dumpPath); err != nil && !os.IsNotExist(err) {
		t.Errorf("err should be NotExist ,but %s", err.Error())
	}

	if !atomic.CompareAndSwapInt32(&pool.status, initial, running) {
		t.Errorf("pool status should be initial, but %d", pool.status)
	}
	pool.dumpTxsToFile()
	if _, err := os.Stat(pool.dumpPath); err != nil && !os.IsNotExist(err) {
		t.Errorf("err should be NotExist ,but %s", err.Error())
	}

	for i := 0; i < 100; i++ {
		tmp := genTx(0, 0, uint64(i+1), uint64(i+1))
		txs = append(txs, tmp.GetTx())
		if err := pool.put(tmp); err != nil {
			t.Errorf("put should succeed, %s", err.Error())
		}
	}

	pool.dumpTxsToFile()
	if _, err := os.Stat(pool.dumpPath); err != nil {
		t.Errorf("dump file should be created but, %s", err.Error())
	}
	deinitTest()

	initTest(t)
	pool.dumpPath = "./mempool_dump_test"
	ready, orphan := pool.Size()
	if ready != 0 || orphan != 0 {
		t.Errorf("size wrong:%d, %d", ready, orphan)
	}
	if !atomic.CompareAndSwapInt32(&pool.status, initial, running) {
		t.Errorf("pool status should be initial, but %d", pool.status)
	}
	pool.loadTxs()
	ready, orphan = pool.Size()
	if ready != 0 || orphan != 0 {
		t.Errorf("size wrong:%d, %d", ready, orphan)
	}

	if !atomic.CompareAndSwapInt32(&pool.status, running, initial) {
		t.Errorf("pool status should be initial, but %d", pool.status)
	}

	pool.loadTxs()
	ready, orphan = pool.Size()
	if ready != 100 || orphan != 0 {
		t.Errorf("size wrong:%d, %d", ready, orphan)
	}
	deinitTest()
	os.Remove(pool.dumpPath) // nolint: errcheck
}

func TestEvictOnProfit(t *testing.T) {
	initTest(t)
	defer deinitTest()

	if err := pool.put(genTx(0, 0, 1, 3)); err != nil {
		t.Errorf("put should succeed, %s", err.Error())
	}
	if err := pool.put(genTx(0, 0, 1, 10)); err == nil {
		t.Errorf("put should failed") //FIXME
	}

	if err := pool.put(genTx(0, 0, 5, 3)); err != nil {
		t.Errorf("put should succeed, %s", err.Error())
	}
	pool.put(genTx(0, 0, 6, 3))
	pool.put(genTx(0, 0, 7, 3))

	if err := pool.put(genTx(0, 0, 6, 10)); err == nil {
		t.Errorf("put should failed") // FIXME
	}
}

func TestDeleteInvokePriceFilterOut(t *testing.T) {
	initTest(t)
	defer deinitTest()

	checkRemainder := func(total int, orphan int) {
		w, o := pool.Size()
		if w != total || o != orphan {
			t.Fatalf("pool should have %d tx(%d orphans) but(%d/%d)\n", total, orphan, w, o)
		}
	}
	txs := make([]types.Transaction, 0)
	txs = append(txs, genTx(0, 1, 1, defaultBalance-6))
	txs = append(txs, genTx(0, 1, 2, 2))
	txs = append(txs, genTx(0, 1, 3, 10))
	txs = append(txs, genTx(0, 1, 4, 5))

	for _, tx := range txs {
		pool.put(tx)
	}
	checkRemainder(len(txs), 0)
	simulateBlockGen(txs[:1]...)

	checkRemainder(2, 1)
	simulateBlockGen(txs[1:2]...)
	checkRemainder(0, 0)
}

func TestRemoveTx(t *testing.T) {
	initTest(t)
	defer deinitTest()

	err := pool.put(genTx(0, 0, 1, 0))
	assert.NoError(t, err, "put")
	two := genTx(0, 0, 2, 0)
	err = pool.put(two)
	assert.NoError(t, err, "put")
	three := genTx(0, 0, 3, 0)
	err = pool.put(three)
	assert.NoError(t, err, "put")
	err = pool.put(genTx(0, 0, 5, 0))
	assert.NoError(t, err, "put")
	err = pool.put(genTx(0, 0, 6, 0))
	assert.NoError(t, err, "put")
	err = pool.put(genTx(0, 0, 7, 0))
	assert.NoError(t, err, "put")
	assert.Equal(t, 6, pool.length, "length")
	assert.Equal(t, 3, pool.orphan, "orphan")
	err = pool.removeTx(two.GetTx())
	assert.Equal(t, 5, pool.length, "length")
	assert.Equal(t, 4, pool.orphan, "orphan")
	err = pool.removeTx(three.GetTx())
	assert.NoError(t, err, "remove")
	assert.Equal(t, 4, pool.length, "length")
	assert.Equal(t, 3, pool.orphan, "orphan")
}

func TestMemPool_GetAddress(t *testing.T) {
	t.Skip("skip test since underlying env is not capable to test this single method")
	initTest(t)
	defer deinitTest()
	pool.testConfig = false

	quirkHash := types.DecodeB58(types.B23994084_001)
	normalHash := types.NewTx().CalculateTxHash()
	dummySender := types.NewAccount(recipient[0])

	type args struct {
		hash []byte
		recp []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"TQuirkWrongRecp", args{quirkHash, []byte("111233.123")}, false},
		{"TQuirkValidRecp", args{quirkHash, recipient[1]}, false},
		{"TNormWrongRecp", args{normalHash, []byte("111233.123")}, true},
		{"TNormValidRecp", args{normalHash, recipient[1]}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := pool

			sampleTX := types.NewTx()
			sampleTX.Body.Nonce = 1
			sampleTX.Body.Type = types.TxType_TRANSFER
			sampleTX.Body.Recipient = tt.args.recp
			sampleTX.Hash = tt.args.recp
			tx := types.NewTransaction(sampleTX)
			if err := mp.validateTx(tx, dummySender.Address); (err != nil) != tt.wantErr {
				t.Errorf("validateTx() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemPool_listHash(t *testing.T) {
	initTest(t)
	defer deinitTest()

	// generate sample txs
	samples := []accTxs{}
	for i := 0; i < 10; i++ {
		acc := accs[i]
		txCnt := i
		if i == 0 {
			txCnt = 101
		}
		at := accTxs{acc, make([]types.Transaction, txCnt)}
		for j := 0; j < txCnt; j++ {
			tx := types.NewTx()
			tx.Body.Nonce = uint64(j + 1)
			at.txs[j] = types.NewTransaction(tx)
		}
		samples = append(samples, at)
	}

	tests := []struct {
		name     string
		txs      []accTxs
		args     int
		wantSize int
		want1    bool
	}{
		{"TEmpty", nil, 15, 0, false},
		{"TSingleAcc", samples[1:2], 15, 1, false},
		{"TSmallMult", samples[1:5], 15, 10, false},
		{"TFitMult", samples[1:6], 15, 15, false},
		{"TBigMult", samples[2:], 15, 15, true},
		{"THugeSingle", samples[:1], 15, 15, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := newTestPool()
			for _, at := range tt.txs {
				list, err := mp.acquireMemPoolList(at.acc)
				if err != nil {
					t.Fatalf("Test error while setting initial env: err %v", err)
				}
				for _, tx := range at.txs {
					_, err := list.Put(tx)
					if err != nil {
						t.Fatalf("Test error while setting initial env: err %v", err)
					}
				}
				mp.releaseMemPoolList(list)
			}
			got, got1 := mp.listHash(tt.args)
			if len(got) != tt.wantSize {
				t.Errorf("listHash() got = %v, want %v", got, tt.wantSize)
			}
			if got1 != tt.want1 {
				t.Errorf("listHash() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

type accTxs struct {
	acc []byte
	txs []types.Transaction
}

func ex(samples []accTxs, idxs ...int) []accTxs {
	sli := make([]accTxs, len(idxs))
	for i, idx := range idxs {
		sli[i] = samples[idx]
	}
	return sli
}

func TestMemPool_Unconfirmed(t *testing.T) {
	req := assert.New(t)

	ti := time.Now().Add(evictPeriod)
	u := newUnconfirmedTxs([]byte("abc"), &ti, 10, 1)
	b, err := json.Marshal(u)
	req.Nil(err)
	t.Log(string(b))
}
