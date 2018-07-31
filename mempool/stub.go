/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package mempool

import (
	"encoding/hex"
	"math/rand"
	"sync"
	"time"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
)

var (
	SAMPLES = [][]byte{
		{0x01, 0x00, 0x00, 0x00},
		{0x02, 0x00, 0x00, 0x00},
		{0x03, 0x00, 0x00, 0x00},
		{0x04, 0x00, 0x00, 0x00},
		{0x05, 0x00, 0x00, 0x00},
	}
)

func (mp *MemPool) generateInfiniteTx() {

	sampleSize := len(SAMPLES)
	var nonce []uint64
	for i := 0; i < sampleSize; i++ {
		ns, _ := mp.getAccountState(SAMPLES[i], false)
		nonce = append(nonce, ns.Nonce+1)
	}

	chunk := 100
	txs := make([]*types.Tx, chunk)
	for {
		for i := 0; i < chunk; i++ {
			acc := rand.Intn(sampleSize)
			tx := &types.Tx{
				Body: &types.TxBody{
					Nonce:     nonce[acc],
					Account:   SAMPLES[acc],
					Recipient: SAMPLES[0],
					Amount:    1,
				},
			}
			nonce[acc]++
			txs[i] = tx

		}
		mp.Hub().RequestFuture(message.MemPoolSvc,
			&message.MemPoolPut{Txs: txs}, time.Second*100).Result()

		//	err := mp.put(tx)

		//mp.Debugf("create temp tx : %s %s", err, tx.GetBody().String())

	}
}

func (mp *MemPool) generateSampleTxs(maxCount int) error {
	sampleSize := len(SAMPLES)
	count := rand.Intn(maxCount)
	account := SAMPLES[rand.Intn(sampleSize)]
	ns, _ := mp.getAccountState(account, false)
	txs := make([]*types.Tx, count+1)
	for i := 1; i < count+1; i++ {
		txs[i] = &types.Tx{
			Body: &types.TxBody{
				Nonce:     ns.Nonce + uint64(i),
				Account:   account,
				Recipient: SAMPLES[rand.Intn(sampleSize)],
				Amount:    (rand.Uint64() % ns.Balance) + 1,
			},
		}
		err := mp.put(txs[i])
		mp.Debugf("create temp tx : %s %s", err, txs[i].GetBody().String())
	}
	return nil
}

func getAccount(tx *types.Tx) string {
	return hex.EncodeToString(tx.GetBody().GetAccount())
}

const defaultBalance = uint64(10000000)

var (
	lock        sync.RWMutex
	balance     = map[string]uint64{}
	nonce       = map[string]uint64{}
	bestBlockNo = types.BlockNo(1)
)

func initStubData() {
	lock.Lock()
	defer lock.Unlock()
	balance = map[string]uint64{}
	nonce = map[string]uint64{}
	bestBlockNo = types.BlockNo(1)

}
func getNonceByAccMock(acc string) uint64 {
	lock.Lock()
	defer lock.Unlock()
	_, ok := nonce[acc]
	if !ok {
		nonce[acc] = 0
	}
	return nonce[acc]
}
func getBalanceByAccMock(acc string) uint64 {
	lock.Lock()
	defer lock.Unlock()
	_, ok := balance[acc]
	if !ok {
		balance[acc] = defaultBalance
	}
	return balance[acc]
}

func getCurrentBestBlockNoMock() types.BlockNo {
	return bestBlockNo
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
