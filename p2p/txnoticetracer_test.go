/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

func Test_newTxNoticeTracer(t *testing.T) {
	logger := log.NewLogger("p2p.test")

	tests := []struct {
		name string
	}{
		{"T1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockActor := p2pmock.NewMockActorService(ctrl)
			got := newTxNoticeTracer(logger, mockActor)

			if got.retryC == nil {
				t.Errorf("member not inited retryC")
			}
			if got.reportC == nil {
				t.Errorf("member not inited reportC")
			}
		})
	}
}

func Test_txNoticeTracer_traceTxNoticeRegister(t *testing.T) {
	logger := log.NewLogger("p2p.test")
	dummyHashes := make([]types.TxID, 0)
	dummyHashes = append(dummyHashes, sampleTxHashes...)

	type args struct {
		peerCnt int
		ids     []types.TxID
	}
	tests := []struct {
		name string

		report args

		wantRetryIDs int
	}{
		{"TSingle", args{5, dummyHashes}, 0},
		{"TNoPeer", args{0, dummyHashes}, len(dummyHashes)},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockActor := p2pmock.NewMockActorService(ctrl)

			tnt := newTxNoticeTracer(logger, mockActor)

			tnt.RegisterTxNotice(tt.report.ids, tt.report.peerCnt)
			select {
			case r := <-tnt.reportC:
				if r.tType != create {
					t.Errorf("RegisterTxNotice() rType %v, want %v", r.tType, create)
				} else {
					tnt.newTrace(r)
				}
			default:
				t.Errorf("RegisterTxNotice()  unexpected behaviour ")
			}
			if len(tnt.retryIDs) != tt.wantRetryIDs {
				t.Errorf("retryIDs = %v , want %v ", len(tnt.retryIDs), tt.wantRetryIDs)
			}

		})
	}
}

func Test_txNoticeTracer_traceTxNoticeRegisterResult(t *testing.T) {
	logger := log.NewLogger("p2p.test")
	dummyHashes := make([]types.TxID, 0)
	dummyHashes = append(dummyHashes, sampleTxHashes...)

	AllSucc := txNoticeSendReport{tType: p2pcommon.Send, peerCnt: 1, hashes: dummyHashes}
	AllFail := txNoticeSendReport{tType: p2pcommon.Fail, peerCnt: 1, hashes: dummyHashes}
	AllSkip := txNoticeSendReport{tType: p2pcommon.Skip, peerCnt: 1, hashes: dummyHashes}
	PartSucc := txNoticeSendReport{tType: p2pcommon.Send, peerCnt: 1, hashes: dummyHashes[1:4]}
	PartFail := txNoticeSendReport{tType: p2pcommon.Fail, peerCnt: 1, hashes: dummyHashes[1:4]}
	PartSkip := txNoticeSendReport{tType: p2pcommon.Skip, peerCnt: 1, hashes: dummyHashes[1:4]}
	tests := []struct {
		name string

		args []txNoticeSendReport

		wantStatCnt  int
		wantRetryIDs int
	}{
		{"TSucc", ad(AllSucc, AllSucc, AllSucc), 0, 0},
		{"TSucc2", ad(AllSucc, AllSucc, AllFail), 0, 0},
		{"TSucc3", ad(AllSkip, AllSucc, AllSucc), 0, 0},
		{"TFail1", ad(AllFail, AllFail, AllFail), 0, 6},
		{"TFail2", ad(AllFail, AllSkip, AllFail), 0, 6},

		{"TPartSucc", ad(PartSucc, PartSucc, PartSucc), 3, 0},
		{"TPartFail1", ad(PartFail, PartFail, PartFail), 3, 3},
		{"TPartFail2", ad(PartSkip, PartSkip, PartSkip), 3, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockActor := p2pmock.NewMockActorService(ctrl)

			tnt := newTxNoticeTracer(logger, mockActor)
			// add initial items
			tnt.newTrace(txNoticeSendReport{create, dummyHashes, 3})

			for _, r := range tt.args {
				tnt.handleReport(r)
			}

			if tnt.txSendStats.Len() != tt.wantStatCnt {
				t.Errorf("stats = %v , want %v ", tnt.txSendStats.Len(), tt.wantStatCnt)
			}

			if len(tnt.retryIDs) != tt.wantRetryIDs {
				t.Errorf("retryIDs = %v , want %v ", len(tnt.retryIDs), tt.wantRetryIDs)
			}

		})
	}
}

func ad(rs ...txNoticeSendReport) []txNoticeSendReport {
	return rs
}

func Test_txNoticeTracer_retryNotice(t *testing.T) {
	logger := log.NewLogger("p2p.test")

	tests := []struct {
		name string

		inStock    []types.TxID
		wantHashes int
	}{
		{"TNothing", nil, 0},
		{"TMulti", sampleTxHashes, len(sampleTxHashes)},
		{"TSameIDs", append(sampleTxHashes, sampleTxHashes[0:3]...), len(sampleTxHashes)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockActor := p2pmock.NewMockActorService(ctrl)

			tnt := newTxNoticeTracer(logger, mockActor)
			tnt.retryIDs = append(tnt.retryIDs, tt.inStock...)
			if tt.wantHashes > 0 {
				mockActor.EXPECT().TellRequest(message.P2PSvc, gomock.Any()).Do(func(a string, b interface{}) {
					bb := b.(notifyNewTXs)
					if len(bb.ids) != tt.wantHashes {
						t.Errorf("send hash count %v, want %v", len(bb.ids), tt.wantHashes)
					}
				})
			}

			tnt.retryNotice()
		})
	}
}

func Test_txNoticeTracer_cleanupStales(t *testing.T) {
	logger := log.NewLogger("p2p.test")
	dummyHashes := make([]types.TxID, 0)
	dummyHashes = append(dummyHashes, sampleTxHashes...)

	n := time.Now()
	o := n.Add(time.Minute * -12)

	tests := []struct {
		name      string
		aTimes    []time.Time
		sents     []int
		wantRetry int
	}{
		{"TAllNew", compT(n, n, n, n, n, n), compI(0, 1, 2, 3, 4, 0), 0},
		{"TOldButSent", compT(o, o, o, n, n, n), compI(3, 1, 2, 0, 4, 0), 0},
		{"TOldUnsent", compT(o, o, o, n, n, n), compI(3, 0, 0, 0, 4, 0), 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockActor := p2pmock.NewMockActorService(ctrl)
			if tt.wantRetry > 0 {
				mockActor.EXPECT().TellRequest(gomock.Any(), gomock.Any()).Do(func(p string, v interface{}) {
					notiReq := v.(notifyNewTXs)
					if len(notiReq.ids) != tt.wantRetry {
						t.Errorf("cleanupStales() retry cnt %v ,want %v", len(notiReq.ids), tt.wantRetry)
					}
				})
			}
			tnt := newTxNoticeTracer(logger, mockActor)
			for i, h := range dummyHashes {
				st := &txNoticeSendStat{hash: h, created: o, accecced: tt.aTimes[i], sent: tt.sents[i]}
				tnt.txSendStats.Add(h, st)
			}

			tnt.cleanupStales()
		})
	}
}

func compT(ts ...time.Time) []time.Time {
	return ts
}
func compI(ts ...int) []int {
	return ts
}
