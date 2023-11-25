/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"math/rand"
	"testing"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
	"github.com/gofrs/uuid"
)

var sampleInputIDs []types.TxID
var Tx5000, Tx1000, Tx100, Tx10, Tx1 []types.TxID

const hashCnt = 5000

func init() {
	buf := make([]byte, types.HashIDLength)
	sampleInputIDs = make([]types.TxID, hashCnt)
	for i := 0; i < hashCnt; i++ {
		rand.Read(buf)
		sampleInputIDs[i] = types.ToTxID(buf)
	}
	Tx5000 = sampleInputIDs
	Tx1000 = make([]types.TxID, 1000)
	copy(Tx1000, sampleInputIDs)
	Tx100 = make([]types.TxID, 100)
	copy(Tx100, sampleInputIDs)
	Tx10 = make([]types.TxID, 10)
	copy(Tx10, sampleInputIDs)
	Tx1 = make([]types.TxID, 1)
	copy(Tx1, sampleInputIDs)
}

func BenchmarkBaseMOFactory_NewMsgTxBroadcastOrder(b *testing.B) {
	dummyP2PS := &P2P{}

	benchmarks := []struct {
		name string
		in   []types.TxID
	}{
		{"B1", Tx1},
		{"B10", Tx10},
		{"B100", Tx100},
		{"B1000", Tx1000},
		{"B5000", Tx5000},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				mf := &baseMOFactory{
					is: dummyP2PS,
				}
				in := bm.in
				hashes := make([][]byte, 0, len(in))
				for _, hash := range in {
					hashes = append(hashes, hash[:])
				}

				_ = mf.NewMsgTxBroadcastOrder(&types.NewTransactionsNotice{TxHashes: hashes})
			}
		})
	}
}

func BenchmarkBaseMOFactory_DiffFunc(b *testing.B) {
	dummyP2PS := &P2P{}

	benchmarks := []struct {
		name string
		in   []types.TxID
	}{
		{"B1", Tx1},
		{"B10", Tx10},
		{"B100", Tx100},
		{"B1000", Tx1000},
		{"B5000", Tx5000},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				mf := &baseMOFactory{
					is: dummyP2PS,
				}
				_ = mf.diffMsgTxBroadcastOrder(bm.in)
			}
		})
	}
}

func (mf *baseMOFactory) diffMsgTxBroadcastOrder(ids []types.TxID) p2pcommon.MsgOrder {
	rmo := &pbTxNoticeOrder{}
	reqID := uuid.Must(uuid.NewV4())
	hashes := make([][]byte, len(ids))
	for i, hash := range ids {
		hashes[i] = hash[:]
	}
	message := &types.NewTransactionsNotice{TxHashes: hashes}
	if mf.fillUpMsgOrder(&rmo.pbMessageOrder, reqID, uuid.Nil, p2pcommon.NewTxNotice, message) {
		rmo.txHashes = ids
		return rmo
	}
	return nil
}
