/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	lru "github.com/hashicorp/golang-lru"
	"time"
)

const (
	RequiredSendCount int = 1
)
const create p2pcommon.ReportType = iota + 1000
var peerIDHolder = []types.PeerID(nil)

type txNoticeTracer struct {
	logger *log.Logger
	actor  p2pcommon.ActorService
	reqCnt int

	txSendStats *lru.Cache
	reportC     chan txNoticeSendReport

	retryIDs []types.TxID
	retryC   <-chan time.Time

	finish chan int
}

var _ p2pcommon.TxNoticeTracer = (*txNoticeTracer)(nil)

func newTxNoticeTracer(logger *log.Logger, actor p2pcommon.ActorService) *txNoticeTracer {
	t := &txNoticeTracer{logger: logger, actor: actor, reqCnt: RequiredSendCount, reportC: make(chan txNoticeSendReport, syncManagerChanSize), finish:make(chan int)}
	var err error
	t.txSendStats, err = lru.New(DefaultGlobalTxCacheSize * 4)
	if err != nil {
		panic("Failed to create p2p trace cache " + err.Error())
	}
	t.retryC = time.NewTicker(time.Minute >> 1).C
	return t
}

type txNoticeSendStat struct {
	hash     types.TxID
	created  time.Time
	accessed time.Time
	remain   int
	sentCnt  int
	sent     []types.PeerID
}

type txNoticeSendReport struct {
	tType  p2pcommon.ReportType
	hashes []types.TxID
	expect int
	peerIDs []types.PeerID
}

func (t *txNoticeTracer) run() {
	t.logger.Info().Msg("starting p2p txNoticeTracer")
	cleanUpT := time.NewTicker(time.Minute * 10)
TRACE_LOOP:
	for {
		select {
		case rep := <-t.reportC:
			if rep.tType == create {
				t.newTrace(rep)
			} else {
				t.handleReport(rep)
			}
		case <-t.retryC:
			t.retryNotice()
		case <-cleanUpT.C:
			t.cleanupStales()
		case <-t.finish:
			break TRACE_LOOP
		}
	}
	t.logger.Info().Msg("txNoticeTracer is finished")
}
func (t *txNoticeTracer) Start() {
	go t.run()
}
func (t *txNoticeTracer) Stop() {
	close(t.finish)
}

func (t *txNoticeTracer) RegisterTxNotice(txIDs []types.TxID, cnt int, alreadySent []types.PeerID) {
	t.reportC <- txNoticeSendReport{create, txIDs, cnt, peerIDHolder}
}

func (t *txNoticeTracer) ReportNotSend(txIDs []types.TxID, cnt int) {
	t.reportC <- txNoticeSendReport{p2pcommon.Fail, txIDs, cnt, peerIDHolder}
}

func (t *txNoticeTracer) ReportSend(txIDs []types.TxID, peerID types.PeerID) {
	t.reportC <- txNoticeSendReport{p2pcommon.Send, txIDs, 0, []types.PeerID{peerID}}
}

func (t *txNoticeTracer) newTrace(report txNoticeSendReport) {
	if report.expect == 0 {
		t.retryIDs = append(t.retryIDs, report.hashes...)
		t.logger.Debug().Array("txs", types.NewLogTxIDsMarshaller(t.retryIDs, 10)).Msg("no active peer to send notice. retrying later")
	} else {
		t.logger.Debug().Array("txs", types.NewLogTxIDsMarshaller(t.retryIDs, 10)).Int("toSendCnt", report.expect).Msg("new tx notice trace")
		ctime := time.Now()
		for _, txHash := range report.hashes {
			t.txSendStats.Add(txHash, &txNoticeSendStat{hash: txHash, created: ctime, accessed: ctime, remain: report.expect})
		}
	}
}

func (t *txNoticeTracer) handleReport(report txNoticeSendReport) {
	//t.logger.Debug().Str("type", report.tType.String()).Array("txs", types.NewLogTxIDsMarshaller(t.retryIDs,10)).Int("peerCnt", report.peerCnt).Msg("new tx notice trace")
	for _, txHash := range report.hashes {
		s, exist := t.txSendStats.Get(txHash)
		if !exist { // evicted
			continue
		}
		stat := s.(*txNoticeSendStat)
		stat.remain--
		if report.tType == p2pcommon.Send {
			stat.sentCnt++
		}
		if stat.remain == 0 {
			t.txSendStats.Remove(txHash)
			if stat.sentCnt < t.reqCnt { // couldn't send any nodes
				t.retryIDs = append(t.retryIDs, txHash)
			}
		} else {
			stat.accessed = time.Now()
		}
	}

}

func (t *txNoticeTracer) retryNotice() {
	if len(t.retryIDs) == 0 {
		return
	}
	t.logger.Debug().Array("txs", types.NewLogTxIDsMarshaller(t.retryIDs, 10)).Msg("retrying to send tx notices")
	hMap := make(map[types.TxID]int, len(t.retryIDs))
	hashes := make([]types.TxID, 0, len(t.retryIDs))
	for _, hash := range t.retryIDs {
		if _, exist := hMap[hash]; !exist {
			hashes = append(hashes, hash)
			hMap[hash] = 1
		}
	}
	// clear
	t.retryIDs = t.retryIDs[:0]
	if len(hashes) > 0 {
		t.actor.TellRequest(message.P2PSvc, notifyNewTXs{hashes, nil})
	}
}

func (t *txNoticeTracer) cleanupStales() {
	t.logger.Debug().Msg("Cleaning up TX notice stats ")
	// It should be nothing or very few stats remains in cleanup time. If not, find bugs .
	expireTime := time.Now().Add(-1 * time.Minute * 10)
	keys := t.txSendStats.Keys()
	size := len(keys)
	if size > 1000 {
		size = 1000
	}
	toRetries := make([]types.TxID, 0, 10)
	for i := 0; i < size; i++ {
		s, found := t.txSendStats.Get(keys[i])
		if !found {
			continue
		}
		stat := s.(*txNoticeSendStat)
		if !stat.accessed.Before(expireTime) {
			break
		}
		if stat.sentCnt == 0 {
			toRetries = append(toRetries, stat.hash)
		}
		t.txSendStats.Remove(keys[i])
	}
	if len(toRetries) > 0 {
		t.logger.Info().Int("cnt", len(toRetries)).Msg("Unsent TX notices are found")
		t.retryIDs = append(t.retryIDs, toRetries...)
	} else {
		t.logger.Debug().Msg("no unsent TX notices are found")
	}
}
