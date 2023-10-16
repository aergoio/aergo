/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package mempool

import (
	"sync"

	"github.com/aergoio/aergo/v2/types"
)

/*
	func (mp *MemPool) generateInfiniteTx() {
		SAMPLES := [][]byte{
			{0x01, 0x00, 0x00, 0x00},
			{0x02, 0x00, 0x00, 0x00},
			{0x03, 0x00, 0x00, 0x00},
		}

		time.Sleep(time.Second * 2)
		sampleSize := len(SAMPLES)
		var nonce []uint64
		for i := 0; i < sampleSize; i++ {
			ns, _ := mp.getAccountState(SAMPLES[i], false)
			nonce = append(nonce, ns.Nonce+1)
		}

		chunk := 1000
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
				tx.Hash = tx.CalculateTxHash()
				nonce[acc]++
				txs[i] = tx

			}
			mp.RequestToFuture(message.MemPoolSvc,
				&message.MemPoolPut{Txs: txs}, time.Second*100).Result() // nolint: errcheck

			//	err := mp.put(tx)
			//mp.Debugf("create temp tx : %s %s", err, tx.GetBody().String())
		}
	}
*/
const defaultBalance = uint64(10000000)

var (
	lock          sync.RWMutex
	balance       = map[string]uint64{}
	nonce         = map[string]uint64{}
	bestBlockInfo = &types.BlockHeaderInfo{No: 1}
)

func initStubData() {
	lock.Lock()
	defer lock.Unlock()
	balance = map[string]uint64{}
	nonce = map[string]uint64{}
	bestBlockInfo = &types.BlockHeaderInfo{No: 1}

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

func getCurrentBestBlockNoMock() types.BlockID {
	return types.ToBlockID(nil)
}

func getCurrentBestBlockInfoMock() *types.BlockHeaderInfo {
	return bestBlockInfo
}
